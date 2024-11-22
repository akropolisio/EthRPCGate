package eth

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/kaonone/eth-rpc-gate/pkg/utils"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

var DefaultGasAmountForKaon = big.NewInt(210000) // 2.1e5 (min gas limit)

// Kaon default gas value is 2 * CENT
var DefaultGasPriceInWei = big.NewInt(2000000000)

type (
	SendTransactionResponse string

	// SendTransactionRequest eth_sendTransaction
	SendTransactionRequest struct {
		From     string  `json:"from"`
		To       string  `json:"to"`
		Gas      *ETHInt `json:"gas"`      // optional
		GasPrice *ETHInt `json:"gasPrice"` // optional
		Value    string  `json:"value"`    // optional
		Data     string  `json:"data"`     // optional
		Nonce    string  `json:"nonce"`    // optional
	}
)

func (r *SendTransactionRequest) UnmarshalJSON(data []byte) error {
	type Request SendTransactionRequest

	var params []Request
	if err := json.Unmarshal(data, &params); err != nil {
		return err
	}

	*r = SendTransactionRequest(params[0])

	if r.Gas == nil {
		// ETH: (optional, default: 90000) Integer of the gas provided for the transaction execution. It will return unused gas.
		// Kaon: (numeric or string, optional) gasLimit, default: 6 * CENT, max: 60000 * CENT (CENT is gwei)
		r.Gas = &ETHInt{DefaultGasAmountForKaon}
	}

	if r.GasPrice == nil {
		// ETH: (optional, default: To-Be-Determined) Integer of the gasPrice used for each paid gas
		// Kaon: (numeric or string, optional) gasPrice KAON price per gas unit, default: 20 * CENT, 1 * CENT (CENT is gwei)
		r.GasPrice = &ETHInt{DefaultGasPriceInWei}
	}

	return nil
}

// see: https://ethereum.stackexchange.com/questions/8384/transfer-an-amount-between-two-ethereum-accounts-using-json-rpc
func (t *SendTransactionRequest) IsSendEther() bool {
	// data must be empty
	return t.Value != "" && t.To != "" && t.From != "" && t.Data == ""
}

func (t *SendTransactionRequest) IsCreateContract() bool {
	return t.To == "" && t.Data != ""
}

func (t *SendTransactionRequest) IsCallContract() bool {
	return t.To != "" && t.Data != ""
}

func (t *SendTransactionRequest) GasHex() string {
	if t.Gas == nil {
		return ""
	}

	return t.Gas.Hex()
}

func (t *SendTransactionRequest) GasPriceHex() string {
	if t.GasPrice == nil {
		return ""
	}
	return t.GasPrice.Hex()
}

// ========== eth_sendRawTransaction ============= //
type ResultType interface {
	IsResultType()
}

type (
	// Presents hexed string of a raw transaction
	SendRawTransactionRequest [1]string

	// Presents hexed string of a transaction hash
	SendRawTransactionResponse string
	/*
	   {
	   "txid": "d0fe0caa1b798c36da37e9118a06a7d151632d670b82d1c7dc3985577a71880f",
	   "sender": "qTKrsHUrzutdCVu3qi3iV1upzB2QpuRsRb",
	   "hash160": "6b22910b1e302cf74803ffd1691c2ecb858d3712",
	   "address": "c89a5d225f578d84a94741490c1b40889b4f7a00"
	   }
	*/
	CreateContractResponse struct {
		Txid    string `json:"txid"`
		Sender  string `json:"sender"`
		Hash160 string `json:"hash160"`
		Address string `json:"address"`
	}
	/*
		{
		  "address": "1e6f89d7399081b4f8f8aa1ae2805a5efff2f960",
		  "executionResult": {
		    "gasUsed": 21678,
		    "excepted": "None",
		    "newAddress": "1e6f89d7399081b4f8f8aa1ae2805a5efff2f960",
		    "output": "0000000000000000000000000000000000000000000000000000000000000001",
		    "codeDeposit": 0,
		    "gasRefunded": 0,
		    "depositSize": 0,
		    "gasForDeposit": 0
		  },
		  "transactionReceipt": {
		    "stateRoot": "d44fc5ad43bae52f01ff7eb4a7bba904ee52aea6c41f337aa29754e57c73fba6",
		    "gasUsed": 21678,
		    "bloom": "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
		    "log": []
		  }
		}
	*/
	CallContractResponse struct {
		Address         string `json:"address"`
		ExecutionResult struct {
			GasUsed         big.Int `json:"gasUsed"`
			Excepted        string  `json:"excepted"`
			ExceptedMessage string  `json:"exceptedMessage"`
			NewAddress      string  `json:"newAddress"`
			Output          string  `json:"output"`
			CodeDeposit     int     `json:"codeDeposit"`
			GasRefunded     big.Int `json:"gasRefunded"`
			DepositSize     int     `json:"depositSize"`
			GasForDeposit   big.Int `json:"gasForDeposit"`
		} `json:"executionResult"`
		TransactionReceipt struct {
			StateRoot string        `json:"stateRoot"`
			GasUsed   big.Int       `json:"gasUsed"`
			Bloom     string        `json:"bloom"`
			Log       []interface{} `json:"log"`
		} `json:"transactionReceipt"`
	}
	/*
		{
		  "txid": "6b7f70d8520e1ec87ba7f1ee559b491cc3028b77ae166e789be882b5d370eac9",
		  "sender": "qTKrsHUrzutdCVu3qi3iV1upzB2QpuRsRb",
		  "hash160": "6b22910b1e302cf74803ffd1691c2ecb858d3712"
		}
	*/
	SendToContractResponse struct {
		Txid    string `json:"txid"`
		Sender  string `json:"sender"`
		Hash160 string `json:"hash160"`
	}
)

type GeneralSendRawTransactionResponse struct {
	Result interface{} `json:"result"`
	// It may have in the Data
	// SendRawTransactionResponse / CreateContractResponse / CallContractResponse or SendToContractResponse
	Data interface{} `json:"-"`
}

func (r GeneralSendRawTransactionResponse) MarshalJSON() ([]byte, error) {
	switch result := r.Result.(type) {
	case string:
		return json.Marshal(result)
	case *CreateContractResponse:
		return json.Marshal(result)
	case *CallContractResponse:
		return json.Marshal(result)
	case *SendToContractResponse:
		return json.Marshal(result)
	default:
		type Alias GeneralSendRawTransactionResponse
		return json.Marshal(&struct {
			*Alias
		}{
			Alias: (*Alias)(&r),
		})
	}
}

