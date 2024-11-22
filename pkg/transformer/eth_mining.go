package transformer

import (
	"context"

	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/kaonone/eth-rpc-gate/pkg/kaon"
	"github.com/labstack/echo"
)

// ProxyETHGetHashrate implements ETHProxy
type ProxyETHMining struct {
	*kaon.Kaon
}

func (p *ProxyETHMining) Method() string {
	return "eth_mining"
}

func (p *ProxyETHMining) Request(_ *eth.JSONRPCRequest, c echo.Context) (interface{}, *eth.JSONRPCError) {
	return p.request(c.Request().Context())
}

func (p *ProxyETHMining) request(ctx context.Context) (*eth.MiningResponse, *eth.JSONRPCError) {
	kaonresp, err := p.Kaon.GetMining(ctx)
	if err != nil {
		return nil, eth.NewCallbackError(err.Error())
	}

	// kaon res -> eth res
	return p.ToResponse(kaonresp), nil
}

func (p *ProxyETHMining) ToResponse(kaonresp *kaon.GetMiningResponse) *eth.MiningResponse {
	ethresp := eth.MiningResponse(kaonresp.Staking)
	return &ethresp
}
