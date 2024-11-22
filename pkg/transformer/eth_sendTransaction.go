package transformer

import (
	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/kaonone/eth-rpc-gate/pkg/kaon"
	"github.com/kaonone/eth-rpc-gate/pkg/utils"
	"github.com/labstack/echo"
	"github.com/shopspring/decimal"
)

var MinimumGasLimit = int64(22000)

// ProxyETHSendTransaction implements ETHProxy
type ProxyETHSendTransaction struct {
	*kaon.Kaon
}

func (p *ProxyETHSendTransaction) Method() string {
	return "eth_sendTransaction"
}

func (p *ProxyETHSendTransaction) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, *eth.JSONRPCError) {
	var req eth.SendTransactionRequest
	err := unmarshalRequest(rawreq.Params, &req)
	if err != nil {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError(err.Error())
	}

	if req.Gas != nil && req.Gas.Int64() < MinimumGasLimit {
		p.GetLogger().Log("msg", "Gas limit is too low", "gasLimit", req.Gas.String())
	}

	var result interface{}
	var jsonErr *eth.JSONRPCError

	if req.IsCreateContract() {
		result, jsonErr = p.requestCreateContract(&req)
	} else if req.IsSendEther() {
		result, jsonErr = p.requestSendToAddress(&req)
	} else if req.IsCallContract() {
		result, jsonErr = p.requestSendToContract(&req)
	} else {
		return nil, eth.NewInvalidParamsError("Unknown operation")
	}

	if err == nil {
		p.GenerateIfPossible()
	}

	return result, jsonErr
}

func (p *ProxyETHSendTransaction) requestSendToContract(ethtx *eth.SendTransactionRequest) (*eth.SendTransactionResponse, *eth.JSONRPCError) {
	gasLimit, gasPrice, err := EthGasToKaon(ethtx)
	if err != nil {
		return nil, eth.NewInvalidParamsError(err.Error())
	}

	amount := decimal.NewFromFloat(0.0)
	if ethtx.Value != "" {
		var err error
		amount, err = EthValueToKaonAmount(ethtx.Value, ZeroSatoshi)
		if err != nil {
			return nil, eth.NewInvalidParamsError(err.Error())
		}
	}

	kaonreq := kaon.SendToContractRequest{
		ContractAddress: utils.RemoveHexPrefix(ethtx.To),
		Datahex:         utils.RemoveHexPrefix(ethtx.Data),
		Amount:          kaon.TransformAmount(amount),
		GasLimit:        gasLimit,
		GasPrice:        gasPrice,
	}

	if from := ethtx.From; from != "" && utils.IsEthHexAddress(from) {
		from, err = p.FromHexAddress(from)
		if err != nil {
			return nil, eth.NewCallbackError(err.Error())
		}
		kaonreq.SenderAddress = from
	}

	var resp *kaon.SendToContractResponse
	if err := p.Kaon.Request(kaon.MethodSendToContract, &kaonreq, &resp); err != nil {
		return nil, eth.NewCallbackError(err.Error())
	}

	ethresp := eth.SendTransactionResponse(utils.AddHexPrefix(resp.Txid))
	return &ethresp, nil
}

func (p *ProxyETHSendTransaction) requestSendToAddress(req *eth.SendTransactionRequest) (*eth.SendTransactionResponse, *eth.JSONRPCError) {
	getKaonWalletAddress := func(addr string) (string, error) {
		if utils.IsEthHexAddress(addr) {
			return p.FromHexAddress(utils.RemoveHexPrefix(addr))
		}
		return addr, nil
	}

	from, err := getKaonWalletAddress(req.From)
	if err != nil {
		return nil, eth.NewInvalidParamsError(err.Error())
	}

	to, err := getKaonWalletAddress(req.To)
	if err != nil {
		return nil, eth.NewInvalidParamsError(err.Error())
	}

	amount, err := EthValueToKaonAmount(req.Value, ZeroSatoshi)
	if err != nil {
		return nil, eth.NewInvalidParamsError(err.Error())
	}

	p.GetDebugLogger().Log("msg", "successfully converted from wei to KAON", "wei", req.Value, "kaon", amount)

	kaonreq := kaon.SendToAddressRequest{
		Address:       to,
		Amount:        kaon.TransformAmount(amount),
		SenderAddress: from,
	}

	var kaonresp kaon.SendToAddressResponse
	if err := p.Kaon.Request(kaon.MethodSendToAddress, &kaonreq, &kaonresp); err != nil {
		// this can fail with:
		// "error": {
		//   "code": -3,
		//   "message": "Sender address does not have any unspent outputs"
		// }
		// this can happen if there are enough coins but some required are untrusted
		// you can get the trusted coin balance via getbalances rpc call
		return nil, eth.NewCallbackError(err.Error())
	}

	ethresp := eth.SendTransactionResponse(utils.AddHexPrefix(string(kaonresp)))

	return &ethresp, nil
}

func (p *ProxyETHSendTransaction) requestCreateContract(req *eth.SendTransactionRequest) (*eth.SendTransactionResponse, *eth.JSONRPCError) {
	gasLimit, gasPrice, err := EthGasToKaon(req)
	if err != nil {
		return nil, eth.NewInvalidParamsError(err.Error())
	}

	kaonreq := &kaon.CreateContractRequest{
		ByteCode: utils.RemoveHexPrefix(req.Data),
		GasLimit: gasLimit,
		GasPrice: gasPrice,
	}

	if req.From != "" {
		from := req.From
		if utils.IsEthHexAddress(from) {
			from, err = p.FromHexAddress(from)
			if err != nil {
				return nil, eth.NewCallbackError(err.Error())
			}
		}

		kaonreq.SenderAddress = from
	}

	var resp *kaon.CreateContractResponse
	if err := p.Kaon.Request(kaon.MethodCreateContract, kaonreq, &resp); err != nil {
		return nil, eth.NewCallbackError(err.Error())
	}

	ethresp := eth.SendTransactionResponse(utils.AddHexPrefix(string(resp.Txid)))

	return &ethresp, nil
}
