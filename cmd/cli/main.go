package main

import (
	"fmt"
	"os"

	"github.com/nixpig/syringe.sh/cmd/cli/clicmd"
)

const (
	host = "localhost"
	port = 23234
)

// TODO: sort this out!!
func main() {
	if err := clicmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
