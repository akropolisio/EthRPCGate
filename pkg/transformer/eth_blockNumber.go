package transformer

import (
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/kaonone/eth-rpc-gate/pkg/kaon"
	"github.com/labstack/echo"
)

// ProxyETHBlockNumber implements ETHProxy
type ProxyETHBlockNumber struct {
	*kaon.Kaon
}

func (p *ProxyETHBlockNumber) Method() string {
	return "eth_blockNumber"
}

func (p *ProxyETHBlockNumber) Request(_ *eth.JSONRPCRequest, c echo.Context) (interface{}, *eth.JSONRPCError) {
	return p.request(c, 5)
}

func (p *ProxyETHBlockNumber) request(c echo.Context, retries int) (*eth.BlockNumberResponse, *eth.JSONRPCError) {
	kaonresp, err := p.Kaon.GetBlockCount(c.Request().Context())
	if err != nil {
		if retries > 0 && strings.Contains(err.Error(), kaon.ErrTryAgain.Error()) {
			ctx := c.Request().Context()
			t := time.NewTimer(500 * time.Millisecond)
			select {
			case <-ctx.Done():
				return nil, eth.NewCallbackError(err.Error())
			case <-t.C:
				// fallthrough
			}
			return p.request(c, retries-1)
		}
		return nil, eth.NewCallbackError(err.Error())
	}

	// kaon res -> eth res
	return p.ToResponse(kaonresp), nil
}

func (p *ProxyETHBlockNumber) ToResponse(kaonresp *kaon.GetBlockCountResponse) *eth.BlockNumberResponse {
	hexVal := hexutil.EncodeBig(kaonresp.Int)
	ethresp := eth.BlockNumberResponse(hexVal)
	return &ethresp
}
