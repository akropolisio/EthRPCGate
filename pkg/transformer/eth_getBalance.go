package transformer

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/kaonone/eth-rpc-gate/pkg/kaon"
	"github.com/kaonone/eth-rpc-gate/pkg/utils"
	"github.com/labstack/echo"
)

// ProxyETHGetBalance implements ETHProxy
type ProxyETHGetBalance struct {
	*kaon.Kaon
}

func (p *ProxyETHGetBalance) Method() string {
	return "eth_getBalance"
}

func (p *ProxyETHGetBalance) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, *eth.JSONRPCError) {
	var req eth.GetBalanceRequest
	if err := unmarshalRequest(rawreq.Params, &req); err != nil {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError(err.Error())
	}

	if req.Address == "0x0000000000000000000000000000000000000000" { // ErrInvalidAddress
		return "0x0", nil
	}

	addr := utils.RemoveHexPrefix(req.Address)
	{
		// is address a contract or an account?
		kaonreq := kaon.GetAccountInfoRequest(addr)
		kaonresp, err := p.GetAccountInfo(c.Request().Context(), &kaonreq)

		// the address is a contract
		if err == nil {
			// the unit of the balance Satoshi
			p.GetDebugLogger().Log("method", p.Method(), "address", req.Address, "msg", "is a contract")
			return hexutil.EncodeBig(&kaonresp.Balance), nil
		}
	}

	{
		// try account
		base58Addr, err := p.FromHexAddress(addr)
		if err != nil {
			p.GetDebugLogger().Log("method", p.Method(), "address", req.Address, "msg", "error parsing address", "error", err)
			return nil, eth.NewCallbackError(err.Error())
		}

		kaonreq := kaon.GetAddressBalanceRequest{Address: base58Addr}
		kaonresp, err := p.GetAddressBalance(c.Request().Context(), &kaonreq)
		if err != nil {
			if err == kaon.ErrInvalidAddress {
				// invalid address should return 0x0
				return "0x0", nil
			}
			p.GetDebugLogger().Log("method", p.Method(), "address", req.Address, "msg", "error getting address balance", "error", err)
			return nil, eth.NewCallbackError(err.Error())
		}

		// 1 KAON = 10 ^ 18 Satoshi
		balance := kaonresp.Balance

		return hexutil.EncodeBig(&balance), nil
	}
}
