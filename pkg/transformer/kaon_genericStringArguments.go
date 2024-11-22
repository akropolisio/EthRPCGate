package transformer

import (
	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/kaonone/eth-rpc-gate/pkg/kaon"
	"github.com/labstack/echo"
)

type ProxyKAONGenericStringArguments struct {
	*kaon.Kaon
	prefix string
	method string
}

var _ ETHProxy = (*ProxyKAONGenericStringArguments)(nil)

func (p *ProxyKAONGenericStringArguments) Method() string {
	return p.prefix + "_" + p.method
}

func (p *ProxyKAONGenericStringArguments) Request(req *eth.JSONRPCRequest, c echo.Context) (interface{}, *eth.JSONRPCError) {
	var params eth.StringsArguments
	if err := unmarshalRequest(req.Params, &params); err != nil {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError("couldn't unmarshal request parameters")
	}

	if len(params) != 1 {
		return nil, eth.NewInvalidParamsError("require 1 argument: the base58 Kaon address")
	}

	return p.request(params)
}

func (p *ProxyKAONGenericStringArguments) request(params eth.StringsArguments) (*string, *eth.JSONRPCError) {
	var response string
	err := p.Client.Request(p.method, params, &response)
	if err != nil {
		return nil, eth.NewInvalidRequestError(err.Error())
	}

	return &response, nil
}
