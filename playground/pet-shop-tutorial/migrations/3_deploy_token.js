var URC20 = artifacts.require("URC20Token");

module.exports = async function(deployer) {
  await deployer.deploy(URC20);
};
