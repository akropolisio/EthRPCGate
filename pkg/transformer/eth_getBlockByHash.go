package transformer

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/kaonone/eth-rpc-gate/pkg/blockhash"
	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/kaonone/eth-rpc-gate/pkg/kaon"
	"github.com/kaonone/eth-rpc-gate/pkg/utils"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

var ErrBlockHashNotConfigured = errors.New("BlockHash database not configured")
var ErrBlockHashUnknown = errors.New("BlockHash unknown")

// ProxyETHGetBlockByHash implements ETHProxy
type ProxyETHGetBlockByHash struct {
	*kaon.Kaon
}

func (p *ProxyETHGetBlockByHash) Method() string {
	return "eth_getBlockByHash"
}

func (p *ProxyETHGetBlockByHash) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, *eth.JSONRPCError) {
	req := new(eth.GetBlockByHashRequest)
	if err := unmarshalRequest(rawreq.Params, req); err != nil {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError(err.Error())
	}

	blockHash := c.Get("blockHash")
	bh, ok := blockHash.(*blockhash.BlockHash)
	if !ok {
		// ok, do nothing
	}

	req.BlockHash = utils.RemoveHexPrefix(req.BlockHash)

	resultChan := make(chan *eth.GetBlockByHashResponse, 2)
	errorChan := make(chan *eth.JSONRPCError, 1)
	kaonBlockErrorChan := make(chan error, 1)
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	go func() {
		result, err := p.request(ctx, req)
		if err != nil {
			errorChan <- err
			return
		}

		resultChan <- result
	}()

	if bh == nil {
		kaonBlockErrorChan <- ErrBlockHashNotConfigured
	} else {
		go func() {
			kaonBlockHash, err := bh.GetKaonBlockHashContext(ctx, req.BlockHash)
			if err != nil {
				kaonBlockErrorChan <- err
				return
			}

			if kaonBlockHash == nil {
				kaonBlockErrorChan <- ErrBlockHashUnknown
				return
			}

			request := &eth.GetBlockByHashRequest{
				BlockHash:       utils.RemoveHexPrefix(*kaonBlockHash),
				FullTransaction: req.FullTransaction,
			}

			result, jsonErr := p.request(ctx, request)
			if jsonErr != nil {
				kaonBlockErrorChan <- jsonErr.Error()
				return
			}

			resultChan <- result
		}()
	}

	select {
	case result := <-resultChan:
		// TODO: Stop remaining request
		if result == nil {
			select {
			case result := <-resultChan:
				// backup succeeded
				return result, nil
			case <-kaonBlockErrorChan:
				// backup failed, return original request
				return nil, nil
			}
		} else {
			return result, nil
		}
	case err := <-errorChan:
		// the main request failed, wait for backup to finish
		select {
		case result := <-resultChan:
			// backup succeeded
			return result, nil
		case <-kaonBlockErrorChan:
			// backup failed, return original request
			return nil, err
		}
	}
}

