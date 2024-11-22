# Simple VUE project to switch to Kaon network via Metamask

## Project setup
```
npm install
```

### Compiles and hot-reloads for development
```
npm run serve
```

### Compiles and minifies for production
```
npm run build
```

### Customize configuration
See [Configuration Reference](https://cli.vuejs.org/config/).

### wallet_addEthereumChain
```
// request account access
window.kaon.request({ method: 'eth_requestAccounts' })
    .then(() => {
        // add chain
        window.kaon.request({
            method: "wallet_addEthereumChain",
            params: [{
                {
                    chainId: '0x2ED5',
                    chainName: 'Kaon Testnet',
                    rpcUrls: ['https://testnet.kaon.one/'],
                    blockExplorerUrls: ['https://testnet.kaon.one/'],
                    iconUrls: [
                        'https://kaon.one/images/metamask_icon.svg',
                        'https://kaon.one/images/metamask_icon.png',
                    ],
                    nativeCurrency: {
                        decimals: 18,
                        symbol: 'KAON',
                    },
                }
            }],
        }
    });
```

# Known issues
- Metamask requires https for `rpcUrls` so that must be enabled
  - Either directly through eth-rpc-gate with `--https-key ./path --https-cert ./path2` see [SSL](../README.md#ssl)
  - Through the Makefile `make docker-configure-https && make run-ethrpcgate-https`
  - Or do it yourself with a proxy (eg, nginx)
