package cli

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/btcsuite/btcutil"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/kaonone/eth-rpc-gate/pkg/analytics"
	"github.com/kaonone/eth-rpc-gate/pkg/kaon"
	"github.com/kaonone/eth-rpc-gate/pkg/notifier"
	"github.com/kaonone/eth-rpc-gate/pkg/params"
	"github.com/kaonone/eth-rpc-gate/pkg/server"
	"github.com/kaonone/eth-rpc-gate/pkg/transformer"
	"github.com/natefinch/lumberjack"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app = kingpin.New("ethrpcgate", "Kaon adapter to Ethereum JSON RPC")

	accountsFile = app.Flag("accounts", "account private keys (in WIF) returned by eth_accounts").Envar("ACCOUNTS").File()

	kaonRPC             = app.Flag("kaon-rpc", "URL of Kaon RPC service").Envar("KAON_RPC").Default("").String()
	kaonNetwork         = app.Flag("kaon-network", "if 'regtest' (or connected to a regtest node with 'auto') eth-rpc-gate will generate blocks").Envar("KAON_NETWORK").Default("auto").String()
	generateToAddressTo = app.Flag("generateToAddressTo", "[regtest only] configure address to mine blocks to when mining new transactions in blocks").Envar("GENERATE_TO_ADDRESS").Default("").String()
	bind                = app.Flag("bind", "network interface to bind to (e.g. 0.0.0.0) ").Envar("GATE_BIND").Default("").String()
	port                = app.Flag("port", "port to serve proxy").Envar("GATE_PORT").Default("").Int()
	httpsKey            = app.Flag("https-key", "https keyfile").Envar("GATE_CERT_KEY").Default("").String()
	httpsCert           = app.Flag("https-cert", "https certificate").Envar("GATE_CERT").Default("").String()
	logFile             = app.Flag("log-file", "write logs to a file").Envar("LOG_FILE").Default("").String()
	matureBlockHeight   = app.Flag("mature-block-height-override", "override how old a coinbase/coinstake needs to be to be considered mature enough for spending (KAON uses 2000 blocks after the 32s block fork) - if this value is incorrect transactions can be rejected").Envar("BLOCKS_MATURITY").Default("21").Int()
	healthCheckPercent  = app.Flag("health-check-healthy-request-amount", "configure the minimum request success rate for healthcheck").Envar("HEALTH_CHECK_REQUEST_PERCENT").Default("80").Int()

	sqlHost     = app.Flag("sql-host", "database hostname").Envar("SQL_HOST").Default("").String()
	sqlPort     = app.Flag("sql-port", "database port").Envar("SQL_PORT").Default("").Int()
	sqlUser     = app.Flag("sql-user", "database username").Envar("SQL_USER").Default("").String()
	sqlPassword = app.Flag("sql-password", "database password").Envar("SQL_PASSWORD").Default("").String()
	sqlSSL      = app.Flag("sql-ssl", "use SSL to connect to database").Envar("SQL_SSL").Bool()
	sqlDbname   = app.Flag("sql-dbname", "database name").Envar("SQL_DBNAME").Default("").String()

	dbConnectionString = app.Flag("dbstring", "database connection string").Envar("GATE_DBSTRING").Default("").String()

	devMode        = app.Flag("dev", "[Insecure] Developer mode").Envar("DEV").Default("false").Bool()
	singleThreaded = app.Flag("singleThreaded", "[Non-production] Process RPC requests in a single thread").Envar("SINGLE_THREADED").Default("false").Bool()

	ignoreUnknownTransactions = app.Flag("ignoreTransactions", "[Development] Ignore transactions inside blocks we can't fetch and return responses instead of failing").Default("false").Bool()
	disableSnipping           = app.Flag("disableSnipping", "[Development] Disable ...snip... in logs").Default("false").Bool()
	hideKaondLogs             = app.Flag("hideKaondLogs", "[Development] Hide KAOND debug logs").Envar("HIDE_KAOND_LOGS").Default("false").Bool()
	hideTCPLogs               = app.Flag("hide-tcp-logs", "Hide logs containing TCP bind and port information").Envar("HIDE_TCP_LOGS").Default("false").Bool()
)

