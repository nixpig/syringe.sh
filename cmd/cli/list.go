package cli

import (
	"context"
	"fmt"

	"github.com/nixpig/syringe.sh/api"
	"github.com/nixpig/syringe.sh/internal/items"
)

func list(
	ctx context.Context,
	a api.API,
) ([]items.Item, error) {
	list, err := a.List()
	if err != nil {
		return nil, fmt.Errorf("list of items: %w", err)
	}

	return list, nil
}
