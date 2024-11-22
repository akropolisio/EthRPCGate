package transformer

import (
	"context"
	"encoding/json"

	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/kaonone/eth-rpc-gate/pkg/kaon"
	"github.com/kaonone/eth-rpc-gate/pkg/utils"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

// ProxyKAONFromHexAddress implements ETHProxy
type ProxyKAONFromHexAddress struct {
	*kaon.Kaon
}

func (p *ProxyKAONFromHexAddress) Method() string {
	return "kaon_fromhexaddress"
}

func (p *ProxyKAONFromHexAddress) Request(req *eth.JSONRPCRequest, c echo.Context) (interface{}, *eth.JSONRPCError) {
	var address string

	var err error
	var params []json.RawMessage
	if err = json.Unmarshal(req.Params, &params); err != nil {
		return nil, eth.NewCallbackError(errors.Wrap(err, "json unmarshal").Error())
	}

	if len(params) == 0 {
		return nil, eth.NewInvalidParamsError("params must be set")
	}

	if err := json.Unmarshal(params[0], &address); err != nil {
		return nil, eth.NewCallbackError(errors.Wrap(err, "json unmarshal").Error())
	}

	if address == "" {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError("address is empty")
	}

	if !utils.IsEthHexAddress(address) {
		return nil, eth.NewInvalidParamsError("address is invalid")
	}
	address = utils.RemoveHexPrefix(address)

	return p.request(c.Request().Context(), &address)
}

func (p *ProxyKAONFromHexAddress) request(ctx context.Context, req *string) (interface{}, *eth.JSONRPCError) {
	kaonresp, err := p.Kaon.FromHexAddressWithContext(ctx, *req)
	if err != nil {
		return nil, eth.NewCallbackError(err.Error())
	}
	// kaon res -> eth res
	return p.ToResponse(kaonresp), nil
}

func (p *ProxyKAONFromHexAddress) ToResponse(hexaddr string) interface{} {
	return utils.RemoveHexPrefix(hexaddr)
}
