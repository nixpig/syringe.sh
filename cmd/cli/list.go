package cli

import (
	"context"
	"fmt"

	"github.com/nixpig/syringe.sh/api"
)

func List(
	ctx context.Context,
	a api.API,
) ([]string, error) {
	items, err := a.List()
	if err != nil {
		return nil, fmt.Errorf("list of items: %w", err)
	}

	keys := make([]string, len(items))

	for i, item := range items {
		keys[i] = item.Key
	}

	return keys, nil
}
