package transformer

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/kaonone/eth-rpc-gate/pkg/kaon"
	"github.com/kaonone/eth-rpc-gate/pkg/utils"
	"github.com/labstack/echo"
)

// ProxyETHCall implements ETHProxy
type ProxyETHCall struct {
	*kaon.Kaon
}

func (p *ProxyETHCall) Method() string {
	return "eth_call"
}

func (p *ProxyETHCall) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, *eth.JSONRPCError) {
	var req eth.CallRequest
	if err := unmarshalRequest(rawreq.Params, &req); err != nil {
		// TODO: Is this correct error code?
		return nil, eth.NewInvalidParamsError(err.Error())
	}

	return p.request(c.Request().Context(), &req)
}

func (p *ProxyETHCall) request(ctx context.Context, ethreq *eth.CallRequest) (interface{}, *eth.JSONRPCError) {
	// eth req -> kaon req
	kaonreq, jsonErr := p.ToRequest(ethreq)
	if jsonErr != nil {
		return nil, jsonErr
	}

	// Check for specific cases based on the data and to fields
	if strings.ToLower(ethreq.To) == "0x0000000000000000000000000000000000000000" {
		switch {
		case strings.HasPrefix(ethreq.Data, "0x70a08231"):
			// Handle balance request (already implemented)
			address := "0x" + ethreq.Data[34:]
			balanceReq := &eth.GetBalanceRequest{
				Address: address,
			}

			params, err := json.Marshal([]interface{}{balanceReq.Address, "latest"})
			if err != nil {
				return nil, eth.NewCallbackError("failed to marshal params: " + err.Error())
			}

			rawreq := &eth.JSONRPCRequest{
				Method: "eth_getBalance",
				Params: json.RawMessage(params),
			}

			c := echo.New().AcquireContext()
			balanceProxy := ProxyETHGetBalance{Kaon: p.Kaon}
			balanceResult, balanceErr := balanceProxy.Request(rawreq, c)

			if balanceErr != nil {
				return nil, balanceErr
			}

			callResp := eth.CallResponse(balanceResult.(string))
			return &callResp, nil

		case strings.HasPrefix(ethreq.Data, "0x06fdde03"):
			// Handle token name request (0x06fdde03 corresponds to "name()")
			name := "KAON"
			encodedName := hex.EncodeToString([]byte(name))
			// Pad the encoded name to 32 bytes
			encodedName = encodedName + strings.Repeat("0", 64-len(encodedName))
			return eth.CallResponse("0x" + encodedName), nil

		case strings.HasPrefix(ethreq.Data, "0x18160ddd"):
			// Handle total supply request (0x18160ddd corresponds to "totalSupply()")
			unlimitedSupply := new(big.Int)
			unlimitedSupply.SetString("000000000000000000000000000000000000ffffffffffffffffffffffffffff", 16)
			return eth.CallResponse(hexutil.EncodeBig(unlimitedSupply)), nil

		case strings.HasPrefix(ethreq.Data, "0x313ce567"):
			// Handle token decimals request (0x313ce567 corresponds to "decimals()")
			decimals := big.NewInt(18)
			return eth.CallResponse(hexutil.EncodeBig(decimals)), nil

		case strings.HasPrefix(ethreq.Data, "0x95d89b41"):
			// Handle token symbol request (0x95d89b41 corresponds to "symbol()")
			symbol := "KAON"
			encodedSymbol := hex.EncodeToString([]byte(symbol))
			// Pad the encoded symbol to 32 bytes
			encodedSymbol = encodedSymbol + strings.Repeat("0", 64-len(encodedSymbol))
			return eth.CallResponse("0x" + encodedSymbol), nil
		}
	}

	if kaonreq.GasLimit != nil && kaonreq.GasLimit.Cmp(big.NewInt(90000000)) > 0 {
		kaonresp := eth.CallResponse("0x")
		p.Kaon.GetLogger().Log("msg", "Caller gas above allowance, capping", "requested", kaonreq.GasLimit.Int64(), "cap", "90,000,000")
		return &kaonresp, nil
	}

	kaonresp, err := p.CallContract(ctx, kaonreq)
	if err != nil {
		if err == kaon.ErrInvalidAddress {
			kaonresp := eth.CallResponse("0x")
			return &kaonresp, nil
		}

		return nil, eth.NewCallbackError(err.Error())
	}

	// kaon res -> eth res
	return p.ToResponse(kaonresp), nil
}

func (p *ProxyETHCall) ToRequest(ethreq *eth.CallRequest) (*kaon.CallContractRequest, *eth.JSONRPCError) {
	from := ethreq.From
	var err error
	if utils.IsEthHexAddress(from) {
		from, err = p.FromHexAddress(from)
		if err != nil {
			return nil, eth.NewCallbackError(err.Error())
		}
	}

	var gasLimit *big.Int
	if ethreq.Gas != nil {
		gasLimit = ethreq.Gas.Int
	}

	if gasLimit != nil && gasLimit.Int64() < MinimumGasLimit {
		p.GetLogger().Log("msg", "Gas limit is too low", "gasLimit", gasLimit.String())
	}

	return &kaon.CallContractRequest{
		To:       ethreq.To,
		From:     from,
		Data:     ethreq.Data,
		GasLimit: gasLimit,
	}, nil
}

func (p *ProxyETHCall) ToResponse(qresp *kaon.CallContractResponse) interface{} {
	// TODO: research under which sircumstances this may work
	// if qresp.ExecutionResult.Output == "" {
	// 	return eth.NewJSONRPCError(
	// 		-32000,
	// 		"Revert: executionResult output is empty",
	// 		nil,
	// 	)
	// }

	data := utils.AddHexWithLengthPrefix(qresp.ExecutionResult.Output)
	kaonresp := eth.CallResponse(data)
	return &kaonresp

}
