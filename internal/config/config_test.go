package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	scenarios := map[string]func(t *testing.T){
		"test get config (success)": testGetConfigSuccess,
	}

	for scenario, fn := range scenarios {
		t.Run(scenario, fn)
	}

}

func testGetConfigSuccess(t *testing.T) {
	config, err := GetConfig("/home/username/.config")

	require.Equal(t, Config{
		ConfigFilePath:   "/home/username/.config/syringe/config",
		DatabaseFilePath: "/home/username/.config/syringe/database.db",
	}, config, "should construct config correctly")

	require.NoError(t, err, "should not return error")
}
