package database

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect"
	"github.com/nkust-monitor-iot-project-2024/central/ent"
	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"go.uber.org/fx"

	_ "github.com/lib/pq"
)

var EntFx = fx.Module("ent", fx.Provide(func(lifecycle fx.Lifecycle, config utils.Config) (*ent.Client, error) {
	host := config.String("postgres.host")
	if host == "" {
		return nil, errors.New("postgres.host is required")
	}

	port := config.String("postgres.port")
	if port == "" {
		port = "5432"
	}

	user := config.String("postgres.user")
	if user == "" {
		return nil, errors.New("postgres.user is required")
	}

	password := config.String("postgres.password")
	if password == "" {
		return nil, errors.New("postgres.password is required")
	}

	dbname := config.String("postgres.dbname")
	if dbname == "" {
		return nil, errors.New("postgres.dbname is required")
	}

	sslmode := config.String("postgres.sslmode")
	if sslmode == "" {
		sslmode = "disable"
	}

	client, err := ent.Open(dialect.Postgres, fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbname, sslmode))
	if err != nil {
		return nil, fmt.Errorf("open connection to postgres: %w", err)
	}

	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Run the auto migration tool.
			if err := client.Schema.Create(ctx); err != nil {
				return fmt.Errorf("failed creating schema resources: %w", err)
			}

			return nil
		},
		OnStop: func(context.Context) error {
			return client.Close()
		},
	})

	return client, nil
}))
