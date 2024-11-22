package transformer

import (
	"context"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/kaonone/eth-rpc-gate/pkg/conversion"
	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/kaonone/eth-rpc-gate/pkg/kaon"
	"github.com/kaonone/eth-rpc-gate/pkg/utils"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

var STATUS_SUCCESS = "0x1"
var STATUS_FAILURE = "0x0"

// ProxyETHGetTransactionReceipt implements ETHProxy
type ProxyETHGetTransactionReceipt struct {
	*kaon.Kaon
}

func (p *ProxyETHGetTransactionReceipt) Method() string {
	return "eth_getTransactionReceipt"
}

func (p *ProxyETHGetTransactionReceipt) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, *eth.JSONRPCError) {
	var req eth.GetTransactionReceiptRequest
	if err := unmarshalRequest(rawreq.Params, &req); err != nil {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError(err.Error())
	}
	if req == "" {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError("empty transaction hash")
	}
	var (
		txHash  = utils.RemoveHexPrefix(string(req))
		kaonReq = kaon.GetTransactionReceiptRequest(txHash)
	)
	return p.request(c.Request().Context(), &kaonReq)
}

func (p *ProxyETHGetTransactionReceipt) request(ctx context.Context, req *kaon.GetTransactionReceiptRequest) (*eth.GetTransactionReceiptResponse, *eth.JSONRPCError) {
	kaonReceipt, err := p.Kaon.GetTransactionReceipt(ctx, string(*req))
	if err != nil {
		kaonHash, err := p.GetTransactionHashByEthHash(ctx, string(*req))

		var getRewardTransactionErr error
		var ethTx *eth.GetTransactionByHashResponse
		if err == nil {
			ethTx, _, getRewardTransactionErr = getRewardTransactionByHash(ctx, p.Kaon, string(*kaonHash))
		} else {
			ethTx, _, getRewardTransactionErr = getRewardTransactionByHash(ctx, p.Kaon, string(*req))
		}

		if getRewardTransactionErr != nil {
			errCause := errors.Cause(err)
			if errCause == kaon.EmptyResponseErr {
				return nil, nil
			}
			p.Kaon.GetDebugLogger().Log("msg", "Transaction does not exist", "txid", string(*req))
			return nil, eth.NewCallbackError(err.Error())
		}
		if ethTx == nil {
			// unconfirmed tx, return nil
			// https://github.com/openethereum/parity-ethereum/issues/3482
			return nil, nil
		}
		return &eth.GetTransactionReceiptResponse{
			TransactionHash:   ethTx.Hash,
			TransactionIndex:  ethTx.TransactionIndex,
			BlockHash:         ethTx.BlockHash,
			BlockNumber:       ethTx.BlockNumber,
			CumulativeGasUsed: ethTx.CumulativeGas,
			EffectiveGasPrice: ethTx.GasPrice,
			GasUsed:           ethTx.Gas,
			From:              ethTx.From,
			To:                ethTx.To,
			Logs:              []eth.Log{},
			LogsBloom:         eth.EmptyLogsBloom,
			Status:            STATUS_SUCCESS,
		}, nil
	}

	ethReceipt := &eth.GetTransactionReceiptResponse{
		TransactionHash:   utils.AddHexPrefix(kaonReceipt.TransactionHash),
		TransactionIndex:  hexutil.EncodeUint64(kaonReceipt.TransactionIndex),
		BlockHash:         utils.AddHexPrefix(kaonReceipt.BlockHash),
		BlockNumber:       hexutil.EncodeUint64(kaonReceipt.BlockNumber),
		ContractAddress:   utils.AddHexPrefixIfNotEmpty(kaonReceipt.ContractAddress),
		CumulativeGasUsed: hexutil.EncodeBig(&kaonReceipt.CumulativeGasUsed),
		EffectiveGasPrice: hexutil.EncodeBig(&kaonReceipt.EffectiveGasPrice),
		GasUsed:           hexutil.EncodeBig(&kaonReceipt.GasUsed),
		From:              utils.AddHexPrefixIfNotEmpty(kaonReceipt.From),
		To:                utils.AddHexPrefixIfNotEmpty(kaonReceipt.To),

		// TODO: researching
		// ! Temporary accept this value to be always zero, as it is at eth logs
		LogsBloom: eth.EmptyLogsBloom,
	}

	status := STATUS_FAILURE
	if kaonReceipt.Excepted == "None" {
		status = STATUS_SUCCESS
	} else {
		p.Kaon.GetDebugLogger().Log("transaction", ethReceipt.TransactionHash, "msg", "transaction excepted", "message", kaonReceipt.Excepted)
	}
	ethReceipt.Status = status

	r := kaon.TransactionReceipt(*kaonReceipt)
	ethReceipt.Logs = conversion.ExtractETHLogsFromTransactionReceipt(&r, r.Log)

	kaonTx, err := p.Kaon.GetRawTransaction(ctx, kaonReceipt.TransactionHash, false)
	if err != nil {
		p.GetDebugLogger().Log("msg", "couldn't get transaction", "err", err)
		return nil, eth.NewCallbackError("couldn't get transaction")
	}
	decodedRawKaonTx, err := p.Kaon.DecodeRawTransaction(ctx, kaonTx.Hex)
	if err != nil {
		p.GetDebugLogger().Log("msg", "couldn't decode raw transaction", "err", err)
		return nil, eth.NewCallbackError("couldn't decode raw transaction")
	}
	if decodedRawKaonTx.IsContractCreation() {
		ethReceipt.To = ""
	} else {
		ethReceipt.ContractAddress = ""
	}

	if ethReceipt.BlockHash == "0x0000000000000000000000000000000000000000000000000000000000000000" { // nullify pending txs
		ethReceipt.ContractAddress = ""
		ethReceipt.BlockNumber = ""
		ethReceipt.BlockHash = ""
	}

	// TODO: researching
	// - The following code reason is unknown (see original comment)
	// - Code temporary commented, until an error occures
	// ! Do not remove
	// // contractAddress : DATA, 20 Bytes - The contract address created, if the transaction was a contract creation, otherwise null.
	// if status != "0x1" {
	// 	// if failure, should return null for contractAddress, instead of the zero address.
	// 	ethTxReceipt.ContractAddress = ""
	// }

	return ethReceipt, nil
}
