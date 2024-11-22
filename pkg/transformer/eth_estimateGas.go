package transformer

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/kaonone/eth-rpc-gate/pkg/kaon"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

// 2.1e5
var NonContractVMGasLimit = "0x33450"
var ErrExecutionReverted = errors.New("execution reverted")

var GAS_BUFFER = 1.10

// ProxyETHEstimateGas implements ETHProxy
type ProxyETHEstimateGas struct {
	*ProxyETHCall
}

func (p *ProxyETHEstimateGas) Method() string {
	return "eth_estimateGas"
}

func (p *ProxyETHEstimateGas) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, *eth.JSONRPCError) {
	var ethreq eth.CallRequest
	if jsonErr := unmarshalRequest(rawreq.Params, &ethreq); jsonErr != nil {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError(jsonErr.Error())
	}

	// when supplying this parameter to callcontract to estimate gas in the kaon api
	// if there isn't enough gas specified here, the result will be an exception
	// Excepted = "OutOfGasIntrinsic"
	// Gas = "the supplied value"
	// this is different from geth's behavior
	// which will return a used gas value that is higher than the incoming gas parameter
	// so we set this to nil so that callcontract will return the actual gas estimate
	ethreq.Gas = nil

	// eth req -> kaon req
	kaonreq, jsonErr := p.ToRequest(&ethreq)
	if jsonErr != nil {
		return nil, jsonErr
	}

	// kaon [code: -5] Incorrect address occurs here
	kaonresp, err := p.CallContract(c.Request().Context(), kaonreq)
	if err != nil {
		return nil, eth.NewCallbackError(err.Error())
	}

	return p.toResp(kaonresp)
}

func multiplyGasUsedByBuffer(gasUsedIn big.Int, gasBuffer float64) *big.Int {
	// Convert GAS_BUFFER to a big.Float
	buffer := new(big.Float).SetFloat64(gasBuffer)

	// Convert big.Int to big.Float
	gasUsed := new(big.Float).SetInt(&gasUsedIn)

	// Multiply big.Float values
	result := new(big.Float).Mul(gasUsed, buffer)

	// Convert the result back to big.Int
	finalResult := new(big.Int)
	result.Int(finalResult)

	return finalResult
}

func (p *ProxyETHEstimateGas) toResp(kaonresp *kaon.CallContractResponse) (*eth.EstimateGasResponse, *eth.JSONRPCError) {
	// TODO: research under which circumstances it may work
	// p.Kaon.GetLogger().Log("msg", "!!!!!!!", "requested", marshalToString(kaonresp))
	// if kaonresp.ExecutionResult.Excepted != "None" {
	// 	return nil, eth.NewCallbackError(ErrExecutionReverted.Error())
	// }
	gas := eth.EstimateGasResponse(hexutil.EncodeBig(multiplyGasUsedByBuffer(kaonresp.ExecutionResult.GasUsed, GAS_BUFFER)))
	p.GetDebugLogger().Log(p.Method(), gas)
	return &gas, nil
}