func loadAccounts(r io.Reader, l log.Logger) kaon.Accounts {
	var accounts kaon.Accounts

	if accountsFile != nil {
		s := bufio.NewScanner(*accountsFile)
		for s.Scan() {
			line := s.Text()

			wif, err := btcutil.DecodeWIF(line)
			if err != nil {
				level.Error(l).Log("msg", "Failed to parse account", "err", err.Error())
				continue
			}

			accounts = append(accounts, wif)
		}
	}

	if len(accounts) > 0 {
		level.Info(l).Log("msg", fmt.Sprintf("Loaded %d accounts", len(accounts)))
	} else {
		level.Warn(l).Log("msg", "No accounts loaded from account file")
	}

	return accounts
}

type multiErrorWriter struct {
	mainWriter  io.Writer
	errorWriter io.Writer
}

func (m *multiErrorWriter) Write(p []byte) (n int, err error) {
	// Write to the main log file
	n, err = m.mainWriter.Write(p)
	if err != nil {
		return n, err
	}

	// Check if the log line contains an error level and write to the error log file
	if bytes.Contains(p, []byte("level=error")) || bytes.Contains(p, []byte("level=warn")) {
		_, err = m.errorWriter.Write(p)
		if err != nil {
			return n, err
		}
	}

	return n, nil
}

type filteredWriter struct {
	writer      io.Writer
	hideTCPLogs bool
	hostAddr    string
	kaonRPC     string
}

func (f *filteredWriter) Write(p []byte) (n int, err error) {
	// Filter out logs containing the TCP information if hideTCPLogs is enabled
	// Pretend to have written the bytes, but actually discard them if it is enabled
	if f.hideTCPLogs && bytes.Contains(p, []byte(f.hostAddr)) {
		return len(p), nil
	}
	if f.hideTCPLogs && bytes.Contains(p, []byte(f.kaonRPC)) {
		return len(p), nil
	}

	return f.writer.Write(p)
}

func setupLogger(logFile *string, devMode *bool, hideTCPLogs *bool, bind *string, port *int) (io.Writer, log.Logger, error) {
	// Create the filter string
	filterStr := fmt.Sprintf("%s:%d", *bind, *port)

	// Main log file writer
	var mainWriter io.Writer = os.Stdout

	// Error log file writer (if required)
	var errorWriter io.Writer = os.Stdout

	if logFile != nil && (*logFile) != "" {
		// Main log file with rotation
		logRotator := &lumberjack.Logger{
			Filename:   *logFile,
			MaxSize:    50,    // Max size in MB before rotation
			MaxBackups: 10,    // Maximum number of old log files to keep
			MaxAge:     28,    // Maximum number of days to retain old log files (optional)
			Compress:   false, // Compress the rotated log files (optional)
		}
		mainWriter = io.MultiWriter(mainWriter, logRotator)

		// Error log file with rotation
		errorLogFile := *logFile + ".error"
		errorRotator := &lumberjack.Logger{
			Filename:   errorLogFile,
			MaxSize:    50,    // Max size in MB before rotation
			MaxBackups: 10,    // Maximum number of old log files to keep
			MaxAge:     28,    // Maximum number of days to retain old log files (optional)
			Compress:   false, // Compress the rotated log files (optional)
		}
		errorWriter = io.MultiWriter(errorWriter, errorRotator)
	}

	// Wrap mainWriter with filteredWriter to filter out TCP logs if needed
	filteredMainWriter := &filteredWriter{
		writer:      mainWriter,
		hideTCPLogs: *hideTCPLogs,
		hostAddr:    filterStr,
		kaonRPC:     *kaonRPC,
	}

	// Wrap errorWriter with filteredWriter to filter out TCP logs if needed
	filteredErrorWriter := &filteredWriter{
		writer:      errorWriter,
		hideTCPLogs: *hideTCPLogs,
		hostAddr:    filterStr,
		kaonRPC:     *kaonRPC,
	}

	// Custom multi writer that writes to both filtered main log file and filtered error log file
	multiWriter := &multiErrorWriter{
		mainWriter:  filteredMainWriter,
		errorWriter: filteredErrorWriter,
	}

	// Create a logger that uses the multiWriter
	logger := log.NewLogfmtLogger(multiWriter)

	// Filter log levels if not in devMode
	if !*devMode {
		logger = level.NewFilter(logger, level.AllowWarn())
	}

	return multiWriter, logger, nil
}

