package cli

import (
	"context"
	"fmt"

	"github.com/nixpig/syringe.sh/api"
	"github.com/nixpig/syringe.sh/pkg/ssh"
	"github.com/nixpig/syringe.sh/stores"
)

func Set(
	ctx context.Context,
	a api.API,
	encrypt ssh.Cryptor,
	key string,
	value string,
) error {
	encryptedValue, err := encrypt(value)
	if err != nil {
		return fmt.Errorf("encrypt '%s': %w", value, err)
	}

	if err := a.Set(stores.StoreItem{
		Key:   key,
		Value: encryptedValue,
	}); err != nil {
		return fmt.Errorf("set '%s' in store: %w", key, err)
	}

	return nil
}
