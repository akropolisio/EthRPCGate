package transformer

import (
	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/labstack/echo"
)

type ETHGetCompilers struct {
}

func (p *ETHGetCompilers) Method() string {
	return "eth_getCompilers"
}

func (p *ETHGetCompilers) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, *eth.JSONRPCError) {
	// hardcoded to empty
	return []string{}, nil
}
