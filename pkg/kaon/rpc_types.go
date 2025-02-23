package kaon

import (
	"encoding/json"
	"math/big"
	"strconv"
	"strings"

	"github.com/kaonone/eth-rpc-gate/pkg/utils"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

// Amount is a custom type to handle big integers with automatic scaling.
type Amount struct {
	decimal.Decimal
}

var ZeroAmount = NewAmount(0, 1)

// UnmarshalJSON is a custom unmarshaler that scales the number by 10^18.
func (a *Amount) UnmarshalJSON(data []byte) error {
	// Temporary variable for unmarshaling
	var d decimal.Decimal
	if err := json.Unmarshal(data, &d); err != nil {
		return err
	}

	// Scale by 10^18
	scaleFactor := decimal.NewFromInt(1e18)
	scaled := d.Mul(scaleFactor)

	// Assign the scaled value to Amount
	a.Decimal = scaled

	return nil
}

func NewAmount(value int64, exp int32) Amount {
	d := decimal.New(value, exp)
	return Amount{Decimal: d}
}

func TransformAmount(d decimal.Decimal) Amount {
	return Amount{Decimal: d}
}

// NewFromFloat creates a new Amount from a float64.
func NewFromFloat(value float64) Amount {
	d := decimal.NewFromFloat(value)
	// Apply scaling if needed, e.g., multiply by 10^18
	scaleFactor := decimal.NewFromInt(1e18)
	scaled := d.Mul(scaleFactor)
	return Amount{Decimal: scaled}
}

// NewFromString creates a new Amount from a string.
func NewFromString(value string) (Amount, error) {
	d, err := decimal.NewFromString(value)
	if err != nil {
		return Amount{}, err
	}
	// Apply scaling if needed, e.g., multiply by 10^18
	scaleFactor := decimal.NewFromInt(1e18)
	scaled := d.Mul(scaleFactor)
	return Amount{Decimal: scaled}, nil
}

// NewFromString creates a new Amount from a string.
func NewFromInt(value int64) Amount {
	d := decimal.NewFromInt(value)

	// Apply scaling if needed, e.g., multiply by 10^18
	scaleFactor := decimal.NewFromInt(1e18)
	scaled := d.Mul(scaleFactor)
	return Amount{Decimal: scaled}
}

// TODO: Wipe these out when it comes time to change over from floats to integers, and change SendToContractRequest to not use strings where numerics will do
// Todo: Go and fix the need for a custom json unmarshall in the non raw versions of these types

const (
	genesisBlockHeight = 0

	// Is hex representation of 40M value, which is the block gas limit, 20M is tx gas limit
	DefaultBlockGasLimit = "2625A00"

	// Is a zero wallet address, which is used as a stub, when
	// original value cannot be defined in such cases as generated
	// transaction
	ZeroAddress = "0000000000000000000000000000000000000000"

	// Is a zero user_input/label, that usually may be send along
	// with a transaction or contract. Primarly usage is as stub,
	// when original value has not been provided
	//
	// This value has the minimum length, which is acceptable by
	// graph-node
	ZeroUserInput = "00"
)

type SendToContractRawRequest struct {
	ContractAddress string   `json:"contractAddress"`
	Datahex         string   `json:"data"`
	Amount          Amount   `json:"amount"`
	GasLimit        *big.Int `json:"gasLimit"`
	GasPrice        string   `json:"gasPrice"`
	SenderAddress   string   `json:"senderaddress"`
}

type CreateContractRawRequest struct {
	ByteCode      string   `json:"bytecode"`
	GasLimit      *big.Int `json:"gasLimit"`
	GasPrice      string   `json:"gasPrice"`
	SenderAddress string   `json:"senderaddress"`
}

type (
	Log struct {
		Address string   `json:"address"`
		Topics  []string `json:"topics"`
		Data    string   `json:"data"`
		Index   int      `json:"-"` // Keep track of which index the log is at
	}

	LogBlockData interface {
		GetTransactionHash() string
		GetTransactionIndex() uint64
		GetBlockHash() string
		GetBlockNumber() uint64
	}

	/*
		{
		  "chain": "regtest",
		  "blocks": 4137,
		  "headers": 4137,
		  "bestblockhash": "3863e43665ab15af1167df2f30a1c6f658c64704a3a2903bb0c5afde7e5d54ff",
		  "difficulty": 4.656542373906925e-10,
		  "mediantime": 1533096368,
		  "verificationprogress": 1,
		  "initialblockdownload": false,
		  "chainwork": "0000000000000000000000000000000000000000000000000000000000002054",
		  "size_on_disk": 3103111212,
		  "pruned": false,
		  "pruneheight": 0,
		  "automatic_pruning": false,
		  "prune_target_size": 0,
		  "softforks": {
			"csv": {
				"type": "buried|bip9|other"
				"bip9": {
					"status": "defined|started|lockedIn|active|failed",
					"bit": 1,
					"start_time": 0,
					"timeout": 999999999999,
					"since": 432
					"statistics": {
						"period": 2,
						"threshold": 0,
						"elapsed": 1,
						"count": 10223,
						"possible": true
					}
				},
				"height": 10,
				"active": false
			},
			"segwit": {
				"type": "buried|bip9|other"
				"bip9": {
					"status": "defined|started|lockedIn|active|failed",
					"bit": 1,
					"start_time": 0,
					"timeout": 999999999999,
					"since": 432
					"statistics": {
						"period": 2,
						"threshold": 0,
						"elapsed": 1,
						"count": 10223,
						"possible": true
					}
				},
				"height": 10,
				"active": false
			}.
			"XXXX": {...}
		  },
		  "warnings" : "str"
		}
	*/
	GetBlockChainInfoResponse struct {
		Chain                string  `json:"chain"`
		Blocks               int64   `json:"blocks"`
		Headers              int64   `json:"headers"`
		Bestblockhash        string  `json:"bestblockhash"`
		Difficulty           float64 `json:"difficulty"`
		Mediantime           int64   `json:"mediantime"`
		Verificationprogress float64 `json:"verificationprogress"`
		Initialblockdownload bool    `json:"initialblockdownload"`
		Chainwork            string  `json:"chainwork"`
		SizeOnDisk           int64   `json:"size_on_disk"`
		Pruned               bool    `json:"pruned"`
		PruneHeight          int64   `json:"pruneheight"`
		AutomaticPruning     bool    `json:"automatic_pruning"`
		PruneTargetSize      int64   `json:"prune_target_size"`
		Softforks            map[string]struct {
			Segwit struct {
				Type string `json:"type"`
				Bip9 struct {
					Status     string `json:"status"`
					Bit        int64  `json:"bit"`
					StartTime  int64  `json:"start_time"`
					Timout     int64  `json:"timeout"`
					Since      int64  `json:"since"`
					Statistics struct {
						Period    int64 `json:"period"`
						Threshold int64 `json:"threshold"`
						Elapsed   int64 `json:"elapsed"`
						Count     int64 `json:"count"`
						Possible  bool  `json:"possible"`
					} `json:"statistics"`
				} `json:"bip9"`
				Height int64 `json:"height"`
				Active bool  `json:"active"`
			} `json:"segwit"`
		} `json:"softforks"`
		Warnings string `json:"warnings"`
	}
)

func (l Log) GetAddress() string {
	return l.Address
}

func (l Log) GetTopics() []string {
	return l.Topics
}

func (l Log) GetData() string {
	return l.Data
}

// ========== SendToAddress ============= //

type (
	SendToAddressRequest struct {
		Address       string
		Amount        Amount
		SenderAddress string
	}
	SendToAddressResponse string
)

func (r *SendToAddressRequest) MarshalJSON() ([]byte, error) {
	/*
		1. "address"            (string, required) The Kaon address to send to.
		2. "amount"             (numeric or string, required) The amount in Kaon to send. eg 0.1
		3. "comment"            (string, optional) A comment used to store what the transaction is for.
		                             This is not part of the transaction, just kept in your wallet.
		4. "comment_to"         (string, optional) A comment to store the name of the person or organization
		                             to which you're sending the transaction. This is not part of the
		                             transaction, just kept in your wallet.
		5. subtractfeefromamount  (boolean, optional, default=false) The fee will be deducted from the amount being sent.
		                             The recipient will receive less KAONs than you enter in the amount field.
		6. replaceable            (boolean, optional) Allow this transaction to be replaced by a transaction with higher fees via BIP 125
		7. conf_target            (numeric, optional) Confirmation target (in blocks)
		8. "estimate_mode"      (string, optional, default=UNSET) The fee estimate mode, must be one of:
		       "UNSET"
		       "ECONOMICAL"
		       "CONSERVATIVE"
		9. "avoid_reuse" 	(boolean, optional, default=true) Avoid spending from dirty addresses;
					addresses are considered dirty if they have previously been used in a transaction
		10. "senderaddress"      (string, optional) The Kaon address that will be used to send money from.
		11."changeToSender"     (bool, optional, default=false) Return the change to the sender.
	*/
	return json.Marshal([]interface{}{
		r.Address,
		r.Amount,
		"", // comment
		"", // comment_to
		false,
		nil,
		nil,
		nil,
		false,
		r.SenderAddress,
		true,
	})
}

// ========== SendToContract ============= //

type (
	SendToContractRequest struct {
		ContractAddress string
		Datahex         string
		Amount          Amount
		GasLimit        *big.Int
		GasPrice        string
		SenderAddress   string
	}

	/*
		{
		  "txid": "6b7f70d8520e1ec87ba7f1ee559b491cc3028b77ae166e789be882b5d370eac9",
		  "sender": "qTKrsHUrzutdCVu3qi3iV1upzB2QpuRsRb",
		  "hash160": "6b22910b1e302cf74803ffd1691c2ecb858d3712"
		}
	*/
	SendToContractResponse struct {
		Txid    string `json:"txid"`
		Sender  string `json:"sender"`
		Hash160 string `json:"hash160"`
	}
)

func (r *SendToContractRequest) MarshalJSON() ([]byte, error) {
	/*
	   1. "contractaddress" (string, required) The contract address that will receive the funds and data.
	   2. "datahex"  (string, required) data to send.
	   3. "amount"      (numeric or string, optional) The amount in KAON to send. eg 0.1, default: 0
	   4. gasLimit  (numeric or string, optional) gasLimit, default: 250000, max: 40000000
	   5. gasPrice  (numeric or string, optional) gasPrice KAON price per gas unit, default: 0.0000004, min:0.0000004
	   6. "senderaddress" (string, optional) The Kaon address that will be used as sender.
	   7. "broadcast" (bool, optional, default=true) Whether to broadcast the transaction or not.
	   8. "changeToSender" (bool, optional, default=true) Return the change to the sender.
	*/

	return json.Marshal([]interface{}{
		r.ContractAddress,
		r.Datahex,
		r.Amount,
		r.GasLimit,
		r.GasPrice,
		r.SenderAddress,
	})
}

// ========== CreateContract ============= //

type (
	CreateContractRequest struct {
		ByteCode      string
		GasLimit      *big.Int
		GasPrice      string
		SenderAddress string
	}
	/*
	   {
	   "txid": "d0fe0caa1b798c36da37e9118a06a7d151632d670b82d1c7dc3985577a71880f",
	   "sender": "qTKrsHUrzutdCVu3qi3iV1upzB2QpuRsRb",
	   "hash160": "6b22910b1e302cf74803ffd1691c2ecb858d3712",
	   "address": "c89a5d225f578d84a94741490c1b40889b4f7a00"
	   }
	*/
	CreateContractResponse struct {
		Txid    string `json:"txid"`
		Sender  string `json:"sender"`
		Hash160 string `json:"hash160"`
		Address string `json:"address"`
	}
)

func (r *CreateContractRequest) MarshalJSON() ([]byte, error) {
	/*
		1. "bytecode"  (string, required) contract bytcode.
		2. gasLimit  (numeric or string, optional) gasLimit, default: 2500000, max: 40000000
		3. gasPrice  (numeric or string, optional) gasPrice KAON price per gas unit, default: 0.0000004, min:0.0000004
		4. "senderaddress" (string, optional) The Kaon address that will be used to create the contract.
		5. "broadcast" (bool, optional, default=true) Whether to broadcast the transaction or not.
		6. "changeToSender" (bool, optional, default=true) Return the change to the sender.
	*/
	return json.Marshal([]interface{}{
		r.ByteCode,
		r.GasLimit,
		r.GasPrice,
		r.SenderAddress,
	})
}

// ========== CallContract ============= //

type (
	CallContractRequest struct {
		From     string
		To       string
		Data     string
		GasLimit *big.Int
	}

	/*
		{
		  "address": "1e6f89d7399081b4f8f8aa1ae2805a5efff2f960",
		  "executionResult": {
		    "gasUsed": 21678,
		    "excepted": "None",
		    "newAddress": "1e6f89d7399081b4f8f8aa1ae2805a5efff2f960",
		    "output": "0000000000000000000000000000000000000000000000000000000000000001",
		    "codeDeposit": 0,
		    "gasRefunded": 0,
		    "depositSize": 0,
		    "gasForDeposit": 0
		  },
		  "transactionReceipt": {
		    "stateRoot": "d44fc5ad43bae52f01ff7eb4a7bba904ee52aea6c41f337aa29754e57c73fba6",
		    "gasUsed": 21678,
		    "bloom": "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
		    "log": []
		  }
		}
	*/
	CallContractResponse struct {
		Address         string `json:"address"`
		ExecutionResult struct {
			GasUsed         big.Int `json:"gasUsed"`
			Excepted        string  `json:"excepted"`
			ExceptedMessage string  `json:"exceptedMessage"`
			NewAddress      string  `json:"newAddress"`
			Output          string  `json:"output"`
			CodeDeposit     int     `json:"codeDeposit"`
			GasRefunded     big.Int `json:"gasRefunded"`
			DepositSize     int     `json:"depositSize"`
			GasForDeposit   big.Int `json:"gasForDeposit"`
		} `json:"executionResult"`
		TransactionReceipt struct {
			StateRoot string        `json:"stateRoot"`
			GasUsed   big.Int       `json:"gasUsed"`
			Bloom     string        `json:"bloom"`
			Log       []interface{} `json:"log"`
		} `json:"transactionReceipt"`
	}
)

func (r *CallContractRequest) MarshalJSON() ([]byte, error) {
	params := []interface{}{
		utils.RemoveHexPrefix(r.To),
		utils.RemoveHexPrefix(r.Data),
		r.From,
	}
	if r.GasLimit != nil {
		// optional parameter, null will not work
		params = append(params, r.GasLimit)
	}
	/*
		1. "address" (string, required) The account address
		2. "data"    (string, required) The data hex string
		3. address   (string, optional) The sender address hex string
		4. gasLimit  (string, optional) The gas limit for executing the contract
	*/

	return json.Marshal(params)
}

// ========== FromHexAddress ============= //

type (
	FromHexAddressRequest  string
	FromHexAddressResponse string
)

func (r FromHexAddressRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal([]interface{}{
		string(r),
	})
}

// ========== GetHexAddress ============= //

type (
	GetHexAddressRequest  string
	GetHexAddressResponse string
)

func (r GetHexAddressRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal([]interface{}{
		string(r),
	})
}

// ========== DecodeRawTransaction ============= //
func (r DecodeRawTransactionRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal([]interface{}{
		string(r),
	})
}

type (
	DecodeRawTransactionRequest string

	/*
		{
		  "txid": "d0fe0caa1b798c36da37e9118a06a7d151632d670b82d1c7dc3985577a71880f",
		  "hash": "d0fe0caa1b798c36da37e9118a06a7d151632d670b82d1c7dc3985577a71880f",
		  "version": 2,
		  "size": 552,
		  "vsize": 552,
		  "locktime": 608,
		  "vin": [
		    {
		      "txid": "7f5350dc474f2953a3f30282c1afcad2fb61cdcea5bd949c808ecc6f64ce1503",
		      "vout": 0,
		      "scriptSig": {
		        "asm": "3045022100af4de764705dbd3c0c116d73fe0a2b78c3fab6822096ba2907cfdae2bb28784102206304340a6d260b364ef86d6b19f2b75c5e55b89fb2f93ea72c05e09ee037f60b[ALL] 03520b1500a400483f19b93c4cb277a2f29693ea9d6739daaf6ae6e971d29e3140",
		        "hex": "483045022100af4de764705dbd3c0c116d73fe0a2b78c3fab6822096ba2907cfdae2bb28784102206304340a6d260b364ef86d6b19f2b75c5e55b89fb2f93ea72c05e09ee037f60b012103520b1500a400483f19b93c4cb277a2f29693ea9d6739daaf6ae6e971d29e3140"
		      },
		      "sequence": 4294967294
		    }
		  ],
		  "vout": [
		    {
		      "value": 0,
		      "n": 0,
		      "scriptPubKey": {
		        "asm": "4 2500000 40 608060405234801561001057600080fd5b50604051602080610131833981016040525160005560fe806100336000396000f30060806040526004361060485763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166360fe47b18114604d5780636d4ce63c146064575b600080fd5b348015605857600080fd5b5060626004356088565b005b348015606f57600080fd5b50607660cc565b60408051918252519081900360200190f35b600054604080513381526020810192909252805183927f61ec51fdd1350b55fc6e153e60509e993f8dcb537fe4318c45a573243d96cab492908290030190a2600055565b600054905600a165627a7a723058200541c7c0da642ef9004daeb68d281a3c2341e765336f10b4a0ab45dbb7b7f83c00290000000000000000000000000000000000000000000000000000000000000064 OP_CREATE",
		        "hex": "010403a0252601284d5101608060405234801561001057600080fd5b50604051602080610131833981016040525160005560fe806100336000396000f30060806040526004361060485763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166360fe47b18114604d5780636d4ce63c146064575b600080fd5b348015605857600080fd5b5060626004356088565b005b348015606f57600080fd5b50607660cc565b60408051918252519081900360200190f35b600054604080513381526020810192909252805183927f61ec51fdd1350b55fc6e153e60509e993f8dcb537fe4318c45a573243d96cab492908290030190a2600055565b600054905600a165627a7a723058200541c7c0da642ef9004daeb68d281a3c2341e765336f10b4a0ab45dbb7b7f83c00290000000000000000000000000000000000000000000000000000000000000064c1",
		        "type": "create"
		      }
		    },
		    {
		      "value": 19996.59434,
		      "n": 1,
		      "scriptPubKey": {
		        "asm": "OP_DUP OP_HASH160 ce7137386121f7531f716d2d4ff36805bc65b3ec OP_EQUALVERIFY OP_CHECKSIG",
		        "hex": "76a914ce7137386121f7531f716d2d4ff36805bc65b3ec88ac",
		        "reqSigs": 1,
		        "type": "pubkeyhash",
		        "addresses": [
		          "qcNwyuvvPhiN4JVgwPp4QWPiK1p7YGvkf1"
		        ]
		      }
		    }
		  ]
		}
	*/
	DecodedRawTransactionResponse struct {
		ID       string                       `json:"txid"`
		Hash     string                       `json:"hash"`
		Size     int64                        `json:"size"`
		Vsize    int64                        `json:"vsize"`
		Version  int64                        `json:"version"`
		Locktime int64                        `json:"locktime"`
		Vins     []*DecodedRawTransactionInV  `json:"vin"`
		Vouts    []*DecodedRawTransactionOutV `json:"vout"`
	}
	DecodedRawTransactionInV struct {
		Address            string                         `json:"address"`
		TxID               string                         `json:"txid"`
		Vout               int64                          `json:"vout"`
		PreviousVoutPubkey DecodedRawTransactionScriptSig `json:"previousPubkey"`
		ScriptSig          DecodedRawTransactionScriptSig `json:"scriptSig"`
		Txinwitness        []string                       `json:"txinwitness"`
		Sequence           int64                          `json:"sequence"`
	}

	DecodedRawTransactionOutV struct {
		Value        Amount                            `json:"value"`
		N            int64                             `json:"n"`
		ScriptPubKey DecodedRawTransactionScriptPubKey `json:"scriptPubKey"`
	}

	// TODO: Make these two generic? Same struct is also present in other RPC data types
	DecodedRawTransactionScriptSig struct {
		ASM string `json:"asm"`
		Hex string `json:"hex"`
	}

	DecodedRawTransactionScriptPubKey struct {
		ASM       string   `json:"asm"`
		Hex       string   `json:"hex"`
		ReqSigs   int64    `json:"reqSigs"`
		Type      string   `json:"type"`
		Addresses []string `json:"addresses"`
	}
)

// Calculates transaction total amount of KAON
func (resp *DecodedRawTransactionResponse) CalcAmount() decimal.Decimal {
	var amount decimal.Decimal
	for _, out := range resp.Vouts {
		amount = amount.Add(out.Value.Decimal)
	}
	return amount
}

type ContractInfo struct {
	From      string
	To        string
	GasLimit  string
	GasPrice  string
	UserInput string
}

// TODO: complete
func (resp *DecodedRawTransactionResponse) ExtractContractInfo() (_ ContractInfo, isContractTx bool, _ error) {
	// Break if Vouts have several contracts

	var info *ContractInfo

	for _, vout := range resp.Vouts {
		var (
			script  = strings.Split(vout.ScriptPubKey.ASM, " ")
			finalOp = script[len(script)-1]
		)

		switch finalOp {
		case "OP_CALL":
			if info != nil {
				return ContractInfo{}, false, errors.New("Duplicate OP_CALL/OP_CREATE vouts")
			}
			callInfo, err := ParseCallSenderASM(script)
			// OP_CALL with OP_SENDER has the script type "nonstandard"
			if err != nil {
				// Check for OP_CALL without OP_SENDER
				callInfo, err = ParseCallASM(script)
				if err != nil {
					return ContractInfo{}, false, errors.WithMessage(err, "couldn't parse call sender ASM")
				}
			}
			info = &ContractInfo{
				From:      callInfo.From,
				To:        callInfo.To,
				GasLimit:  callInfo.GasLimit,
				GasPrice:  callInfo.GasPrice,
				UserInput: callInfo.CallData,
			}

			return *info, true, nil

		case "OP_CREATE":
			if info != nil {
				return ContractInfo{}, false, errors.New("Duplicate OP_CALL/OP_CREATE vouts")
			}
			// OP_CALL with OP_SENDER has the script type "create_sender"
			createInfo, err := ParseCreateSenderASM(script)
			if err != nil {
				// Check for OP_CREATE without OP_SENDER
				createInfo, err = ParseCreateASM(script)
				if err != nil {
					return ContractInfo{}, false, errors.WithMessage(err, "couldn't parse create sender ASM")
				}
			}
			info = &ContractInfo{
				From:      createInfo.From,
				To:        createInfo.To,
				GasLimit:  createInfo.GasLimit,
				GasPrice:  createInfo.GasPrice,
				UserInput: createInfo.CallData,
			}

			return *info, true, nil

		case "OP_SPEND":
			// TODO: complete
			return ContractInfo{}, true, errors.New("OP_SPEND contract parsing partially implemented")
		}
	}

	if info != nil {
		return *info, true, nil
	}

	// Let's try a standard trx parsing method
	for _, vin := range resp.Vins {
		fromAddress := ""
		if vin.Address == "" {
			var (
				script  = strings.Split(vin.PreviousVoutPubkey.ASM, " ")
				finalOp = script[len(script)-1]
			)
			if finalOp != "OP_CHECKSIG" {
				continue // We do not support non-standard money transfers
			}
			sender, err := ParseP2PKHReciever(script)
			if err != nil {
				continue
			}

			// P2Sh address(such as MUrenj2sPqEVTiNbHQ2RARiZYyTAAeKiDX) and BECH32 address (such as qc1qkt33x6hkrrlwlr6v59wptwy6zskyrjfe40y0lx)
			// will cause ConvertKaonAddress to fall
			// but we not support them for now
			fromAddressIn, err := utils.ConvertKaonAddress(*sender)

			if err != nil {
				continue
			}
			fromAddress = fromAddressIn
		} else {

			// P2Sh address(such as MUrenj2sPqEVTiNbHQ2RARiZYyTAAeKiDX) and BECH32 address (such as qc1qkt33x6hkrrlwlr6v59wptwy6zskyrjfe40y0lx)
			// will cause ConvertKaonAddress to fall
			// but we not support them for now
			fromAddressIn, err := utils.ConvertKaonAddress(vin.Address)

			if err != nil {
				continue
			}
			fromAddress = fromAddressIn
		}

		if int(vin.Vout) >= len(resp.Vouts) {
			continue

			// TODO: probably we still may be able to solve it in that case
			// research
			// return ContractInfo{
			// 	From: fromAddress,
			// }, false, nil
		}

		var (
			vout     = resp.Vouts[vin.Vout]
			reciever = ""
		)
		for _, address := range vout.ScriptPubKey.Addresses {
			if address != "" {
				// P2Sh address(such as MUrenj2sPqEVTiNbHQ2RARiZYyTAAeKiDX) and BECH32 address (such as qc1qkt33x6hkrrlwlr6v59wptwy6zskyrjfe40y0lx)
				// will cause ConvertKaonAddress to fall
				// but we not support them for now
				hex, err := utils.ConvertKaonAddress(address)
				if err == nil {
					reciever = hex
					break
				}
			}
		}

		if reciever == "" {
			var (
				script  = strings.Split(vout.ScriptPubKey.ASM, " ")
				finalOp = script[len(script)-1]
			)
			if finalOp != "OP_CHECKSIG" {
				continue // We do not support non-standard money transfers
			}
			recieverParsed, err := ParseP2PKHReciever(script)
			if err != nil {
				continue
			}

			// P2Sh address(such as MUrenj2sPqEVTiNbHQ2RARiZYyTAAeKiDX) and BECH32 address (such as qc1qkt33x6hkrrlwlr6v59wptwy6zskyrjfe40y0lx)
			// will cause ConvertKaonAddress to fall
			// but we not support them for now
			recieverIn, err := utils.ConvertKaonAddress(*recieverParsed)

			if err != nil {
				continue
			}
			reciever = recieverIn
		}

		return ContractInfo{
			From: fromAddress,
			To:   reciever,
		}, false, nil
	}

	return ContractInfo{}, false, nil
}

func (resp *DecodedRawTransactionResponse) IsContractCreation() bool {
	for _, vout := range resp.Vouts {
		if strings.HasSuffix(vout.ScriptPubKey.ASM, "OP_CREATE") {
			return true
		}
	}
	return false
}

// ========== GetTransactionOut ============= //
type (
	// GetTransactionOutRequest struct {
	// 	Hash            string `json:"txid"`
	// 	VoutNumber      int    `json:"n"`
	// 	MempoolIncluded bool   `json:"include_mempool"`
	// }
	GetTransactionOutRequest []interface{}
	// TODO: Make ScriptPubKey into a separate struct (or use generic variant?) for ease of use?
	GetTransactionOutResponse struct {
		BestBlockHash    string `json:"bestblock"`
		ConfirmationsNum int    `json:"confirmations"`
		Amount           Amount `json:"value"`
		ScriptPubKey     struct {
			ASM        string   `json:"asm"`
			Hex        string   `json:"hex"`
			ReqSigsNum int      `json:"reqSigs"`
			Type       string   `json:"type"`
			Addresses  []string `json:"addresses"`
		} `json:"scriptPubKey"`
		IsReward    bool `json:"coinbase"`
		IsCoinstake bool `json:"coinstake"`
	}
)

// ========== GetTransactionReceipt ============= //
type (
	GetTransactionReceiptRequest  string
	GetTransactionReceiptResponse TransactionReceipt
	/*
	   {
	     "blockHash": "975326b65c20d0b8500f00a59f76b08a98513fff7ce0484382534a47b55f8985",
	     "blockNumber": 4063,
	     "transactionHash": "c1816e5fbdd4d1cc62394be83c7c7130ccd2aadefcd91e789c1a0b33ec093fef",
	     "transactionIndex": 2,
	     "from": "6b22910b1e302cf74803ffd1691c2ecb858d3712",
	     "to": "db46f738bf32cdafb9a4a70eb8b44c76646bcaf0",
	     "cumulativeGasUsed": 68572,
	     "gasUsed": 68572,
	     "contractAddress": "db46f738bf32cdafb9a4a70eb8b44c76646bcaf0",
	     "excepted": "None",
	     "log": [
	       {
	         "address": "db46f738bf32cdafb9a4a70eb8b44c76646bcaf0",
	         "topics": [
	           "0f6798a560793a54c3bcfe86a93cde1e73087d944c0ea20544137d4121396885",
	           "0000000000000000000000006b22910b1e302cf74803ffd1691c2ecb858d3712"
	         ],
	         "data": "0000000000000000000000000000000000000000000000000000000000000001"
	       },
	       {
	         "address": "db46f738bf32cdafb9a4a70eb8b44c76646bcaf0",
	         "topics": [
	           "ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
	           "0000000000000000000000000000000000000000000000000000000000000000",
	           "0000000000000000000000006b22910b1e302cf74803ffd1691c2ecb858d3712"
	         ],
	         "data": "0000000000000000000000000000000000000000000000000000000000000001"
	       }
	     ]
	   }
	*/
	TransactionReceipt struct {
		BlockHash         string  `json:"blockHash"`
		BlockNumber       uint64  `json:"blockNumber"`
		TransactionHash   string  `json:"transactionHash"`
		TransactionIndex  uint64  `json:"transactionIndex"`
		From              string  `json:"from"`
		To                string  `json:"to"`
		CumulativeGasUsed big.Int `json:"cumulativeGasUsed"`
		GasUsed           big.Int `json:"gasUsed"`
		EffectiveGasPrice big.Int `json:"effectiveGasPrice"`
		// NOTE: will be null for a contract creation transaction

		// TODO: discuss
		// 	? May be a contract transaction created by non-contract
		//
		// The created contract address. If this tx is created by the contract,
		// return the contract address, else return null
		ContractAddress string `json:"contractAddress"`

		// May has "None" value, which means, that transaction is not executed
		Excepted string `json:"excepted"`

		Log         []Log `json:"log"`
		OutputIndex int64 `json:"outputIndex"`
	}
)

func (r TransactionReceipt) GetTransactionHash() string {
	return r.TransactionHash
}

func (r TransactionReceipt) GetTransactionIndex() uint64 {
	return r.TransactionIndex
}

func (r TransactionReceipt) GetBlockHash() string {
	return r.BlockHash
}

func (r TransactionReceipt) GetBlockNumber() uint64 {
	return r.BlockNumber
}

func (r GetTransactionReceiptRequest) MarshalJSON() ([]byte, error) {
	/*
		1. "hash"          (string, required) The transaction hash
	*/
	return json.Marshal([]interface{}{
		string(r),
	})
}

var EmptyResponseErr = errors.New("result is empty")

func (resp *GetTransactionReceiptResponse) UnmarshalJSON(data []byte) error {
	// NOTE: do not use `GetTransactionReceiptResponse`, 'cause
	// it may violate to infinite loop, while calling
	// UnmarshalJSON interface
	var receipts []TransactionReceipt
	if err := json.Unmarshal(data, &receipts); err != nil {
		return err
	}
	if receiptsNum := len(receipts); receiptsNum != 1 {
		return EmptyResponseErr
	}
	*resp = GetTransactionReceiptResponse(receipts[0])
	return nil
}

// ========== GetBlockCount ============= //

type (
	GetBlockCountResponse struct {
		*big.Int
	}
)

func (r *GetBlockCountResponse) UnmarshalJSON(data []byte) error {
	var i *big.Int
	err := json.Unmarshal(data, &i)
	if err != nil {
		return err
	}

	r.Int = i
	return nil
}

// ========== GetGasPrice ============= //

type (
	GetGasPricetResponse struct {
		*big.Int
	}
)

func (r *GetGasPricetResponse) UnmarshalJSON(data []byte) error {
	var i *big.Int
	err := json.Unmarshal(data, &i)
	if err != nil {
		return err
	}

	r.Int = i
	return nil
}

// ========== GetHashrate & GetMining ============= //

type (
	//Switch things up to use Staking infor only
	//Pass the reponse to their respective calls
	GetHashrateResponse StakingInfo
	GetMiningResponse   StakingInfo

	StakingInfo struct {
		Enabled        bool     `json:"enabled"`
		Staking        bool     `json:"staking"`
		Errors         string   `json:"errors"`
		CurrentBlockTx *big.Int `json:"currentblocktx"`
		PooledTx       *big.Int `json:"pooledtx"`
		Difficulty     float64  `json:"difficulty"`
		SearchInterval *big.Int `json:"search-interval"`
		Weight         *big.Int `json:"weight"`
		NetSakeWeight  *big.Int `json:"netstakeweight"`
		ExpectedTime   *big.Int `json:"expectedtime"`
	}
)

func (resp *GetHashrateResponse) UnmarshalJSON(data []byte) error {

	var stakingInfo StakingInfo
	if err := json.Unmarshal(data, &stakingInfo); err != nil {
		return err
	}

	resp.Difficulty = stakingInfo.Difficulty
	return nil
}

func (resp *GetMiningResponse) UnmarshalJSON(data []byte) error {

	var stakingInfo StakingInfo
	if err := json.Unmarshal(data, &stakingInfo); err != nil {
		return err
	}

	resp.Staking = stakingInfo.Staking
	return nil
}

// ========== GetRawTransaction ============= //

type (
	GetRawTransactionRequest struct {
		TxID    string
		Verbose bool
	}
	GetRawTransactionResponse struct {
		Hex       string `json:"hex"`
		ID        string `json:"txid"`
		Hash      string `json:"hash"`
		Size      int64  `json:"size"`
		Vsize     int64  `json:"vsize"`
		Version   int64  `json:"version"`
		Weight    int64  `json:"weight"`
		Generated bool   `json:"generated"`

		BlockHash        string `json:"blockhash"`
		BlockIndex       int    `json:"blockindex"`
		TransactionIndex int    `json:"index"`
		Confirmations    int64  `json:"confirmations"`
		Time             int64  `json:"time"`
		BlockTime        int64  `json:"blocktime"`

		OP_SENDER string `json:"OP_SENDER"`

		Vins  []RawTransactionVin  `json:"vin"`
		Vouts []RawTransactionVout `json:"vout"`

		GasPrice big.Int `json:"gasPrice"`
		Locktime int64   `json:"locktime"`

		// Unused fields:
		// - "in_active_chain"

	}
	RawTransactionVin struct {
		ID      string `json:"txid"`
		VoutN   int64  `json:"vout"`
		Amount  Amount `json:"value"`
		Address string `json:"address"`
		// TODO: temporary solution
		ScriptSig DecodedRawTransactionScriptSig `json:"scriptSig"`

		// Additional fields:
		// - "sequence"
		// - "txinwitness"
	}
	// TODO: Make details into a separate struct (or use generic scriptPubKey?) for ease of use?
	RawTransactionVout struct {
		Amount  Amount                    `json:"value"`
		Details RawTransactionVoutDetails `json:"scriptPubKey"`

		// Additional fields:
		// - "n"
	}

	RawTransactionVoutDetails struct {
		Address   string   `json:"address"`
		Addresses []string `json:"addresses"`
		ASM       string   `json:"asm"`
		Hex       string   `json:"hex"`
		// ReqSigs   interface{} `json:"reqSigs"`
		Type string `json:"type"`
	}
)

func (d *RawTransactionVoutDetails) GetAddresses() []string {
	if len(d.Address) != 0 {
		return []string{d.Address}
	}
	return d.Addresses[:]
}

func (r *GetRawTransactionRequest) MarshalJSON() ([]byte, error) {
	/*
		1. "txid"      (string, required) The transaction id
		2. verbose     (bool, optional, default=false) If false, return a string, otherwise return a json object
		3. "blockhash" (string, optional) The block in which to look for the transaction
	*/
	return json.Marshal([]interface{}{
		r.TxID,
		r.Verbose,
	})
}

// Calculates transaction total amount of KAON
func (resp *GetRawTransactionResponse) CalcAmount() decimal.Decimal {
	var amount decimal.Decimal
	for _, out := range resp.Vouts {
		amount = amount.Add(out.Amount.Decimal)
	}
	return amount
}

func (r *GetRawTransactionResponse) IsPending() bool {
	return r.BlockHash == ""
}

func (r *GetRawTransactionResponse) GetMiningFeeInKAON() (*big.Int, error) {
	var fee decimal.Decimal

	// when sending KAON, the first vout will be the target
	// the second will be change from the vin, it will be returned to the same account
	if len(r.Vouts) > 0 {

		var valueIn decimal.Decimal
		var valueOut decimal.Decimal

		for _, vin := range r.Vins {
			valueIn = vin.Amount.Add(valueIn)
		}

		for _, vout := range r.Vouts {
			valueOut = vout.Amount.Add(valueOut)

		}

		if valueIn.Cmp(decimal.Zero) == 0 { // generated
			fee = decimal.Zero
		} else {

			fee = valueIn.Sub(valueOut)
		}
	} else {
		fee = decimal.Zero
	}

	feeBigInt, errEncode := utils.ToBigInt(&fee)
	if errEncode != nil {
		return nil, errEncode
	}

	return feeBigInt, nil
}

// ========== GetTransactionHashByEthHash ============= //

type (
	GetTransactionHashByEthHashRequest struct {
		TxID string
	}

	/*
		KAON trx hash
	*/
	GetTransactionHashByEthHashResponse string
)

func (r GetTransactionHashByEthHashRequest) MarshalJSON() ([]byte, error) {
	/*
		1. "TxID"          (string, required) The transaction hash
	*/
	return json.Marshal([]interface{}{
		string(r.TxID),
	})
}

// ========== GetTransaction ============= //

type (
	GetTransactionRequest struct {
		TxID string
	}

	/*
		{
		    "amount": 0,
		    "fee": -0.2012,
		    "confirmations": 2,
		    "blockhash": "ea26fd59a2145dcecd0e2f81b701019b51ca754b6c782114825798973d8187d6",
		    "blockindex": 2,
		    "blocktime": 1533092896,
		    "txid": "11e97fa5877c5df349934bafc02da6218038a427e8ed081f048626fa6eb523f5",
		    "walletconflicts": [],
		    "time": 1533092879,
		    "timereceived": 1533092879,
		    "bip125-replaceable": "no",
		    "details": [
		      {
		        "account": "",
		        "category": "send",
		        "amount": 0,
		        "vout": 0,
		        "fee": -0.2012,
		        "abandoned": false
		      }
		    ],
		    "hex": "020000000159c0514feea50f915854d9ec45bc6458bb14419c78b17e7be3f7fd5f563475b5010000006a473044022072d64a1f4ea2d54b7b05050fc853ab192c91cc5ca17e23007867f92f2ab59d9202202b8c9ab9348c8edbb3b98b1788382c8f37642ec9bd6a4429817ab79927319200012103520b1500a400483f19b93c4cb277a2f29693ea9d6739daaf6ae6e971d29e3140feffffff02000000000000000063010403400d0301644440c10f190000000000000000000000006b22910b1e302cf74803ffd1691c2ecb858d3712000000000000000000000000000000000000000000000000000000000000000a14be528c8378ff082e4ba43cb1baa363dbf3f577bfc260e66272970100001976a9146b22910b1e302cf74803ffd1691c2ecb858d371288acb00f0000"
		  }
	*/
	GetTransactionResponse struct {
		Amount            Amount               `json:"amount"`
		Fee               Amount               `json:"fee"`
		Confirmations     int64                `json:"confirmations"`
		BlockHash         string               `json:"blockhash"`
		BlockIndex        int64                `json:"blockindex"`
		TransactionIndex  int64                `json:"index"`
		BlockTime         int64                `json:"blocktime"`
		ID                string               `json:"txid"`
		Time              int64                `json:"time"`
		ReceivedAt        int64                `json:"timereceived"`
		Bip125Replaceable string               `json:"bip125-replaceable"`
		Details           []*TransactionDetail `json:"details"`
		Hex               string               `json:"hex"`
		Generated         bool                 `json:"generated"`
		GasPrice          big.Int              `json:"gasPrice"`
	}
	TransactionDetail struct {
		// TODO: research/discuss
		Address string `json:"address"`
		// Represents transaction direction: `send` or `receive`
		Category string `json:"category"`
		Amount   Amount `json:"amount"`
		// Comment value
		Label string `json:"label"`
		Vout  int64  `json:"vout"`
		// NOTE:
		// 	- Negative value
		// 	- Presetned only for `send` transaction category
		Fee Amount `json:"fee"`
		// TODO: discuss
		// 	? What's the meaning
		//
		// NOTE: presetned only for `send` transaction category
		Abandoned bool `json:"abandoned"`
	}
)

func (r *GetTransactionRequest) MarshalJSON() ([]byte, error) {
	/*
		1. "txid"                  (string, required) The transaction id
		2. "include_watchonly"     (bool, optional, default=false) Whether to include watch-only addresses in balance calculation and details[]
		3. "waitconf"              (int, optional, default=0) Wait for enough confirmations before returning
	*/
	return json.Marshal([]interface{}{
		r.TxID,
	})
}

func (r *GetTransactionResponse) UnmarshalJSON(data []byte) error {
	if string(data) == "[]" {
		return EmptyResponseErr
	}
	type Response GetTransactionResponse
	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}

	*r = GetTransactionResponse(resp)

	return nil
}

