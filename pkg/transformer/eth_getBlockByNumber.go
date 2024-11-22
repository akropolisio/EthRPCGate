package transformer

import (
	"context"
	"math/big"

	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/kaonone/eth-rpc-gate/pkg/kaon"
	"github.com/labstack/echo"
)

// ProxyETHGetBlockByNumber implements ETHProxy
type ProxyETHGetBlockByNumber struct {
	*kaon.Kaon
}

func (p *ProxyETHGetBlockByNumber) Method() string {
	return "eth_getBlockByNumber"
}

func (p *ProxyETHGetBlockByNumber) Request(rpcReq *eth.JSONRPCRequest, c echo.Context) (interface{}, *eth.JSONRPCError) {
	req := new(eth.GetBlockByNumberRequest)
	if err := unmarshalRequest(rpcReq.Params, req); err != nil {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError(err.Error())
	}
	return p.request(c.Request().Context(), req)
}

func (p *ProxyETHGetBlockByNumber) request(ctx context.Context, req *eth.GetBlockByNumberRequest) (*eth.GetBlockByNumberResponse, *eth.JSONRPCError) {
	blockNum, err := getBlockNumberByRawParam(ctx, p.Kaon, req.BlockNumber, false)
	if err != nil {
		return nil, eth.NewCallbackError("couldn't get block number by parameter")
	}

	blockHash, jsonErr := proxyETHGetBlockByHash(ctx, p, p.Kaon, blockNum)
	if jsonErr != nil {
		return nil, eth.NewInvalidParamsError(jsonErr.Message())
	}
	if blockHash == nil {
		return nil, nil
	}

	var (
		getBlockByHashReq = &eth.GetBlockByHashRequest{
			BlockHash:       string(*blockHash),
			FullTransaction: req.FullTransaction,
		}
		proxy = &ProxyETHGetBlockByHash{Kaon: p.Kaon}
	)
	block, jsonErr := proxy.request(ctx, getBlockByHashReq)
	if jsonErr != nil {
		p.GetDebugLogger().Log("function", p.Method(), "msg", "couldn't get block by hash", "jsonErr", jsonErr.Message())
		return nil, eth.NewCallbackError("couldn't get block by hash")
	}
	if blockNum != nil {
		p.GetDebugLogger().Log("function", p.Method(), "request", string(req.BlockNumber), "msg", "Successfully got block by number", "result", blockNum.String())
	}
	return block, nil
}

func proxyETHGetBlockByHash(ctx context.Context, p ETHProxy, q *kaon.Kaon, blockNum *big.Int) (*kaon.GetBlockHashResponse, *eth.JSONRPCError) {
	// Attempt to get the block hash from Kaon
	resp, err := q.GetBlockHash(ctx, blockNum)
	if err != nil {
		// Handle specific known errors
		if err == kaon.ErrInvalidParameter {
			// block doesn't exist; return null as per ETH RPC spec
			/**
			{
				"jsonrpc": "2.0",
				"id": 1234,
				"result": null
			}
			**/
			q.GetDebugLogger().Log(
				"function", p.Method(),
				"request", blockNum.String(),
				"msg", "Unknown block",
				"error", err.Error(),
			)
			return nil, nil
		}

		// Catch-all for any other unknown errors
		q.GetDebugLogger().Log(
			"function", p.Method(),
			"request", blockNum.String(),
			"msg", "Unexpected error occurred while getting block hash",
			"error", err.Error(),
		)
		return nil, eth.NewCallbackError("unexpected error: " + err.Error())
	}

	// Successfully retrieved block hash
	return &resp, nil
}
