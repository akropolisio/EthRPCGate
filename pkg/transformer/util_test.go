package transformer

import (
	"fmt"
	"testing"

	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/kaonone/eth-rpc-gate/pkg/internal"
	"github.com/kaonone/eth-rpc-gate/pkg/kaon"
	"github.com/kaonone/eth-rpc-gate/pkg/utils"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestEthValueToKaonAmount(t *testing.T) {
	cases := []map[string]interface{}{
		{
			"in":   "0xde0b6b3a7640000",
			"want": decimal.NewFromFloat(1),
		},
		{

			"in":   "0x6f05b59d3b20000",
			"want": decimal.NewFromFloat(0.5),
		},
		{
			"in":   "0x2540be400",
			"want": decimal.NewFromFloat(0.00000001),
		},
		{
			"in":   "0x1",
			"want": decimal.NewFromInt(0),
		},
	}
	for _, c := range cases {
		in := c["in"].(string)
		want := c["want"].(decimal.Decimal)
		got, err := EthValueToKaonAmount(in, MinimumGas)
		if err != nil {
			t.Error(err)
		}

		// TODO: Refactor to use new testing utilities?
		if !got.Equal(want) {
			t.Errorf("in: %s, want: %v, got: %v", in, want, got)
		}
	}
}

func TestKaonValueToEthAmount(t *testing.T) {
	cases := []decimal.Decimal{
		decimal.NewFromFloat(1),
		decimal.NewFromFloat(0.5),
		decimal.NewFromFloat(0.00000001),
		MinimumGas,
	}
	for _, c := range cases {
		in := c
		eth := KaonDecimalValueToETHAmount(in)
		out := EthDecimalValueToKaonAmount(eth)

		// TODO: Refactor to use new testing utilities?
		if !in.Equals(out) {
			t.Errorf("in: %s, eth: %v, kaon: %v", in, eth, out)
		}
	}
}

func TestKaonAmountToEthValue(t *testing.T) {
	in, want := decimal.NewFromFloat(0.1), "0x16345785d8a0000"
	got, err := formatKaonAmount(in)
	if err != nil {
		t.Error(err)
	}

	internal.CheckTestResultUnspecifiedInputMarshal(in, want, got, t, false)
}

func TestLowestKaonAmountToEthValue(t *testing.T) {
	in, want := decimal.NewFromFloat(0.00000001), "0x2540be400"
	got, err := formatKaonAmount(in)
	if err != nil {
		t.Error(err)
	}

	internal.CheckTestResultUnspecifiedInputMarshal(in, want, got, t, false)
}

func TestAddressesConversion(t *testing.T) {
	t.Parallel()

	inputs := []struct {
		kaonChain   string
		ethAddress  string
		kaonAddress string
	}{
		{
			kaonChain:   kaon.ChainTest,
			ethAddress:  "6c89a1a6ca2ae7c00b248bb2832d6f480f27da68",
			kaonAddress: "uTTH1Yr2eKCuDLqfxUyBLCAjmomQ8pyrBt",
		},

		// Test cases for addresses defined here:
		// 	- https://github.com/hayeah/openzeppelin-solidity/blob/kaon/QTUM-NOTES.md#create-test-accounts
		//
		// NOTE: Ethereum addresses are without `0x` prefix, as it expects by conversion functions
		{
			kaonChain:   kaon.ChainTest,
			ethAddress:  "1CE507204a6fC8fd6aA7e54D1481d30ACB0Dbead",
			kaonAddress: "ar2SzdHghSgeacypPn7zfDe3qfKAEwimus",
		},
		{
			kaonChain:   kaon.ChainTest,
			ethAddress:  "3f501c368cb9ddb5f27ed72ac0d602724adfa175",
			kaonAddress: "auASFMxv45WgjCW6wkpDuHWjxXhzNA9mjP",
		},
		{
			kaonChain:   kaon.ChainTest,
			ethAddress:  "57ed9afd4668ab81b648e68d2a76227434d6a8ee",
			kaonAddress: "awQb8vf21idkFoZiYPA4hWgtuPyko2qUaR",
		},
		{
			kaonChain:   kaon.ChainTest,
			ethAddress:  "1dd46713aa54541c74f4ef391b59b55133f675ec",
			kaonAddress: "ar7PkgNdY1HkDtUo3D4GTsYrcqoHBJygNQ",
		},
		{
			kaonChain:   kaon.ChainTest,
			ethAddress:  "c3530fe16dd1cc69dae31dc6f029ca57feab5536",
			kaonAddress: "b7CSynDNwb2LQcCWXs8Qn79LUkgMdsK61S",
		},
		{
			kaonChain:   kaon.ChainTest,
			ethAddress:  "6c880fa6feb2a5917bcc1afc8afa0e4f61776a8f",
			kaonAddress: "ayHXgXugbHDDR8cBjX2ZVLfkGE78QTeW2Z",
		},
	}

	for i, in := range inputs {
		var (
			in       = in
			testDesc = fmt.Sprintf("#%d", i)
		)
		// TODO: Investigate why this testing setup is so different
		t.Run(testDesc, func(t *testing.T) {
			kaonAddress, err := convertETHAddress(in.ethAddress, in.kaonChain)
			require.NoError(t, err, "couldn't convert Ethereum address to Kaon address")
			require.Equal(t, in.kaonAddress, kaonAddress, "unexpected converted Kaon address value")

			ethAddress, err := utils.ConvertKaonAddress(in.kaonAddress)
			require.NoError(t, err, "couldn't convert Kaon address to Ethereum address")
			require.Equal(t, in.ethAddress, ethAddress, "unexpected converted Ethereum address value")
		})
	}
}

func TestSendTransactionRequestHasDefaultGasPriceAndAmount(t *testing.T) {
	var req eth.SendTransactionRequest
	err := unmarshalRequest([]byte(`[{}]`), &req)
	if err != nil {
		t.Fatal(err)
	}
	defaultGasPriceInWei := req.GasPrice.Int
	defaultGasPriceInKAON := EthDecimalValueToKaonAmount(decimal.NewFromBigInt(defaultGasPriceInWei, 1))

	// TODO: Refactor to use new testing utilities?
	if !defaultGasPriceInKAON.Equals(MinimumGas) {
		t.Fatalf("Default gas price does not convert to KAON minimum gas price, got: %s want: %s", defaultGasPriceInKAON.String(), MinimumGas.String())
	}
	if eth.DefaultGasAmountForKaon.String() != req.Gas.Int.String() {
		t.Fatalf("Default gas amount does not match expected default, got: %s want: %s", req.Gas.Int.String(), eth.DefaultGasAmountForKaon.String())
	}
}
