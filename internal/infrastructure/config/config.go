package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/openmarkers/openmarkers-cli/internal/shared/constants"
)

type Config struct {
	Server         string `json:"server,omitempty"`
	DefaultProfile string `json:"default_profile,omitempty"`
}

func ConfigDir() string {
	if d := os.Getenv("OPENMARKERS_CONFIG_DIR"); d != "" {
		return d
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", constants.ConfigDirName)
}

func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.json")
}

func Load() *Config {
	c := &Config{}
	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		return c
	}
	_ = json.Unmarshal(data, c)
	return c
}

func (c *Config) Save() error {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigPath(), data, 0600)
}

func ResolveServer(flagVal string, cfg *Config) string {
	if flagVal != "" {
		return flagVal
	}
	if env := os.Getenv("OPENMARKERS_SERVER"); env != "" {
		return env
	}
	if cfg != nil && cfg.Server != "" {
		return cfg.Server
	}
	return constants.DefaultServer
}

func ResolveProfile(flagVal string, cfg *Config) string {
	if flagVal != "" {
		return flagVal
	}
	if env := os.Getenv("OPENMARKERS_PROFILE"); env != "" {
		return env
	}
	if cfg != nil && cfg.DefaultProfile != "" {
		return cfg.DefaultProfile
	}
	return ""
}
