package transformer

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/btcsuite/btcutil"
	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/kaonone/eth-rpc-gate/pkg/internal"
	"github.com/kaonone/eth-rpc-gate/pkg/kaon"
)

func TestGetAccountInfoRequest(t *testing.T) {
	//prepare request
	requestParams := []json.RawMessage{[]byte(`"0x1e6f89d7399081b4f8f8aa1ae2805a5efff2f960"`), []byte(`"123"`)}
	requestRPC, err := internal.PrepareEthRPCRequest(1, requestParams)
	if err != nil {
		t.Fatal(err)
	}
	//prepare client
	mockedClientDoer := internal.NewDoerMappedMock()
	kaonClient, err := internal.CreateMockedClient(mockedClientDoer)
	if err != nil {
		t.Fatal(err)
	}

	//prepare account
	account, err := btcutil.DecodeWIF("5JK4Gu9nxCvsCxiq9Zf3KdmA9ACza6dUn5BRLVWAYEtQabdnJ89")
	if err != nil {
		t.Fatal(err)
	}
	kaonClient.Accounts = append(kaonClient.Accounts, account)

	//prepare responses
	getAccountInfoResponse := kaon.GetAccountInfoResponse{
		Address: "1e6f89d7399081b4f8f8aa1ae2805a5efff2f960",
		Balance: *big.NewInt(12431243),
		// Storage json.RawMessage `json:"storage"`,
		Code: "606060405236156100ad576000357c0100000000000000000...",
	}
	err = mockedClientDoer.AddResponseWithRequestID(3, kaon.MethodGetAccountInfo, getAccountInfoResponse)
	if err != nil {
		t.Fatal(err)
	}

	//preparing proxy & executing request
	proxyEth := ProxyETHGetCode{kaonClient}
	got, jsonErr := proxyEth.Request(requestRPC, internal.NewEchoContext())
	if jsonErr != nil {
		t.Fatal(jsonErr)
	}

	want := eth.GetCodeResponse("0x606060405236156100ad576000357c0100000000000000000...")

	internal.CheckTestResultEthRequestRPC(*requestRPC, want, got, t, false)
}

func TestGetCodeInvalidAddressRequest(t *testing.T) {
	//prepare request
	requestParams := []json.RawMessage{[]byte(`"0x0000000000000000000000000000000000000000"`), []byte(`"123"`)}
	requestRPC, err := internal.PrepareEthRPCRequest(1, requestParams)
	if err != nil {
		t.Fatal(err)
	}
	//prepare client
	mockedClientDoer := internal.NewDoerMappedMock()
	kaonClient, err := internal.CreateMockedClient(mockedClientDoer)
	if err != nil {
		t.Fatal(err)
	}

	//prepare responses
	getAccountInfoErrorResponse := kaon.GetErrorResponse(kaon.ErrInvalidAddress)
	if getAccountInfoErrorResponse == nil {
		panic("mocked error response is nil")
	}
	err = mockedClientDoer.AddError(kaon.MethodGetAccountInfo, getAccountInfoErrorResponse)
	if err != nil {
		t.Fatal(err)
	}

	//preparing proxy & executing request
	proxyEth := ProxyETHGetCode{kaonClient}
	got, jsonErr := proxyEth.Request(requestRPC, internal.NewEchoContext())
	if jsonErr != nil {
		t.Fatal(jsonErr)
	}

	want := eth.GetCodeResponse("0x")

	internal.CheckTestResultEthRequestRPC(*requestRPC, want, got, t, false)
}
