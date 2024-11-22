package kaon

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
)

type Accounts []*btcutil.WIF

func (as Accounts) FindByHexAddress(addr string) *btcutil.WIF {
	for _, a := range as {
		acc := &Account{a}

		if addr == acc.ToHexAddress() {
			return a
		}
	}

	return nil
}

type Account struct {
	*btcutil.WIF
}

func (a *Account) ToHexAddress() string {
	// wif := (*btcutil.WIF)(a)

	keyid := btcutil.Hash160(a.SerializePubKey())
	return hex.EncodeToString(keyid)
}

var kaonMainNetParams = chaincfg.MainNetParams
var kaonTestNetParams = chaincfg.MainNetParams

func init() {
	kaonMainNetParams.PubKeyHashAddrID = 58
	kaonMainNetParams.ScriptHashAddrID = 50

	kaonTestNetParams.PubKeyHashAddrID = 120
	kaonTestNetParams.ScriptHashAddrID = 110
}

func (a *Account) ToBase58Address(isMain bool) (string, error) {
	params := &kaonMainNetParams
	if !isMain {
		params = &kaonTestNetParams
	}

	addr, err := btcutil.NewAddressPubKey(a.SerializePubKey(), params)
	if err != nil {
		return "", err
	}

	return addr.AddressPubKeyHash().String(), nil
}
