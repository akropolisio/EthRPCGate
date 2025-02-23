package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/heptiolabs/healthcheck"
	"github.com/kaonone/eth-rpc-gate/pkg/analytics"
	"github.com/kaonone/eth-rpc-gate/pkg/blockhash"
	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/kaonone/eth-rpc-gate/pkg/kaon"
	"github.com/kaonone/eth-rpc-gate/pkg/transformer"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/pkg/errors"
)

type Server struct {
	address       string
	transformer   *transformer.Transformer
	kaonRPCClient *kaon.Kaon
	logWriter     io.Writer
	logger        log.Logger
	httpsKey      string
	httpsCert     string
	debug         bool
	mutex         *sync.Mutex
	echo          *echo.Echo
	blockHash     *blockhash.BlockHash

	healthCheckPercent   *int
	kaonRequestAnalytics *analytics.Analytics
	ethRequestAnalytics  *analytics.Analytics

	blocksMutex     sync.RWMutex
	lastBlock       int64
	nextBlockCheck  *time.Time
	lastBlockStatus error
}

func New(
	kaonRPCClient *kaon.Kaon,
	transformer *transformer.Transformer,
	addr string,
	opts ...Option,
) (*Server, error) {
	requests := 50

	p := &Server{
		logger:              log.NewNopLogger(),
		echo:                echo.New(),
		address:             addr,
		kaonRPCClient:       kaonRPCClient,
		transformer:         transformer,
		ethRequestAnalytics: analytics.NewAnalytics(requests),
	}

	blockHashProcessor, err := blockhash.NewBlockHash(
		kaonRPCClient.GetContext(),
		func() log.Logger {
			return p.kaonRPCClient.GetLogger()
		},
	)
	if err != nil {
		return nil, err
	}

	p.blockHash = blockHashProcessor

	for _, opt := range opts {
		if err = opt(p); err != nil {
			return nil, err
		}
	}

	return p, nil
}

func (s *Server) Start() error {
	logWriter := s.logWriter
	e := s.echo

	health := healthcheck.NewHandler()
	health.AddLivenessCheck("kaond-connection", func() error { return s.testConnectionToKaond() })
	health.AddLivenessCheck("kaond-logevents-enabled", func() error { return s.testLogEvents() })
	health.AddLivenessCheck("kaond-blocks-syncing", func() error { return s.testBlocksSyncing() })
	health.AddLivenessCheck("kaond-error-rate", func() error { return s.testKaondErrorRate() })
	health.AddLivenessCheck("ethrpcgate-error-rate", func() error { return s.testEthRPCGateErrorRate() })

	e.Use(middleware.CORS())
	e.Use(middleware.BodyDump(func(c echo.Context, req []byte, res []byte) {
		myctx := c.Get("myctx")
		cc, ok := myctx.(*myCtx)
		if !ok {
			return
		}

		if s.debug {
			reqBody, reqErr := kaon.ReformatJSON(req)
			resBody, resErr := kaon.ReformatJSON(res)
			if reqErr == nil && resErr == nil {
				cc.GetDebugLogger().Log("msg", "ETH RPC")
				fmt.Fprintf(logWriter, "=> ETH request\n%s\n", reqBody)
				fmt.Fprintf(logWriter, "<= ETH response\n%s\n", resBody)
			} else if reqErr != nil {
				cc.GetErrorLogger().Log("msg", "Error reformatting request json", "error", reqErr, "body", string(req))
			} else {
				cc.GetErrorLogger().Log("msg", "Error reformatting response json", "error", resErr, "body", string(res))
			}
		}
	}))

	e.Use(func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &myCtx{
				Context:       c,
				logWriter:     logWriter,
				logger:        s.logger,
				transformer:   s.transformer,
				blockHash:     s.blockHash,
				kaonAnalytics: s.kaonRequestAnalytics,
				ethAnalytics:  s.ethRequestAnalytics,
			}

			c.Set("myctx", cc)
			c.Set("blockHash", cc.blockHash)

			return h(c)
		}
	})

	// support batch requests
	e.Use(batchRequestsMiddleware)

	e.HTTPErrorHandler = errorHandler
	e.HideBanner = true
	if health != nil {
		e.GET("/live", func(c echo.Context) error {
			health.LiveEndpoint(c.Response(), c.Request())
			return nil
		})
		e.GET("/ready", func(c echo.Context) error {
			health.ReadyEndpoint(c.Response(), c.Request())
			return nil
		})
	}

	if s.mutex == nil {
		e.POST("/*", httpHandler)
		e.GET("/*", websocketHandler)
	} else {
		level.Info(s.logger).Log("msg", "Processing RPC requests single threaded")
		e.POST("/*", func(c echo.Context) error {
			s.mutex.Lock()
			defer s.mutex.Unlock()
			return httpHandler(c)
		})
		e.GET("/*", websocketHandler)
	}

	https := (s.httpsKey != "" && s.httpsCert != "")
	url := s.kaonRPCClient.GetURL().Redacted()
	level.Info(s.logger).Log("listen", s.address, "KAON_RPC", url, "msg", "proxy started", "https", https)

	var err error

	// shutdown echo server when context ends
	go func(ctx context.Context, e *echo.Echo) {
		<-ctx.Done()
		e.Close()
	}(s.kaonRPCClient.GetContext(), e)

	if s.kaonRPCClient.DbConfig.String() == "" {
		level.Warn(s.logger).Log("msg", "Database not configured - won't be able to respond to Ethereum block hash requests")
	} else {
		chainIdChan := make(chan int, 1)
		err := s.blockHash.Start(&s.kaonRPCClient.DbConfig, chainIdChan)
		if err != nil {
			level.Error(s.logger).Log("msg", "Failed to launch block hash converter", "error", err)
		}

		go func() {
			chainIdChan <- s.kaonRPCClient.ChainId()
		}()
	}

	if https {
		level.Info(s.logger).Log("msg", "SSL enabled")
		err = e.StartTLS(s.address, s.httpsCert, s.httpsKey)
	} else {
		err = e.Start(s.address)
	}

	return err
}

