package transformer

import (
	"context"
	"encoding/json"

	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/kaonone/eth-rpc-gate/pkg/kaon"
	"github.com/kaonone/eth-rpc-gate/pkg/utils"
	"github.com/labstack/echo"
)

// ProxyKAONGetHexAddress implements ETHProxy
type ProxyKAONGetHexAddress struct {
	*kaon.Kaon
}

func (p *ProxyKAONGetHexAddress) Method() string {
	return "kaon_gethexaddress"
}

func (p *ProxyKAONGetHexAddress) Request(req *eth.JSONRPCRequest, c echo.Context) (interface{}, *eth.JSONRPCError) {
	var address string
	if err := json.Unmarshal(req.Params, &address); err != nil {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError("couldn't unmarshal request")
	}
	if address == "" {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError("address is empty")
	}

	return p.request(c.Request().Context(), &address)
}

func (p *ProxyKAONGetHexAddress) request(ctx context.Context, req *string) (interface{}, *eth.JSONRPCError) {
	addr, err := utils.ConvertKaonAddress(*req)
	if err == nil {
		return p.ToResponse(addr), nil
	} else {
		// P2Sh address(such as MUrenj2sPqEVTiNbHQ2RARiZYyTAAeKiDX) and BECH32 address (such as qc1qkt33x6hkrrlwlr6v59wptwy6zskyrjfe40y0lx)
		// will cause ConvertKaonAddress to fall
		addr, rerr := p.Kaon.GetHexAddress(ctx, *req)
		if rerr == nil {
			return p.ToResponse(addr), nil
		}
		return nil, eth.NewCallbackError(rerr.Error())
	}
}

func (p *ProxyKAONGetHexAddress) ToResponse(hexaddr string) interface{} {
	return utils.AddHexPrefix(hexaddr)
}
