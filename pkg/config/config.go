package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Secret string

// Value returns the resolved value of the secret. If the value is in the form SECRET:VARNAME, it loads from the environment.
func (s Secret) Value() string {
	str := string(s)
	if strings.HasPrefix(str, "SECRET:") {
		return os.Getenv(strings.TrimPrefix(str, "SECRET:"))
	}
	return str
}

// Config holds database and migration settings.
type Config struct {
	Host           Secret
	Port           Secret
	User           Secret
	Password       Secret
	DBName         Secret
	SSLMode        string
	MigrationsPath string
}

// LoadConfig loads config from a TOML file and environment variables.
// If the path is relative, it will search from the current working directory and project root.
// Environment variables override file values (e.g. MIGRATOR_HOST, MIGRATOR_DBNAME).
func LoadConfig(path string) (*Config, error) {
	resolvedPath := path
	if !filepath.IsAbs(path) {
		// Try current working directory
		cwd, _ := os.Getwd()
		tryPath := filepath.Join(cwd, path)
		if _, err := os.Stat(tryPath); err == nil {
			resolvedPath = tryPath
		} else {
			// Try project root (one level up from cwd)
			rootPath := filepath.Join(cwd, "..", path)
			if _, err := os.Stat(rootPath); err == nil {
				resolvedPath = rootPath
			}
		}
	}

	v := viper.New()
	v.SetConfigFile(resolvedPath)
	v.SetConfigType("toml")
	v.SetEnvPrefix("migrator")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Validate checks that all required config fields are set (after resolving secrets).
func (c *Config) Validate() error {
	if c.Host.Value() == "" {
		return errors.New("config: host is required")
	}
	if c.Port.Value() == "" {
		return errors.New("config: port is required")
	}
	if c.User.Value() == "" {
		return errors.New("config: user is required")
	}
	if c.DBName.Value() == "" {
		return errors.New("config: dbname is required")
	}
	if c.MigrationsPath == "" {
		return errors.New("config: migrationspath is required")
	}
	return nil
}
