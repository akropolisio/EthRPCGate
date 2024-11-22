package transformer

import (
	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/kaonone/eth-rpc-gate/pkg/kaon"
	"github.com/kaonone/eth-rpc-gate/pkg/utils"
	"github.com/labstack/echo"
)

// ProxyETHAccounts implements ETHProxy
type ProxyETHAccounts struct {
	*kaon.Kaon
}

func (p *ProxyETHAccounts) Method() string {
	return "eth_accounts"
}

func (p *ProxyETHAccounts) Request(_ *eth.JSONRPCRequest, c echo.Context) (interface{}, *eth.JSONRPCError) {
	return p.request()
}

func (p *ProxyETHAccounts) request() (eth.AccountsResponse, *eth.JSONRPCError) {
	var accounts eth.AccountsResponse

	for _, acc := range p.Accounts {
		acc := kaon.Account{WIF: acc}
		addr := acc.ToHexAddress()

		accounts = append(accounts, utils.AddHexPrefix(addr))
	}

	return accounts, nil
}

func (p *ProxyETHAccounts) ToResponse(ethresp *kaon.CallContractResponse) *eth.CallResponse {
	data := utils.AddHexPrefix(ethresp.ExecutionResult.Output)
	kaonresp := eth.CallResponse(data)
	return &kaonresp
}