func (p *ProxyETHGetBlockByHash) request(ctx context.Context, req *eth.GetBlockByHashRequest) (*eth.GetBlockByHashResponse, *eth.JSONRPCError) {
	blockHeader, err := p.GetBlockHeader(ctx, req.BlockHash)
	if err != nil {
		if err == kaon.ErrInvalidAddress {
			// unknown block hash should return {result: null}
			p.GetDebugLogger().Log("msg", "Unknown block hash", "blockHash", req.BlockHash)
			return nil, nil
		}
		p.GetDebugLogger().Log("msg", "couldn't get block header", "blockHash", req.BlockHash)
		return nil, eth.NewCallbackError("couldn't get block header")
	}
	block, err := p.GetBlock(ctx, req.BlockHash, req.FullTransaction)
	if err != nil {
		p.GetDebugLogger().Log("msg", "couldn't get block", "blockHash", req.BlockHash)
		return nil, eth.NewCallbackError("couldn't get block")
	}
	nonce := hexutil.EncodeUint64(uint64(block.Nonce))
	// left pad nonce with 0 to length 16, eg: 0x0000000000000042
	nonce = utils.AddHexPrefix(fmt.Sprintf("%016v", utils.RemoveHexPrefix(nonce)))
	resp := &eth.GetBlockByHashResponse{
		// TODO: researching
		// * If ETH block has pending status, then the following values must be null
		// ? Is it possible case for Kaon
		Hash:   utils.AddHexPrefix(req.BlockHash),
		Number: hexutil.EncodeUint64(uint64(block.Height)),

		// TODO: researching
		// ! Not found
		// ! Has incorrect value for compatability
		ReceiptsRoot: utils.AddHexPrefix(block.Merkleroot),

		// TODO: researching
		// ! Not found
		// ! Probably, may be calculated by huge amount of requests
		TotalDifficulty: hexutil.EncodeUint64(uint64(blockHeader.Difficulty)),

		// TODO: researching
		// ! Not found
		// ? Expect it always to be null
		Uncles: []string{},

		// TODO: check value correctness
		Sha3Uncles: eth.DefaultSha3Uncles,

		// TODO: backlog
		// ! Not found
		// - Temporary expect this value to be always zero, as Etherium logs are usually zeros
		LogsBloom: eth.EmptyLogsBloom,

		// TODO: researching
		// ? What value to put
		// - Temporary set this value to be always zero
		// - the graph requires this to be of length 64
		ExtraData: "0x0000000000000000000000000000000000000000000000000000000000000000",

		Nonce:            nonce,
		Size:             hexutil.EncodeUint64(uint64(block.Size)),
		Difficulty:       hexutil.EncodeUint64(uint64(blockHeader.Difficulty)),
		StateRoot:        utils.AddHexPrefix(blockHeader.HashStateRoot),
		TransactionsRoot: utils.AddHexPrefix(block.Merkleroot),
		Transactions:     make([]interface{}, 0, len(block.Txs)),
		Timestamp:        hexutil.EncodeUint64(blockHeader.Time),
	}

	if blockHeader.IsGenesisBlock() {
		resp.ParentHash = "0x0000000000000000000000000000000000000000000000000000000000000000"
		resp.Miner = utils.AddHexPrefix(kaon.ZeroAddress)
	} else {
		resp.ParentHash = utils.AddHexPrefix(blockHeader.Previousblockhash)
		// ! Not found

		if blockHeader.Proposer == "" {
			resp.Miner = utils.AddHexPrefix(kaon.ZeroAddress)
		} else {
			resp.Miner = utils.AddHexPrefix(blockHeader.Proposer)
		}
	}

	resp.GasLimit = utils.AddHexPrefix(kaon.DefaultBlockGasLimit)
	resp.GasUsed = utils.AddHexPrefix(blockHeader.GasUsed.String())

	var cumulativeGas *big.Int = big.NewInt(0)

	if req.FullTransaction {
		for i, txHash := range block.Txs {
			switch txData := txHash.(type) {
			case string:
				// Fallback to legacy trx processingg
				tx, err := getTransactionByHashKAON(ctx, p.Kaon, txData)
				if err != nil {
					p.GetDebugLogger().Log("msg", "Couldn't get transaction by hash", "hash", txData, "err", err)
					return nil, eth.NewCallbackError("couldn't get transaction by hash")
				}
				if tx == nil {
					if block.Height == 0 {
						// Error Invalid address - The genesis block coinbase is not considered an ordinary transaction and cannot be retrieved
						// the coinbase we can ignore since its not a real transaction, mainnet ethereum also doesn't return any data about the genesis coinbase
						p.GetDebugLogger().Log("msg", "Failed to get transaction in genesis block, probably the coinbase which we can't get")
					} else {
						p.GetDebugLogger().Log("msg", "Failed to get transaction by hash included in a block", "hash", txData)
						if !p.GetFlagBool(kaon.FLAG_IGNORE_UNKNOWN_TX) {
							return nil, eth.NewCallbackError("couldn't get transaction by hash included in a block")
						}
					}
				} else {
					resp.Transactions = append(resp.Transactions, *tx)
				}
			case *kaon.BlockTransactionDetails:
				var ethTx *eth.GetTransactionByHashResponse
				tx, err := formatTransactionInternal(ctx, p.Kaon, txData, block.Height, i, ethTx)
				if err != nil {
					p.GetDebugLogger().Log("msg", "Couldn't get transaction by hash", "hash", txData, "err", err)
					return nil, eth.NewCallbackError("couldn't get transaction by hash")
				}

				tx.GasUsed = tx.Gas // TODO
				tx.CumulativeGas = utils.AddHexPrefix(hexutil.EncodeBig(cumulativeGas))
				var newValue *big.Int
				newValue, _ = hexutil.DecodeBig(tx.Gas)
				cumulativeGas = new(big.Int).Add(cumulativeGas, newValue)

				resp.Transactions = append(resp.Transactions, *tx)

			default:

				// Create variables for potential responses
				var ethTxData kaon.BlockTransactionDetails

				// Try to unmarshal the response data into each potential response type
				err := json.Unmarshal([]byte(marshalToString(txData)), &ethTxData)
				if err != nil {
					p.GetDebugLogger().Log("msg", "Couldn't get transaction by hash", "hash", txData, "err", err)
					continue
					// return nil, eth.NewCallbackError("couldn't get transaction by hash")
				}

				var ethTx *eth.GetTransactionByHashResponse
				tx, serr := formatTransactionInternal(ctx, p.Kaon, &ethTxData, block.Height, i, ethTx)
				if serr != nil {
					p.GetDebugLogger().Log("msg", "Couldn't get transaction by hash", "hash", ethTxData.ID, "err", serr.Message())
					continue
					// return nil, eth.NewCallbackError("couldn't get transaction by hash")
				}

				tx.GasUsed = tx.Gas // TODO
				tx.CumulativeGas = utils.AddHexPrefix(hexutil.EncodeBig(cumulativeGas))
				var newValue *big.Int
				newValue, _ = hexutil.DecodeBig(tx.Gas)
				cumulativeGas = new(big.Int).Add(cumulativeGas, newValue)

				resp.Transactions = append(resp.Transactions, *tx)

			}
			resp.GasLimit = utils.AddHexPrefix(kaon.DefaultBlockGasLimit) // TODO: replace by dynamic
			resp.GasUsed = utils.AddHexPrefix(blockHeader.GasUsed.String())
		}
	} else {
		for _, txHash := range block.Txs {
			switch txData := txHash.(type) {
			case string:
				// NOTE:
				// 	Etherium RPC API doc says, that tx hashes must be of [32]byte,
				// 	however it doesn't seem to be correct, 'cause Etherium tx hash
				// 	has [64]byte just like Kaon tx hash has. In this case we do no
				// 	additional convertations now, while everything works fine
				resp.Transactions = append(resp.Transactions, utils.AddHexPrefix(txData))
			default:
				// Create variables for potential responses
				var ethTxData kaon.BlockTransactionDetails

				// Try to unmarshal the response data into each potential response type
				err := json.Unmarshal([]byte(marshalToString(txData)), &ethTxData)
				if err != nil {
					p.GetDebugLogger().Log("msg", "Couldn't get transaction by hash", "hash", txData, "err", err)
					continue
					// return nil, eth.NewCallbackError("couldn't get transaction by hash")
				}
				resp.Transactions = append(resp.Transactions, utils.AddHexPrefix(ethTxData.ID))
			}
		}
	}

	return resp, nil
}
