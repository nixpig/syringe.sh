package cli

import (
	"context"
	"fmt"

	"github.com/nixpig/syringe.sh/api"
	"github.com/nixpig/syringe.sh/internal/items"
	"github.com/nixpig/syringe.sh/pkg/ssh"
)

func Set(
	ctx context.Context,
	a api.API,
	encrypt ssh.Cryptor,
	item *items.Item,
) error {
	encryptedValue, err := encrypt(item.Value)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	if err := a.Set(items.New(item.Key, encryptedValue)); err != nil {
		return fmt.Errorf("set '%s' in store: %w", item.Key, err)
	}

	return nil
}
