package transformer

import (
	"context"

	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/kaonone/eth-rpc-gate/pkg/kaon"
	"github.com/kaonone/eth-rpc-gate/pkg/utils"
	"github.com/labstack/echo"
)

// ProxyETHGetCode implements ETHProxy
type ProxyETHGetCode struct {
	*kaon.Kaon
}

func (p *ProxyETHGetCode) Method() string {
	return "eth_getCode"
}

func (p *ProxyETHGetCode) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, *eth.JSONRPCError) {
	var req eth.GetCodeRequest
	if err := unmarshalRequest(rawreq.Params, &req); err != nil {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError(err.Error())
	}

	return p.request(c.Request().Context(), &req)
}

func (p *ProxyETHGetCode) request(ctx context.Context, ethreq *eth.GetCodeRequest) (eth.GetCodeResponse, *eth.JSONRPCError) {
	kaonreq := kaon.GetAccountInfoRequest(utils.RemoveHexPrefix(ethreq.Address))

	kaonresp, err := p.GetAccountInfo(ctx, &kaonreq)
	if err != nil {
		if err == kaon.ErrInvalidAddress {
			/**
			// correct response for an invalid address
			{
				"jsonrpc": "2.0",
				"id": 123,
				"result": "0x"
			}
			**/
			return "0x", nil
		} else {
			return "", eth.NewCallbackError(err.Error())
		}
	}

	// kaon res -> eth res
	return eth.GetCodeResponse(utils.AddHexPrefix(kaonresp.Code)), nil
}
