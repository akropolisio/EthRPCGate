package transformer

import (
	"github.com/go-kit/kit/log"
	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/kaonone/eth-rpc-gate/pkg/kaon"
	"github.com/kaonone/eth-rpc-gate/pkg/notifier"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

type Transformer struct {
	kaonClient   *kaon.Kaon
	debugMode    bool
	logger       log.Logger
	transformers map[string]ETHProxy
}

// New creates a new Transformer
func New(kaonClient *kaon.Kaon, proxies []ETHProxy, opts ...Option) (*Transformer, error) {
	if kaonClient == nil {
		return nil, errors.New("kaonClient cannot be nil")
	}

	t := &Transformer{
		kaonClient: kaonClient,
		logger:     log.NewNopLogger(),
	}

	var err error
	for _, p := range proxies {
		if err = t.Register(p); err != nil {
			return nil, err
		}
	}

	for _, opt := range opts {
		if err := opt(t); err != nil {
			return nil, err
		}
	}

	return t, nil
}

// Register registers an ETHProxy to a Transformer
func (t *Transformer) Register(p ETHProxy) error {
	if t.transformers == nil {
		t.transformers = make(map[string]ETHProxy)
	}

	m := p.Method()
	if _, ok := t.transformers[m]; ok {
		return errors.Errorf("method already exist: %s ", m)
	}

	t.transformers[m] = p

	return nil
}

// Transform takes a Transformer and transforms the request from ETH request and returns the proxy request
func (t *Transformer) Transform(req *eth.JSONRPCRequest, c echo.Context) (interface{}, *eth.JSONRPCError) {
	proxy, err := t.getProxy(req.Method)
	if err != nil {
		return nil, err
	}
	resp, err := proxy.Request(req, c)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (t *Transformer) getProxy(method string) (ETHProxy, *eth.JSONRPCError) {
	proxy, ok := t.transformers[method]
	if !ok {
		return nil, eth.NewMethodNotFoundError(method)
	}
	return proxy, nil
}

func (t *Transformer) IsDebugEnabled() bool {
	return t.debugMode
}

// DefaultProxies are the default proxy methods made available
func DefaultProxies(kaonRPCClient *kaon.Kaon, agent *notifier.Agent) []ETHProxy {
	filter := eth.NewFilterSimulator()
	getFilterChanges := &ProxyETHGetFilterChanges{Kaon: kaonRPCClient, filter: filter}
	ethCall := &ProxyETHCall{Kaon: kaonRPCClient}

	ethProxies := []ETHProxy{
		ethCall,
		&ProxyNetListening{Kaon: kaonRPCClient},
		&ProxyETHPersonalUnlockAccount{},
		&ProxyETHChainId{Kaon: kaonRPCClient},
		&ProxyETHBlockNumber{Kaon: kaonRPCClient},
		&ProxyETHHashrate{Kaon: kaonRPCClient},
		&ProxyETHMining{Kaon: kaonRPCClient},
		&ProxyETHNetVersion{Kaon: kaonRPCClient},
		&ProxyETHGetTransactionByHash{Kaon: kaonRPCClient},
		&ProxyETHGetTransactionByBlockNumberAndIndex{Kaon: kaonRPCClient},
		&ProxyETHGetLogs{Kaon: kaonRPCClient},
		&ProxyETHGetTransactionReceipt{Kaon: kaonRPCClient},
		&ProxyETHSendTransaction{Kaon: kaonRPCClient},
		&ProxyETHDebugTraceBlockByNumber{Kaon: kaonRPCClient},
		&ProxyETHTraceBlock{Kaon: kaonRPCClient},
		&ProxyETHDebugTraceTransaction{Kaon: kaonRPCClient},
		&ProxyETHAccounts{Kaon: kaonRPCClient},
		&ProxyETHGetCode{Kaon: kaonRPCClient},

		&ProxyETHNewFilter{Kaon: kaonRPCClient, filter: filter},
		&ProxyETHNewBlockFilter{Kaon: kaonRPCClient, filter: filter},
		getFilterChanges,
		&ProxyETHGetFilterLogs{ProxyETHGetFilterChanges: getFilterChanges},
		&ProxyETHUninstallFilter{Kaon: kaonRPCClient, filter: filter},

		&ProxyETHEstimateGas{ProxyETHCall: ethCall},
		&ProxyETHGetBlockByNumber{Kaon: kaonRPCClient},
		&ProxyETHGetBlockByHash{Kaon: kaonRPCClient},
		&ProxyETHGetBalance{Kaon: kaonRPCClient},
		&ProxyETHGetStorageAt{Kaon: kaonRPCClient},
		&ETHGetCompilers{},
		&ETHProtocolVersion{},
		&ETHGetUncleByBlockHashAndIndex{},
		&ETHGetUncleCountByBlockHash{},
		&ETHGetUncleCountByBlockNumber{},
		&Web3ClientVersion{},
		&Web3Sha3{},
		&ProxyETHSign{Kaon: kaonRPCClient},
		&ProxyETHGasPrice{Kaon: kaonRPCClient},
		&ProxyETHTxCount{Kaon: kaonRPCClient},
		&ProxyETHSignTransaction{Kaon: kaonRPCClient},
		&ProxyETHSendRawTransaction{Kaon: kaonRPCClient},

		&ETHSubscribe{Kaon: kaonRPCClient, Agent: agent},
		&ETHUnsubscribe{Kaon: kaonRPCClient, Agent: agent},

		&ProxyKAONGetUTXOs{Kaon: kaonRPCClient},
		&ProxyKAONGenerateToAddress{Kaon: kaonRPCClient},

		&ProxyNetPeerCount{Kaon: kaonRPCClient},
	}

	permittedKaonCalls := []string{
		kaon.MethodGetHexAddress,
		kaon.MethodFromHexAddress,
	}

	for _, kaonMethod := range permittedKaonCalls {
		ethProxies = append(
			ethProxies,
			&ProxyKAONGenericStringArguments{
				Kaon:   kaonRPCClient,
				prefix: "dev",
				method: kaonMethod,
			},
		)
	}

	return ethProxies
}

func SetDebug(debug bool) func(*Transformer) error {
	return func(t *Transformer) error {
		t.debugMode = debug
		return nil
	}
}

func SetLogger(l log.Logger) func(*Transformer) error {
	return func(t *Transformer) error {
		t.logger = log.WithPrefix(l, "component", "transformer")
		return nil
	}
}
