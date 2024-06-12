package config

import (
	"context"
	"log/slog"
	"strings"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common"
	runesconfig "github.com/gaze-network/indexer-network/modules/runes/config"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
	"github.com/gaze-network/indexer-network/pkg/middleware/requestcontext"
	"github.com/gaze-network/indexer-network/pkg/middleware/requestlogger"
	"github.com/gaze-network/indexer-network/pkg/reportingclient"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	isInit bool
	mu     sync.Mutex
	config = &Config{
		Logger: logger.Config{
			Output: "TEXT",
		},
		Network: common.NetworkMainnet,
		HTTPServer: HTTPServerConfig{
			Port: 8080,
		},
		BitcoinNode: BitcoinNodeClient{
			User: "user",
			Pass: "pass",
		},
		Modules: Modules{
			Runes: runesconfig.Config{
				Datasource: "bitcoin-node",
				Database:   "postgres",
			},
		},
	}
)

type Config struct {
	EnableModules []string               `mapstructure:"enable_modules"`
	APIOnly       bool                   `mapstructure:"api_only"`
	Logger        logger.Config          `mapstructure:"logger"`
	BitcoinNode   BitcoinNodeClient      `mapstructure:"bitcoin_node"`
	Network       common.Network         `mapstructure:"network"`
	HTTPServer    HTTPServerConfig       `mapstructure:"http_server"`
	Modules       Modules                `mapstructure:"modules"`
	Reporting     reportingclient.Config `mapstructure:"reporting"`
}

type BitcoinNodeClient struct {
	Host       string `mapstructure:"host"`
	User       string `mapstructure:"user"`
	Pass       string `mapstructure:"pass"`
	DisableTLS bool   `mapstructure:"disable_tls"`
}

type Modules struct {
	Runes runesconfig.Config `mapstructure:"runes"`
}

type HTTPServerConfig struct {
	Port      int                               `mapstructure:"port"`
	Logger    requestlogger.Config              `mapstructure:"logger"`
	RequestIP requestcontext.WithClientIPConfig `mapstructure:"requestip"`
}

// Parse parse the configuration from environment variables
func Parse(configFile ...string) Config {
	mu.Lock()
	defer mu.Unlock()
	return parse(configFile...)
}

// Load returns the loaded configuration
func Load() Config {
	mu.Lock()
	defer mu.Unlock()
	if isInit {
		return *config
	}
	return parse()
}

// BindPFlag binds a specific key to a pflag (as used by cobra).
// Example (where serverCmd is a Cobra instance):
//
//	serverCmd.Flags().Int("port", 1138, "Port to run Application server on")
//	Viper.BindPFlag("port", serverCmd.Flags().Lookup("port"))
func BindPFlag(key string, flag *pflag.Flag) {
	if err := viper.BindPFlag(key, flag); err != nil {
		logger.Panic("Something went wrong, failed to bind flag for config", slog.String("package", "config"), slogx.Error(err))
	}
}

// SetDefault sets the default value for this key.
// SetDefault is case-insensitive for a key.
// Default only used when no value is provided by the user via flag, config or ENV.
func SetDefault(key string, value any) { viper.SetDefault(key, value) }

func parse(configFile ...string) Config {
	ctx := logger.WithContext(context.Background(), slog.String("package", "config"))

	if len(configFile) > 0 && configFile[0] != "" {
		viper.SetConfigFile(configFile[0])
	} else {
		viper.AddConfigPath("./")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if err := viper.ReadInConfig(); err != nil {
		var errNotfound viper.ConfigFileNotFoundError
		if errors.As(err, &errNotfound) {
			logger.WarnContext(ctx, "Config file not found, use default config value", slogx.Error(err))
		} else {
			logger.PanicContext(ctx, "Invalid config file", slogx.Error(err))
		}
	}

	if err := viper.Unmarshal(&config); err != nil {
		logger.PanicContext(ctx, "Something went wrong, failed to unmarshal config", slogx.Error(err))
	}

	isInit = true
	return *config
}
