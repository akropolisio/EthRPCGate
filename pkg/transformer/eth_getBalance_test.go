package transformer

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/btcsuite/btcutil"
	"github.com/kaonone/eth-rpc-gate/pkg/internal"
	"github.com/kaonone/eth-rpc-gate/pkg/kaon"
)

func TestGetBalanceRequestAccount(t *testing.T) {
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
	fromHexAddressResponse := kaon.FromHexAddressResponse("5JK4Gu9nxCvsCxiq9Zf3KdmA9ACza6dUn5BRLVWAYEtQabdnJ89")
	err = mockedClientDoer.AddResponseWithRequestID(2, kaon.MethodFromHexAddress, fromHexAddressResponse)
	if err != nil {
		t.Fatal(err)
	}

	getAddressBalanceResponse := kaon.GetAddressBalanceResponse{Balance: *big.NewInt(100000000), Received: *big.NewInt(100000000), Immature: *big.NewInt(100000000)}
	err = mockedClientDoer.AddResponseWithRequestID(3, kaon.MethodGetAddressBalance, getAddressBalanceResponse)
	if err != nil {
		t.Fatal(err)
	}

	// TODO: Need getaccountinfo to return an account for unit test
	// if getaccountinfo returns nil
	// then address is contract, else account

	//preparing proxy & executing request
	proxyEth := ProxyETHGetBalance{kaonClient}
	got, jsonErr := proxyEth.Request(requestRPC, internal.NewEchoContext())
	if jsonErr != nil {
		t.Fatal(jsonErr)
	}

	want := string("0xde0b6b3a7640000") //1 Kaon represented in Wei

	internal.CheckTestResultEthRequestRPC(*requestRPC, want, got, t, false)
}

func TestGetBalanceRequestContract(t *testing.T) {
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
		// Code    string          `json:"code"`,
	}
	err = mockedClientDoer.AddResponseWithRequestID(3, kaon.MethodGetAccountInfo, getAccountInfoResponse)
	if err != nil {
		t.Fatal(err)
	}

	//preparing proxy & executing request
	proxyEth := ProxyETHGetBalance{kaonClient}
	got, jsonErr := proxyEth.Request(requestRPC, internal.NewEchoContext())
	if jsonErr != nil {
		t.Fatal(jsonErr)
	}

	want := string("0xbdaf8b")

	internal.CheckTestResultEthRequestRPC(*requestRPC, want, got, t, false)
}
