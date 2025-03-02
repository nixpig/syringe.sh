package cmd

import (
	"context"
	"fmt"

	"github.com/nixpig/syringe.sh/internal/store"
)

func List(
	ctx context.Context,
	storeImpl store.Store,
) ([]string, error) {
	items, err := storeImpl.List()
	if err != nil {
		return nil, fmt.Errorf("list of items: %w", err)
	}

	keys := make([]string, len(items))

	for i, item := range items {
		keys[i] = item.Key
	}

	return keys, nil
}