func action(pc *kingpin.ParseContext) error {
	addr := fmt.Sprintf("%s:%d", *bind, *port)

	logWriter, logger, err := setupLogger(logFile, devMode, hideTCPLogs, bind, port)
	if err != nil {
		return errors.Wrap(err, "Failed to set up logger")
	}

	var accounts kaon.Accounts
	if *accountsFile != nil {
		accounts = loadAccounts(*accountsFile, logger)
		(*accountsFile).Close()
	}

	isMain := *kaonNetwork == kaon.ChainMain

	ctx, shutdownKaon := context.WithCancel(context.Background())
	defer shutdownKaon()

	kaonRequestAnalytics := analytics.NewAnalytics(50)

	kaonJSONRPC, err := kaon.NewClient(
		isMain,
		*kaonRPC,
		kaon.SetDebug(*devMode),
		kaon.SetLogWriter(logWriter),
		kaon.SetLogger(logger),
		kaon.SetAccounts(accounts),
		kaon.SetGenerateToAddress(*generateToAddressTo),
		kaon.SetIgnoreUnknownTransactions(*ignoreUnknownTransactions),
		kaon.SetDisableSnippingKaonRpcOutput(*disableSnipping),
		kaon.SetHideKaondLogs(*hideKaondLogs),
		kaon.SetMatureBlockHeight(matureBlockHeight),
		kaon.SetContext(ctx),
		kaon.SetSqlHost(*sqlHost),
		kaon.SetSqlPort(*sqlPort),
		kaon.SetSqlUser(*sqlUser),
		kaon.SetSqlPassword(*sqlPassword),
		kaon.SetSqlSSL(*sqlSSL),
		kaon.SetSqlDatabaseName(*sqlDbname),
		kaon.SetSqlConnectionString(*dbConnectionString),
		kaon.SetAnalytics(kaonRequestAnalytics),
	)
	if err != nil {
		return errors.Wrap(err, "Failed to setup KAON client")
	}

	kaonClient, err := kaon.New(kaonJSONRPC, *kaonNetwork)
	if err != nil {
		return errors.Wrap(err, "Failed to setup KAON chain")
	}

	agent := notifier.NewAgent(context.Background(), kaonClient, nil)
	proxies := transformer.DefaultProxies(kaonClient, agent)
	t, err := transformer.New(
		kaonClient,
		proxies,
		transformer.SetDebug(*devMode),
		transformer.SetLogger(logger),
	)
	if err != nil {
		return errors.Wrap(err, "transformer#New")
	}
	agent.SetTransformer(t)

	httpsKeyFile := getEmptyStringIfFileDoesntExist(*httpsKey, logger)
	httpsCertFile := getEmptyStringIfFileDoesntExist(*httpsCert, logger)

	s, err := server.New(
		kaonClient,
		t,
		addr,
		server.SetLogWriter(logWriter),
		server.SetLogger(logger),
		server.SetDebug(*devMode),
		server.SetSingleThreaded(*singleThreaded),
		server.SetHttps(httpsKeyFile, httpsCertFile),
		server.SetKaonAnalytics(kaonRequestAnalytics),
		server.SetHealthCheckPercent(healthCheckPercent),
	)
	if err != nil {
		return errors.Wrap(err, "server#New")
	}

	return s.Start()
}

func getEmptyStringIfFileDoesntExist(file string, l log.Logger) string {
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		l.Log("file does not exist", file)
		return ""
	}
	return file
}

func Run() {
	app.Version(params.VersionWithGitSha)
	kingpin.MustParse(app.Parse(os.Args[1:]))
}

func init() {
	app.Action(action)
}
