package main

import (
	"fmt"
	"os"

	"github.com/peaqnetwork/peaq-network-ev-charging-sim-be-p2p/cmd/server"
)

func main() {

	err := server.Run()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
