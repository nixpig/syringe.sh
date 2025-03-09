package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func initialiseConfig(configPath string, v *viper.Viper) error {
	if configPath == "" {
		userConfigDir, err := os.UserConfigDir()
		if err != nil {
			return fmt.Errorf("get user config dir: %w", err)
		}

		configPath = filepath.Join(userConfigDir, "syringe")
	}

	if err := os.MkdirAll(configPath, os.ModePerm); err != nil {
		return err
	}

	configFile, err := os.OpenFile(
		filepath.Join(configPath, "settings"),
		os.O_RDWR|os.O_CREATE,
		0666,
	)
	if err != nil {
		return fmt.Errorf("open config file (%s): %w", configPath, err)
	}
	configFile.Close()

	v.SetConfigFile(configFile.Name())
	v.SetConfigType("env")
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("read in config: %w", err)
	}

	return nil
}
