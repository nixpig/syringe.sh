package main

import (
	"fmt"
	"os"

	"github.com/nixpig/syringe.sh/cli/cmd"
)

const (
	host = "localhost"
	port = 23234
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
