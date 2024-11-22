package transformer

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/kaonone/eth-rpc-gate/pkg/kaon"
	"github.com/labstack/echo"
)

// ProxyETHTxCount implements ETHProxy
type ProxyETHTxCount struct {
	*kaon.Kaon
}

func (p *ProxyETHTxCount) Method() string {
	return "eth_getTransactionCount"
}

func (p *ProxyETHTxCount) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, *eth.JSONRPCError) {
	kaonresp, err := p.Kaon.GetTransactionCount(c.Request().Context(), "", "")
	if err != nil {
		return nil, eth.NewCallbackError(err.Error())
	}

	// kaon res -> eth res
	return p.response(kaonresp), nil
}

func (p *ProxyETHTxCount) response(kaonresp *big.Int) string {
	return hexutil.EncodeBig(kaonresp)
}
