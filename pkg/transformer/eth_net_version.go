package transformer

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/kaonone/eth-rpc-gate/pkg/kaon"
	"github.com/labstack/echo"
)

// ProxyETHNetVersion implements ETHProxy
type ProxyETHNetVersion struct {
	*kaon.Kaon
}

func (p *ProxyETHNetVersion) Method() string {
	return "net_version"
}

func (p *ProxyETHNetVersion) Request(_ *eth.JSONRPCRequest, c echo.Context) (interface{}, *eth.JSONRPCError) {
	return p.request()
}

func (p *ProxyETHNetVersion) request() (*eth.NetVersionResponse, *eth.JSONRPCError) {
	networkID, err := getChainId(p.Kaon)
	if err != nil {
		return nil, err
	}
	response := eth.NetVersionResponse(hexutil.EncodeBig(networkID))
	return &response, nil
}
