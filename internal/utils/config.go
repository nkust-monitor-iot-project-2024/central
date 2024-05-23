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

	parser := NewEnvInterpolation(toml.Parser())

	// Docker friendly
	err := conf.Load(file.Provider("/etc/iotmonitor/config.toml"), parser)
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

		err = conf.Load(file.Provider(iotMonitorConfigPath), parser)
		if err != nil {
			slog.Debug(
				"cannot find config file from user directory",
				slog.String("path", iotMonitorConfigPath),
				slog.String("error", err.Error()),
			)
		}
	}

	// Debug friendly
	err = conf.Load(file.Provider("config.toml"), parser)
	if err != nil {
		initLogger.Debug(
			"cannot find config file in the current directory",
			slog.String("path", "config.toml"),
			slog.String("error", err.Error()),
		)
	}

	// Env
	err = conf.Load(env.ProviderWithValue("IOT_MONITOR_", "_", func(k string, v string) (string, interface{}) {
		return strings.ToLower(strings.TrimPrefix(k, "IOT_MONITOR_")), os.ExpandEnv(v)
	}), nil)
	if err != nil {
		initLogger.Debug(
			"cannot find environment variables",
			slog.String("error", err.Error()),
		)
	}

	return Config{Koanf: conf}
}

type EnvInterpolation struct {
	parser koanf.Parser
}

func (ei *EnvInterpolation) Marshal(m map[string]interface{}) ([]byte, error) {
	return ei.parser.Marshal(m)
}

func (ei *EnvInterpolation) Unmarshal(bytes []byte) (map[string]interface{}, error) {
	expanded := os.ExpandEnv(string(bytes))

	unmarshalled, err := ei.parser.Unmarshal([]byte(expanded))
	if err != nil {
		return nil, err
	}

	return unmarshalled, nil
}

func NewEnvInterpolation(parser koanf.Parser) koanf.Parser {
	return &EnvInterpolation{parser: parser}
}
