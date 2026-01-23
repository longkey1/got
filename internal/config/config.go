package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/viper"
)

const (
	DefaultGolangUrl  = "https://golang.org"
	DefaultGorootsDir = "goroots"
	DefaultTempDir    = "tmp"
)

type Config struct {
	GolangUrl  string   `mapstructure:"golang_url"`
	GorootsDir string   `mapstructure:"goroots_dir"`
	TempDir    string   `mapstructure:"temp_dir"`
	Versions   []string `mapstructure:"versions"`
}

// DefaultConfigPath returns the default configuration directory path.
func DefaultConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config dir: %w", err)
	}

	if runtime.GOOS == "darwin" {
		configDir = os.Getenv("XDG_CONFIG_HOME")
		if configDir == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("failed to get user home dir: %w", err)
			}

			if home == "" {
				return "", fmt.Errorf("$XDG_CONFIG_HOME or $HOME are not defined")
			}
			configDir = filepath.Join(home, ".config")
		}
	}

	return filepath.Join(configDir, "got"), nil
}

// Load loads the configuration from the specified file or default location.
func Load(cfgFile string) (*Config, error) {
	defaultPath, err := DefaultConfigPath()
	if err != nil {
		return nil, err
	}

	// SetDefault must be called before ReadInConfig
	viper.SetDefault("golang_url", DefaultGolangUrl)
	viper.SetDefault("goroots_dir", filepath.Join(defaultPath, DefaultGorootsDir))
	viper.SetDefault("temp_dir", filepath.Join(defaultPath, DefaultTempDir))
	viper.SetDefault("versions", []string{})

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in config directory with name "config" (without extension).
		viper.AddConfigPath(defaultPath)
		viper.SetConfigType("toml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	err = viper.ReadInConfig()
	if err != nil {
		// If --config flag was explicitly set, fail on error
		if cfgFile != "" {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Otherwise, it's okay if the default config file doesn't exist
	} else {
		_, _ = fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
