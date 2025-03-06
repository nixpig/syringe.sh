package cmd

import (
	"context"
	"fmt"

	"github.com/nixpig/syringe.sh/internal/store"
)

func Delete(ctx context.Context, storeImpl store.Store, key string) error {
	if err := storeImpl.Delete(key); err != nil {
		return fmt.Errorf("delete '%s': %w", key, err)
	}

	return nil
}
