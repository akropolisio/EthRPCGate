package transformer

import (
	"context"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/kaonone/eth-rpc-gate/pkg/kaon"
	"github.com/kaonone/eth-rpc-gate/pkg/utils"
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
	var req eth.GetTransactionCountRequest
	if err := unmarshalRequest(rawreq.Params, &req); err != nil {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError(err.Error())
	}

	kaonAddress := utils.RemoveHexPrefix(req.Address)

	return p.request(
		c.Request().Context(),
		&kaon.GetTransactionCountRequest{
			Address:     kaonAddress,
			BlockNumber: req.Tag,
		},
	)
}

func (p *ProxyETHTxCount) request(ctx context.Context, ethreq *kaon.GetTransactionCountRequest) (*eth.GetTransactionCountResponse, *eth.JSONRPCError) {
	kaonresp, err := p.Kaon.GetTransactionCount(ctx, ethreq)
	if err != nil {
		return nil, eth.NewCallbackError(err.Error())
	}

	// kaon res -> eth res
	return p.ToResponse(kaonresp), nil
}

func (p *ProxyETHTxCount) ToResponse(kaonresp *kaon.GetTransactionCountResponse) *eth.GetTransactionCountResponse {
	hexVal := hexutil.EncodeBig(kaonresp.Int)
	ethresp := eth.GetTransactionCountResponse(hexVal)
	return &ethresp
}