// CallResponse
type CallResponse string

// CallRequest eth_call
type CallRequest struct {
	From     string  `json:"from"`     // optional|mandatory
	To       string  `json:"to"`       // optional|mandatory
	Gas      *ETHInt `json:"gas"`      // optional
	GasPrice *ETHInt `json:"gasPrice"` // optional
	Value    string  `json:"value"`    // optional
	Data     string  `json:"data"`     // optional
}

func (t *CallRequest) GasHex() string {
	if t.Gas == nil {
		return ""
	}
	return t.Gas.Hex()
}

func (t *CallRequest) GasPriceHex() string {
	if t.GasPrice == nil {
		return ""
	}
	return t.GasPrice.Hex()
}

func (t *CallRequest) UnmarshalJSON(data []byte) error {
	var err error
	var params []json.RawMessage
	if err = json.Unmarshal(data, &params); err != nil {
		return err
	}

	if len(params) == 0 {
		return errors.New("params must be set")
	}

	type txCallObject CallRequest
	var obj txCallObject
	if err = json.Unmarshal(params[0], &obj); err != nil {
		return err
	}
	if obj.To == "0x0000000000000000000000000000000000000000" {
		return nil
	}

	cr := CallRequest(obj)
	*t = cr
	return nil
}

type (
	PersonalUnlockAccountResponse bool
	BlockNumberResponse           string
	NetVersionResponse            string
	HashrateResponse              string
	MiningResponse                bool
)

// ========== eth_sign ============= //

type (
	SignRequest struct {
		Account string
		Message []byte
	}
	SignResponse string
)

func (t *SignRequest) UnmarshalJSON(data []byte) (err error) {
	var params []interface{}

	err = json.Unmarshal(data, &params)
	if err != nil {
		return errors.Wrap(err, "json unmarshalling")
	}

	if len(params) != 2 {
		return errors.New("expects 2 arguments")
	}

	if account, ok := params[0].(string); ok {
		t.Account = account
	} else {
		return errors.New("account address should be a hex string")
	}

	if data, ok := params[1].(string); ok {
		var msg []byte
		if !strings.HasPrefix(data, "0x") {
			msg = []byte(data)
		} else {
			msg, err = hex.DecodeString(utils.RemoveHexPrefix(data))
			if err != nil {
				return errors.Wrap(err, "invalid data format")
			}
		}

		t.Message = msg
	} else {
		return errors.New("data should be a hex string")
	}

	return nil
}

// ========== GetLogs ============= //

type (
	GetLogsRequest struct {
		FromBlock json.RawMessage `json:"fromBlock"`
		ToBlock   json.RawMessage `json:"toBlock"`
		Address   json.RawMessage `json:"address"` // string or []string
		Topics    []interface{}   `json:"topics"`
		Blockhash string          `json:"blockhash"`
	}
	GetLogsResponse []Log
)

func (r *GetLogsRequest) UnmarshalJSON(data []byte) error {
	type Request GetLogsRequest
	var params []Request
	if err := json.Unmarshal(data, &params); err != nil {
		return errors.Wrap(err, "json unmarshalling")
	}

	if len(params) == 0 {
		return errors.New("params must be set")
	}

	*r = GetLogsRequest(params[0])

	return nil
}

// ========== GetTransactionByHash ============= //
type (
	// Presents transaction hash value
	GetTransactionByHashRequest  string
	GetTransactionByHashResponse struct {
		// NOTE: must be null when its pending
		BlockHash string `json:"blockHash"`
		// NOTE: must be null when its pending
		BlockNumber string `json:"blockNumber"`

		// Hex representation of an integer - position in the block
		//
		// NOTE: must be null when its pending
		TransactionIndex string `json:"transactionIndex"`

		Hash string `json:"hash"`

		// The number of transactions made by the sender prior to this one
		// NOTE:
		// 	Unnecessary value, but keep it to be always 0x0, to be
		// 	graph-node compatible
		Nonce string `json:"nonce"`

		// Value transferred in Wei
		Value string `json:"value"`
		// The data send along with the transaction
		Input string `json:"input"`

		From string `json:"from"`
		// NOTE: must be null, if it's a contract creation transaction
		To string `json:"to"`

		// Gas provided by the sender
		Gas string `json:"gas"`
		// Gas price provided by the sender in Wei
		GasPrice string `json:"gasPrice"`

		// Cumulative gas used
		CumulativeGas string `json:"cumulativeGas"`
		// Gas used by the rrx TODO
		GasUsed string `json:"gasUsed"`

		// ECDSA recovery id
		V string `json:"v,omitempty"`
		// ECDSA signature r
		R string `json:"r,omitempty"`
		// ECDSA signature s
		S string `json:"s,omitempty"`
	}
)

func (r *GetTransactionByHashRequest) UnmarshalJSON(data []byte) error {
	var params []interface{}
	if err := json.Unmarshal(data, &params); err != nil {
		return err
	}
	if paramsNum := len(params); paramsNum != 1 {
		return fmt.Errorf("invalid parameters number - %d/1", paramsNum)
	}

	switch t := params[0].(type) {
	case string:
		*r = GetTransactionByHashRequest(t)
		return nil
	default:
		return fmt.Errorf("invalid parameter type %T, but %T is expected", t, "")
	}
}

// ========== GetTransactionByBlockHashAndIndex ========== //

type GetTransactionByBlockHashAndIndex struct {
	BlockHash        string
	TransactionIndex string
}

func (r *GetTransactionByBlockHashAndIndex) UnmarshalJSON(data []byte) error {
	var params []interface{}
	if err := json.Unmarshal(data, &params); err != nil {
		return errors.Wrap(err, "couldn't unmarhsal parameters")
	}
	paramsNum := len(params)
	if paramsNum == 0 {
		return errors.Errorf("missing value for required argument 0")
	} else if paramsNum == 1 {
		return errors.Errorf("missing value for required argument 1")
	} else if paramsNum > 2 {
		return errors.Errorf("too many arguments, want at most 2")
	}

	blockHash, ok := params[0].(string)
	if !ok {
		return newErrInvalidParameterType(1, params[0], "")
	}
	r.BlockHash = blockHash

	transactionIndex, ok := params[1].(string)
	if !ok {
		return newErrInvalidParameterType(2, params[1], "")
	}
	r.TransactionIndex = transactionIndex

	return nil
}

// ========== GetTransactionByBlockNumberAndIndex ========== //

type GetTransactionByBlockNumberAndIndex struct {
	BlockNumber      string
	TransactionIndex string
}

