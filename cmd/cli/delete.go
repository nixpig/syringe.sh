package cli

import (
	"context"
	"fmt"

	"github.com/nixpig/syringe.sh/api"
)

func remove(ctx context.Context, a api.API, key string) error {
	if err := a.Remove(key); err != nil {
		return fmt.Errorf("delete '%s': %w", key, err)
	}

	return nil
}