func (r *GetTransactionResponse) IsPending() bool {
	return r.BlockHash == ""
}

// ========== SearchLogs ============= //

type (
	SearchLogsRequest struct {
		FromBlock            *big.Int
		ToBlock              *big.Int
		Addresses            []string
		Topics               []SearchLogsTopic
		MinimumConfirmations *uint
	}

	SearchLogsResponse []TransactionReceipt
	SearchLogsTopic    []string
)

func NewSearchLogsTopics(topics [][]string) []SearchLogsTopic {
	result := make([]SearchLogsTopic, len(topics))

	for i, topic := range topics {
		result[i] = NewSearchLogsTopic(topic)
	}

	return result
}

func NewSearchLogsTopic(topics []string) SearchLogsTopic {
	result := SearchLogsTopic{}

	for _, topic := range topics {
		result = append(result, topic)
	}

	return result
}

func (t SearchLogsTopic) MarshalJSON() ([]byte, error) {
	if len(t) == 1 {
		return []byte(`"` + t[0] + `"`), nil
	}

	return []byte("null"), nil
}

func (r *SearchLogsRequest) MarshalJSON() ([]byte, error) {
	/*
		1. "fromBlock"        (numeric, required) The number of the earliest block (latest may be given to mean the most recent block).
		2. "toBlock"          (string, required) The number of the latest block (-1 may be given to mean the most recent block).
		3. "address"          (string, optional) An address or a list of addresses to only get logs from particular account(s).
		4. "topics"           (string, optional) An array of values from which at least one must appear in the log entries. The order is important, if you want to leave topics out use null, e.g. ["null", "0x00..."].
		5. "minconf"          (uint, optional, default=0) Minimal number of confirmations before a log is returned
	*/
	var addresses interface{}
	if r.Addresses != nil && len(r.Addresses) != 0 {
		addresses = map[string][]string{
			"addresses": r.Addresses,
		}
	}

	var topics interface{}
	if len(r.Topics) > 0 {
		// if all topics are null, filter them all out
		nullCount := 0
		for _, topic := range r.Topics {
			byts, err := json.Marshal(topic)
			if err != nil {
				return []byte{}, err
			}

			if string(byts) == "null" {
				nullCount++
			}
		}

		if nullCount != len(r.Topics) {
			topics = map[string][]SearchLogsTopic{
				"topics": r.Topics,
			}
		}
	}

	data := []interface{}{
		r.FromBlock,
		r.ToBlock,
		// should be null if not specified
		addresses,
		// should be null if not specified
		topics,
	}

	if r.MinimumConfirmations != nil {
		data = append(data, r.MinimumConfirmations)
	}

	return json.Marshal(data)
}