func (r *GetTransactionByBlockNumberAndIndex) UnmarshalJSON(data []byte) error {
	var params []interface{}
	if err := json.Unmarshal(data, &params); err != nil {
		return errors.Wrap(err, "couldn't unmarhsal parameters")
	}
	paramsNum := len(params)
	if paramsNum == 0 {
		return errors.Errorf("missing value for required argument 0")
	} else if paramsNum == 1 {
		return errors.Errorf("missing value for required argument 1")
	} else if paramsNum > 2 {
		return errors.Errorf("too many arguments, want at most 2")
	}

	blockNumber, ok := params[0].(string)
	if !ok {
		return newErrInvalidParameterType(1, params[0], "")
	}
	r.BlockNumber = blockNumber

	transactionIndex, ok := params[1].(string)
	if !ok {
		return newErrInvalidParameterType(2, params[1], "")
	}
	r.TransactionIndex = transactionIndex

	return nil
}

// ========== GetTransactionReceipt ============= //

type (
	// Presents transaction hash of a contract
	GetTransactionReceiptRequest  string
	GetTransactionReceiptResponse struct {
		TransactionHash  string `json:"transactionHash"`  // DATA, 32 Bytes - hash of the transaction.
		TransactionIndex string `json:"transactionIndex"` // QUANTITY - integer of the transactions index position in the block.
		BlockHash        string `json:"blockHash"`        // DATA, 32 Bytes - hash of the block where this transaction was in.
		BlockNumber      string `json:"blockNumber"`      // QUANTITY - block number where this transaction was in.
		From             string `json:"from,omitempty"`   // DATA, 20 Bytes - address of the sender.
		// NOTE: must be null if it's a contract creation transaction
		To                string `json:"to,omitempty"` // DATA, 20 Bytes - address of the receiver. null when its a contract creation transaction.
		EffectiveGasPrice string `json:"effectiveGasPrice"`
		CumulativeGasUsed string `json:"cumulativeGasUsed"` // QUANTITY - The total amount of gas used when this transaction was executed in the block.
		GasUsed           string `json:"gasUsed"`           // QUANTITY - The amount of gas used by this specific transaction alone.
		// NOTE: must be null if it's NOT a contract creation transaction
		ContractAddress string `json:"contractAddress,omitempty"` // DATA, 20 Bytes - The contract address created, if the transaction was a contract creation, otherwise null.
		Logs            []Log  `json:"logs"`                      // Array - Array of log objects, which this transaction generated.
		LogsBloom       string `json:"logsBloom"`                 // DATA, 256 Bytes - Bloom filter for light clients to quickly retrieve related logs.
		Status          string `json:"status"`                    // QUANTITY either 1 (success) or 0 (failure)

		// TODO: researching
		// ? Do we need this value
		// Root              string `json:"root,omitempty"`
	}

	Log struct {
		Removed          string   `json:"removed,omitempty"` // TAG - true when the log was removed, due to a chain reorganization. false if its a valid log.
		LogIndex         string   `json:"logIndex"`          // QUANTITY - integer of the log index position in the block. null when its pending log.
		TransactionIndex string   `json:"transactionIndex"`  // QUANTITY - integer of the transactions index position log was created from. null when its pending log.
		TransactionHash  string   `json:"transactionHash"`   // DATA, 32 Bytes - hash of the transactions this log was created from. null when its pending log.
		BlockHash        string   `json:"blockHash"`         // DATA, 32 Bytes - hash of the block where this log was in. null when its pending. null when its pending log.
		BlockNumber      string   `json:"blockNumber"`       // QUANTITY - the block number where this log was in. null when its pending. null when its pending log.
		Address          string   `json:"address"`           // DATA, 20 Bytes - address from which this log originated.
		Data             string   `json:"data"`              // DATA - contains one or more 32 Bytes non-indexed arguments of the log.
		Topics           []string `json:"topics"`            // Array of DATA - Array of 0 to 4 32 Bytes DATA of indexed log arguments.
		Type             string   `json:"type,omitempty"`
	}
)

func (r *GetTransactionReceiptRequest) UnmarshalJSON(data []byte) error {
	var params []string
	err := json.Unmarshal(data, &params)
	if err != nil {
		return errors.Wrap(err, "json unmarshalling")
	}

	if len(params) == 0 {
		return errors.New("params must be set")
	}

	*r = GetTransactionReceiptRequest(params[0])
	return nil
}

// ========== debug_traceBlockByNumber ============= //
// Example of the request:  [
//       "0x72ABA",
//       {
//         "timeout": "15s",
//         "tracer": "callTracer"
//       }
// ]

type TraceBlockByNumberRequest struct {
	Address string
	Config  TraceConfig
}

type TraceBlockByNumberConfig struct {
	Timeout string `json:"timeout"`
	Tracer  string `json:"tracer"`
}

// Example of the response:  [
// {
//   "txHash": "0xb42edc1d46932ef34be0ba49402dc94e3d2319c066f02945f6828cd344fcfa7b",
//   "result": {
// 	 "calls": [
// 	  {
// 		"calls": [
// 		  {
// 			"from": "0x00000000000000000000000000000000000001ff",
// 			"gas": "0x595a",
// 			"gasUsed": "0x16",
// 			"input": "0x0100",
// 			"output": "0x0100",
// 			"to": "0x00000000000000000000000000000000000000ff",
// 			"type": "CALL",
// 			"value": "0x0"
// 		  }
// 		],
// 		"from": "0x00000000000000000000000000000000000002ff",
// 		"gas": "0x6525",
// 		"gasUsed": "0xa7b",
// 		"input": "0x000100",
// 		"output": "0x0100",
// 		"to": "0x00000000000000000000000000000000000001ff",
// 		"type": "CALL",
// 		"value": "0x0"
// 	  },
// 	  {
// 		"calls": [
// 		  {
// 			"from": "0x00000000000000000000000000000000000001ff",
// 			"gas": "0x584a",
// 			"gasUsed": "0x10",
// 			"input": "0x",
// 			"to": "0x00000000000000000000000000000000000000ff",
// 			"type": "CALL",
// 			"value": "0x0"
// 		  }
// 		],
// 		"from": "0x00000000000000000000000000000000000002ff",
// 		"gas": "0x5a4c",
// 		"gasUsed": "0xb1",
// 		"input": "0x00",
// 		"to": "0x00000000000000000000000000000000000001ff",
// 		"type": "CALL",
// 		"value": "0x0"
// 	  }
// 	],
// 	"from": "0x71562b71999873db5b286df957af199ec94617f7",
// 	"gas": "0xc350",
// 	"gasUsed": "0x684c",
// 	"input": "0x01000100",
// 	"to": "0x00000000000000000000000000000000000002ff",
// 	"type": "CALL",
// 	"value": "0x0"
//   }
// }]

