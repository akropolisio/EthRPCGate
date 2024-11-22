package transformer

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/labstack/echo"

	"github.com/kaonone/eth-rpc-gate/pkg/conversion"
	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/kaonone/eth-rpc-gate/pkg/kaon"
	"github.com/kaonone/eth-rpc-gate/pkg/utils"
)

// ProxyETHGetFilterChanges implements ETHProxy
type ProxyETHGetFilterChanges struct {
	*kaon.Kaon
	filter *eth.FilterSimulator
}

func (p *ProxyETHGetFilterChanges) Method() string {
	return "eth_getFilterChanges"
}

func (p *ProxyETHGetFilterChanges) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, *eth.JSONRPCError) {

	filter, err := processFilter(p, rawreq)
	if err != nil {
		return nil, err
	}

	switch filter.Type {
	case eth.NewFilterTy:
		return p.requestFilter(c.Request().Context(), filter)
	case eth.NewBlockFilterTy:
		return p.requestBlockFilter(c.Request().Context(), filter)
	case eth.NewPendingTransactionFilterTy:
		fallthrough
	default:
		return nil, eth.NewInvalidParamsError("Unknown filter type")
	}
}

func (p *ProxyETHGetFilterChanges) requestBlockFilter(ctx context.Context, filter *eth.Filter) (kaonresp eth.GetFilterChangesResponse, err *eth.JSONRPCError) {
	kaonresp = make(eth.GetFilterChangesResponse, 0)

	_lastBlockNumber, ok := filter.Data.Load("lastBlockNumber")
	if !ok {
		return kaonresp, eth.NewCallbackError("Could not get lastBlockNumber")
	}
	lastBlockNumber := _lastBlockNumber.(uint64)

	blockCountBigInt, blockErr := p.GetBlockCount(ctx)
	if blockErr != nil {
		return kaonresp, eth.NewCallbackError(blockErr.Error())
	}
	blockCount := blockCountBigInt.Uint64()

	differ := blockCount - lastBlockNumber

	hashes := make(eth.GetFilterChangesResponse, differ)
	for i := range hashes {
		blockNumber := new(big.Int).SetUint64(lastBlockNumber + uint64(i) + 1)

		resp, err := p.GetBlockHash(ctx, blockNumber)
		if err != nil {
			return kaonresp, eth.NewCallbackError(err.Error())
		}

		hashes[i] = utils.AddHexPrefix(string(resp))
	}

	kaonresp = hashes
	filter.Data.Store("lastBlockNumber", blockCount)
	return
}

func (p *ProxyETHGetFilterChanges) requestFilter(ctx context.Context, filter *eth.Filter) (kaonresp eth.GetFilterChangesResponse, err *eth.JSONRPCError) {
	kaonresp = make(eth.GetFilterChangesResponse, 0)

	_lastBlockNumber, ok := filter.Data.Load("lastBlockNumber")
	if !ok {
		return kaonresp, eth.NewCallbackError("Could not get lastBlockNumber")
	}
	lastBlockNumber := _lastBlockNumber.(uint64)

	blockCountBigInt, blockErr := p.GetBlockCount(ctx)
	if blockErr != nil {
		return kaonresp, eth.NewCallbackError(blockErr.Error())
	}
	blockCount := blockCountBigInt.Uint64()

	differ := blockCount - lastBlockNumber

	if differ == 0 {
		return eth.GetFilterChangesResponse{}, nil
	}

	searchLogsReq, err := p.toSearchLogsReq(filter, big.NewInt(int64(lastBlockNumber+1)), big.NewInt(int64(blockCount)))
	if err != nil {
		return nil, err
	}

	return p.doSearchLogs(ctx, searchLogsReq)
}

func (p *ProxyETHGetFilterChanges) doSearchLogs(ctx context.Context, req *kaon.SearchLogsRequest) (eth.GetFilterChangesResponse, *eth.JSONRPCError) {
	resp, err := conversion.SearchLogsAndFilterExtraTopics(ctx, p.Kaon, req)
	if err != nil {
		return nil, err
	}

	receiptToResult := func(receipt *kaon.TransactionReceipt) []interface{} {
		logs := conversion.ExtractETHLogsFromTransactionReceipt(receipt, receipt.Log)
		res := make([]interface{}, len(logs))
		for i := range res {
			res[i] = logs[i]
		}
		return res
	}
	results := make(eth.GetFilterChangesResponse, 0)
	for _, receipt := range resp {
		r := kaon.TransactionReceipt(receipt)
		results = append(results, receiptToResult(&r)...)
	}

	return results, nil
}

func (p *ProxyETHGetFilterChanges) toSearchLogsReq(filter *eth.Filter, from, to *big.Int) (*kaon.SearchLogsRequest, *eth.JSONRPCError) {
	ethreq := filter.Request.(*eth.NewFilterRequest)
	var err error
	var addresses []string
	if ethreq.Address != nil {
		if isBytesOfString(ethreq.Address) {
			var addr string
			if err = json.Unmarshal(ethreq.Address, &addr); err != nil {
				// TODO: Correct error code?
				return nil, eth.NewInvalidParamsError(err.Error())
			}
			addresses = append(addresses, addr)
		} else {
			if err = json.Unmarshal(ethreq.Address, &addresses); err != nil {
				// TODO: Correct error code?
				return nil, eth.NewInvalidParamsError(err.Error())
			}
		}
		for i := range addresses {
			addresses[i] = utils.RemoveHexPrefix(addresses[i])
		}
	}

	kaonreq := &kaon.SearchLogsRequest{
		Addresses: addresses,
		FromBlock: from,
		ToBlock:   to,
	}

	topics, ok := filter.Data.Load("topics")
	if ok {
		kaonreq.Topics = topics.([]kaon.SearchLogsTopic)
	}

	return kaonreq, nil
}
