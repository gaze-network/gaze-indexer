package config

import (
	"log/slog"
	"os"
	"sync"

	"github.com/Cleverse/go-utilities/utils"
	"github.com/caarlos0/env/v10"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/joho/godotenv"
)

const (
	DefaultENVPath = "./.env" // Default path to the .env file
)

var (
	configOnce sync.Once
	config     = &Config{}
)

type Config struct {
	Logger      logger.Config     `envPrefix:"LOGGER_"`
	BitcoinNode BitcoinNodeClient `envPrefix:"BITCOIN_NODE_"`
}

type BitcoinNodeClient struct {
	Host       string `env:"HOST"`
	User       string `env:"USER" default:"user"`
	Pass       string `env:"PASS" default:"pass"`
	DisableTLS bool   `env:"DISABLE_TLS" default:"false"`
}

// LoadConfig loads the configuration from environment variables
//
//   - ENV_PATH: relative path to the .env file (default: .env)
//   - ENV_PREFIX: prefix for environment variables (default is empty), e.g. "APP_" will look for "APP_ENV" instead of "ENV"
func LoadConfig() Config {
	logger := logger.With(slog.String("package", "config"))
	configOnce.Do(func() {
		// Load environment variables from .env file
		envPath := utils.Default(os.Getenv("ENV_PATH"), DefaultENVPath)
		if err := godotenv.Load(envPath); err != nil {
			logger.Warn("failed to load .env file, using environment variables directly",
				slog.Any("error", err),
				slog.String("path", envPath),
			)
		}

		// Env parser options
		opts := env.Options{
			Prefix:                utils.Default(os.Getenv("ENV_PREFIX"), ""),
			UseFieldNameByDefault: true,
			RequiredIfNoDef:       false,
		}

		// Parse environment variables
		if err := env.ParseWithOptions(config, opts); err != nil {
			logger.Error("failed to parse environment variables", slog.Any("error", err))
			panic(err)
		}

		logger.Info("loaded config from environment variables successfully")
	})
	return *config
}
