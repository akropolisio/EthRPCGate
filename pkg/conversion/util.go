package conversion

import (
	"context"
	"strings"

	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/kaonone/eth-rpc-gate/pkg/kaon"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/kaonone/eth-rpc-gate/pkg/utils"
)

func ExtractETHLogsFromTransactionReceipt(blockData kaon.LogBlockData, logs []kaon.Log) []eth.Log {
	result := make([]eth.Log, 0, len(logs))
	for i, log := range logs {
		topics := make([]string, 0, len(log.GetTopics()))
		for _, topic := range log.GetTopics() {
			topics = append(topics, utils.AddHexPrefix(topic))
		}
		result = append(result, eth.Log{
			TransactionHash:  utils.AddHexPrefix(blockData.GetTransactionHash()),
			TransactionIndex: hexutil.EncodeUint64(blockData.GetTransactionIndex()),
			BlockHash:        utils.AddHexPrefix(blockData.GetBlockHash()),
			BlockNumber:      hexutil.EncodeUint64(blockData.GetBlockNumber()),
			Data:             utils.AddHexPrefix(log.GetData()),
			Address:          utils.AddHexPrefix(log.GetAddress()),
			Topics:           topics,
			LogIndex:         hexutil.EncodeUint64(uint64(i)),
		})
	}
	return result
}

func ConvertLogTopicsToStringArray(topics []interface{}) []string {
	var requestedTopics []string
	for _, topic := range topics {
		requestedTopic, ok := topic.(string)
		if ok {
			requestedTopics = append(requestedTopics, requestedTopic)
		}
	}

	return requestedTopics
}

func SearchLogsAndFilterExtraTopics(ctx context.Context, q *kaon.Kaon, req *kaon.SearchLogsRequest) (kaon.SearchLogsResponse, *eth.JSONRPCError) {
	receipts, err := q.SearchLogs(ctx, req)
	if err != nil {
		return nil, eth.NewCallbackError(err.Error())
	}

	hasTopics := len(req.Topics) != 0
	hasAddresses := len(req.Addresses) != 0

	if !hasTopics && !hasAddresses {
		return receipts, nil
	}

	if !hasTopics && !hasAddresses {
		// no actual string topics or addresses, probably weird inputs
		return receipts, nil
	}

	requestedAddressesMap := populateLoopUpMapWithToLower(req.Addresses)

	var filteredReceipts kaon.SearchLogsResponse

	for _, receipt := range receipts {
		var logs []kaon.Log
		for index, log := range receipt.Log {
			log.Index = index
			if hasAddresses && !requestedAddressesMap[strings.ToLower(log.Address)] {
				continue
			}

			if DoFiltersMatch(req.Topics, log.Topics) {
				logs = append(logs, log)
			}
		}
		receipt.Log = logs
		if len(receipt.Log) != 0 {
			filteredReceipts = append(filteredReceipts, receipt)
		}
	}

	return filteredReceipts, nil
}

func FilterKaonLogs(addresses []string, filters []kaon.SearchLogsTopic, logs []kaon.Log) []kaon.Log {
	hasTopics := len(filters) != 0
	hasAddresses := len(addresses) != 0

	if !hasTopics && !hasAddresses {
		return logs
	}

	if !hasTopics && !hasAddresses {
		// no actual string topics or addresses, probably weird inputs
		return logs
	}

	requestedAddressesMap := populateLoopUpMapWithToLower(addresses)

	filteredLogs := []kaon.Log{}

	for _, log := range logs {
		if hasAddresses && !requestedAddressesMap[strings.ToLower(strings.TrimPrefix(log.Address, "0x"))] {
			continue
		}

		if DoFiltersMatch(filters, log.Topics) {
			filteredLogs = append(filteredLogs, log)
			break
		}
	}

	return filteredLogs
}

func DoFiltersMatch(filters []kaon.SearchLogsTopic, topics []string) bool {
	filterCount := len(filters)
	for i, topic := range topics {
		if i >= filterCount {
			break
		}

		filter := filters[i]

		if len(filter) == 0 {
			// nil, accept all
			continue
		} else if len(filter) == 1 {
			if strings.ToLower(filter[0]) == strings.ToLower(topic) {
				// match
				continue
			} else {
				// not a match
				return false
			}
		} else {
			// or
			match := false

			for _, orFilter := range filter {
				match = strings.ToLower(orFilter) == strings.ToLower(topic)
				if match {
					break
				}
			}

			if match {
				continue
			} else {
				return false
			}
		}
	}

	return true
}

func populateLoopUpMapWithToLower(inputs []string) map[string]bool {
	lookupMap := make(map[string]bool)

	for _, input := range inputs {
		lookupMap[strings.ToLower(input)] = true
	}

	return lookupMap
}