// ========== GetAccountInfo ============= //

type (
	// the account address
	GetAccountInfoRequest string

	/*
		{
			"address": "1adf95f5c60cdc0dfd99c3d2857cd01419be521c",
			"balance": 0,
			"storage": {
				"8a35acfbc15ff81a39ae7d344fd709f28e8600b4aa8c65c6b64bfe7fe36bd19b": {
					"0000000000000000000000000000000000000000000000000000000000000004": "000000000000000000000000000000000000000000000000000000000000000a"
				},
				"c2575a0e9e593c00f959f8c92f12db2869c3395a3b0502d05e2516446f71f85b": {
					"0000000000000000000000000000000000000000000000000000000000000003": "0000000000000000000000001CE507204a6fC8fd6aA7e54D1481d30ACB0Dbead"
				}
			},
			"code": "0x..."
		}
	*/
	GetAccountInfoResponse struct {
		Address string          `json:"address"`
		Balance big.Int         `json:"balance"`
		Storage json.RawMessage `json:"storage"`
		Code    string          `json:"code"`
	}
)

func (r *GetAccountInfoRequest) MarshalJSON() ([]byte, error) {
	/*
		1. "address"          (string, required) The account address
	*/
	return json.Marshal([]interface{}{
		string(*r),
	})
}

