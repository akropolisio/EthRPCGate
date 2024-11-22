<template>
  <div class="hello">
    <div v-if="web3Detected">
      <b-button v-if="kaonConnected">Connected to Kaon Network</b-button>
      <b-button v-else-if="connected" v-on:click="connectToKAON()">Connect to Kaon Network</b-button>
      <b-button v-else v-on:click="connectToWeb3()">Connect</b-button>
    </div>
    <b-button v-else>No Web3 detected - Install metamask</b-button>
  </div>
</template>

<script>
let KAONMainnet = {
  chainId: '0x2ED3', // 11987
  chainName: 'Kaon Mainnet',
  rpcUrls: ['https://mainnet.kaon.one/'],
  blockExplorerUrls: ['https://mainnet.kaon.one/'],
  iconUrls: [
    'https://kaon.one/images/metamask_icon.svg',
    'https://kaon.one/images/metamask_icon.png',
  ],
  nativeCurrency: {
    decimals: 18,
    symbol: 'KAON',
  },
};
let KAONRegTest = {
  chainId: '0x2ED4', // 11988
  chainName: 'Kaon Regtest',
  rpcUrls: ['https://localhost:25991'],
  // blockExplorerUrls: [''],
  iconUrls: [
    'https://kaon.one/images/metamask_icon.svg',
    'https://kaon.one/images/metamask_icon.png',
  ],
  nativeCurrency: {
    decimals: 18,
    symbol: 'KAON',
  },
};
let KAONTestNet = {
  chainId: '0x2ED5', // 11989
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
};
let config = {
  "0x2ED3": KAONMainnet,
  "0x2ED5": KAONTestNet,
  "0x2ED4": KAONRegTest,
};

export default {
  name: 'Web3Button',
  props: {
    msg: String,
    connected: Boolean,
    kaonConnected: Boolean,
  },
  computed: {
    web3Detected: function() {
      return !!this.Web3;
    },
  },
  methods: {
    getChainId: function() {
      return window.kaon.chainId;
    },
    isOnKaonChainId: function() {
      let chainId = this.getChainId();
      return chainId == KAONMainnet.chainId || chainId == KAONTestNet.chainId;
    },
    connectToWeb3: function(){
      if (this.connected) {
        return;
      }
      let self = this;
      window.kaon.request({ method: 'eth_requestAccounts' })
        .then(() => {
          console.log("Emitting web3Connected event");
          let kaonConnected = self.isOnKaonChainId();
          let currentlykaonConnected = self.kaonConnected;
          self.$emit("web3Connected", true);
          if (currentlykaonConnected != kaonConnected) {
            console.log("ChainID matches Kaon Network, not prompting to add network to web3, already connected.");
            self.$emit("kaonConnected", true);
          }
        })
        .catch((e) => {
          console.log("Connecting to web3 failed", arguments, e);
        })
    },
    connectToKAON: function() {
      console.log("Connecting to Kaon Network, current chainID is", this.getChainId());

      let self = this;
      let kaonConfig = config[this.getChainId()] || KAONTestNet;
      console.log("Adding network to Metamask", kaonConfig);
      window.kaon.request({
        method: "wallet_addEthereumChain",
        params: [kaonConfig],
      })
        .then(() => {
          self.$emit("kaonConnected", true);
        })
        .catch(() => {
          console.log("Adding network failed", arguments);
        })
    },
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
</style>
