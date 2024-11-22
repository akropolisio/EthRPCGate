package eth

import (
	"encoding/json"
)

var EmptyLogsBloom = "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
var DefaultSha3Uncles = "0x0000000000000000000000000000000000000000000000000000000000000000" // We don't need uncles in our system since we are POS

const (
	RPCVersion = "2.0"
)

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	ID      json.RawMessage `json:"id"`
	Params  json.RawMessage `json:"params"`
}

type JSONRPCResult struct {
	JSONRPC   string          `json:"jsonrpc"`
	RawResult json.RawMessage `json:"result,omitempty"`
	Error     *JSONRPCError   `json:"error,omitempty"`
	ID        json.RawMessage `json:"id,omitempty"`
}

func NewJSONRPCResult(id json.RawMessage, res interface{}) (*JSONRPCResult, error) {
	rawResult, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}

	return &JSONRPCResult{
		JSONRPC:   RPCVersion,
		ID:        id,
		RawResult: rawResult,
	}, nil
}

type JSONRPCNotification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params"`
}

func NewJSONRPCNotification(method string, params interface{}) (*JSONRPCNotification, error) {
	rawParams, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	return &JSONRPCNotification{
		JSONRPC: RPCVersion,
		Method:  method,
		Params:  rawParams,
	}, nil
}