// ========== GetAddressByAccount ============= //

type (
	// the account name
	GetAddressesByAccountRequest string

	/*
		[
			"ar2SzdHghSgeacypPn7zfDe3qfKAEwimus"
		]
	*/
	GetAddressesByAccountResponse []string
)

func (r *GetAddressesByAccountRequest) MarshalJSON() ([]byte, error) {
	/*
		1. "account"        (string, required) The account name.
	*/
	return json.Marshal([]interface{}{
		string(*r),
	})
}

// ========== GetBlockHash ============= //
type (
	GetBlockHashRequest struct {
		*big.Int
	}
	GetBlockHashResponse string
)

func (r *GetBlockHashRequest) MarshalJSON() ([]byte, error) {
	/*
		1. height         (numeric, required) The height index
	*/
	return json.Marshal([]interface{}{
		r.Uint64(),
	})
}

// ========== Generate ============= //
type (
	GenerateRequest struct {
		BlockNum int
		Address  string
		MaxTries *int
	}
	GenerateResponse []string
)

func (r *GenerateRequest) MarshalJSON() ([]byte, error) {
	/*
		1. nblocks      (numeric, required) How many blocks are generated immediately.
		2. maxtries     (numeric, optional) How many iterations to try (default = 1000000).
	*/
	params := []interface{}{
		r.BlockNum,
		r.Address,
	}

	if r.MaxTries != nil {
		params = append(params, r.MaxTries)
	}

	return json.Marshal(params)
}

