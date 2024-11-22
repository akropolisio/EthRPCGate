package kaon

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/kaonone/eth-rpc-gate/pkg/utils"
	"github.com/pkg/errors"
)

type (
	ContractInvokeInfo struct {
		// VMVersion string
		From     string
		GasLimit string
		GasPrice string
		CallData string
		To       string
	}
)

// Given a hex string, reverse the order of the bytes
// Used in special cases when gasLimit and gasPrice are 'bigendian hex encoded'
func reversePartToHex(s string) (string, error) {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+2, j-2 {
		runes[i], runes[i+1], runes[j-1], runes[j] = runes[j-1], runes[j], runes[i], runes[i+1]
	}
	partInt, err := strconv.ParseInt(string(runes), 16, 64)
	if err != nil {
		return "", err
	}
	partHex := strconv.FormatInt(partInt, 16)
	return partHex, nil
}

func ParseP2PKHReciever(parts []string) (*string, error) {
	// OP_DUP
	// OP_HASH160_COMPAT
	// OP_PUBKEYHASH  - zZkaonAddr - base58 reciever addr
	// OP_EQUALVERIFY
	// OP_CHECKSIG

	if len(parts) != 5 {
		return nil, errors.New(fmt.Sprintf("invalid Pay2PKH script for parts 5: %v", parts))
	}
	return &parts[2], nil
}

func ParseCallASM(parts []string) (*ContractInvokeInfo, error) {
	// "4 25548 40 8588b2c50000000000000000000000000000000000000000000000000000000000000000 57946bb437560b13275c32a468c6fd1e0c2cdd48 OP_CAL"

	// 4                     // EVM version
	// 100000                // gas limit
	// 10                    // gas price
	// 1234                  // data to be sent by the contract
	// Contract Address      // contract address
	// OP_CALL

	if len(parts) != 6 {
		return nil, errors.New(fmt.Sprintf("invalid OP_CALL script for parts 6: %v", parts))
	}

	//! For some TXs we get the following string with GasPrice and/or GasLimit encoded as 'big endian' hex:
	//"4 90d0030000000000 2800000000000000 a9059cbb0000000000000000000000008e60e0b8371c0312cfc703e5e28bc57dfa0674fd0000000000000000000000000000000000000000000000000000000005f5e100 f2703e93f87b846a7aacec1247beaec1c583daa4 OP_CALL"
	//! Current fix checks if GasLimit or GasPrice are hex encoded, then reverts the order of the bytes
	//! in GasPrice and GasLimit fields and returns the correct hex values.
	// i.e. gasLimit = "90d0030000000000" is returned as "0x3d090"
	//! This approach will fail to detect the case where the GasPrice and GasLimit are encoded as hex but 'stringBase10ToHex' does not return an error.
	// TODO: research alternative approach to fix this.
	gasLimit, gasPrice, err := parseGasFields(parts[1], parts[2])
	if err != nil {
		return nil, err
	}

	return &ContractInvokeInfo{
		// From:     parts[1],
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		CallData: parts[3],
		To:       parts[4],
	}, nil

}

func ParseCallSenderASM(parts []string) (*ContractInvokeInfo, error) {
	// See: https://github.com/qtumproject/qips/issues/6

	// "1 1CE507204a6fC8fd6aA7e54D1481d30ACB0Dbead 69463043021f3ba540f52e0bae0c608c3d7135424fb683c77ee03217fcfe0af175c586aadc02200222e460a42268f02f130bc46f3ef62f228dd8051756dc13693332423515fcd401210299d391f528b9edd07284c7e23df8415232a8ce41531cf460a390ce32b4efd112 OP_SENDER 4 40000000 40 60fe47b10000000000000000000000000000000000000000000000000000000000000319 9e11fba86ee5d0ba4996b0d1973de6b694f4fc95 OP_CALL"
	// 1                     // address type of the pubkeyhash (public key hash)
	// Address               // sender's pubkeyhash address
	// {signature, pubkey}   // serialized scriptSig
	// OP_SENDER
	// 4                     // EVM version
	// 100000                // gas limit
	// 10                    // gas price
	// 1234                  // data to be sent by the contract
	// Contract Address      // contract address
	// OP_CALL

	if len(parts) != 10 {
		return nil, errors.New(fmt.Sprintf("invalid create_sender script for parts 10: %v", parts))
	}

	gasLimit, gasPrice, err := parseGasFields(parts[5], parts[6])
	if err != nil {
		return nil, err
	}

	return &ContractInvokeInfo{
		From:     parts[1],
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		CallData: parts[7],
		To:       parts[8],
	}, nil

}

