package transformer

import (
	"context"
	"fmt"
	"math/big"

	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/kaonone/eth-rpc-gate/pkg/kaon"
	"github.com/kaonone/eth-rpc-gate/pkg/utils"
	"github.com/labstack/echo"
	"github.com/shopspring/decimal"
)

type ProxyKAONGetUTXOs struct {
	*kaon.Kaon
}

var _ ETHProxy = (*ProxyKAONGetUTXOs)(nil)

func (p *ProxyKAONGetUTXOs) Method() string {
	return "kaon_getUTXOs"
}

func (p *ProxyKAONGetUTXOs) Request(req *eth.JSONRPCRequest, c echo.Context) (interface{}, *eth.JSONRPCError) {
	var params eth.GetUTXOsRequest
	if err := unmarshalRequest(req.Params, &params); err != nil {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError("couldn't unmarshal request parameters")
	}

	err := params.CheckHasValidValues()
	if err != nil {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError("couldn't validate parameters value")
	}

	return p.request(c.Request().Context(), params)
}

func (p *ProxyKAONGetUTXOs) request(ctx context.Context, params eth.GetUTXOsRequest) (*eth.GetUTXOsResponse, *eth.JSONRPCError) {
	address, err := convertETHAddress(utils.RemoveHexPrefix(params.Address), p.Chain())
	if err != nil {
		return nil, eth.NewInvalidParamsError("couldn't convert Ethereum address to Kaon address")
	}

	req := kaon.GetAddressUTXOsRequest{
		Addresses: []string{address},
	}

	resp, err := p.Kaon.GetAddressUTXOs(ctx, &req)
	if err != nil {
		return nil, eth.NewCallbackError(err.Error())
	}

	blockCount, err := p.Kaon.GetBlockCount(ctx)
	if err != nil {
		return nil, eth.NewCallbackError(err.Error())
	}

	matureBlockHeight := big.NewInt(int64(p.Kaon.GetMatureBlockHeight()))

	//Convert minSumAmount to Satoshis
	minimumSum := convertFromKaonToSatoshis(params.MinSumAmount)
	queryingAll := minimumSum.Equal(decimal.Zero)

	allUtxoTypes := false
	if len(params.Types) > 0 {
		if params.Types[0] == eth.ALL_UTXO_TYPES {
			allUtxoTypes = true
		}
	} else {
		allUtxoTypes = true
	}

	utxoTypes := map[eth.UTXOScriptType]bool{}
	for _, typ := range params.Types {
		utxoTypes[typ] = true
	}

	var utxos []eth.KaonUTXO
	var minUTXOsSum decimal.Decimal
	for _, utxo := range *resp {
		ethUTXO := toEthResponseType(utxo)
		ethUTXO.Height = uint64(utxo.Height.Int64())
		ethUTXO.ScriptPubKey = utxo.Script
		utxoType := ethUTXO.GetType()
		ethUTXO.Type = utxoType.String()
		ethUTXO.Safe = true
		if !allUtxoTypes {
			if _, ok := utxoTypes[utxoType]; !ok {
				continue
			}
		}

		if utxo.IsStake {
			matureAt := big.NewInt(utxo.Height.Int64()).Add(
				big.NewInt(utxo.Height.Int64()),
				matureBlockHeight,
			)
			if blockCount.Int.Cmp(matureAt) <= 0 {
				// immature
				ethUTXO.Safe = false
				if !allUtxoTypes {
					if _, ok := utxoTypes[eth.IMMATURE]; !ok {
						continue
					}
				}
			}
		}

		ethUTXO.Confirmations = blockCount.Int64() - utxo.Height.Int64()
		if ethUTXO.Confirmations < 0 {
			panic(fmt.Sprintf("Computed negative confirmations: %d - %d = %d\n", blockCount.Int64(), utxo.Height.Int64(), ethUTXO.Confirmations))
		}
		ethUTXO.Spendable = true

		if ethUTXO.Safe {
			minUTXOsSum = minUTXOsSum.Add(utxo.Satoshis.Decimal)
		}
		utxos = append(utxos, ethUTXO)
		if !queryingAll && minUTXOsSum.GreaterThanOrEqual(minimumSum) {
			return (*eth.GetUTXOsResponse)(&utxos), nil
		}
	}

	if queryingAll {
		return (*eth.GetUTXOsResponse)(&utxos), nil
	}

	return nil, eth.NewCallbackError("required minimum amount is greater than total amount of UTXOs")
}

func toEthResponseType(utxo kaon.UTXO) eth.KaonUTXO {
	return eth.KaonUTXO{
		Address: utxo.Address,
		TXID:    utxo.TXID,
		Vout:    utxo.OutputIndex,
		Amount:  convertFromSatoshisToKaon(utxo.Satoshis.Decimal).String(),
	}
}