// ========== GetBlockHeader ============= //
type (
	GetBlockHeaderRequest struct {
		Hash       string
		NotVerbose bool
	}

	/*
		{
		  "hash": "bba11e1bacc69ba535d478cf1f2e542da3735a517b0b8eebaf7e6bb25eeb48c5",
		  "confirmations": 1,
		  "height": 3983,
		  "version": 536870912,
		  "versionHex": "20000000",
		  "merkleroot": "0b5f03dc9d456c63c587cc554b70c1232449be43d1df62bc25a493b04de90334",
		  "time": 1536551888,
		  "mediantime": 1536551728,
		  "nonce": 0,
		  "bits": "207fffff",
		  "difficulty": 4.656542373906925e-10,
		  "chainwork": "0000000000000000000000000000000000000000000000000000000000001f20",
		  "hashStateRoot": "3e49216e58f1ad9e6823b5095dc532f0a6cc44943d36ff4a7b1aa474e172d672",
		  "hashUTXORoot": "130a3e712d9f8b06b83f5ebf02b27542fb682cdff3ce1af1c17b804729d88a47",
		  "previousblockhash": "6d7d56af09383301e1bb32a97d4a5c0661d62302c06a778487d919b7115543be",
		  "flags": "proof-of-stake",
		  "proofhash": "15bd6006ecbab06708f705ecf68664b78b388e4d51416cdafb019d5b90239877",
		  "modifier": "a79c00d1d570743ca8135a173d535258026d26bafbc5f3d951c3d33486b1f120",
		  "gasUsed": 2698989
		}
	*/
	GetBlockHeaderResponse struct {
		Hash              string  `json:"hash"`
		Confirmations     int     `json:"confirmations"`
		Height            int     `json:"height"`
		Version           int     `json:"version"`
		VersionHex        string  `json:"versionHex"`
		Merkleroot        string  `json:"merkleroot"`
		Time              uint64  `json:"time"`
		Mediantime        int     `json:"mediantime"`
		Nonce             int     `json:"nonce"`
		Bits              string  `json:"bits"`
		Difficulty        float64 `json:"difficulty"`
		Chainwork         string  `json:"chainwork"`
		HashStateRoot     string  `json:"hashStateRoot"`
		HashUTXORoot      string  `json:"hashUTXORoot"`
		Previousblockhash string  `json:"previousblockhash"`
		Flags             string  `json:"flags"`
		Proofhash         string  `json:"proofhash"`
		Modifier          string  `json:"modifier"`
		Proposer          string  `json:"proposer"`
		GasUsed           big.Int `json:"gasUsed"`
	}
)

func (r *GetBlockHeaderRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal([]interface{}{
		r.Hash,
		!r.NotVerbose,
	})
}

func (r *GetBlockHeaderResponse) IsGenesisBlock() bool {
	return r.Height == genesisBlockHeight
}

// ========== GetBlock ============= //
type (
	GetBlockRequest struct {
		Hash      string
		Verbosity bool
	}

	/*
		{
		  "hash": "bba11e1bacc69ba535d478cf1f2e542da3735a517b0b8eebaf7e6bb25eeb48c5",
		  "confirmations": 57,
		  "strippedsize": 584,
		  "size": 620,
		  "weight": 2372,
		  "height": 3983,
		  "version": 536870912,
		  "versionHex": "20000000",
		  "merkleroot": "0b5f03dc9d456c63c587cc554b70c1232449be43d1df62bc25a493b04de90334",
		  "hashStateRoot": "3e49216e58f1ad9e6823b5095dc532f0a6cc44943d36ff4a7b1aa474e172d672",
		  "hashUTXORoot": "130a3e712d9f8b06b83f5ebf02b27542fb682cdff3ce1af1c17b804729d88a47",

		  "tx": [
		    "3208dc44733cbfa11654ad5651305428de473ef1e61a1ec07b0c1a5f4843be91",
		    "8fcd819194cce6a8454b2bec334d3448df4f097e9cdc36707bfd569900268950"
		  ],
		  ... OR
		  "tx": [
			{
			"txid": "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			"hash": "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			"version": 1,
			"size": 234,
			"vsize": 123,
			"weight": 567,
			"locktime": 0,
			"vin": [
				{
				"coinbase": "031234567890abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				"sequence": 4294967295
				}
			],
			"vout": [
					{
					"value": 12.345,
					"n": 0,
					"scriptPubKey": ...
					// ... additional vout transactions if available
					}
			],
			// ... additional details of the transaction if available
			}
			// ... additional transactions in the block if available
		  ],

		  "time": 1536551888,
		  "mediantime": 1536551728,
		  "nonce": 0,
		  "bits": "207fffff",
		  "difficulty": 4.656542373906925e-10,
		  "chainwork": "0000000000000000000000000000000000000000000000000000000000001f20",
		  "previousblockhash": "6d7d56af09383301e1bb32a97d4a5c0661d62302c06a778487d919b7115543be",
		  "nextblockhash": "d7758774cfdd6bab7774aa891ae035f1dc5a2ff44240784b5e7bdfd43a7a6ec1",
		  "flags": "proof-of-stake",
		  "proofhash": "15bd6006ecbab06708f705ecf68664b78b388e4d51416cdafb019d5b90239877",
		  "modifier": "a79c00d1d570743ca8135a173d535258026d26bafbc5f3d951c3d33486b1f120",
		  "signature": "3045022100a6ab6c2b14b1f73e734f1a61d4d22385748e48836492723a6ab37cdf38525aba022014a51ecb9e51f5a7a851641683541fec6f8f20205d0db49e50b2a4e5daed69d2"
		}
	*/
	GetBlockResponse struct {
		Hash              string        `json:"hash"`
		Confirmations     int           `json:"confirmations"`
		Strippedsize      int           `json:"strippedsize"`
		Size              int           `json:"size"`
		Weight            int           `json:"weight"`
		Height            int           `json:"height"`
		Version           int           `json:"version"`
		VersionHex        string        `json:"versionHex"`
		Merkleroot        string        `json:"merkleroot"`
		HashStateRoot     string        `json:"hashStateRoot"`
		HashUTXORoot      string        `json:"hashUTXORoot"`
		Txs               []interface{} `json:"tx"`
		Time              int           `json:"time"`
		Mediantime        int           `json:"mediantime"`
		Nonce             int           `json:"nonce"`
		Bits              string        `json:"bits"`
		Difficulty        float64       `json:"difficulty"`
		Chainwork         string        `json:"chainwork"`
		Previousblockhash string        `json:"previousblockhash"`
		Proposer          string        `json:"Proposer"`
		Nextblockhash     string        `json:"nextblockhash"`
		Flags             string        `json:"flags"`
		Proofhash         string        `json:"proofhash"`
		Modifier          string        `json:"modifier"`
		Signature         string        `json:"signature"`
	}

	BlockTransactionDetails struct {
		Hex       string `json:"hex"`
		ID        string `json:"txid"`
		Hash      string `json:"hash"`
		Size      int64  `json:"size"`
		Vsize     int64  `json:"vsize"`
		Version   int64  `json:"version"`
		Weight    int64  `json:"weight"`
		Generated bool   `json:"generated"`

		BlockHash        string `json:"blockhash"`
		BlockIndex       int    `json:"blockindex"`
		TransactionIndex int    `json:"index"`
		Confirmations    int64  `json:"confirmations"`
		Time             int64  `json:"time"`
		BlockTime        int64  `json:"blocktime"`

		OP_SENDER string `json:"OP_SENDER"`

		Vins  []BlockTransactionVin  `json:"vin"`
		Vouts []BlockTransactionVout `json:"vout"`

		GasPrice big.Int `json:"gasPrice"`
		Locktime int64   `json:"locktime"`

		// Unused fields:
		// - "in_active_chain"

	}

	BlockTransactionVin struct {
		Address            string                    `json:"address"`
		TxID               string                    `json:"txid"`
		Vout               int64                     `json:"vout"`
		PreviousVoutPubkey BlockTransactionScriptSig `json:"previousPubkey"`
		ScriptSig          BlockTransactionScriptSig `json:"scriptSig"`
		Txinwitness        []string                  `json:"txinwitness"`
		Sequence           int64                     `json:"sequence"`
		Value              Amount                    `json:"value"`
	}

	BlockTransactionVout struct {
		Value        Amount                       `json:"value"`
		N            int64                        `json:"n"`
		ScriptPubKey BlockTransactionScriptPubKey `json:"scriptPubKey"`
	}

	// TODO: Make these two generic? Same struct is also present in other RPC data types
	BlockTransactionScriptSig struct {
		ASM string `json:"asm"`
		Hex string `json:"hex"`
	}

	BlockTransactionScriptPubKey struct {
		ASM       string   `json:"asm"`
		Hex       string   `json:"hex"`
		ReqSigs   int64    `json:"reqSigs"`
		Type      string   `json:"type"`
		Addresses []string `json:"addresses"`
	}
)

