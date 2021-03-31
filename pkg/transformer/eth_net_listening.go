package transformer

import (
	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"
)

// ProxyETHGetCode implements ETHProxy
type ProxyNetListening struct {
	*qtum.Qtum
}

func (p *ProxyNetListening) Method() string {
	return "net_listening"
}

func (p *ProxyNetListening) Request(rawreq *eth.JSONRPCRequest) (interface{}, error) {
	p.GetDebugLogger().Log("method", p.Method(), "msg", "Is hardcoded true")
	return true, nil
}
