package transformer

import (
	"context"
	"encoding/json"

	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/kaonone/eth-rpc-gate/pkg/kaon"
	"github.com/kaonone/eth-rpc-gate/pkg/utils"
	"github.com/labstack/echo"
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
	if err := json.Unmarshal(req.Params, &address); err != nil {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError("couldn't unmarshal request")
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
	return utils.AddHexPrefix(hexaddr)
}
