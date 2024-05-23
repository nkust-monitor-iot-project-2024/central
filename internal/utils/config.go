package utils

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"go.uber.org/fx"
)

var ConfigFxModule = fx.Module("config", fx.Provide(
	fx.Annotate(NewConfig, fx.ParamTags(`name:"initLogger"`))),
)

type Config struct {
	*koanf.Koanf
}

func NewConfig(initLogger *slog.Logger) Config {
	conf := koanf.New(".")

	// Docker friendly
	err := conf.Load(file.Provider("/etc/iotmonitor/config.toml"), toml.Parser())
	if err != nil {
		initLogger.Debug(
			"cannot find config file from /etc",
			slog.String("path", "/etc/iotmonitor/config.toml"),
			slog.String("error", err.Error()),
		)
	}

	// User friendly
	configDir, err := os.UserConfigDir()
	if err != nil {
		initLogger.Debug(
			"cannot find user home directory",
			slog.String("error", err.Error()),
		)
	} else {
		iotMonitorConfigPath := filepath.Join(configDir, "iotmonitor", "config.toml")
		initLogger.Info("finding config file from user directory", slog.String("path", iotMonitorConfigPath))

		err = conf.Load(file.Provider(iotMonitorConfigPath), toml.Parser())
		if err != nil {
			slog.Debug(
				"cannot find config file from user directory",
				slog.String("path", iotMonitorConfigPath),
				slog.String("error", err.Error()),
			)
		}
	}

	// Debug friendly
	err = conf.Load(file.Provider("config.toml"), toml.Parser())
	if err != nil {
		initLogger.Debug(
			"cannot find config file in the current directory",
			slog.String("path", "config.toml"),
			slog.String("error", err.Error()),
		)
	}

	// Env
	err = conf.Load(env.Provider("IOT_MONITOR_", "_", func(s string) string {
		return strings.ToLower(strings.TrimPrefix(s, "IOT_MONITOR_"))
	}), nil)
	if err != nil {
		initLogger.Debug(
			"cannot find environment variables",
			slog.String("error", err.Error()),
		)
	}

	return Config{Koanf: conf}
}
