package main

import (
	"fmt"
	"os"
)

const (
	host = "localhost"
	port = 23234
)

// TODO: sort this out!!
func main() {
	if err := execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
