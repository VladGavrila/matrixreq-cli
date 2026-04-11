package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds the application configuration.
type Config struct {
	URL            string `mapstructure:"url"`
	Token          string `mapstructure:"token"`
	DefaultProject string `mapstructure:"default_project"`
	JiraBaseURL    string `mapstructure:"jira_base_url"`
}

// configDir returns the XDG-compliant config directory for mxreq.
func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".config", "mxreq"), nil
}

// ConfigPath returns the full path to the config file.
func ConfigPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// Load reads configuration from file, environment, and merges with viper.
func Load() (*Config, error) {
	dir, err := configDir()
	if err != nil {
		return nil, err
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(dir)

	viper.SetEnvPrefix("MATRIX")
	viper.AutomaticEnv()
	_ = viper.BindEnv("url")
	_ = viper.BindEnv("token")
	_ = viper.BindEnv("default_project")
	_ = viper.BindEnv("jira_base_url")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &cfg, nil
}

// Save writes the given config to disk.
func Save(cfg *Config) error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	viper.Set("url", cfg.URL)
	viper.Set("token", cfg.Token)
	viper.Set("default_project", cfg.DefaultProject)
	viper.Set("jira_base_url", cfg.JiraBaseURL)

	path := filepath.Join(dir, "config.yaml")
	return viper.WriteConfigAs(path)
}

// Validate checks that required fields are present.
func (c *Config) Validate() error {
	if c.URL == "" {
		return fmt.Errorf("instance URL is required (set via --url, MATRIX_URL, or 'mxreq config init')")
	}
	if c.Token == "" {
		return fmt.Errorf("API token is required (set via --token, MATRIX_TOKEN, or 'mxreq config init')")
	}
	return nil
}
