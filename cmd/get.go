package cmd

import (
	"context"
	"fmt"

	"github.com/nixpig/syringe.sh/internal/store"
	"github.com/nixpig/syringe.sh/pkg/ssh"
)

func Get(
	ctx context.Context,
	storeImpl *store.SqliteStore,
	decrypt ssh.Cryptor,
	key string,
) (string, error) {
	item, err := storeImpl.Get(key)
	if err != nil {
		return "", fmt.Errorf("get '%s' from store: %w", err)
	}

	decryptedValue, err := decrypt(item.Value)
	if err != nil {
		return "", fmt.Errorf("decrypt '%s': %w", key, err)
	}

	return decryptedValue, nil
}
