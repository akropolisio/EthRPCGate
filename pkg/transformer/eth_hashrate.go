package transformer

import (
	"context"
	"math"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/kaonone/eth-rpc-gate/pkg/kaon"
	"github.com/labstack/echo"
)

// ProxyETHGetHashrate implements ETHProxy
type ProxyETHHashrate struct {
	*kaon.Kaon
}

func (p *ProxyETHHashrate) Method() string {
	return "eth_hashrate"
}

func (p *ProxyETHHashrate) Request(_ *eth.JSONRPCRequest, c echo.Context) (interface{}, *eth.JSONRPCError) {
	return p.request(c.Request().Context())
}

func (p *ProxyETHHashrate) request(ctx context.Context) (*eth.HashrateResponse, *eth.JSONRPCError) {
	kaonresp, err := p.Kaon.GetHashrate(ctx)
	if err != nil {
		return nil, eth.NewCallbackError(err.Error())
	}

	// kaon res -> eth res
	return p.ToResponse(kaonresp), nil
}

func (p *ProxyETHHashrate) ToResponse(kaonresp *kaon.GetHashrateResponse) *eth.HashrateResponse {
	hexVal := hexutil.EncodeUint64(math.Float64bits(kaonresp.Difficulty))
	ethresp := eth.HashrateResponse(hexVal)
	return &ethresp
}
