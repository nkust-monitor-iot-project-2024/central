package database

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect"
	"github.com/nkust-monitor-iot-project-2024/central/ent"
	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"go.uber.org/fx"
)

var EntFx = fx.Module("ent", fx.Provide(func(lifecycle fx.Lifecycle, config utils.Config) (*ent.Client, error) {
	dataSourceName := config.String("postgres.dataSourceName")
	if dataSourceName == "" {
		return nil, errors.New("postgres.dataSourceName is required")
	}

	client, err := ent.Open(dialect.Postgres, dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("open connection to postgres: %w", err)
	}

	lifecycle.Append(fx.Hook{
		OnStop: func(context.Context) error {
			return client.Close()
		},
	})

	// Run the auto migration tool.
	if err := client.Schema.Create(context.Background()); err != nil {
		return nil, fmt.Errorf("failed creating schema resources: %w", err)
	}

	return client, nil
}))
