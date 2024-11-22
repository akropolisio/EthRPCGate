package utils

import (
	"math/big"

	"github.com/shopspring/decimal"
)

func InStrSlice(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

// Specific decimal.Decimal to big.Int conversion to support big values
func ToBigInt(value *decimal.Decimal) (*big.Int, error) {
	return value.BigInt(), nil
}

// Specific decimal.Decimal to big.Int conversion to support big values
func ToDecimal(value *big.Int) (*decimal.Decimal, error) {
	dst := new(decimal.Decimal)

	str, errEncode := value.MarshalText()
	if errEncode != nil {
		return dst, errEncode
	}

	dst.UnmarshalText(str)
	return dst, nil
}
