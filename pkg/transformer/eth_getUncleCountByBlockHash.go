package transformer

import (
	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/labstack/echo"
)

type ETHGetUncleCountByBlockHash struct {
}

func (p *ETHGetUncleCountByBlockHash) Method() string {
	return "eth_getUncleCountByBlockHash"
}

func (p *ETHGetUncleCountByBlockHash) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, *eth.JSONRPCError) {
	// hardcoded to 0
	return 0, nil
}
