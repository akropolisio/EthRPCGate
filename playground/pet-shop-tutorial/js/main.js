import "core-js/stable"
import "regenerator-runtime/runtime"
import {providers, Contract, ethers} from "ethers"
import {KaonProvider, KaonWallet} from "kaon-ethers-wrapper"
import {utils} from "web3"
var $ = require( "jquery" );
import AdoptionArtifact from './Adoption.json'
import Pets from './pets.json'
window.$ = $;
window.jQuery = $;

let KAONMainnet = {
  chainId: '0x2ED3', // 11987
  chainName: 'KAON Mainnet',
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
  chainName: 'KAON Regtest',
  rpcUrls: ['https://localhost:25996'],
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
  chainName: 'KAON Testnet',
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
  11987: KAONMainnet,
  "0x2ED5": KAONTestNet,
  11989: KAONTestNet,
  "0x2ED4": KAONRegTest,
  11988: KAONRegTest,
};
config[KAONMainnet.chainId] = KAONMainnet;
config[KAONTestNet.chainId] = KAONTestNet;
config[KAONRegTest.chainId] = KAONRegTest;

const metamask = true;
window.App = {
  web3Provider: null,
  contracts: {},
  account: "",

  init: function() {
    // Load pets.
    var petsRow = $('#petsRow');
    var petTemplate = $('#petTemplate');

    for (let i = 0; i < Pets.length; i ++) {
      petTemplate.find('.panel-title').text(Pets[i].name);
      petTemplate.find('img').attr('src', Pets[i].picture);
      petTemplate.find('.pet-breed').text(Pets[i].breed);
      petTemplate.find('.pet-age').text(Pets[i].age);
      petTemplate.find('.pet-location').text(Pets[i].location);
      petTemplate.find('.btn-adopt').attr('pets-id', Pets[i].id);

      petsRow.append(petTemplate.html());
    }

    App.login()
    if (!metamask) {
      return App.initEthers();
    }
    return App.initWeb3();
  },

  getChainId: function() {
    return (window.kaon || {}).chainId || 11989;
  },
  isOnKaonChainId: function() {
    let chainId = this.getChainId();
    return chainId == KAONMainnet.chainId ||
        chainId == KAONTestNet.chainId ||
        chainId == KAONRegTest.chainId;
  },

  initEthers: function() {
    throw "TODO: USE STANDARD ETHERS-JS";
    // let kaonRpcProvider = new KaonProvider((config[this.getChainId()] || {}).rpcUrls[0]);
    // let kaonWallet = new KaonWallet(privKey, kaonRpcProvider);

    // window.kaonWallet = kaonWallet;
    // App.account = kaonWallet.address
    // App.web3Provider = kaonWallet;
    // return App.initContract();
  },

  initWeb3: function() {
    let self = this;
    let kaonConfig = config[this.getChainId()] || KAONRegTest;
    console.log("Adding network to Metamask", kaonConfig);
    window.kaon.request({
      method: "wallet_addEthereumChain",
      params: [kaonConfig],
    })
      .then(() => {
        console.log("Successfully connected to kaon")
        window.kaon.request({ method: 'eth_requestAccounts' })
          .then((accounts) => {
            console.log("Successfully logged into metamask", accounts);
            let kaonConnected = self.isOnKaonChainId();
            let currentlykaonConnected = self.kaonConnected;
            if (accounts && accounts.length > 0) {
              App.account = accounts[0];
            }
            if (currentlykaonConnected != kaonConnected) {
              console.log("ChainID matches KAON, not prompting to add network to web3, already connected.");
            }
            throw "TODO: USE STANDARD ETHERS-JS";
            // let kaonRpcProvider = new KaonProvider(KAONTestNet.rpcUrls[0]);
            // let kaonWallet = new KaonWallet("", kaonRpcProvider);
            // App.account = kaonWallet.address
            // if (!metamask) {
            //   App.web3Provider = kaonWallet;
            // } else {
            //   App.web3Provider = new providers.Web3Provider(window.kaon);
            // }

            return App.initContract();
          })
          .catch((e) => {
            console.log("Connecting to web3 failed", e);
          })
      })
      .catch(() => {
        console.log("Adding network failed", arguments);
      })
  },

  initContract: async function() {
    let chainId = utils.hexToNumber(this.getChainId())
    console.log("chainId", chainId)
    const artifacts = AdoptionArtifact.networks[''+chainId];
    if (!artifacts) {
      alert("Contracts are not deployed on chain " + chainId);
      return
    }
    if (!metamask) {
      App.contracts.Adoption = new Contract(artifacts.address, AdoptionArtifact.abi, App.web3Provider)
    } else {
      App.contracts.Adoption = new Contract(artifacts.address, AdoptionArtifact.abi, App.web3Provider.getSigner())
    }


    // Set the provider for our contract
    // App.contracts.Adoption.setProvider(App.web3Provider);

    // Use our contract to retrieve and mark the adopted pets
    await App.markAdopted();
    return App.bindEvents();
  },

  bindEvents: function() {
    $(document).on('click', '.btn-adopt', App.handleAdopt);
  },

  markAdopted: function(adopters, account) {
    var adoptionInstance;
    return new Promise((resolve, reject) => {
      let deployed = App.contracts.Adoption.deployed();
      deployed.then(function(instance) {
        adoptionInstance = instance;
        return adoptionInstance.getAdopters.call()
          .then(function(adopters) {
            console.log("Current adopters", adopters)
            for (var i = 0; i < adopters.length; i++) {
              const adopter = adopters[i];
              if (adopter !== '0x0000000000000000000000000000000000000000') {
                $('.panel-pet').eq(i).find('button').text('Adopted').attr('disabled', true);
                $('.panel-pet').eq(i).find('.pet-adopter-container').css('display', 'block');
                let adopterLabel = adopter;
                if (adopter === App.account) {
                  adopterLabel = "You"
                }
                $('.panel-pet').eq(i).find('.pet-adopter-address').text(adopterLabel);
              } else {
                $('.panel-pet').eq(i).find('.pet-adopter-container').css('display', 'none');
              }
            }
            resolve()
            console.log("Successfully marked as adopted")
          }).catch(function(err) {
            console.log(err);
            reject(err)
          });
      }).catch(function(err) {
        console.error(err)
      })
    });
  },

  handleAdopt: function(event) {
    event.preventDefault();

    var petId = parseInt($(event.target).data('id'));

    var adoptionInstance;

    App.contracts.Adoption.deployed().then(function(instance) {
      adoptionInstance = instance;

      return adoptionInstance.adopt(petId/*, {from: App.account}*/);
    }).then(function(result) {
      console.log("Successfully adopted")
      return App.markAdopted();
    }).catch(function(err) {
      console.error("Adoption failed", err)
      console.error(err.message);
    });
  },

  login: function() {
  },

  handleLogout: function() {
    localStorage.removeItem("userWalletAddress");

    App.login();
    App.markAdopted();
  }
};

$(function() {
  $(document).ready(function() {
    App.init();
  });
});
