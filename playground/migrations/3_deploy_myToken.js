var MyToken = artifacts.require("./MyToken.sol");

module.exports = async function(deployer) {
  await deployer.deploy(MyToken, 9999999999999, {from: "0x1CE507204a6fC8fd6aA7e54D1481d30ACB0Dbead"});
};
