package utils

import (
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/btcsuite/btcutil/base58"
	// "github.com/decred/base58"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
)

func RemoveHexPrefix(hex string) string {
	if strings.HasPrefix(hex, "0x") {
		return hex[2:]
	}
	return hex
}

func IsEthHexAddress(str string) bool {
	return strings.HasPrefix(str, "0x") || common.IsHexAddress("0x"+str)
}

func AddHexPrefix(hex string) string {
	if strings.HasPrefix(hex, "0x") {
		return hex
	}
	return "0x" + hex
}

func AddHexWithLengthPrefix(hex string) string {
	// Remove existing "0x" prefix if present
	if strings.HasPrefix(hex, "0x") {
		hex = hex[2:]
	}

	// Calculate the number of leading zeroes needed
	requiredLength := 64
	currentLength := len(hex)
	zeroesNeeded := requiredLength - currentLength

	// Add leading zeroes
	for i := 0; i < zeroesNeeded; i++ {
		hex = "0" + hex
	}

	// Add "0x" prefix
	return "0x" + hex
}

func AddHexPrefixIfNotEmpty(hex string) string {
	if hex == "" {
		return hex
	}
	return AddHexPrefix(hex)
}

func AddHexWithLengthPrefixIfNotEmpty(hex string) string {
	if hex == "" {
		return hex
	}
	if hex == "0" {
		return ""
	}
	return AddHexWithLengthPrefix(hex)
}

// DecodeBig decodes a hex string whether input is with 0x prefix or not.
func DecodeBig(input string) (*big.Int, error) {
	input = AddHexPrefix(input)
	if input == "0x00" {
		return big.NewInt(0), nil
	}
	return hexutil.DecodeBig(input)
}

// Converts Kaon address to an Ethereum address
func ConvertKaonAddress(address string) (ethAddress string, _ error) {
	if n := len(address); n < 22 {
		return "", errors.Errorf("invalid address: length is less than 22 bytes - %d", n)
	}

	_, _, err := base58.CheckDecode(address)
	if err != nil {
		return "", errors.Errorf("invalid address")
	}

	// Drop Kaon chain prefix and checksum suffix
	ethAddrBytes := base58.Decode(address)[1:21]

	return hex.EncodeToString(ethAddrBytes), nil
}
