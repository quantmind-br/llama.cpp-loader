// Package config loads and persists the application TOML config.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// AppConfig is the in-memory representation of the user config.
type AppConfig struct {
	Paths  PathsConfig  `mapstructure:"paths"`
	Models ModelsConfig `mapstructure:"models"`
	UI     UIConfig     `mapstructure:"ui"`
}

type PathsConfig struct {
	ProfilesDir string `mapstructure:"profiles_dir"`
	LogDir      string `mapstructure:"log_dir"`
	StateDir    string `mapstructure:"state_dir"`
}

type ModelsConfig struct {
	SearchPaths []string `mapstructure:"search_paths"`
}

type UIConfig struct {
	DefaultTab  string `mapstructure:"default_tab"`
	Keybindings string `mapstructure:"keybindings"`
}

// DefaultConfigPath returns ~/.config/llama-cpp-loader/config.toml.
func DefaultConfigPath() (string, error) {
	home, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locate user config dir: %w", err)
	}
	return filepath.Join(home, "llama-cpp-loader", "config.toml"), nil
}

// Load reads the config from the default location, creating defaults if missing.
func Load() (AppConfig, error) {
	path, err := DefaultConfigPath()
	if err != nil {
		return AppConfig{}, err
	}
	return LoadFrom(path)
}

// LoadFrom reads the config from the given path. If the file does not exist,
// it is created with defaults.
func LoadFrom(path string) (AppConfig, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("toml")

	applyDefaults(v)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return AppConfig{}, fmt.Errorf("mkdir config dir: %w", err)
		}
		if err := v.SafeWriteConfigAs(path); err != nil {
			return AppConfig{}, fmt.Errorf("write default config: %w", err)
		}
	}

	if err := v.ReadInConfig(); err != nil {
		return AppConfig{}, fmt.Errorf("read config: %w", err)
	}

	var cfg AppConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return AppConfig{}, fmt.Errorf("unmarshal config: %w", err)
	}
	return cfg, nil
}

func applyDefaults(v *viper.Viper) {
	home, _ := os.UserHomeDir()
	v.SetDefault("paths.profiles_dir", filepath.Join(home, ".config", "llama-cpp-loader", "profiles"))
	v.SetDefault("paths.log_dir", filepath.Join(home, ".local", "state", "llama-cpp-loader", "logs"))
	v.SetDefault("paths.state_dir", filepath.Join(home, ".local", "state", "llama-cpp-loader"))
	v.SetDefault("models.search_paths", []string{
		filepath.Join(home, ".lmstudio", "models"),
		filepath.Join(home, "models"),
	})
	v.SetDefault("ui.default_tab", "profiles")
	v.SetDefault("ui.keybindings", "default")
}
