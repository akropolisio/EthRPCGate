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

func TestAccountRequest(t *testing.T) {
	requestParams := []json.RawMessage{}
	request, err := internal.PrepareEthRPCRequest(1, requestParams)
	if err != nil {
		t.Fatal(err)
	}

	mockedClientDoer := internal.NewDoerMappedMock()
	kaonClient, err := internal.CreateMockedClient(mockedClientDoer)
	if err != nil {
		t.Fatal(err)
	}

	exampleAcc1, err := btcutil.DecodeWIF("5JK4Gu9nxCvsCxiq9Zf3KdmA9ACza6dUn5BRLVWAYEtQabdnJ89")
	if err != nil {
		t.Fatal(err)
	}
	exampleAcc2, err := btcutil.DecodeWIF("5JwvXtv6YCa17XNDHJ6CJaveg4mrpqFvcjdrh9FZWZEvGFpUxec")
	if err != nil {
		t.Fatal(err)
	}

	kaonClient.Accounts = append(kaonClient.Accounts, exampleAcc1, exampleAcc2)

	//preparing proxy & executing request
	proxyEth := ProxyETHAccounts{kaonClient}
	got, jsonErr := proxyEth.Request(request, internal.NewEchoContext())
	if jsonErr != nil {
		t.Fatal(jsonErr.Error())
	}

	want := eth.AccountsResponse{"0x6d358cf96533189dd5a602d0937fddf0888ad3ae", "0x7e22630f90e6db16283af2c6b04f688117a55db4"}

	internal.CheckTestResultEthRequestRPC(*request, want, got, t, false)
}

func TestAccountMethod(t *testing.T) {
	mockedClientDoer := internal.NewDoerMappedMock()
	kaonClient, err := internal.CreateMockedClient(mockedClientDoer)
	if err != nil {
		t.Fatal(err)
	}
	//preparing proxy & executing request
	proxyEth := ProxyETHAccounts{kaonClient}
	got := proxyEth.Method()

	want := string("eth_accounts")

	internal.CheckTestResultDefault(want, got, t, false)
}
func TestAccountToResponse(t *testing.T) {
	mockedClientDoer := internal.NewDoerMappedMock()
	kaonClient, err := internal.CreateMockedClient(mockedClientDoer)
	if err != nil {
		t.Fatal(err)
	}
	proxyEth := ProxyETHAccounts{kaonClient}
	callResponse := kaon.CallContractResponse{
		ExecutionResult: struct {
			GasUsed         big.Int `json:"gasUsed"`
			Excepted        string  `json:"excepted"`
			ExceptedMessage string  `json:"exceptedMessage"`
			NewAddress      string  `json:"newAddress"`
			Output          string  `json:"output"`
			CodeDeposit     int     `json:"codeDeposit"`
			GasRefunded     big.Int `json:"gasRefunded"`
			DepositSize     int     `json:"depositSize"`
			GasForDeposit   big.Int `json:"gasForDeposit"`
		}{
			Output: "0x0000000000000000000000000000000000000000000000000000000000000002",
		},
	}

	got := *proxyEth.ToResponse(&callResponse)
	want := eth.CallResponse("0x0000000000000000000000000000000000000000000000000000000000000002")

	internal.CheckTestResultDefault(want, got, t, false)
}
