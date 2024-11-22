package transformer

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/kaonone/eth-rpc-gate/pkg/kaon"
	"github.com/labstack/echo"
)

// ProxyETHEstimateGas implements ETHProxy
type ProxyETHGasPrice struct {
	*kaon.Kaon
}

func (p *ProxyETHGasPrice) Method() string {
	return "eth_gasPrice"
}

func (p *ProxyETHGasPrice) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, *eth.JSONRPCError) {
	kaonresp, err := p.Kaon.GetGasPrice(c.Request().Context())
	if err != nil {
		return nil, eth.NewCallbackError(err.Error())
	}

	// kaon res -> eth res
	return p.response(kaonresp), nil
}

func (p *ProxyETHGasPrice) response(kaonresp *big.Int) string {
	// 34 GWEI is the minimum price that KAON will confirm tx with
	return hexutil.EncodeBig(convertFromSatoshiToWei(kaonresp))
}