func ParseCreateASM(parts []string) (*ContractInvokeInfo, error) {
	// OP_CREATE prefixed by OP_SENDER always have this structure:
	//
	// 1                     // address type of the pubkeyhash (public key hash)
	// Address               // sender's pubkeyhash address
	// {signature, pubkey}   // serialized scriptSig
	// OP_SENDER
	// 4                     // EVM version
	// 100000                // gas limit
	// 10                    // gas price
	// 1234                  // data to be sent by the contract
	// OP_CREATE

	if len(parts) < 5 {
		return nil, nil
	}

	gasLimit, gasPrice, err := parseGasFields(parts[1], parts[2])
	if err != nil {
		return nil, err
	}
	info := &ContractInvokeInfo{
		From:     parts[1],
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		CallData: parts[3],
	}
	return info, nil
}

func ParseCreateSenderASM(parts []string) (*ContractInvokeInfo, error) {
	// See: https://github.com/qtumproject/qips/issues/6
	// https://blog.qtum.org/qip-5-add-op-sender-opcode-571511802938

	// "1 1CE507204a6fC8fd6aA7e54D1481d30ACB0Dbead 6a473044022067ca66b0308ae16aeca7a205ce0490b44a61feebe5632710b52aabde197f9e4802200e8beec61a58dbe1279a9cdb68983080052ae7b9997bc863b7c5623e4cb55fd
	// b01210299d391f528b9edd07284c7e23df8415232a8ce41531cf460a390ce32b4efd112 OP_SENDER 4 6721975 100 6060604052341561000f57600080fd5b60008054600160a060020a033316600160a060020a03199091161790556101de8061003b6000
	// 396000f300606060405263ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416630900f010811461005d578063445df0ac1461007e5780638da5cb5b146100a3578063fdacd576146100d257600080fd5b341561
	// 006857600080fd5b61007c600160a060020a03600435166100e8565b005b341561008957600080fd5b61009161017d565b60405190815260200160405180910390f35b34156100ae57600080fd5b6100b6610183565b604051600160a060020a039091168152
	// 60200160405180910390f35b34156100dd57600080fd5b61007c600435610192565b6000805433600160a060020a03908116911614156101795781905080600160a060020a031663fdacd5766001546040517c01000000000000000000000000000000000000
	// 0000000000000000000063ffffffff84160281526004810191909152602401600060405180830381600087803b151561016457600080fd5b6102c65a03f1151561017557600080fd5b5050505b5050565b60015481565b600054600160a060020a031681565b
	// 60005433600160a060020a03908116911614156101af5760018190555b505600a165627a7a72305820b6a912c5b5115d1a5412235282372dc4314f325bac71ee6c8bd18f658d7ed1ad0029 OP_CREATE"

	// 1    // address type of the pubkeyhash (public key hash)
	// Address               // sender's pubkeyhash address
	// {signature, pubkey}   //serialized scriptSig
	// OP_SENDER
	// 4                     // EVM version
	// 100000                //gas limit
	// 10                    //gas price
	// 1234                  // data to be sent by the contract
	// OP_CREATE

	if len(parts) != 9 {
		return nil, errors.New(fmt.Sprintf("invalid create_sender script for parts 9: %v", len(parts)))
	}

	gasLimit, gasPrice, err := parseGasFields(parts[5], parts[6])
	if err != nil {
		return nil, err
	}
	info := &ContractInvokeInfo{
		From:     parts[1],
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		CallData: parts[7],
	}
	return info, nil
}

func stringBase10ToHex(str string) (string, error) {
	var v big.Int
	_, ok := v.SetString(str, 0)
	if !ok {
		return "", errors.Errorf("Failed to parse big.Int: %s", str)
	}

	return utils.AddHexPrefix(v.Text(16)), nil
}

func parseGasFields(gasLimitPart, gasPricePart string) (string, string, error) {
	gasLimit, err1 := stringBase10ToHex(utils.AddHexPrefix(gasLimitPart))
	gasPrice, err2 := stringBase10ToHex(utils.AddHexPrefix(gasPricePart))
	if err1 != nil || err2 != nil {
		gasLimit, err1 = reversePartToHex(gasLimitPart)
		if err1 != nil {
			return "", "", err1
		}
		gasPrice, err2 = reversePartToHex(gasPricePart)
		if err2 != nil {
			return "", "", err2
		}
	}
	return gasLimit, gasPrice, nil
}