type Option func(*Server) error

func SetLogWriter(logWriter io.Writer) Option {
	return func(p *Server) error {
		p.logWriter = logWriter
		return nil
	}
}

func SetLogger(l log.Logger) Option {
	return func(p *Server) error {
		p.logger = l
		return nil
	}
}

func SetDebug(debug bool) Option {
	return func(p *Server) error {
		p.debug = debug
		return nil
	}
}

func SetSingleThreaded(singleThreaded bool) Option {
	return func(p *Server) error {
		if singleThreaded {
			p.mutex = &sync.Mutex{}
		} else {
			p.mutex = nil
		}
		return nil
	}
}

func SetHttps(key string, cert string) Option {
	return func(p *Server) error {
		p.httpsKey = key
		p.httpsCert = cert
		return nil
	}
}

func SetKaonAnalytics(analytics *analytics.Analytics) Option {
	return func(p *Server) error {
		p.kaonRequestAnalytics = analytics
		return nil
	}
}

func SetHealthCheckPercent(percent *int) Option {
	return func(p *Server) error {
		p.healthCheckPercent = percent
		return nil
	}
}

func batchRequestsMiddleware(h echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		myctx := c.Get("myctx")
		cc, ok := myctx.(*myCtx)
		if !ok {
			return errors.New("Could not find myctx")
		}

		// Request
		reqBody := []byte{}
		if c.Request().Body != nil { // Read
			var err error
			reqBody, err = ioutil.ReadAll(c.Request().Body)
			if err != nil {
				panic(fmt.Sprintf("%v", err))
			}
		}
		isBatchRequests := func(msg json.RawMessage) bool {
			return len(msg) != 0 && msg[0] == '['
		}
		c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(reqBody)) // Reset

		if !isBatchRequests(reqBody) {
			return h(c)
		}

		var rpcReqs []*eth.JSONRPCRequest
		if err := c.Bind(&rpcReqs); err != nil {

			return err
		}

		results := make([]*eth.JSONRPCResult, 0, len(rpcReqs))

		for _, req := range rpcReqs {
			result, err := callHttpHandler(cc, req)
			if err != nil {
				return err
			}

			results = append(results, result)
		}

		return c.JSON(http.StatusOK, results)
	}
}

func callHttpHandler(cc *myCtx, req *eth.JSONRPCRequest) (*eth.JSONRPCResult, error) {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpreq := httptest.NewRequest(echo.POST, "/", ioutil.NopCloser(bytes.NewReader(reqBytes)))
	httpreq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	newCtx := cc.Echo().NewContext(httpreq, rec)
	myCtx := &myCtx{
		Context:       newCtx,
		logWriter:     cc.GetLogWriter(),
		logger:        cc.logger,
		transformer:   cc.transformer,
		blockHash:     cc.blockHash,
		kaonAnalytics: cc.kaonAnalytics,
		ethAnalytics:  cc.ethAnalytics,
	}
	newCtx.Set("myctx", myCtx)
	if err = httpHandler(myCtx); err != nil {
		errorHandler(err, myCtx)
	}

	var result *eth.JSONRPCResult
	if err = json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		return nil, err
	}

	return result, nil
}
