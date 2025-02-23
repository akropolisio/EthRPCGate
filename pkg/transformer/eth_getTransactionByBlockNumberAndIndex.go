package transformer

import (
	"context"
	"encoding/json"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/kaonone/eth-rpc-gate/pkg/kaon"
	"github.com/labstack/echo"
)

// ProxyETHGetTransactionByBlockNumberAndIndex implements ETHProxy
type ProxyETHGetTransactionByBlockNumberAndIndex struct {
	*kaon.Kaon
}

func (p *ProxyETHGetTransactionByBlockNumberAndIndex) Method() string {
	return "eth_getTransactionByBlockNumberAndIndex"
}

func (p *ProxyETHGetTransactionByBlockNumberAndIndex) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, *eth.JSONRPCError) {
	var req eth.GetTransactionByBlockNumberAndIndex
	if err := json.Unmarshal(rawreq.Params, &req); err != nil {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError("couldn't unmarshal request")
	}
	if req.BlockNumber == "" {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError("invalid argument 0: empty hex string")
	}

	return p.request(c.Request().Context(), &req)
}

func (p *ProxyETHGetTransactionByBlockNumberAndIndex) request(ctx context.Context, req *eth.GetTransactionByBlockNumberAndIndex) (interface{}, *eth.JSONRPCError) {
	// Decoded by ProxyETHGetTransactionByBlockHashAndIndex, quickly decode so we can fail cheaply without making any calls
	_, decodeErr := hexutil.DecodeUint64(req.TransactionIndex)
	if decodeErr != nil {
		return nil, eth.NewInvalidParamsError("invalid argument 1")
	}

	blockNum, err := getBlockNumberByParam(ctx, p.Kaon, req.BlockNumber, false)
	if err != nil {
		return nil, eth.NewCallbackError("couldn't get block number by parameter")
	}

	blockHash, err := proxyETHGetBlockByHash(ctx, p, p.Kaon, blockNum)
	if err != nil {
		return nil, err
	}
	if blockHash == nil {
		return nil, nil
	}

	var (
		getBlockByHashReq = &eth.GetTransactionByBlockHashAndIndex{
			BlockHash:        string(*blockHash),
			TransactionIndex: req.TransactionIndex,
		}
		proxy = &ProxyETHGetTransactionByBlockHashAndIndex{Kaon: p.Kaon}
	)
	return proxy.request(ctx, getBlockByHashReq)
}
