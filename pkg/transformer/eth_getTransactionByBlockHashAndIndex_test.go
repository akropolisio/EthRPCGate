package transformer

import (
	"encoding/json"
	"testing"

	"github.com/kaonone/eth-rpc-gate/pkg/internal"
	"github.com/kaonone/eth-rpc-gate/pkg/kaon"
)

func initializeProxyETHGetTransactionByBlockHashAndIndex(kaonClient *kaon.Kaon) ETHProxy {
	return &ProxyETHGetTransactionByBlockHashAndIndex{kaonClient}
}

func TestGetTransactionByBlockHashAndIndex(t *testing.T) {
	testETHProxyRequest(
		t,
		initializeProxyETHGetTransactionByBlockHashAndIndex,
		[]json.RawMessage{[]byte(`"` + internal.GetTransactionByHashBlockHash + `"`), []byte(`"0x0"`)},
		internal.GetTransactionByHashResponseData,
	)
}
