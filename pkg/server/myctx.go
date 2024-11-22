package server

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/kaonone/eth-rpc-gate/pkg/analytics"
	"github.com/kaonone/eth-rpc-gate/pkg/blockhash"
	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/kaonone/eth-rpc-gate/pkg/transformer"
	"github.com/labstack/echo"
)

type myCtx struct {
	echo.Context
	rpcReq      *eth.JSONRPCRequest
	logWriter   io.Writer
	logger      log.Logger
	transformer *transformer.Transformer
	blockHash   *blockhash.BlockHash

	healthCheckPercent *int
	kaonAnalytics      *analytics.Analytics
	ethAnalytics       *analytics.Analytics
}

func (c *myCtx) GetJSONRPCResult(result interface{}) (*eth.JSONRPCResult, error) {
	return eth.NewJSONRPCResult(c.rpcReq.ID, result)
}

func (c *myCtx) JSONRPCResult(result interface{}) error {
	response, err := c.GetJSONRPCResult(result)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response)
}

func (c *myCtx) GetJSONRPCError(err *eth.JSONRPCError) *eth.JSONRPCResult {
	var id json.RawMessage
	if c.rpcReq != nil && c.rpcReq.ID != nil {
		id = c.rpcReq.ID
	}
	return &eth.JSONRPCResult{
		ID:      id,
		Error:   err,
		JSONRPC: eth.RPCVersion,
	}
}

func (c *myCtx) JSONRPCError(errIn *eth.JSONRPCError) error {
	resp := c.GetJSONRPCError(errIn)

	if !c.Response().Committed {
		err := c.JSON(http.StatusOK, resp)
		c.logger.Log("Internal server error", err)
		return err
	}

	return nil
}

func (c *myCtx) SetLogWriter(logWriter io.Writer) {
	c.logWriter = logWriter
}

func (c *myCtx) GetLogWriter() io.Writer {
	return c.logWriter
}

func (c *myCtx) SetLogger(l log.Logger) {
	c.logger = log.WithPrefix(l, "component", "context")
}

func (c *myCtx) GetLogger() log.Logger {
	return c.logger
}

func (c *myCtx) GetDebugLogger() log.Logger {
	if !c.IsDebugEnabled() {
		return log.NewNopLogger()
	}
	return log.With(level.Debug(c.logger))
}

func (c *myCtx) GetErrorLogger() log.Logger {
	return log.With(level.Error(c.logger))
}

func (c *myCtx) IsDebugEnabled() bool {
	return c.transformer.IsDebugEnabled()
}