func (r *GetBlockRequest) MarshalJSON() ([]byte, error) {

	return json.Marshal([]interface{}{
		r.Hash,
		r.Verbosity,
	})
}

func (r *BlockTransactionDetails) UnmarshalJSON(data []byte) error {
	if string(data) == "[]" {
		return EmptyResponseErr
	}
	type Response BlockTransactionDetails
	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}

	*r = BlockTransactionDetails(resp)

	return nil
}

func (r *BlockTransactionDetails) IsPending() bool {
	return r.BlockHash == ""
}

func (r *BlockTransactionDetails) GetMiningFeeInKAON() (*big.Int, error) {
	if r.Generated {
		return big.NewInt(0), nil
	}

	var fee decimal.Decimal

	// when sending KAON, the first vout will be the target
	// the second will be change from the vin, it will be returned to the same account
	if len(r.Vouts) > 0 {

		var valueIn decimal.Decimal
		var valueOut decimal.Decimal

		for _, vin := range r.Vins {
			valueIn = vin.Value.Add(valueIn)
		}

		for _, vout := range r.Vouts {
			valueOut = vout.Value.Add(valueOut)

		}
		if valueIn.Cmp(decimal.Zero) == 0 { // generated
			fee = decimal.Zero
		} else {

			fee = valueIn.Sub(valueOut)
		}
	} else {
		fee = decimal.Zero
	}

	feeBigInt, errEncode := utils.ToBigInt(&fee)
	if errEncode != nil {
		return nil, errEncode
	}

	return feeBigInt, nil
}

// Calculates transaction total amount of KAON
func (resp *BlockTransactionDetails) CalcAmount() decimal.Decimal {
	var amount decimal.Decimal
	for _, out := range resp.Vouts {
		amount = amount.Add(out.Value.Decimal)
	}
	return amount
}

func (resp *BlockTransactionDetails) ExtractContractInfo() (_ ContractInfo, isContractTx bool, _ error) {
	// Break if Vouts have several contracts

	var info *ContractInfo

	for _, vout := range resp.Vouts {
		var (
			script  = strings.Split(vout.ScriptPubKey.ASM, " ")
			finalOp = script[len(script)-1]
		)

		switch finalOp {
		case "OP_CALL":
			if info != nil {
				return ContractInfo{}, false, errors.New("Duplicate OP_CALL/OP_CREATE vouts")
			}
			callInfo, err := ParseCallSenderASM(script)
			// OP_CALL with OP_SENDER has the script type "nonstandard"
			if err != nil {
				// Check for OP_CALL without OP_SENDER
				callInfo, err = ParseCallASM(script)
				if err != nil {
					return ContractInfo{}, false, errors.WithMessage(err, "couldn't parse call sender ASM")
				}
			}
			info = &ContractInfo{
				From:      callInfo.From,
				To:        callInfo.To,
				GasLimit:  callInfo.GasLimit,
				GasPrice:  callInfo.GasPrice,
				UserInput: callInfo.CallData,
			}

			return *info, true, nil

		case "OP_CREATE":
			if info != nil {
				return ContractInfo{}, false, errors.New("Duplicate OP_CALL/OP_CREATE vouts")
			}
			// OP_CALL with OP_SENDER has the script type "create_sender"
			createInfo, err := ParseCreateSenderASM(script)
			if err != nil {
				// Check for OP_CREATE without OP_SENDER
				createInfo, err = ParseCreateASM(script)
				if err != nil {
					return ContractInfo{}, false, errors.WithMessage(err, "couldn't parse create sender ASM")
				}
			}
			info = &ContractInfo{
				From:      createInfo.From,
				To:        createInfo.To,
				GasLimit:  createInfo.GasLimit,
				GasPrice:  createInfo.GasPrice,
				UserInput: createInfo.CallData,
			}

			return *info, true, nil

		case "OP_SPEND":
			// TODO: complete
			return ContractInfo{}, true, errors.New("OP_SPEND contract parsing partially implemented")

		case "OP_CHECKSIG":
			if !resp.Generated { // generated trxes does not have valuable vins data so process what do we have
				continue
			}

			if vout.ScriptPubKey.Type != "pubkey" && vout.ScriptPubKey.Type != "pubkeyhash" {
				continue // do not support other types
			}

			var (
				reciever = ""
			)
			for _, address := range vout.ScriptPubKey.Addresses {
				if address != "" {
					// P2Sh address(such as MUrenj2sPqEVTiNbHQ2RARiZYyTAAeKiDX) and BECH32 address (such as qc1qkt33x6hkrrlwlr6v59wptwy6zskyrjfe40y0lx)
					// will cause ConvertKaonAddress to fall
					// but we not support them for now
					hex, err := utils.ConvertKaonAddress(address)
					if err == nil {
						reciever = hex
						break
					}
				}
			}

			return ContractInfo{
				From: "",
				To:   reciever,
			}, false, nil
		}

	}

	if info != nil {
		return *info, true, nil
	}

	if resp.Generated {
		return ContractInfo{}, false, nil
	}

	// Let's try a standard trx parsing method
	for i, vin := range resp.Vins {
		fromAddress := ""
		if vin.Address == "" {
			var (
				script  = strings.Split(vin.PreviousVoutPubkey.ASM, " ")
				finalOp = script[len(script)-1]
			)
			if finalOp != "OP_CHECKSIG" {
				continue // We do not support non-standard money transfers
			}
			sender, err := ParseP2PKHReciever(script)
			if err != nil {
				continue
			}

			// P2Sh address(such as MUrenj2sPqEVTiNbHQ2RARiZYyTAAeKiDX) and BECH32 address (such as qc1qkt33x6hkrrlwlr6v59wptwy6zskyrjfe40y0lx)
			// will cause ConvertKaonAddress to fall
			// but we not support them for now
			fromAddressIn, err := utils.ConvertKaonAddress(*sender)

			if err != nil {
				continue
			}
			fromAddress = fromAddressIn
		} else {

			// P2Sh address(such as MUrenj2sPqEVTiNbHQ2RARiZYyTAAeKiDX) and BECH32 address (such as qc1qkt33x6hkrrlwlr6v59wptwy6zskyrjfe40y0lx)
			// will cause ConvertKaonAddress to fall
			// but we not support them for now
			fromAddressIn, err := utils.ConvertKaonAddress(vin.Address)

			if err != nil {
				continue
			}
			fromAddress = fromAddressIn
		}

		if int(vin.Vout) >= len(resp.Vouts) {
			continue

			// TODO: probably we still may be able to solve it in that case
			// research
			// return ContractInfo{
			// 	From: fromAddress,
			// }, false, nil
		}

		var (
			vout     = resp.Vouts[i]
			reciever = ""
		)
		for _, address := range vout.ScriptPubKey.Addresses {
			if address != "" {
				// P2Sh address(such as MUrenj2sPqEVTiNbHQ2RARiZYyTAAeKiDX) and BECH32 address (such as qc1qkt33x6hkrrlwlr6v59wptwy6zskyrjfe40y0lx)
				// will cause ConvertKaonAddress to fall
				// but we not support them for now
				hex, err := utils.ConvertKaonAddress(address)
				if err == nil {
					reciever = hex
					break
				}
			}
		}

		if reciever == "" {
			var (
				script  = strings.Split(vout.ScriptPubKey.ASM, " ")
				finalOp = script[len(script)-1]
			)
			if finalOp != "OP_CHECKSIG" {
				continue // We do not support non-standard money transfers
			}
			recieverParsed, err := ParseP2PKHReciever(script)
			if err != nil {
				continue
			}

			// P2Sh address(such as MUrenj2sPqEVTiNbHQ2RARiZYyTAAeKiDX) and BECH32 address (such as qc1qkt33x6hkrrlwlr6v59wptwy6zskyrjfe40y0lx)
			// will cause ConvertKaonAddress to fall
			// but we not support them for now
			recieverIn, err := utils.ConvertKaonAddress(*recieverParsed)

			if err != nil {
				continue
			}
			reciever = recieverIn
		}

		return ContractInfo{
			From: fromAddress,
			To:   reciever,
		}, false, nil
	}

	return ContractInfo{}, false, nil
}

