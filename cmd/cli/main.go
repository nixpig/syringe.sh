package main

import (
	"context"

	"github.com/charmbracelet/log"
	"github.com/nixpig/syringe.sh/internal/cli"
	"github.com/spf13/viper"
)

func main() {
	v := viper.New()
	// TODO: sort out config properly
	// if err := initialiseConfig(os.Getenv("SYRINGE_CONFIG_PATH"), v); err != nil {
	// 	log.Fatal(err)
	// }

	v.SetDefault("identity", "/home/nixpig/.ssh/id_rsa_test")

	ctx := context.Background()

	if err := cli.New(v).ExecuteContext(ctx); err != nil {
		log.Fatal("execute command", "err", err)
	}
}
