package config

import (
	"path"

	"github.com/denisbrodbeck/machineid"
)

const (
	APP_NAME          = "syringe"
	DATABASE_FILENAME = "database.db"
	CONFIG_FILENAME   = "config.cfg"
)

type Config struct {
	ConfigFilePath   string
	DatabaseFilePath string
}

func GetConfig(configDir string) (Config, error) {
	syringeConfigDir := path.Join(configDir, APP_NAME)

	syringeConfigFilePath := path.Join(syringeConfigDir, CONFIG_FILENAME)
	syringeDatabaseFilePath := path.Join(syringeConfigDir, DATABASE_FILENAME)

	return Config{
		ConfigFilePath:   syringeConfigFilePath,
		DatabaseFilePath: syringeDatabaseFilePath,
	}, nil
}

func GetUid() (string, error) {
	return machineid.ProtectedID(APP_NAME)
}
