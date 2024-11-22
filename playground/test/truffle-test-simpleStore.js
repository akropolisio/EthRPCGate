/*const artifacts = require('./build/contracts/SimpleStore.json');
const contract = require('truffle-contract');*/
const SimpleStore = artifacts.require("SimpleStore");

contract("SimpleStore", async accounts => {

  it("Has been deployed", async () => {
    const simpleStoreDeployed = await SimpleStore.deployed();
    assert(simpleStoreDeployed, "contract has been deployed");
    console.log("Address is: ", simpleStoreDeployed.address)
  });

  it("should return 100", async () => {
    const instance = await SimpleStore.deployed();
    console.log("executing: get(): ");
    const balance = await instance.get();
    assert.equal(balance.toNumber(), 100);
    console.log("value: ", balance.toNumber())
  });

  it("should return 184", async () => {
    const instance = await SimpleStore.deployed();
    console.log("executing: set(150): ");
    await instance.set(184).then((receipt) => { console.log("receipt: ", receipt)});
    const balance = await instance.get();
    assert.equal(balance.toNumber(), 184);
    console.log("value: ", balance.toNumber())
    

  })


});
/*
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
})*/