type TraceBlockByNumberResponse []TxTraceResult

// TxTraceResult is the result of a single transaction trace.
type TxTraceResult struct {
	TxHash common.Hash `json:"txHash"`           // transaction hash
	Result CallTrace   `json:"result,omitempty"` // Trace results produced by the tracer
	Error  string      `json:"error,omitempty"`  // Trace failure produced by the tracer
}

func (r *TraceBlockByNumberRequest) UnmarshalJSON(data []byte) error {
	var params []interface{}
	err := json.Unmarshal(data, &params)
	if err != nil {
		return errors.Wrap(err, "json unmarshalling")
	}

	if len(params) != 2 {
		return errors.New("params must be set and contain exactly two elements")
	}

	address, ok := params[0].(string)
	if !ok {
		return errors.New("first parameter should be a string")
	}

	configMap, ok := params[1].(map[string]interface{})
	if !ok {
		return errors.New("second parameter should be an object")
	}

	var config TraceConfig
	configBytes, err := json.Marshal(configMap)
	if err != nil {
		return errors.Wrap(err, "json marshalling")
	}

	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		return errors.Wrap(err, "json unmarshalling config")
	}

	*r = TraceBlockByNumberRequest{
		Address: address,
		Config:  config,
	}

	return nil
}

// ========== trace_block ============= //
// TraceBlockRequest represents the request for the trace_block method (numbers of the block, e.g. "0x2")
type TraceBlockRequest string

func (r *TraceBlockRequest) UnmarshalJSON(data []byte) error {
	var params []string
	err := json.Unmarshal(data, &params)
	if err != nil {
		return errors.Wrap(err, "json unmarshalling")
	}

	if len(params) == 0 {
		return errors.New("params must be set")
	}

	*r = TraceBlockRequest(params[0])
	return nil
}

/*
Example of the response: [
    {
      "action": {
        "callType": "call",
        "from": "0xaa7b131dc60b80d3cf5e59b5a21a666aa039c951",
        "gas": "0x0",
        "input": "0x",
        "to": "0xd40aba8166a212d6892125f079c33e6f5ca19814",
        "value": "0x4768d7effc3fbe"
      },
      "blockHash": "0x7eb25504e4c202cf3d62fd585d3e238f592c780cca82dacb2ed3cb5b38883add",
      "blockNumber": 3068185,
      "result": {
        "gasUsed": "0x0",
        "output": "0x"
      },
      "subtraces": 0,
      "traceAddress": [],
      "transactionHash": "0x07da28d752aba3b9dd7060005e554719c6205c8a3aea358599fc9b245c52f1f6",
      "transactionPosition": 0,
      "type": "call"
    },
    {
      "action": {
        "callType": "call",
        "from": "0x4f11ba23bb526c0486d83c6a8f18f632f3fc172a",
        "gas": "0x0",
        "input": "0x",
        "to": "0x7ed1e469fcb3ee19c0366d829e291451be638e59",
        "value": "0x446cde325fbfbe"
      },
      "blockHash": "0x7eb25504e4c202cf3d62fd585d3e238f592c780cca82dacb2ed3cb5b38883add",
      "blockNumber": 3068185,
      "result": {
        "gasUsed": "0x0",
        "output": "0x"
      },
      "subtraces": 0,
      "traceAddress": [],
      "transactionHash": "0x056f11efb5da4ff7cf8523cfcef08393e5dd2ff3ab3223e4324426d285d7ae92",
      "transactionPosition": 1,
      "type": "call"
    },
	{
      "action": {
        "author": "0x7ed1e469fcb3ee19c0366d829e291451be638e59",
        "rewardType": "block",
        "value": "0x446cde325fbfbe"
      },
      "blockHash": "0x7eb25504e4c202cf3d62fd585d3e238f592c780cca82dacb2ed3cb5b38883add",
      "blockNumber": 3068185,
      "result": null,
      "subtraces": 0,
      "traceAddress": [],
      "transactionHash": "0x056f11efb5da4ff7cf8523cfcef08393e5dd2ff3ab3223e4324426d285d7ae92",
      "transactionPosition": 1,
      "type": "reward"
    },
    {
     ...
    }
  ]

  Note: action may contain specific block reward info thus it'll jave Author and RewardType fields filled (and others may be empty in that case)
*/

type BlockTraceAction struct {
	From       string `json:"from,omitempty"`
	Type       string `json:"callType,omitempty"`
	Gas        string `json:"gas,omitempty"`
	Input      string `json:"input,omitempty"`
	To         string `json:"to,omitempty"`
	Value      string `json:"value"`
	Author     string `json:"author,omitempty"`
	RewardType string `json:"rewardType,omitempty"`
}

type BlockTraceResult struct {
	GasUsed string `json:"gasUsed"`
	Output  string `json:"output"`
}

type BlockTrace struct {
	Action              BlockTraceAction  `json:"action"`
	BlockHash           string            `json:"blockHash"`
	BlockNumber         int               `json:"blockNumber"`
	Result              *BlockTraceResult `json:"result"`
	Subtraces           int               `json:"subtraces"`
	TraceAddress        []int             `json:"traceAddress"`
	TransactionHash     string            `json:"transactionHash"`
	TransactionPosition int               `json:"transactionPosition"`
	Type                string            `json:"type"`
}

type TraceBlockResponse []BlockTrace

// ========== debug_traceTransaction ============= //
// DebugTraceTransactionRequest represents the request for the debug_traceTransaction method.
type DebugTraceTransactionRequest struct {
	Hash   string      `json:"hash"`   // The hash of the transaction
	Config TraceConfig `json:"config"` // The configuration for tracing (optional)
}

// TraceConfig represents the tracing options for debug_traceTransaction.
type TraceConfig struct {
	Tracer string `json:"tracer,omitempty"` // The tracer to use (e.g., "callTracer")
	// Add other tracing options here as needed, such as:
	// Reexec uint64 `json:"reexec,omitempty"`
	// DisableStack bool `json:"disableStack,omitempty"`
	// DisableStorage bool `json:"disableStorage,omitempty"`
	// DisableMemory bool `json:"disableMemory,omitempty"`
	// FullStorage bool `json:"fullStorage,omitempty"`
	// Add other relevant fields based on the tracing options you want to support
}

