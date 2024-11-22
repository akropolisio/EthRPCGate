package transformer

import (
	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/labstack/echo"
)

type ETHProtocolVersion struct {
}

func (p *ETHProtocolVersion) Method() string {
	return "eth_protocolVersion"
}

func (p *ETHProtocolVersion) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, *eth.JSONRPCError) {
	return "0x41", nil
}
