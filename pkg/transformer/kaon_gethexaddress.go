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

// ProxyKAONGetHexAddress implements ETHProxy
type ProxyKAONGetHexAddress struct {
	*kaon.Kaon
}

func (p *ProxyKAONGetHexAddress) Method() string {
	return "kaon_gethexaddress"
}

func (p *ProxyKAONGetHexAddress) Request(req *eth.JSONRPCRequest, c echo.Context) (interface{}, *eth.JSONRPCError) {
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
