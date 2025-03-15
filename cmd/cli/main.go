package main

import (
	"context"
	"os"

	"github.com/nixpig/syringe.sh/internal/cli"
	"github.com/spf13/viper"
)

func main() {
	v := viper.New()
	// TODO: sort out config properly
	// if err := initialiseConfig(os.Getenv("SYRINGE_CONFIG_PATH"), v); err != nil {
	// 	log.Fatal(err)
	// }

	ctx := context.Background()

	if err := cli.New(v).ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
