package transformer

import (
	"runtime"

	"github.com/kaonone/eth-rpc-gate/pkg/eth"
	"github.com/kaonone/eth-rpc-gate/pkg/params"
	"github.com/labstack/echo"
)

// Web3ClientVersion implements web3_clientVersion
type Web3ClientVersion struct {
	// *kaon.Kaon
}

func (p *Web3ClientVersion) Method() string {
	return "web3_clientVersion"
}

func (p *Web3ClientVersion) Request(_ *eth.JSONRPCRequest, c echo.Context) (interface{}, *eth.JSONRPCError) {
	return "eth-rpc-gate/" + params.VersionWithGitSha + "/" + runtime.GOOS + "-" + runtime.GOARCH + "/" + runtime.Version(), nil
}

// func (p *Web3ClientVersion) ToResponse(ethresp *kaon.CallContractResponse) *eth.CallResponse {
// 	data := utils.AddHexPrefix(ethresp.ExecutionResult.Output)
// 	kaonresp := eth.CallResponse(data)
// 	return &kaonresp
// }
