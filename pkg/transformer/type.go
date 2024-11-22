package transformer

import (
	"errors"

	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/labstack/echo"
)

var UnmarshalRequestErr = errors.New("Input is invalid")

type Option func(*Transformer) error

type ETHProxy interface {
	Request(*eth.JSONRPCRequest, echo.Context) (interface{}, *eth.JSONRPCError)
	Method() string
}
