package cmd

import (
	"context"
	"fmt"

	"github.com/nixpig/syringe.sh/internal/store"
	"github.com/nixpig/syringe.sh/pkg/ssh"
)

func Set(
	ctx context.Context,
	storeImpl store.Store,
	encrypt ssh.Cryptor,
	key string,
	value string,
) error {
	encryptedValue, err := encrypt(value)
	if err != nil {
		return fmt.Errorf("encrypt '%s': %w", value, err)
	}

	if err := storeImpl.Set(store.StoreItem{
		Key:   key,
		Value: encryptedValue,
	}); err != nil {
		return fmt.Errorf("set '%s' in store: %w", err)
	}

	return nil
}
