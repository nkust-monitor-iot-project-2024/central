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

// ConfigFxModule is the fx module for the config that provides the Config.
var ConfigFxModule = fx.Module("config", fx.Provide(NewConfig))

// Config is the global configuration of the application.
type Config struct {
	*koanf.Koanf
}

// NewConfig creates a new Config.
//
// It loads the configuration from "/etc/iotmonitor/config.toml", "<userConfigDir>/iotmonitor/config.toml",
// "config.toml", and the environment variables started with "IOT_MONITOR_".
// The later one has the highest priority.
//
// The environment variables are case-insensitive and converts with the following rules (wip: not sure yet):
//
// - "IOT_MONITOR_FOO_BAR" -> "foo.bar"
// - "IOT_MONITOR_FOO_BAR_BAZ" -> "foo.bar.baz"
//
// The TOML configuration file supports the environment variables interpolation.
// For example:
//
// ```
// foo = "${FOO}"
// bar = "${BAR}"
// ```
//
// If the environment variables "FOO" and "BAR" are set to, for example, "foo" and "bar" correspondingly,
// the configuration will be:
//
// ```
// foo = "foo"
// bar = "bar"
// ```
func NewConfig() Config {
	conf := koanf.New(".")

	parser := NewEnvInterpolation(toml.Parser())

	// Docker friendly
	err := conf.Load(file.Provider("/etc/iotmonitor/config.toml"), parser)
	if err != nil {
		slog.Debug(
			"cannot find config file from /etc",
			slog.String("path", "/etc/iotmonitor/config.toml"),
			slog.String("error", err.Error()),
		)
	}

	// User friendly
	configDir, err := os.UserConfigDir()
	if err != nil {
		slog.Debug(
			"cannot find user home directory",
			slog.String("error", err.Error()),
		)
	} else {
		iotMonitorConfigPath := filepath.Join(configDir, "iotmonitor", "config.toml")
		slog.Info("finding config file from user directory", slog.String("path", iotMonitorConfigPath))

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
		slog.Debug(
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
		slog.Debug(
			"cannot find environment variables",
			slog.String("error", err.Error()),
		)
	}

	return Config{Koanf: conf}
}

// EnvInterpolation is the parser that interpolates the environment variables.
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

// NewEnvInterpolation creates a new EnvInterpolation.
func NewEnvInterpolation(parser koanf.Parser) koanf.Parser {
	return &EnvInterpolation{parser: parser}
}
