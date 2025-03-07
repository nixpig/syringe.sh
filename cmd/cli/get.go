package cli

import (
	"context"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/nixpig/syringe.sh/api"
	"github.com/nixpig/syringe.sh/pkg/ssh"
)

func Get(
	ctx context.Context,
	a api.API,
	decrypt ssh.Cryptor,
	key string,
) (string, error) {
	item, err := a.Get(key)
	if err != nil {
		return "", fmt.Errorf("get '%s' from store: %w", key, err)
	}

	log.Debug("get", "item", item)

	decryptedValue, err := decrypt(item.Value)
	if err != nil {
		return "", fmt.Errorf("decrypt '%s': %w", key, err)
	}

	return decryptedValue, nil
}
