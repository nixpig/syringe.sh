package config

import (
	"os"
	"path"
)

type Config struct {
	ConfigFilePath   string
	DatabaseFilePath string
}

func GetConfig() (*Config, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	syringeConfigDir := path.Join(userConfigDir, "syringe")

	syringeConfigFilePath := path.Join(syringeConfigDir, "config")
	syringeDatabaseFilePath := path.Join(syringeConfigDir, "database.db")

	return &Config{
		ConfigFilePath:   syringeConfigFilePath,
		DatabaseFilePath: syringeDatabaseFilePath,
	}, nil
}
