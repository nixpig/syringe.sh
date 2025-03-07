package cli

import (
	"context"
	"fmt"

	"github.com/nixpig/syringe.sh/api"
)

func Delete(ctx context.Context, a api.API, key string) error {
	if err := a.Delete(key); err != nil {
		return fmt.Errorf("delete '%s': %w", key, err)
	}

	return nil
}
