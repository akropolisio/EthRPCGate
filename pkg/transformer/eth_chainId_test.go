package transformer

import (
	"encoding/json"
	"testing"

	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/kaonone/eth-rpc-gate/pkg/internal"
	"github.com/kaonone/eth-rpc-gate/pkg/kaon"
)

func TestChainIdMainnet(t *testing.T) {
	testChainIdsImpl(t, "main", "0x2ED3")
}

func TestChainIdTestnet(t *testing.T) {
	testChainIdsImpl(t, "test", "0x2ED5")
}

func TestChainIdRegtest(t *testing.T) {
	testChainIdsImpl(t, "regtest", "0x2ED4")
}

func TestChainIdUnknown(t *testing.T) {
	testChainIdsImpl(t, "???", "0x22ba")
}

func testChainIdsImpl(t *testing.T, chain string, expected string) {
	//preparing request
	requestParams := []json.RawMessage{}
	request, err := internal.PrepareEthRPCRequest(1, requestParams)
	if err != nil {
		t.Fatal(err)
	}

	mockedClientDoer := internal.NewDoerMappedMock()

	//preparing client response
	getBlockCountResponse := kaon.GetBlockChainInfoResponse{Chain: chain}
	err = mockedClientDoer.AddResponseWithRequestID(2, kaon.MethodGetBlockChainInfo, getBlockCountResponse)
	if err != nil {
		t.Fatal(err)
	}

	kaonClient, err := internal.CreateMockedClientForNetwork(mockedClientDoer, kaon.ChainAuto)
	if err != nil {
		t.Fatal(err)
	}

	//preparing proxy & executing request
	proxyEth := ProxyETHChainId{kaonClient}
	got, jsonErr := proxyEth.Request(request, internal.NewEchoContext())
	if jsonErr != nil {
		t.Fatal(jsonErr)
	}

	want := eth.ChainIdResponse(expected)

	internal.CheckTestResultEthRequestRPC(*request, want, got, t, false)
}
