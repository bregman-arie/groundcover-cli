package gc

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type DiskConfig struct {
	APIKey    string `json:"api_key"`
	BackendID string `json:"backend_id"`
	BaseURL   string `json:"base_url"`
}

func DefaultConfigPath() (string, error) {
	base := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME"))
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home dir: %w", err)
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "groundcover-cli", "config.json"), nil
}

func LoadDiskConfig(path string) (DiskConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return DiskConfig{}, err
	}
	var dc DiskConfig
	if err := json.Unmarshal(data, &dc); err != nil {
		return DiskConfig{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return dc, nil
}

func WriteDiskConfig(path string, dc DiskConfig) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("config path is empty")
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create config dir %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(dc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	data = append(data, '\n')

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("write temp config %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("persist config %s: %w", path, err)
	}
	return nil
}

func ApplyDiskConfig(cfg Config, dc DiskConfig) Config {
	if strings.TrimSpace(cfg.APIKey) == "" {
		cfg.APIKey = strings.TrimSpace(dc.APIKey)
	}
	if strings.TrimSpace(cfg.BackendID) == "" {
		cfg.BackendID = strings.TrimSpace(dc.BackendID)
	}
	if strings.TrimSpace(cfg.BaseURL) == "" {
		cfg.BaseURL = strings.TrimSpace(dc.BaseURL)
	}
	return cfg
}