func (r *DebugTraceTransactionRequest) UnmarshalJSON(data []byte) error {
	var params []json.RawMessage
	err := json.Unmarshal(data, &params)
	if err != nil {
		return errors.Wrap(err, "json unmarshalling")
	}

	if len(params) == 0 {
		return errors.New("params must be set")
	}

	var txHash string
	err = json.Unmarshal(params[0], &txHash)
	if err != nil {
		return errors.Wrap(err, "unmarshalling transaction hash")
	}
	r.Hash = txHash

	// If a second parameter is present, unmarshal it into TraceConfig
	if len(params) > 1 {
		var config TraceConfig
		err = json.Unmarshal(params[1], &config)
		if err != nil {
			return errors.Wrap(err, "unmarshalling trace config")
		}
		r.Config = config
	}

	return nil
}

type CallTrace struct {
	Type    string      `json:"type,omitempty"`
	From    string      `json:"from,omitempty"`
	To      string      `json:"to,omitempty"`
	Value   string      `json:"value,omitempty"`
	Gas     string      `json:"gas,omitempty"`
	GasUsed string      `json:"gasUsed,omitempty"`
	Input   string      `json:"input,omitempty"`
	Output  string      `json:"output,omitempty"`
	Calls   []CallTrace `json:"calls,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ========== eth_accounts ============= //
type AccountsResponse []string

// ========== eth_getCode ============= //
type (
	GetCodeRequest struct {
		Address     string
		BlockNumber string
	}
	// the code from the given address.
	GetCodeResponse string
)

func (r *GetCodeRequest) UnmarshalJSON(data []byte) error {
	var params []string
	err := json.Unmarshal(data, &params)
	if err != nil {
		return errors.Wrap(err, "json unmarshalling")
	}

	if len(params) == 0 {
		return errors.New("params must be set")
	}

	r.Address = params[0]
	if len(params) > 1 {
		r.BlockNumber = params[1]
	}

	return nil
}

// ========== eth_newBlockFilter ============= //
// a filter id
type NewBlockFilterResponse string

// ========== eth_uninstallFilter ============= //
// the filter id
type UninstallFilterRequest string

// true if the filter was successfully uninstalled, otherwise false.
type UninstallFilterResponse bool

func (r *UninstallFilterRequest) UnmarshalJSON(data []byte) error {
	var params []string
	err := json.Unmarshal(data, &params)
	if err != nil {
		return errors.Wrap(err, "json unmarshalling")
	}

	if len(params) == 0 {
		return errors.New("params must be set")
	}

	*r = UninstallFilterRequest(params[0])

	return nil
}

// ========== eth_getFilterChanges ============= //
// the filter id
type GetFilterChangesRequest string

type GetFilterChangesResponse []interface{}

func (r *GetFilterChangesRequest) UnmarshalJSON(data []byte) error {
	var params []string
	err := json.Unmarshal(data, &params)
	if err != nil {
		return errors.Wrap(err, "json unmarshalling")
	}

	if len(params) == 0 {
		return errors.New("params must be set")
	}

	*r = GetFilterChangesRequest(params[0])

	return nil
}

// ========== eth_estimateGas ============= //

type EstimateGasResponse string

// ========== eth_gasPrice ============= //

type GasPriceResponse *ETHInt

// ========== eth_getBlockByNumber ============= //

type (
	GetBlockByNumberRequest struct {
		BlockNumber     json.RawMessage
		FullTransaction bool
	}

	/*
	 {
	    "number": "0x1b4",
	    "hash": "0xe670ec64341771606e55d6b4ca35a1a6b75ee3d5145a99d05921026d1527331",
	    "parentHash": "0x9646252be9520f6e71339a8df9c55e4d7619deeb018d2a3f2d21fc165dde5eb5",
	    "nonce": "0xe04d296d2460cfb8472af2c5fd05b5a214109c25688d3704aed5484f9a7792f2",
	    "sha3Uncles": "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
	    "logsBloom": "0xe670ec64341771606e55d6b4ca35a1a6b75ee3d5145a99d05921026d1527331",
	    "transactionsRoot": "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
	    "stateRoot": "0xd5855eb08b3387c0af375e9cdb6acfc05eb8f519e419b874b6ff2ffda7ed1dff",
	    "miner": "0x4e65fda2159562a496f9f3522f89122a3088497a",
	    "difficulty": "0x027f07",
	    "totalDifficulty":  "0x027f07",
	    "extraData": "0x0000000000000000000000000000000000000000000000000000000000000000",
	    "size":  "0x027f07",
	    "gasLimit": "0x9f759",
	    "gasUsed": "0x9f759",
	    "timestamp": "0x54e34e8e",
	    "transactions": [{}],
	    "uncles": ["0x1606e5...", "0xd5145a9..."]
	  }
	*/
	GetBlockByNumberResponse = GetBlockByHashResponse
)

// ========== eth_getBlockByHash ============= //

type (
	GetBlockByHashRequest struct {
		BlockHash       string
		FullTransaction bool
	}

	/*
	 {
	    "number": "0x1b4",
	    "hash": "0xe670ec64341771606e55d6b4ca35a1a6b75ee3d5145a99d05921026d1527331",
	    "parentHash": "0x9646252be9520f6e71339a8df9c55e4d7619deeb018d2a3f2d21fc165dde5eb5",
	    "nonce": "0xe04d296d2460cfb8472af2c5fd05b5a214109c25688d3704aed5484f9a7792f2",
	    "sha3Uncles": "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
	    "logsBloom": "0xe670ec64341771606e55d6b4ca35a1a6b75ee3d5145a99d05921026d1527331",
	    "transactionsRoot": "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
	    "stateRoot": "0xd5855eb08b3387c0af375e9cdb6acfc05eb8f519e419b874b6ff2ffda7ed1dff",
	    "miner": "0x4e65fda2159562a496f9f3522f89122a3088497a",
	    "difficulty": "0x027f07",
	    "totalDifficulty":  "0x027f07",
	    "extraData": "0x0000000000000000000000000000000000000000000000000000000000000000",
	    "size":  "0x027f07",
	    "gasLimit": "0x9f759",
	    "gasUsed": "0x9f759",
	    "timestamp": "0x54e34e8e",
	    "transactions": [{}],
	    "uncles": ["0x1606e5...", "0xd5145a9..."]
	  }
	*/
	GetBlockByHashResponse struct {
		Number     string `json:"number"`
		Hash       string `json:"hash"`
		ParentHash string `json:"parentHash"`
		Nonce      string `json:"nonce"`
		Size       string `json:"size"`
		Miner      string `json:"miner"`
		LogsBloom  string `json:"logsBloom"`
		Timestamp  string `json:"timestamp"`
		ExtraData  string `json:"extraData"`
		//Different type of response []string, []GetTransactionResponse
		Transactions     []interface{} `json:"transactions"`
		StateRoot        string        `json:"stateRoot"`
		TransactionsRoot string        `json:"transactionsRoot"`
		ReceiptsRoot     string        `json:"receiptsRoot"`
		Difficulty       string        `json:"difficulty"`
		// Represents a sum of all blocks difficulties until current block includingly
		TotalDifficulty string `json:"totalDifficulty"`
		GasLimit        string `json:"gasLimit"`
		GasUsed         string `json:"gasUsed"`
		// Represents sha3 hash value based on uncles slice
		Sha3Uncles string   `json:"sha3Uncles"`
		Uncles     []string `json:"uncles"`
	}
)

func (r *GetBlockByNumberRequest) UnmarshalJSON(data []byte) error {
	var params []interface{}
	if err := json.Unmarshal(data, &params); err != nil {
		return errors.Wrap(err, "couldn't unmarhsal data")
	}
	if paramsNum := len(params); paramsNum < 2 {
		return errors.Errorf("invalid parameters number - %d/2", paramsNum)
	}

	blockNumber, ok := params[0].(string)
	if !ok {
		return newErrInvalidParameterType(1, params[0], "")
	}
	// TODO: think of changing []byte type to string type
	r.BlockNumber = json.RawMessage(fmt.Sprintf("\"%s\"", blockNumber))

	fullTxWanted, ok := params[1].(bool)
	if !ok {
		return newErrInvalidParameterType(2, params[1], false)
	}
	r.FullTransaction = fullTxWanted

	return nil
}

func (r *GetBlockByHashRequest) UnmarshalJSON(data []byte) error {
	var params []interface{}
	if err := json.Unmarshal(data, &params); err != nil {
		return errors.Wrap(err, "couldn't unmarhsal parameters")
	}
	if paramsNum := len(params); paramsNum < 2 {
		return errors.Errorf("invalid parameters number - %d/2", paramsNum)
	}

	blockHash, ok := params[0].(string)
	if !ok {
		return newErrInvalidParameterType(1, params[0], "")
	}
	r.BlockHash = blockHash

	fullTxWanted, ok := params[1].(bool)
	if !ok {
		return newErrInvalidParameterType(2, params[1], false)
	}
	r.FullTransaction = fullTxWanted

	return nil
}

// TODO: think of moving it into a separate file
func newErrInvalidParameterType(idx int, gotType interface{}, wantedType interface{}) error {
	return errors.Errorf("invalid %d parameter of %T type, but %T type is expected", idx, gotType, wantedType)
}

// ========== eth_subscribe ============= //

type (
	EthLogSubscriptionParameter struct {
		Address interface{}   `json:"address"`
		Topics  []interface{} `json:"topics"`
	}

	EthSubscriptionRequest struct {
		Method string
		Params *EthLogSubscriptionParameter
	}

	EthSubscriptionResponse string

	/*
	   {
	     "jsonrpc": "2.0",
	     "method": "eth_subscription",
	     "params": {
	       "result": {
	         "difficulty": "0x15d9223a23aa",
	         "extraData": "0xd983010305844765746887676f312e342e328777696e646f7773",
	         "gasLimit": "0x47e7c4",
	         "gasUsed": "0x38658",
	         "logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
	         "miner": "0xf8b483dba2c3b7176a3da549ad41a48bb3121069",
	         "nonce": "0x084149998194cc5f",
	         "number": "0x1348c9",
	         "parentHash": "0x7736fab79e05dc611604d22470dadad26f56fe494421b5b333de816ce1f25701",
	         "receiptRoot": "0x2fab35823ad00c7bb388595cb46652fe7886e00660a01e867824d3dceb1c8d36",
	         "sha3Uncles": "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
	         "stateRoot": "0xb3346685172db67de536d8765c43c31009d0eb3bd9c501c9be3229203f15f378",
	         "timestamp": "0x56ffeff8",
	         "transactionsRoot": "0x0167ffa60e3ebc0b080cdb95f7c0087dd6c0e61413140e39d94d3468d7c9689f"
	       },
	       "subscription": "0x9ce59a13059e417087c02d3236a0b1cc"
	     }
	   }
	*/

	EthSubscriptionNewHeadResponse struct {
		Hash             string `json:"hash"` //! added for go-ethereum client support
		Number           string `json:"number"`
		Difficulty       string `json:"difficulty"`
		ExtraData        string `json:"extraData"`
		GasLimit         string `json:"gasLimit"`
		GasUsed          string `json:"gasUsed"`
		LogsBloom        string `json:"logsBloom"`
		Miner            string `json:"miner"`
		Nonce            string `json:"nonce"`
		ParentHash       string `json:"parentHash"`
		ReceiptRoot      string `json:"receiptsRoot"` //! replaced 'receipt' with 'receipts' for go-ethereum client support
		Sha3Uncles       string `json:"sha3Uncles"`
		StateRoot        string `json:"stateRoot"`
		Timestamp        string `json:"timestamp"`
		TransactionsRoot string `json:"transactionsRoot"`
		MixHash          string `json:"mixHash"` //! added for go-ethereum client support
		// BaseFeePerGas    string `json:"baseFeePerGas"` // added for go-ethereum client support
	}
)

var ErrInvalidAddresses = errors.New("Invalid addresses")

func (s *EthLogSubscriptionParameter) GetAddresses() ([]ETHAddress, error) {
	// can be a string or a string array
	if s.Address == nil {
		return []ETHAddress{}, nil
	} else if address, ok := s.Address.(string); ok {
		ethAddress, err := NewETHAddress(address)
		return []ETHAddress{ethAddress}, err
	} else if addresss, ok := s.Address.([]string); ok {
		ethAddresses := []ETHAddress{}
		for _, address := range addresss {
			ethAddress, err := NewETHAddress(address)
			if err != nil {
				return []ETHAddress{}, err
			}
			ethAddresses = append(ethAddresses, ethAddress)
		}
		return ethAddresses, nil
	} else if addresss, ok := s.Address.([]interface{}); ok {
		ethAddresses := []ETHAddress{}
		for _, address := range addresss {
			if addressString, ok := address.(string); ok {
				ethAddress, err := NewETHAddress(addressString)
				if err != nil {
					return []ETHAddress{}, err
				}
				ethAddresses = append(ethAddresses, ethAddress)
			} else {
				return []ETHAddress{}, ErrInvalidAddresses
			}
		}
		return ethAddresses, nil
	}

	return []ETHAddress{}, ErrInvalidAddresses
}

func (r *EthSubscriptionRequest) UnmarshalJSON(data []byte) error {
	var params []interface{}
	if err := json.Unmarshal(data, &params); err != nil {
		return errors.Wrap(err, "couldn't unmarhsal data")
	}

	method, ok := params[0].(string)
	if !ok {
		return newErrInvalidParameterType(1, params[0], "")
	}
	r.Method = method

	if len(params) >= 2 {
		param, err := json.Marshal(params[1])
		if err != nil {
			return err
		}
		var subscriptionParameter EthLogSubscriptionParameter
		err = json.Unmarshal(param, &subscriptionParameter)
		if err != nil {
			return err
		}
		r.Params = &subscriptionParameter
	}

	return nil
}

func (r EthSubscriptionRequest) MarshalJSON() ([]byte, error) {
	output := []interface{}{}
	output = append(output, r.Method)
	if r.Params != nil {
		output = append(output, r.Params)
	}

	return json.Marshal(output)
}

func NewEthSubscriptionNewHeadResponse(block *GetBlockByHashResponse) *EthSubscriptionNewHeadResponse {
	return &EthSubscriptionNewHeadResponse{
		Difficulty:       block.Difficulty,
		ExtraData:        block.ExtraData,
		GasLimit:         block.GasLimit,
		GasUsed:          block.GasUsed,
		LogsBloom:        block.LogsBloom,
		Miner:            block.Miner,
		Nonce:            block.Nonce,
		Number:           block.Number,
		ParentHash:       block.ParentHash,
		ReceiptRoot:      block.ReceiptsRoot,
		Sha3Uncles:       block.Sha3Uncles,
		Timestamp:        block.Timestamp,
		TransactionsRoot: block.TransactionsRoot,
		//! This is ok?
		MixHash:   "0x0000000000000000000000000000000000000000000000000000000000000000", // Added for go-ethereum client support
		StateRoot: block.StateRoot,                                                      // Added for go-ethereum client support
		Hash:      block.Hash,
	}
}

// ========== eth_unsubscribe =========== //

type (
	EthUnsubscribeRequest []string

	EthUnsubscribeResponse bool
)

// ========== eth_newFilter ============= //

type NewFilterRequest struct {
	FromBlock json.RawMessage `json:"fromBlock"`
	ToBlock   json.RawMessage `json:"toBlock"`
	Address   json.RawMessage `json:"address"`
	Topics    []interface{}   `json:"topics"`
}

func (r *NewFilterRequest) UnmarshalJSON(data []byte) error {
	var params []json.RawMessage
	if err := json.Unmarshal(data, &params); err != nil {
		return errors.Wrap(err, "json unmarshalling")
	}

	if len(params) == 0 {
		return errors.New("params must be set")
	}
	type Req NewFilterRequest
	var req Req

	if err := json.Unmarshal(params[0], &req); err != nil {
		return errors.Wrap(err, "json unmarshalling")
	}

	*r = NewFilterRequest(req)

	return nil
}

type NewFilterResponse string

// ========== eth_getBalance ============= //

type GetBalanceRequest struct {
	Address string
	Block   json.RawMessage
}

func (r *GetBalanceRequest) UnmarshalJSON(data []byte) error {
	tmp := []interface{}{&r.Address, &r.Block}

	return json.Unmarshal(data, &tmp)
}

type GetBalanceResponse string

// =======GetTransactionCount ============= //
type (
	GetTransactionCountRequest struct {
		Address string
		Tag     string
	}
)

// ========== getstorage ============= //
type (
	GetStorageRequest struct {
		Address     string
		Index       string
		BlockNumber string
	}
	GetStorageResponse string
)

func (r *GetStorageRequest) UnmarshalJSON(data []byte) error {
	tmp := []interface{}{&r.Address, &r.Index, &r.BlockNumber}
	return json.Unmarshal(data, &tmp)
}

// ======= eth_chainId ============= //
type ChainIdResponse string

// ======= eth_subscription ======== //
type EthSubscription struct {
	SubscriptionID string                `json:"subscription"`
	Result         interface{}           `json:"result"`
	Params         EthSubscriptionParams `json:"params"`            // Added for go-ethereum client support
	Version        string                `json:"jsonrpc,omitempty"` // Added for go-ethereum client support
	Method         string                `json:"method,omitempty"`  // Added for go-ethereum client support
}

type EthSubscriptionParams struct {
	Result         interface{} `json:"result"`
	SubscriptionID string      `json:"subscription"`
}

// ======= kaon_getUTXOs ============= //

type UTXOScriptType int

const (
	ALL_UTXO_TYPES UTXOScriptType = iota
	UNKNOWN_UTXO
	OPRETURN_UTXO
	IMMATURE
	// len(spk)==35 and (spk[0:1] + spk[34:35]).hex()=='21ac'
	P2PK // pay to pubkey
	// len(spk)==25 and (spk[0:3] + spk[23:25]).hex()=='76a91488ac'
	P2PKH // pay to public key hash
	// len(spk) == 23 and (spk[0:2] + spk[22:23]).hex() == 'a91487'
	P2SH // pay to script hash
	// len(spk) == 22 and (spk[0:2]).hex() == '0014'
	P2WPKH // pay to witness public key hash
	// len(spk) == 34 and (spk[0:2]).hex() == '0020'
	P2WSH // pay to witness script hash
	// is_p2sh() and len(ss) == 23 and (ss[0:3]).hex() == '160014'
	P2SHP2WPKH // P2SH Encapsulating Pay to Witness Public Key Hash
	// is_p2sh() and len(ss) == 35 and (ss[0:3]).hex() == '220020'
	P2SHP2WSH // P2SH Encapsulating Pay to Witness Script Hash
	P2MS      // pay to multisig
)

var AllUTXOScriptTypes = []UTXOScriptType{
	ALL_UTXO_TYPES,
	UNKNOWN_UTXO,
	OPRETURN_UTXO,
	IMMATURE,
	P2PK,
	P2PKH,
	P2SH,
	P2WPKH,
	P2WSH,
	P2SHP2WPKH,
	P2SHP2WSH,
	P2MS,
}

func (utxo UTXOScriptType) String() string {
	return []string{"all", "unknown", "opreturn", "immature", "P2PK", "P2PKH", "P2SH", "P2WPKH", "P2WSH", "P2SHP2WPKH", "P2SHP2WSH", "P2MS"}[utxo]
}

type (
	GetUTXOsRequest struct {
		Address      string
		MinSumAmount decimal.Decimal
		Types        []UTXOScriptType
	}

	KaonUTXO struct {
		Address   string `json:"address"`
		TXID      string `json:"txid"`
		Vout      uint   `json:"vout"`
		Amount    string `json:"amount"`
		Safe      bool   `json:"safe"`
		Spendable bool   `json:"spendable"`
		// Solvable bool `json:"solvable"`
		Confirmations int64  `json:"confirmations"`
		Height        uint64 `json:"height"`
		Type          string `json:"type"`
		ScriptPubKey  string `json:"scriptPubKey"`
		RedeemScript  string `json:"redeemScript,omitempty"`
	}

	GetUTXOsResponse []KaonUTXO
)

func (req *GetUTXOsRequest) UnmarshalJSON(params []byte) error {
	paramsBytesNum := len(params)
	if paramsBytesNum < 2 {
		return fmt.Errorf("bytes number < 2")
	}

	var parameters []string
	err := json.Unmarshal(params, &parameters)
	if err != nil {
		return err
	}

	validTypes := map[string]UTXOScriptType{}
	for _, scriptType := range AllUTXOScriptTypes {
		validTypes[strings.ToLower(scriptType.String())] = scriptType
	}

	if len(parameters) >= 1 {
		typesStartAt := 2
		req.Address = parameters[0]
		req.Types = []UTXOScriptType{}
		allTypes := false
		if len(parameters) >= 2 {
			req.MinSumAmount, err = decimal.NewFromString(parameters[1])
			if err != nil {
				parameter := strings.ToLower(parameters[1])
				if _, ok := validTypes[parameter]; ok {
					// send all
					typesStartAt = 1
					req.MinSumAmount = decimal.NewFromInt(0)
				} else {
					return err
				}
			}
		} else {
			allTypes = true
		}

		for i := typesStartAt; i < len(parameters); i++ {
			parameter := strings.ToLower(parameters[i])
			if typ, ok := validTypes[parameter]; ok {
				if typ == ALL_UTXO_TYPES {
					allTypes = true
				}
				req.Types = append(req.Types, typ)
			}
		}
		if allTypes {
			req.Types = []UTXOScriptType{ALL_UTXO_TYPES}
		}

		if len(parameters) > 3 && len(req.Types) == 0 {
			return fmt.Errorf("unknown script type requested")
		}

		return nil
	}

	return fmt.Errorf("Address required")
}

func (req GetUTXOsRequest) CheckHasValidValues() error {
	if !common.IsHexAddress(req.Address) {
		return errors.Errorf("invalid Ethereum address - %q", req.Address)
	}
	return nil
}

func (utxo KaonUTXO) IsP2PK() bool {
	// len(spk)==35 and (spk[0:1] + spk[34:35]).hex()=='21ac'
	return len(utxo.ScriptPubKey) == 70 && strings.ToLower((utxo.ScriptPubKey[0:2]+utxo.ScriptPubKey[68:70])) == "21ac"
}

func (utxo KaonUTXO) IsP2PKH() bool {
	// len(spk)==25 and (spk[0:3] + spk[23:25]).hex()=='76a91488ac'
	return len(utxo.ScriptPubKey) == 50 && strings.ToLower((utxo.ScriptPubKey[0:6]+utxo.ScriptPubKey[46:50])) == "76a91488ac"
}

func (utxo KaonUTXO) IsP2SH() bool {
	// 76a9143ade697fc8030489727bbb6af6a68f0a9eab2ec188ac
	// len(spk) == 23 and (spk[0:2] + spk[22:23]).hex() == 'a91487'
	return len(utxo.ScriptPubKey) == 46 && strings.ToLower((utxo.ScriptPubKey[0:4]+utxo.ScriptPubKey[44:46])) == "a91487"
}

func (utxo KaonUTXO) IsP2WPKH() bool {
	// len(spk) == 22 and (spk[0:2]).hex() == '0014'
	return len(utxo.ScriptPubKey) == 44 && strings.ToLower(utxo.ScriptPubKey[0:4]) == "0014"
}

func (utxo KaonUTXO) IsP2WSH() bool {
	// len(spk) == 34 and (spk[0:2]).hex() == '0020'
	return len(utxo.ScriptPubKey) == 68 && strings.ToLower(utxo.ScriptPubKey[0:4]) == "0020"
}

func (utxo KaonUTXO) IsP2SHP2WPKH() bool {
	// is_p2sh() and len(ss) == 23 and (ss[0:3]).hex() == '160014'
	return utxo.IsP2SH() && len(utxo.ScriptPubKey) == 46 && strings.ToLower(utxo.ScriptPubKey[0:6]) == "160014"
}

func (utxo KaonUTXO) IsP2SHP2WSH() bool {
	// is_p2sh() and len(ss) == 35 and (ss[0:3]).hex() == '220020'
	return utxo.IsP2SH() && len(utxo.ScriptPubKey) == 70 && strings.ToLower(utxo.ScriptPubKey[0:6]) == "220020"
}

func (utxo KaonUTXO) GetType() UTXOScriptType {
	if utxo.IsP2PK() {
		return P2PK
	} else if utxo.IsP2PKH() {
		return P2PKH
	} else if utxo.IsP2SH() {
		if utxo.IsP2SHP2WPKH() {
			return P2SHP2WPKH
		} else if utxo.IsP2SHP2WSH() {
			return P2SHP2WSH
		} else {
			return P2SH
		}
	} else if utxo.IsP2WPKH() {
		return P2WPKH
	} else if utxo.IsP2WSH() {
		return P2WSH
	} else {
		return UNKNOWN_UTXO
	}
	// TODO: OP_RETURN
}

type (
	StringsArguments []string
	StringResponse   string
)

// ======= web3_sha3 ======= //
type Web3Sha3Request struct {
	Message string
}

func (r *Web3Sha3Request) UnmarshalJSON(data []byte) error {
	var params []interface{}
	if err := json.Unmarshal(data, &params); err != nil {
		return errors.Wrap(err, "couldn't unmarhsal parameters")
	}
	paramsNum := len(params)
	if paramsNum == 0 {
		return errors.Errorf("missing value for required argument 0")
	} else if paramsNum > 1 {
		return errors.Errorf("too many arguments, want at most 1")
	}

	message, ok := params[0].(string)
	if !ok {
		return newErrInvalidParameterType(1, params[0], "")
	}
	r.Message = message

	return nil
}

type NetPeerCountResponse string
