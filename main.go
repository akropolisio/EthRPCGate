package main

import (
	"log"

	"github.com/kaonone/eth-rpc-gate/cli"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	cli.Run()
}
