package main

import (
	"context"

	"github.com/charmbracelet/log"
	cmd "github.com/nixpig/syringe.sh/cmd/cli"
	"github.com/spf13/viper"
)

func main() {
	v := viper.New()
	// if err := initialiseConfig(os.Getenv("SYRINGE_CONFIG_PATH"), v); err != nil {
	// 	log.Fatal(err)
	// }

	if err := cmd.New(v).ExecuteContext(context.Background()); err != nil {
		log.Fatal("execute command", "err", err)
	}
}
