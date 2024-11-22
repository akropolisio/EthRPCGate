package transformer

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/kaonone/eth-rpc-gate/pkg/internal"
	"github.com/kaonone/eth-rpc-gate/pkg/kaon"
)

func TestEstimateGasRequest(t *testing.T) {
	request := eth.CallRequest{
		From: "0x1e6f89d7399081b4f8f8aa1ae2805a5efff2f960",
		To:   "0x1e6f89d7399081b4f8f8aa1ae2805a5efff2f960",
		Data: "0x0",
	}
	requestRaw, err := json.Marshal(&request)
	if err != nil {
		t.Fatal(err)
	}
	requestParamsArray := []json.RawMessage{requestRaw}
	requestRPC, err := internal.PrepareEthRPCRequest(1, requestParamsArray)

	if err != nil {
		t.Fatal(err)
	}

	mockedClientDoer := internal.NewDoerMappedMock()
	kaonClient, err := internal.CreateMockedClient(mockedClientDoer)
	if err != nil {
		t.Fatal(err)
	}

	//preparing responses
	fromHexAddressResponse := kaon.FromHexAddressResponse("0x1e6f89d7399081b4f8f8aa1ae2805a5efff2f960")
	err = mockedClientDoer.AddResponseWithRequestID(2, kaon.MethodFromHexAddress, fromHexAddressResponse)
	if err != nil {
		t.Fatal(err)
	}

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
			GasUsed:  *big.NewInt(216780),
			Excepted: "None",
		},
	}
	err = mockedClientDoer.AddResponseWithRequestID(1, kaon.MethodCallContract, callContractResponse)
	if err != nil {
		t.Fatal(err)
	}

	//preparing proxy & executing request
	proxyEth := ProxyETHCall{kaonClient}
	proxyEthEstimateGas := ProxyETHEstimateGas{&proxyEth}
	got, jsonErr := proxyEthEstimateGas.Request(requestRPC, internal.NewEchoContext())
	if jsonErr != nil {
		t.Fatal(jsonErr)
	}

	want := eth.EstimateGasResponse("0x659d")

	internal.CheckTestResultEthRequestCall(request, &want, got, t, false)
}

func TestEstimateGasRequestExecutionReverted(t *testing.T) {
	request := eth.CallRequest{
		From: "0x1e6f89d7399081b4f8f8aa1ae2805a5efff2f960",
		To:   "0x1e6f89d7399081b4f8f8aa1ae2805a5efff2f960",
		Data: "0x0",
	}
	requestRaw, err := json.Marshal(&request)
	if err != nil {
		t.Fatal(err)
	}
	requestParamsArray := []json.RawMessage{requestRaw}
	requestRPC, err := internal.PrepareEthRPCRequest(1, requestParamsArray)

	if err != nil {
		t.Fatal(err)
	}

	mockedClientDoer := internal.NewDoerMappedMock()
	kaonClient, err := internal.CreateMockedClient(mockedClientDoer)
	if err != nil {
		t.Fatal(err)
	}

	//preparing responses
	fromHexAddressResponse := kaon.FromHexAddressResponse("0x1e6f89d7399081b4f8f8aa1ae2805a5efff2f960")
	err = mockedClientDoer.AddResponseWithRequestID(2, kaon.MethodFromHexAddress, fromHexAddressResponse)
	if err != nil {
		t.Fatal(err)
	}

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
			GasUsed:  *big.NewInt(216780),
			Excepted: "OutOfGas",
		},
	}
	err = mockedClientDoer.AddResponseWithRequestID(1, kaon.MethodCallContract, callContractResponse)
	if err != nil {
		t.Fatal(err)
	}

	//preparing proxy & executing request
	proxyEth := ProxyETHCall{kaonClient}
	proxyEthEstimateGas := ProxyETHEstimateGas{&proxyEth}

	_, got := proxyEthEstimateGas.Request(requestRPC, internal.NewEchoContext())

	want := eth.NewCallbackError(ErrExecutionReverted.Error())

	internal.CheckTestResultDefault(want, got, t, false)
}

func TestEstimateGasNonVMRequest(t *testing.T) {
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

	if err != nil {
		t.Fatal(err)
	}

	mockedClientDoer := internal.NewDoerMappedMock()
	kaonClient, err := internal.CreateMockedClient(mockedClientDoer)
	if err != nil {
		t.Fatal(err)
	}

	//preparing responses
	fromHexAddressResponse := kaon.FromHexAddressResponse("0x1e6f89d7399081b4f8f8aa1ae2805a5efff2f960")
	err = mockedClientDoer.AddResponseWithRequestID(2, kaon.MethodFromHexAddress, fromHexAddressResponse)
	if err != nil {
		t.Fatal(err)
	}

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
			GasUsed:  *big.NewInt(216780),
			Excepted: "None",
		},
	}
	err = mockedClientDoer.AddResponseWithRequestID(1, kaon.MethodCallContract, callContractResponse)
	if err != nil {
		t.Fatal(err)
	}

	//preparing proxy & executing request
	proxyEth := ProxyETHCall{kaonClient}
	proxyEthEstimateGas := ProxyETHEstimateGas{&proxyEth}
	got, jsonErr := proxyEthEstimateGas.Request(requestRPC, internal.NewEchoContext())
	if jsonErr != nil {
		t.Fatal(jsonErr)
	}

	want := eth.EstimateGasResponse(NonContractVMGasLimit)

	internal.CheckTestResultEthRequestCall(request, &want, got, t, false)
}
