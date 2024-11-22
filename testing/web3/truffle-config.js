module.exports = {
  networks: {
    development: {
      host: "127.0.0.1",
      port: 25996, //Switch to 25996 for local HTTP Server, look at Makefile run-ethrpcgate
      network_id: "*",
      gasPrice: "0x64"
    },
    ganache: {
      host: "127.0.0.1",
      port: 8545,
      network_id: "*"
    },
    testnet: {
      host: "https://testnet.kaon.one/",
      port: 80,
      network_id: "*",
      from: "0x1CE507204a6fC8fd6aA7e54D1481d30ACB0Dbead",
      gasPrice: "0x64"
    }
  },
  compilers: {
    solc: {
      version: "^0.6.12",
    }
  },
}
