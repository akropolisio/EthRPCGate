const artifacts = require('./build/contracts/SimpleStore.json');
const contract = require('truffle-contract');
const SimpleStore = contract(artifacts);
SimpleStore.setProvider(web3.currentProvider);

function testGet(store) {
  return store.get().then(function(res) {
    console.log("exec: store.get()")
    console.log("value: ", res.toNumber());
  })
}

function testSet(store) {
  var newVal = Math.floor((Math.random() * 1000) + 1);
  console.log(`exec: store.set(${newVal})`)
  return store.set(newVal, {from: "0x1CE507204a6fC8fd6aA7e54D1481d30ACB0Dbead"}).then(function(res) {
    console.log("receipt: ", res)
  }).catch(function(e) {
    console.log(e)
  })
}

var store;
SimpleStore.deployed().then(function(i) {
  store = i;
}).then(function() {
  return testGet(store)
}).then(function() {
  return testSet(store)
}).then(function() {
  return testGet(store)
}).catch(function(e) {
  console.log(e)
})
