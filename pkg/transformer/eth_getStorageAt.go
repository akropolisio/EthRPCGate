package transformer

import (
	"context"
	"fmt"

	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/kaonone/eth-rpc-gate/pkg/kaon"
	"github.com/kaonone/eth-rpc-gate/pkg/utils"
	"github.com/labstack/echo"
)

// ProxyETHGetStorageAt implements ETHProxy
type ProxyETHGetStorageAt struct {
	*kaon.Kaon
}

func (p *ProxyETHGetStorageAt) Method() string {
	return "eth_getStorageAt"
}

func (p *ProxyETHGetStorageAt) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, *eth.JSONRPCError) {
	var req eth.GetStorageRequest
	if err := unmarshalRequest(rawreq.Params, &req); err != nil {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError(err.Error())
	}

	kaonAddress := utils.RemoveHexPrefix(req.Address)
	blockNumber, err := getBlockNumberByParam(c.Request().Context(), p.Kaon, req.BlockNumber, false)
	if err != nil {
		p.GetDebugLogger().Log("msg", fmt.Sprintf("Failed to get block number by param for '%s'", req.BlockNumber), "err", err)
		return nil, err
	}

	return p.request(
		c.Request().Context(),
		&kaon.GetStorageRequest{
			Address:     kaonAddress,
			BlockNumber: blockNumber,
		},
		utils.RemoveHexPrefix(req.Index),
	)
}

func (p *ProxyETHGetStorageAt) request(ctx context.Context, ethreq *kaon.GetStorageRequest, index string) (*eth.GetStorageResponse, *eth.JSONRPCError) {
	kaonresp, err := p.Kaon.GetStorage(ctx, ethreq)
	if err != nil {
		return nil, eth.NewCallbackError(err.Error())
	}

	// kaon res -> eth res
	return p.ToResponse(kaonresp, index), nil
}

func (p *ProxyETHGetStorageAt) ToResponse(kaonresp *kaon.GetStorageResponse, slot string) *eth.GetStorageResponse {
	// the value for unknown anything
	storageData := eth.GetStorageResponse("0x0000000000000000000000000000000000000000000000000000000000000000")
	if len(slot) != 64 {
		slot = leftPadStringWithZerosTo64Bytes(slot)
	}
	for _, outerValue := range *kaonresp {
		kaonStorageData, ok := outerValue[slot]
		if ok {
			storageData = eth.GetStorageResponse(utils.AddHexPrefix(kaonStorageData))
			return &storageData
		}
	}

	return &storageData
}

// left pad a string with leading zeros to fit 64 bytes
func leftPadStringWithZerosTo64Bytes(hex string) string {
	return fmt.Sprintf("%064v", hex)
}
