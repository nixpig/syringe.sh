package cli

import (
	"context"
	"fmt"

	"github.com/nixpig/syringe.sh/api"
	"github.com/nixpig/syringe.sh/internal/items"
	"github.com/nixpig/syringe.sh/pkg/ssh"
)

func get(
	ctx context.Context,
	a api.API,
	decrypt ssh.Cryptor,
	key string,
) (*items.Item, error) {
	item, err := a.Get(key)
	if err != nil {
		return nil, fmt.Errorf("get '%s' from store: %w", key, err)
	}

	decryptedValue, err := decrypt(item.Value)
	if err != nil {
		return nil, fmt.Errorf("decrypt '%s': %w", key, err)
	}

	return items.New(item.Key, decryptedValue), nil
}
