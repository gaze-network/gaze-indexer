package config

import (
	"context"
	"log/slog"
	"strings"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/internal/postgres"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
	"github.com/spf13/viper"
)

var (
	configOnce sync.Once
	config     = &Config{
		Logger: logger.Config{
			Output: "TEXT",
		},
		BitcoinNode: BitcoinNodeClient{
			User: "user",
			Pass: "pass",
		},
	}
)

type Config struct {
	Logger      logger.Config     `mapstructure:"logger"`
	BitcoinNode BitcoinNodeClient `mapstructure:"bitcoin_node"`
	Network     common.Network    `mapstructure:"network"`
	Modules     map[string]Module `mapstructure:"modules"`
}

type BitcoinNodeClient struct {
	Host       string `mapstructure:"host"`
	User       string `mapstructure:"user"`
	Pass       string `mapstructure:"pass"`
	DisableTLS bool   `mapstructure:"disable_tls"`
}

type Module struct {
	Postgres postgres.Config `mapstructure:"postgres"`
}

// LoadConfig loads the configuration from environment variables
func LoadConfig() Config {
	ctx := logger.WithContext(context.Background(), slog.String("package", "config"))
	configOnce.Do(func() {
		// TODO: Get config file from Args: viper.SetConfigFile("./config.yaml")
		viper.AddConfigPath("./")
		viper.SetConfigName("config")

		viper.AutomaticEnv()
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		if err := viper.ReadInConfig(); err != nil {
			var errNotfound viper.ConfigFileNotFoundError
			if errors.As(err, &errNotfound) {
				logger.WarnContext(ctx, "config file not found, use default value", slogx.Error(err))
			} else {
				logger.PanicContext(ctx, "invalid config file", slogx.Error(err))
			}
		}

		if err := viper.Unmarshal(&config); err != nil {
			logger.PanicContext(ctx, "failed to unmarshal config", slogx.Error(err))
		}
		logger.InfoContext(ctx, "loaded config from environment variables successfully")
	})

	return *config
}
