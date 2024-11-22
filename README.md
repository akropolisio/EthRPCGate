[![Github Build Status](https://github.com/kaonone/eth-rpc-gate/workflows/Openzeppelin/badge.svg)](https://github.com/kaonone/eth-rpc-gate/actions)
[![Github Build Status](https://github.com/kaonone/eth-rpc-gate/workflows/Unit%20tests/badge.svg)](https://github.com/kaonone/eth-rpc-gate/actions)

# Kaon adapter to Ethereum JSON RPC
eth-rpc-gate is a web3 proxy adapter that can be used as a web3 provider to interact with Kaon. It supports HTTP(s) and websockets and the current version enables self hosting of keys.

# Table of Contents

- [Kaon adapter to Ethereum JSON RPC](#kaon-adapter-to-ethereum-json-rpc)
- [Table of Contents](#table-of-contents)
  - [Quick start](#quick-start)
    - [Public instances](#public-instances)
      - [You can use public instances if you don't need to use eth\_sendTransaction or eth\_accounts](#you-can-use-public-instances-if-you-dont-need-to-use-eth_sendtransaction-or-eth_accounts)
  - [Requirements](#requirements)
  - [Installation](#installation)
    - [SSL](#ssl)
    - [Self-signed SSL](#self-signed-ssl)
  - [How to use eth-rpc-gate as a Web3 provider](#how-to-use-eth-rpc-gate-as-a-web3-provider)
  - [How to add eth-rpc-gate to Metamask](#how-to-add-eth-rpc-gate-to-metamask)
  - [Truffle support](#truffle-support)
  - [Ethers support](#ethers-support)
  - [Supported ETH methods](#supported-eth-methods)
  - [Websocket ETH methods (endpoint at /)](#websocket-eth-methods-endpoint-at-)
  - [eth-rpc-gate methods](#eth-rpc-gate-methods)
  - [Development methods](#development-methods)
  - [Health checks](#health-checks)
  - [Deploying and Interacting with a contract using RPC calls](#deploying-and-interacting-with-a-contract-using-rpc-calls)
    - [Assumption parameters](#assumption-parameters)
    - [Deploy the contract](#deploy-the-contract)
    - [Get the transaction using the hash from previous the result](#get-the-transaction-using-the-hash-from-previous-the-result)
    - [Get the transaction receipt](#get-the-transaction-receipt)
    - [Calling the set method](#calling-the-set-method)
    - [Calling the get method](#calling-the-get-method)
    - [EVM Versions](#evm-versions)
  - [Future work](#future-work)

## Quick start
### Public instances
#### You can use public instances if you don't need to use eth_sendTransaction or eth_accounts
Mainnet: https://mainnet.kaon.one/

Testnet: https://testnet.kaon.one/

Regtest: run it locally with ```make quick-start-regtest```

If you need to use eth_sendTransaction, you are going to have to run your own instance pointing to your own Kaon instance

Standard eth_sendRawTransaction will work as expected.

See [Differences between EVM chains](#differences-between-evm-chains) below

## Requirements

- Golang
- Docker
- linux commands: `make`, `curl`

## Installation

```
$ sudo apt install make git golang docker-compose
# Configure GOPATH if not configured
$ export GOPATH=`go env GOPATH`
$ mkdir -p $GOPATH/src/github.com/kaonone && \
  cd $GOPATH/src/github.com/kaonone && \
  git clone https://github.com/kaonone/eth-rpc-gate
$ cd $GOPATH/src/github.com/kaonone/eth-rpc-gate
# Generate self-signed SSL cert (optional)
# If you do this step, eth-rpc-gate will respond in SSL
# otherwise, eth-rpc-gate will respond unencrypted
$ make docker-configure-https
# Pick a network to quick-start with
$ make quick-start-regtest
$ make quick-start-testnet
$ make quick-start-mainnet
```
This will build the docker image for the local version of eth-rpc-gate as well as spin up two containers:

-   One named `ethrpcgate` running on port 25996
    
-   Another one named `kaon` running on rpc port 51474
    

`make quick-start` will also fund the tests accounts with KAON in order for you to start testing and developing locally. Additionally, if you need or want to make changes and or additions to eth-rpc-gate, but don't want to go through the hassle of rebuilding the container, you can run the following command at the project root level:
```
$ make run-ethrpcgate
# For https
$ make docker-configure-https && make run-ethrpcgate-https
```
Which will run the most current local version of eth-rpc-gate on port 25996, but without rebuilding the image or the local docker container.

Note that eth-rpc-gate will use the hex address for the test base58 Kaon addresses that belong the the local Kaon node, for example:
  - ar2SzdHghSgeacypPn7zfDe3qfKAEwimus (hex 0x1CE507204a6fC8fd6aA7e54D1481d30ACB0Dbead )
  - auASFMxv45WgjCW6wkpDuHWjxXhzNA9mjP (hex 0x3f501c368cb9ddb5f27ed72ac0d602724adfa175 )

### SSL
SSL keys and certificates go inside the https folder (mounted at `/https` in the container) and use `--https-key` and `--https-cert` parameters. If the specified files do not exist, it will fall back to http.

### Self-signed SSL
To generate self-signed certificates with docker for local development the following script will generate SSL certificates and drop them into the https folder

```
$ make docker-configure-https
```

## How to use eth-rpc-gate as a Web3 provider

Once eth-rpc-gate is successfully running, all one has to do is point your desired framework to eth-rpc-gate in order to use it as your web3 provider. Lets say you want to use truffle for example, in this case all you have to do is go to your truffle-config.js file and add ethrpcgate as a network:
```
module.exports = {
  networks: {
    ethrpcgate: {
      host: "127.0.0.1",
      port: 25996,
      network_id: "*",
      gasPrice: "0x5d21dba000"
    },
    ...
  },
...
}
```

## How to add eth-rpc-gate to Metamask

Getting eth-rpc-gate to work with Metamask requires just one thing:
- [Configuring Metamask to point to eth-rpc-gate](metamask)

## Truffle support

Hosting your own eth-rpc-gate and blockchain instance works similarly to geth and so truffle is completely supported.

## Ethers support

Ethers is supported, please follow Kaon repository since fork of ethers with new chain ID will be published soon.

## Supported ETH methods

-   [web3_clientVersion](pkg/transformer/web3_clientVersion.go)
-   [web3_sha3](pkg/transformer/web3_sha3.go)
-   [net_version](pkg/transformer/eth_net_version.go)
-   [net_listening](pkg/transformer/eth_net_listening.go)
-   [net_peerCount](pkg/transformer/eth_net_peerCount.go)
-   [eth_protocolVersion](pkg/transformer/eth_protocolVersion.go)
-   [eth_chainId](pkg/transformer/eth_chainId.go)
-   [eth_mining](pkg/transformer/eth_mining.go)
-   [eth_hashrate](pkg/transformer/eth_hashrate.go)
-   [eth_gasPrice](pkg/transformer/eth_gasPrice.go)
-   [eth_accounts](pkg/transformer/eth_accounts.go)
-   [eth_blockNumber](pkg/transformer/eth_blockNumber.go)
-   [eth_getBalance](pkg/transformer/eth_getBalance.go)
-   [eth_getStorageAt](pkg/transformer/eth_getStorageAt.go)
-   [eth_getTransactionCount](pkg/transformer/eth_getTransactionCount.go)
-   [eth_getCode](pkg/transformer/eth_getCode.go)
-   [eth_sign](pkg/transformer/eth_sign.go)
-   [eth_signTransaction](pkg/transformer/eth_signTransaction.go)
-   [eth_sendTransaction](pkg/transformer/eth_sendTransaction.go)
-   [eth_sendRawTransaction](pkg/transformer/eth_sendRawTransaction.go)
-   [eth_call](pkg/transformer/eth_call.go)
-   [eth_estimateGas](pkg/transformer/eth_estimateGas.go)
-   [eth_getBlockByHash](pkg/transformer/eth_getBlockByHash.go)
-   [eth_getBlockByNumber](pkg/transformer/eth_getBlockByNumber.go)
-   [eth_getTransactionByHash](pkg/transformer/eth_getTransactionByHash.go)
-   [eth_getTransactionByBlockHashAndIndex](pkg/transformer/eth_getTransactionByBlockHashAndIndex.go)
-   [eth_getTransactionByBlockNumberAndIndex](pkg/transformer/eth_getTransactionByBlockNumberAndIndex.go)
-   [eth_getTransactionReceipt](pkg/transformer/eth_getTransactionReceipt.go)
-   [eth_getUncleByBlockHashAndIndex](pkg/transformer/eth_getUncleByBlockHashAndIndex.go)
-   [eth_getCompilers](pkg/transformer/eth_getCompilers.go)
-   [eth_newFilter](pkg/transformer/eth_newFilter.go)
-   [eth_newBlockFilter](pkg/transformer/eth_newBlockFilter.go)
-   [eth_uninstallFilter](pkg/transformer/eth_uninstallFilter.go)
-   [eth_getFilterChanges](pkg/transformer/eth_getFilterChanges.go)
-   [eth_getFilterLogs](pkg/transformer/eth_getFilterLogs.go)
-   [eth_getLogs](pkg/transformer/eth_getLogs.go)

## Websocket ETH methods (endpoint at /)

-   (All the above methods)
-   [eth_subscribe](pkg/transformer/eth_subscribe.go) (only 'logs' for now)
-   [eth_unsubscribe](pkg/transformer/eth_unsubscribe.go)

## eth-rpc-gate methods

-   [kaon_getUTXOs](pkg/transformer/kaon_getUTXOs.go)

## Development methods
Use these to speed up development, but don't rely on them in your dapp

-   [dev_gethexaddress](https://github.com/kaonone/kaoncore/blob/master/doc/JSON-RPC-interface.md#gethexaddress) Convert Kaon base58 address to hex
-   [dev_fromhexaddress](https://github.com/kaonone/kaoncore/blob/master/doc/JSON-RPC-interface.md#fromhexaddress) Convert from hex to Kaon base58 address for the connected network (strip 0x prefix from address when calling this)
-   [dev_generatetoaddress](https://github.com/kaonone/kaoncore/blob/master/doc/JSON-RPC-interface.md#generatetoaddress) Mines blocks in regtest (accepts hex/base58 addresses)

## Health checks

There are two health check endpoints, `GET /live` and `GET /ready` they return 200 or 503 depending on health (if they can connect to kaond)

## Deploying and Interacting with a contract using RPC calls


### Assumption parameters

Assume that you have a **contract** like this:

```solidity
pragma solidity ^0.4.18;

contract SimpleStore {
  constructor(uint _value) public {
    value = _value;
  }

  function set(uint newValue) public {
    value = newValue;
  }

  function get() public constant returns (uint) {
    return value;
  }

  uint value;
}
```

so that the **bytecode** is

```
solc --optimize --bin contracts/SimpleStore.sol

======= contracts/SimpleStore.sol:SimpleStore =======
Binary:
608060405234801561001057600080fd5b506040516020806100f2833981016040525160005560bf806100336000396000f30060806040526004361060485763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166360fe47b18114604d5780636d4ce63c146064575b600080fd5b348015605857600080fd5b5060626004356088565b005b348015606f57600080fd5b506076608d565b60408051918252519081900360200190f35b600055565b600054905600a165627a7a7230582049a087087e1fc6da0b68ca259d45a2e369efcbb50e93f9b7fa3e198de6402b810029
```

**constructor parameters** is `0000000000000000000000000000000000000000000000000000000000000001`

### Deploy the contract

```
$ curl --header 'Content-Type: application/json' --data \
     '{"id":"10","jsonrpc":"2.0","method":"eth_sendTransaction","params":[{"from":"0x1CE507204a6fC8fd6aA7e54D1481d30ACB0Dbead","gas":"0x6691b7","gasPrice":"0x5d21dba000","data":"0x608060405234801561001057600080fd5b506040516020806100f2833981016040525160005560bf806100336000396000f30060806040526004361060485763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166360fe47b18114604d5780636d4ce63c146064575b600080fd5b348015605857600080fd5b5060626004356088565b005b348015606f57600080fd5b506076608d565b60408051918252519081900360200190f35b600055565b600054905600a165627a7a7230582049a087087e1fc6da0b68ca259d45a2e369efcbb50e93f9b7fa3e198de6402b8100290000000000000000000000000000000000000000000000000000000000000001"}]}' \
     'http://localhost:25996'

{
  "jsonrpc": "2.0",
  "result": "0xa85cacc6143004139fc68808744ea6125ae984454e0ffa6072ac2f2debb0c2e6",
  "id": "10"
}
```

### Get the transaction using the hash from previous the result

```
$ curl --header 'Content-Type: application/json' --data \
     '{"id":"10","jsonrpc":"2.0","method":"eth_getTransactionByHash","params":["0xa85cacc6143004139fc68808744ea6125ae984454e0ffa6072ac2f2debb0c2e6"]}' \
     'localhost:25996'

{
  "jsonrpc":"2.0",
  "result": {
    "blockHash":"0x1e64595e724ea5161c0597d327072074940f519a6fb285ae60e73a4c996b47a4",
    "blockNumber":"0xc9b5",
    "transactionIndex":"0x5",
    "hash":"0xa85cacc6143004139fc68808744ea6125ae984454e0ffa6072ac2f2debb0c2e6",
    "nonce":"0x0",
    "value":"0x0",
    "input":"0x00",
    "from":"0x1CE507204a6fC8fd6aA7e54D1481d30ACB0Dbead",
    "to":"",
    "gas":"0x363639316237",
    "gasPrice":"0x5d21dba000"
  },
  "id":"10"
}
```

### Get the transaction receipt

```
$ curl --header 'Content-Type: application/json' --data \
     '{"id":"10","jsonrpc":"2.0","method":"eth_getTransactionReceipt","params":["0x6da39dc909debf70a536bbc108e2218fd7bce23305ddc00284075df5dfccc21b"]}' \
     'localhost:25996'

{
  "jsonrpc": "2.0",
  "result": {
    "transactionHash": "0xa85cacc6143004139fc68808744ea6125ae984454e0ffa6072ac2f2debb0c2e6",
    "transactionIndex": "0x5",
    "blockHash": "0x1e64595e724ea5161c0597d327072074940f519a6fb285ae60e73a4c996b47a4",
    "from":"0x1CE507204a6fC8fd6aA7e54D1481d30ACB0Dbead"
    "blockNumber": "0xc9b5",
    "cumulativeGasUsed": "0x8c235",
    "gasUsed": "0x1c071",
    "contractAddress": "0x1286595f8683ae074bc026cf0e587177b36842e2",
    "logs": [],
    "logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
    "status": "0x1"
  },
  "id": "10"
}
```

### Calling the set method

the ABI code of set method with param '["2"]' is `60fe47b10000000000000000000000000000000000000000000000000000000000000002`

```
$ curl --header 'Content-Type: application/json' --data \
     '{"id":"10","jsonrpc":"2.0","method":"eth_sendTransaction","params":[{"from":"0x1CE507204a6fC8fd6aA7e54D1481d30ACB0Dbead","gas":"0x6691b7","gasPrice":"0x5d21dba000","to":"0x1286595f8683ae074bc026cf0e587177b36842e2","data":"60fe47b10000000000000000000000000000000000000000000000000000000000000002"}]}' \
     'localhost:25996'

{
  "jsonrpc": "2.0",
  "result": "0x51a286c3bc68335274b9fd255e3988918a999608e305475105385f7ccf838339",
  "id": "10"
}
```

### Calling the get method

get method's ABI code is `6d4ce63c`

```
$ curl --header 'Content-Type: application/json' --data \
     '{"id":"10","jsonrpc":"2.0","method":"eth_call","params":[{"from":"0x1CE507204a6fC8fd6aA7e54D1481d30ACB0Dbead","gas":"0x6691b7","gasPrice":"0x5d21dba000","to":"0x1286595f8683ae074bc026cf0e587177b36842e2","data":"6d4ce63c"},"latest"]}' \
     'localhost:25996'

{
  "jsonrpc": "2.0",
  "result": "0x0000000000000000000000000000000000000000000000000000000000000002",
  "id": "10"
}
```

### EVM Versions
Currently KAON is operating under the Istanbul EVM, so Shanghai EVM may be incompatible since PISH0 is unsupported.
If you are deploying using Remix, please keep in mind that you need to go to the Solidity Compiler panel, open Advanced Configurations section and pick Istanbul there.

## Future work
- For eth_subscribe only the 'logs' type is supported at the moment,
- Complete support debugging and tracing methods for blocks and transactions for all required mods,
- Support of BRC20 and Ordinals transpiling,
- Support of P2SH and other more complex vout signatures,
- Complete support of WETH analogue in native tokens.