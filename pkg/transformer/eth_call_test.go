package transformer

import (
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/kaonone/eth-rpc-gate/pkg/internal"
	"github.com/kaonone/eth-rpc-gate/pkg/kaon"
)

func TestEthCallRequest(t *testing.T) {
	//prepare request
	request := eth.CallRequest{
		From: "0x1e6f89d7399081b4f8f8aa1ae2805a5efff2f960",
		To:   "0x1e6f89d7399081b4f8f8aa1ae2805a5efff2f960",
	}
	requestRaw, err := json.Marshal(&request)
	if err != nil {
		t.Fatal(err)
	}
	requestParamsArray := []json.RawMessage{requestRaw}
	requestRPC, err := internal.PrepareEthRPCRequest(1, requestParamsArray)

	clientDoerMock := internal.NewDoerMappedMock()
	kaonClient, err := internal.CreateMockedClient(clientDoerMock)

	//preparing response
	callContractResponse := kaon.CallContractResponse{
		Address: "1e6f89d7399081b4f8f8aa1ae2805a5efff2f960",
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
			GasUsed:    *big.NewInt(216780),
			Excepted:   "None",
			NewAddress: "1e6f89d7399081b4f8f8aa1ae2805a5efff2f960",
			Output:     "0000000000000000000000000000000000000000000000000000000000000001",
		},
		TransactionReceipt: struct {
			StateRoot string        `json:"stateRoot"`
			GasUsed   big.Int       `json:"gasUsed"`
			Bloom     string        `json:"bloom"`
			Log       []interface{} `json:"log"`
		}{
			StateRoot: "d44fc5ad43bae52f01ff7eb4a7bba904ee52aea6c41f337aa29754e57c73fba6",
			GasUsed:   *big.NewInt(216780),
			Bloom:     "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
		},
	}
	err = clientDoerMock.AddResponseWithRequestID(1, kaon.MethodCallContract, callContractResponse)
	if err != nil {
		t.Fatal(err)
	}

	fromHexAddressResponse := kaon.FromHexAddressResponse("0x1e6f89d7399081b4f8f8aa1ae2805a5efff2f960")
	err = clientDoerMock.AddResponseWithRequestID(2, kaon.MethodFromHexAddress, fromHexAddressResponse)
	if err != nil {
		t.Fatal(err)
	}

	//preparing proxy & executing
	proxyEth := ProxyETHCall{kaonClient}
	if err != nil {
		t.Fatal(err)
	}

	got, jsonErr := proxyEth.Request(requestRPC, internal.NewEchoContext())
	if jsonErr != nil {
		t.Fatal(jsonErr)
	}

	want := eth.CallResponse("0x0000000000000000000000000000000000000000000000000000000000000001")

	internal.CheckTestResultEthRequestCall(request, &want, got, t, false)
}

func TestRetry(t *testing.T) {
	//prepare request
	request := eth.CallRequest{
		From: "0x1e6f89d7399081b4f8f8aa1ae2805a5efff2f960",
		To:   "0x1e6f89d7399081b4f8f8aa1ae2805a5efff2f960",
	}
	requestRaw, err := json.Marshal(&request)
	if err != nil {
		t.Fatal(err)
	}
	requestParamsArray := []json.RawMessage{requestRaw}
	requestRPC, err := internal.PrepareEthRPCRequest(1, requestParamsArray)

	clientDoerMock := internal.NewDoerMappedMock()
	kaonClient, err := internal.CreateMockedClient(clientDoerMock)

	//preparing response
	callContractResponse := kaon.CallContractResponse{
		Address: "1e6f89d7399081b4f8f8aa1ae2805a5efff2f960",
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
			GasUsed:    *big.NewInt(216780),
			Excepted:   "None",
			NewAddress: "1e6f89d7399081b4f8f8aa1ae2805a5efff2f960",
			Output:     "0000000000000000000000000000000000000000000000000000000000000001",
		},
		TransactionReceipt: struct {
			StateRoot string        `json:"stateRoot"`
			GasUsed   big.Int       `json:"gasUsed"`
			Bloom     string        `json:"bloom"`
			Log       []interface{} `json:"log"`
		}{
			StateRoot: "d44fc5ad43bae52f01ff7eb4a7bba904ee52aea6c41f337aa29754e57c73fba6",
			GasUsed:   *big.NewInt(216780),
			Bloom:     "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
		},
	}

	// return Kaon is busy response 4 times
	for i := 0; i < 4; i++ {
		clientDoerMock.AddRawResponse(kaon.MethodCallContract, []byte(kaon.ErrKaonWorkQueueDepth.Error()))
	}
	// on 5th request, return correct value
	clientDoerMock.AddResponseWithRequestID(1, kaon.MethodCallContract, callContractResponse)

	fromHexAddressResponse := kaon.FromHexAddressResponse("0x1e6f89d7399081b4f8f8aa1ae2805a5efff2f960")
	err = clientDoerMock.AddResponseWithRequestID(2, kaon.MethodFromHexAddress, fromHexAddressResponse)
	if err != nil {
		t.Fatal(err)
	}

	//preparing proxy & executing
	proxyEth := ProxyETHCall{kaonClient}
	if err != nil {
		t.Fatal(err)
	}

	before := time.Now()

	got, jsonErr := proxyEth.Request(requestRPC, internal.NewEchoContext())
	if jsonErr != nil {
		t.Fatal(jsonErr)
	}

	after := time.Now()

	want := eth.CallResponse("0x0000000000000000000000000000000000000000000000000000000000000001")

	internal.CheckTestResultEthRequestCall(request, &want, got, t, false)

	if after.Sub(before) < 2*time.Second {
		t.Errorf("Retrying requests was too quick: %v < 2s", after.Sub(before))
	}
}

func TestEthCallRequestOnUnknownContract(t *testing.T) {
	//prepare request
	request := eth.CallRequest{
		From: "0x1e6f89d7399081b4f8f8aa1ae2805a5efff2f960",
		To:   "0x1e6f89d7399081b4f8f8aa1ae2805a5efff2f960",
	}
	requestRaw, err := json.Marshal(&request)
	if err != nil {
		t.Fatal(err)
	}
	requestParamsArray := []json.RawMessage{requestRaw}
	requestRPC, err := internal.PrepareEthRPCRequest(1, requestParamsArray)

	clientDoerMock := internal.NewDoerMappedMock()
	kaonClient, err := internal.CreateMockedClient(clientDoerMock)

	fromHexAddressResponse := kaon.FromHexAddressResponse("0x1e6f89d7399081b4f8f8aa1ae2805a5efff2f960")
	err = clientDoerMock.AddResponse(kaon.MethodFromHexAddress, fromHexAddressResponse)
	if err != nil {
		t.Fatal(err)
	}

	//preparing error response
	unknownAddressResponse := kaon.GetErrorResponse(kaon.ErrInvalidAddress)
	err = clientDoerMock.AddError(kaon.MethodCallContract, unknownAddressResponse)
	if err != nil {
		t.Fatal(err)
	}

	//preparing proxy & executing
	proxyEth := ProxyETHCall{kaonClient}
	if err != nil {
		t.Fatal(err)
	}

	got, jsonErr := proxyEth.Request(requestRPC, internal.NewEchoContext())
	if jsonErr != nil {
		t.Fatal(jsonErr)
	}

	want := eth.CallResponse("0x")

	internal.CheckTestResultEthRequestCall(request, &want, got, t, false)
}