// ========CreateRawTransaction=========//
type (
	/*
				Arguments:
				1. inputs         (json array, required) A json array of json objects
				[
				{                              (json object)
					"txid": "hex",               (string, required) The transaction id
					"vout": n,                   (numeric, required) The output number
					"sequence": n,               (numeric, optional, default=depends on the value of the 'replaceable' and 'locktime' arguments) The sequence number
				},
				...
				]
				2. outputs   	 (json array, required) a json array with outputs (key-value pairs), where none of the keys are duplicated.
		                                      That is, each address can only appear once and there can only be one 'data' object.
		                                      For compatibility reasons, a dictionary, which holds the key-value pairs directly, is also
		                                      accepted as second parameter.
		     [
		       {                              (json object)
		         "address": amount,           (numeric or string, required) A key-value pair. The key (string) is the Kaon address, the value (float or string) is the amount in KAON
		       },
		       {                              (json object)
		         "data": "hex",               (string, required) A key-value pair. The key must be "data", the value is hex-encoded data
		       },
		       {                              (json object) (send to contract)
		         "contractAddress": "hex",    (string, required) Valid contract address (valid hash160 hex data)
		         "data": "hex",               (string, required) Hex data to add in the call output
		         "amount": amount,            (numeric or string, optional, default=0) Value in KAON to send with the call, should be a valid amount, default 0
		         "gasLimit": n,               (numeric) The gas limit for the transaction
		         "gasPrice": n,               (numeric) The gas price for the transaction
		         "senderaddress": "hex",      (string) The Kaon address that will be used to create the contract.
		       },
		       {                              (json object) (create contract)
		         "bytecode": "hex",           (string, required) contract bytcode.
		         "gasLimit": n,               (numeric) The gas limit for the transaction
		         "gasPrice": n,               (numeric) The gas price for the transaction
		         "senderaddress": "hex",      (string) The Kaon address that will be used to create the contract.
		       },
		       ...
		     ]
			3. locktime                           (numeric, optional, default=0) Raw locktime. Non-0 value also locktime-activates inputs
			4. replaceable                        (boolean, optional, default=false) Marks this transaction as BIP125-replaceable.
		                                      Allows this transaction to be replaced by a transaction with higher fees. If provided, it is an error if explicit sequence numbers are incompatible.
	*/

	RawTxInputs struct {
		TxID string `json:"txid"`
		Vout uint   `json:"vout"`
	}
)

// ========SignRawTransactionWithKey=========//
type (
	/*
			Result:
		{
		  "hex" : "value",                  (string) The hex-encoded raw transaction with signature(s)
		  "complete" : true|false,          (boolean) If the transaction has a complete set of signatures
		  "errors" : [                      (json array of objects) Script verification errors (if there are any)
		    {
		      "txid" : "hash",              (string) The hash of the referenced, previous transaction
		      "vout" : n,                   (numeric) The index of the output to spent and used as input
		      "scriptSig" : "hex",          (string) The hex-encoded signature script
		      "sequence" : n,               (numeric) Script sequence number
		      "error" : "text"              (string) Verification or signing error related to the input
		    }
		    ,...
		  ]
		}
	*/
	SignRawTxResponse struct {
		Hex      string `json:"hex"`
		Complete bool   `json:"complete"`
	}

	SigningError struct {
		Txid      string `json:"txid"`
		Vout      uint   `json:"vout"`
		ScriptSig string `json:"scriptSig"`
		Sequence  uint   `json:"sequence"`
		Error     error  `json:"error"`
	}
)

// ========== GetAddressUTXOs ============= //

type (
	/*
		Arguments:
		1. Input params              (json object, required) Json object
			{
			"addresses": [        (json array, required) The Kaon addresses
				"address",          (string) The Kaon address
				...
			],
			"chainInfo": bool,    (boolean, optional) Include chain info with results
			}

		Result:
		{                       (json object)
		"address" : "str",    (string) The address base58check encoded
		"txid" : "hex",       (string) The output txid
		"height" : n,         (numeric) The block height
		"outputIndex" : n,    (numeric) The output index
		"script" : "hex",     (string) The script hex encoded
		"satoshis" : n        (numeric) The number of satoshis of the output
		}
	*/

	GetAddressUTXOsRequest struct {
		Addresses []string `json:"addresses"`
	}

	UTXO struct {
		Address     string   `json:"address"`
		TXID        string   `json:"txid"`
		OutputIndex uint     `json:"outputIndex"`
		Script      string   `json:"script"`
		Satoshis    Amount   `json:"satoshis"`
		Height      *big.Int `json:"height"`
		IsStake     bool     `json:"isStake"`
	}

	GetAddressUTXOsResponse []UTXO
)

func (resp *GetAddressUTXOsResponse) UnmarshalJSON(data []byte) error {
	// NOTE: do not use `GetTransactionReceiptResponse`, 'cause
	// it may violate to infinite loop, while calling
	// UnmarshalJSON interface
	var utxos []UTXO
	if err := json.Unmarshal(data, &utxos); err != nil {
		return err
	}
	*resp = GetAddressUTXOsResponse(utxos)
	return nil
}

func (r *GetAddressUTXOsRequest) MarshalJSON() ([]byte, error) {
	params := []map[string]interface{}{}
	addresses := map[string]interface{}{
		"addresses": r.Addresses,
	}
	params = append(params, addresses)
	return json.Marshal(params)
}

// ========== ListUnspent ============= //
type (

	/*
		Arguments:
		1. minconf          (numeric, optional, default=1) The minimum confirmations to filter
		2. maxconf          (numeric, optional, default=9999999) The maximum confirmations to filter
		3. "addresses"      (string) A json array of Kaon addresses to filter
		    [
		      "address"     (string) Kaon address
		      ,...
		    ]
		4. include_unsafe (bool, optional, default=true) Include outputs that are not safe to spend
		                  See description of "safe" attribute below.
		5. query_options    (json, optional) JSON with query options
		    {
		      "minimumAmount"    (numeric or string, default=0) Minimum value of each UTXO in KAON
		      "maximumAmount"    (numeric or string, default=unlimited) Maximum value of each UTXO in KAON
		      "maximumCount"     (numeric or string, default=unlimited) Maximum number of UTXOs
		      "minimumSumAmount" (numeric or string, default=unlimited) Minimum sum value of all UTXOs in KAON
		    }
	*/
	ListUnspentRequest struct {
		MinConf      int
		MaxConf      int
		Addresses    []string
		QueryOptions ListUnspentQueryOptions
	}
	ListUnspentQueryOptions struct {
		// Applies to each UTXO
		MinAmount Amount
		// Applies to each UTXO
		MaxAmount      Amount
		MaxNumToReturn int
		// Returns only those UTXOs, which total amount
		// is greater than or equal `MinSumAmount`
		//
		// NOTE: it doesn't consider amount of all
		// UTXOs, that is not all UTXOs may be
		// returned, but a limited number of UTXOs
		MinSumAmount Amount
	}

	/*
				[                   (array of json object)
					{
						"txid" : "txid",          (string) the transaction id
						"vout" : n,               (numeric) the vout value
						"address" : "address",    (string) the Kaon address
						"account" : "account",    (string) DEPRECATED. The associated account, or "" for the default account
						"scriptPubKey" : "key",   (string) the script key
						"amount" : x.xxx,         (numeric) the transaction output amount in KAON
						"confirmations" : n,      (numeric) The number of confirmations
						"redeemScript" : n        (string) The redeemScript if scriptPubKey is P2SH
						"spendable" : xxx,        (bool) Whether we have the private keys to spend this output
						"solvable" : xxx,         (bool) Whether we know how to spend this output, ignoring the lack of keys
						"safe" : xxx              (bool) Whether this output is considered safe to spend. Unconfirmed transactions
							  from outside keys and unconfirmed replacement transactions are considered unsafe
							  and are not eligible for spending by fundrawtransaction and sendtoaddress.
					}
					,...
				]

		[
			{
				"txid": "a8d97ae8bb819cd4aa98ed2ddaef4969783aee845461a9ea5a88184ad58f44fe",
				"vout": 2,
				"address": "ar2SzdHghSgeacypPn7zfDe3qfKAEwimus",
				"account": "",
				"scriptPubKey": "210299d391f528b9edd07284c7e23df8415232a8ce41531cf460a390ce32b4efd112ac",
				"amount": 15007.10682200,
				"confirmations": 532,
				"spendable": true,
				"solvable": true,
				"safe": true
			}
		]
	*/
	ListUnspentResponse []struct {
		Address       string `json:"address"`
		Txid          string `json:"txid"`
		Vout          uint   `json:"vout"`
		Amount        Amount `json:"amount"`
		Safe          bool   `json:"safe"`
		Spendable     bool   `json:"spendable"`
		Solvable      bool   `json:"solvable"`
		Label         string `json:"label"`
		Confirmations int    `json:"confirmations"`
		ScriptPubKey  string `json:"scriptPubKey"`
		RedeemScript  string `json:"redeemScript"`
	}
)

func (r *ListUnspentRequest) MarshalJSON() ([]byte, error) {
	params := []interface{}{
		r.MinConf,
		r.MaxConf,
		r.Addresses,
		true, // `include_unsafe`
		r.QueryOptions,
	}
	return json.Marshal(params)
}

func (options ListUnspentQueryOptions) MarshalJSON() ([]byte, error) {
	optionsObj := map[string]string{}

	if !options.MinAmount.IsZero() {
		optionsObj["minimumAmount"] = options.MinAmount.String()
	}
	if !options.MaxAmount.IsZero() {
		optionsObj["maximumAmount"] = options.MaxAmount.String()
	}
	if options.MaxNumToReturn > 1 {
		optionsObj["maximumCount"] = strconv.Itoa(options.MaxNumToReturn)
	}
	if !options.MinSumAmount.IsZero() {
		optionsObj["minimumSumAmount"] = options.MinSumAmount.String()
	}
	return json.Marshal(optionsObj)
}

// ======== getstorage ======== //
type (
	GetStorageRequest struct {
		Address     string   `json:"address"`
		BlockNumber *big.Int `json:"blockNumber"`
		Index       *big.Int `json:"index"`
	}
	GetStorageResponse map[string]map[string]string
)

func (request *GetStorageRequest) MarshalJSON() ([]byte, error) {
	params := []interface{}{request.Address}
	if request.BlockNumber != nil {
		params = append(params, request.BlockNumber)
	}
	if request.Index != nil {
		params = append(params, request.Index)
	}
	return json.Marshal(params)
}

// =======getaddressnounce ============= //
type (
	/*
		Arguments:
		1. "address",	(string) The Kaon address
		2. "blockNumber" (int | "latest", optional, default=null) The block number to start looking for transactions. ()
		Result: n (int) The integer of the number of transactions sent from an address
	*/
	GetTransactionCountRequest struct {
		Address     string
		BlockNumber interface{} `json:"blockNumber"`
	}

	GetTransactionCountResponse struct {
		*big.Int
	}
)

func (request *GetTransactionCountRequest) MarshalJSON() ([]byte, error) {
	params := []interface{}{request.Address}
	if request.BlockNumber != nil {
		params = append(params, request.BlockNumber)
	}
	return json.Marshal(params)
}

// ======== getaddressbalance ========= //
type (

	/*
		Arguments:
		1. addresses       	(json array, required) The Kaon addresses
			[
				"address",	(string) The Kaon address
				...
			]
		Result:
		{					(json object)
			"balance": n 	(numeric) The current balance in satoshis
			"received": n   (numeric) The total number of satoshis received (including change)
		}
	*/
	GetAddressBalanceRequest struct {
		Address string
	}

	GetAddressBalanceResponse struct {
		Balance  big.Int `json:"balance"`
		Received big.Int `json:"received"`
		Immature big.Int `json:"immature"`
	}
)

func (req *GetAddressBalanceRequest) MarshalJSON() ([]byte, error) {
	params := []interface{}{
		req.Address,
	}
	return json.Marshal(params)
}

// ======== getpeerinfo ========= //
type (
	GetPeerInfoResponse struct {
		// Peer index
		Id int `json:"id"`
		// The IP address and port of the peer - host:port
		Address string `json:"addr"`
		// Bind address of the connection to the peer - ip:port
		AddressBind string `json:"addrbind"`
		// Local address as reported by the peer - ip:port
		AddressLocal string `json:"addrlocal"`
		// The services offered
		Services string `json:"services"`
		// Whether peer has asked us to relay transactions to it
		RelayTransactions bool `json:"relaytxes"`
		// The time in seconds since epoch (Jan 1 1970 GMT) of the last send
		LastSend uint64 `json:"lastsend"`
		// The time in seconds since epoch (Jan 1 1970 GMT) of the last receive
		LastReceive uint64 `json:"lastrecv"`
		// The total bytes sent
		BytesSent uint64 `json:"bytessent"`
		// The total bytes received
		BytesReceived uint64 `json:"bytesrecv"`
		// The connection time in seconds since epoch (Jan 1 1970 GMT)
		ConnectionTime uint64 `json:"conntime"`
		// The time offset in seconds
		TimeOffset uint64 `json:"timeoffset"`
		// ping time (if available)
		PingTime decimal.Decimal `json:"pingtime"`
		// minimum observed ping time (if any at all)
		MinimumPing decimal.Decimal `json:"minping"`
		// ping wait (if non-zero)
		PingWait decimal.Decimal `json:"pingwait"`
		// The peer version, such as 70001
		Version int64 `json:"version"`
		// The string version
		Subversion string `json:"subver"`
		// Inbound (true) or Outbound (false)
		Inbound bool `json:"inbound"`
		// Whether connection was due to addnode/-connect or if it was an automatic/inbound connection
		Addnode bool `json:"addnode"`
		// The starting height (block) of the peer
		StartingHeight uint64 `json:"startingheight"`
		// The ban score
		BanScore int64 `json:"banscore"`
		// The last header we have in common with this peer
		SyncedHeaders int64 `json:"synced_headers"`
		// The last block we have in common with this peer
		SyncedBlocks int64 `json:"synced_blocks"`
		// The heights of blocks we're currently asking from this peer
		Inflight []int64 `json:"inflight"`
		// Whether the peer is whitelisted
		Whitelisted bool `json:"whitelisted"`
		// The total bytes sent aggregated by message type
		BytesSentPerMessage PeerInfoBytesPerMessage `json:"bytessent_per_msg"`
		// The total bytes received aggregated by message type
		BytesReceivedPerMessage PeerInfoBytesPerMessage `json:"bytesrecv_per_msg"`
	}

	PeerInfoBytesPerMessage struct {
		Address     int64  `json:"addr"`
		FeeFilter   uint64 `json:"feefilter"`
		GetHeaders  uint64 `json:"getheaders"`
		Headers     uint64 `json:"headers"`
		Ping        uint64 `json:"ping"`
		Pong        uint64 `json:"pong"`
		SendCompact uint64 `json:"sendcmpct"`
		SendHeaders uint64 `json:"sendheaders"`
		Verack      uint64 `json:"verack"`
		Version     int64  `json:"version"`
	}
)

// ========= getnetworkinfo ========== //
type (
	NetworkInfoResponse struct {
		Version            int64                     `json:"version"`
		Subversion         string                    `json:"subversion"`
		ProtocolVersion    int64                     `json:"protocolversion"`
		LocalServices      string                    `json:"localservices"`
		LocalServicesNames []string                  `json:"localservicesnames"`
		LocalRelay         bool                      `json:"localrelay"`
		TimeOffset         int64                     `json:"timeoffset"`
		Connections        int64                     `json:"connections"`
		NetworkActive      bool                      `json:"networkactive"`
		Networks           []NetworkInfoNetworkInfo  `json:"networks"`
		RelayFee           Amount                    `json:"relayfee"`
		IncrementalFee     Amount                    `json:"incrementalfee"`
		LocalAddresses     []NetworkInfoLocalAddress `json:"localaddresses"`
		Warnings           string                    `json:"warnings"`
	}

	NetworkInfoNetworkInfo struct {
		Name                      string `json:"name"`
		Limited                   bool   `json:"limited"`
		Reachable                 bool   `json:"reachable"`
		Proxy                     string `json:"proxy"`
		ProxyRandomizeCredentials bool   `json:"proxy_randomize_credentials"`
	}

	NetworkInfoLocalAddress struct {
		Address string `json:"address"`
		Port    uint64 `json:"port"`
		Score   int64  `json:"score"`
	}
)

// ========= waitforlogs ========== //
type (
	WaitForLogsRequest struct {
		FromBlock            interface{}       `json:"fromBlock"`
		ToBlock              interface{}       `json:"toBlock"`
		Filter               WaitForLogsFilter `json:"filter"`
		MinimumConfirmations int64             `json:"miniconf"`
	}

	WaitForLogsFilter struct {
		Addresses *[]string          `json:"addresses,omitempty"`
		Topics    *[]SearchLogsTopic `json:"topics,omitempty"`
	}

	WaitForLogsEntry struct {
		BlockHash        string `json:"blockHash"`
		BlockNumber      uint64 `json:"blockNumber"`
		Bloom            string `json:"bloom"`
		TransactionHash  string `json:"transactionHash"`
		TransactionIndex uint64 `json:"transactionIndex"`
		From             string `json:"from"`
		// (does this apply to waitforlogs or only searchlogs?)
		// NOTE: will be null for a contract creation transaction
		To                string  `json:"to"`
		CumulativeGasUsed big.Int `json:"cumulativeGasUsed"`
		GasUsed           big.Int `json:"gasUsed"`

		// (does this apply to waitforlogs or only searchlogs?)
		// TODO: discuss -
		// 	? May be a contract transaction created by non-contract
		//
		// The created contract address. If this tx is created by the contract,
		// return the contract address, else return null
		ContractAddress string   `json:"contractAddress"`
		Data            string   `json:"data"`
		Topics          []string `json:"topics"`

		// May has "None" value, which means, that transaction is not executed
		Excepted        string `json:"excepted"`
		ExceptedMessage string `json:"exceptedMessage"`

		OutputIndex int64 `json:"outputIndex"`
	}

	WaitForLogsResponse struct {
		Entries   []WaitForLogsEntry `json:"entries"`
		Count     uint64             `json:"count"`
		NextBlock uint64             `json:"nextBlock"`
	}
)

func (e WaitForLogsEntry) Log() Log {
	return Log{
		Address: e.ContractAddress,
		Topics:  []string(e.Topics),
		Data:    e.Data,
	}
}

func (e WaitForLogsEntry) GetTransactionHash() string {
	return e.TransactionHash
}

func (e WaitForLogsEntry) GetTransactionIndex() uint64 {
	return e.TransactionIndex
}

func (e WaitForLogsEntry) GetBlockHash() string {
	return e.BlockHash
}

func (e WaitForLogsEntry) GetBlockNumber() uint64 {
	return e.BlockNumber
}

func (e WaitForLogsEntry) GetAddress() string {
	return e.ContractAddress
}

func (e WaitForLogsEntry) GetTopics() []string {
	return e.Topics
}

func (e WaitForLogsEntry) GetData() string {
	return e.Data
}

func (r *WaitForLogsRequest) MarshalJSON() ([]byte, error) {
	/*
		1. fromBlock (int | "latest", optional, default=null) The block number to start looking for logs. ()
		2. toBlock   (int | "latest", optional, default=null) The block number to stop looking for logs. If null, will wait indefinitely into the future.
		3. filter    ({ addresses?: Hex160String[], topics?: Hex256String[] }, optional default={}) Filter conditions for logs. Addresses and topics are specified as array of hexadecimal strings
		4. minconf   (uint, optional, default=6) Minimal number of confirmations before a log is returned
	*/
	return json.Marshal([]interface{}{
		r.FromBlock,
		r.ToBlock,
		r.Filter,
		r.MinimumConfirmations,
	})
}

// ========= createwallet ========== //
type (
	CreateWalletRequest  []string
	CreateWalletResponse struct {
		Name    string `json:"name"`
		Warning string `json:"warning"`
	}
)

// ========= loadwallet ========== //
type (
	LoadWalletRequest  []string
	LoadWalletResponse struct {
		Name    string `json:"name"`
		Warning string `json:"warning"`
	}
)

// ======== sendrawtransaction ========= //

type (
	// Presents hexed string of a raw transcation
	SendRawTransactionRequest [1]string

	GeneralSendRawTransactionResponse struct {
		// Error  *string `json:"error"`
		// ID     string  `json:"id"`
		Result string `json:"result"`

		// It may have in the Data
		// SendRawTransactionResponse / CreateContractResponse / CallContractResponse or SendToContractResponse
		Data interface{} `json:"-"`
	}

	// Presents hexed string of a transaction hash
	SendRawTransactionResponse struct {
		Result string `json:"result"`
	}
)

func (r *GeneralSendRawTransactionResponse) UnmarshalJSON(data []byte) error {
	var err error

	// Check for unique identifiers in `data` and unmarshal the specific parts
	if strings.Contains(string(data), "txid") && strings.Contains(string(data), "address") {
		var createContractResponse CreateContractResponse
		err = json.Unmarshal(data, &createContractResponse)
		if err != nil {
			return err
		}
		r.Data = createContractResponse
	} else if strings.Contains(string(data), "executionResult") {
		var callContractResponse CallContractResponse
		err = json.Unmarshal(data, &callContractResponse)
		if err != nil {
			return err
		}
		r.Data = callContractResponse
	} else if strings.Contains(string(data), "txid") {
		var sendToContractResponse SendToContractResponse
		err = json.Unmarshal(data, &sendToContractResponse)
		if err != nil {
			return err
		}
		r.Data = sendToContractResponse
	} else {
		// SendRawTransactionResponse is a string
		// If no unique identifier is present, assume it's SendRawTransactionResponse
		var sendRawTransactionResponse string
		err = json.Unmarshal(data, &sendRawTransactionResponse)
		if err != nil {
			return err
		}
		r.Result = sendRawTransactionResponse
		r.Data = sendRawTransactionResponse
	}

	return nil
}
