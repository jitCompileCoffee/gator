// Package config handles reading, updating, and writing the local
// gator CLI application configuration settings to disk.
package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Config struct {
	DBURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

const configFileName = ".gatorconfig.json"

func Read() (Config, error) {
	var cfg Config
	filePath, err := getConfigFilePath()
	if err != nil {
		return cfg, err
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return cfg, err
	}

	if err := json.Unmarshal(content, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func (cfg *Config) SetUser(username string) error {
	if username == "" {
		return errors.New("no username provided")
	}
	cfg.CurrentUserName = username

	return write(*cfg)
}

func write(cfg Config) error {
	filePath, err := getConfigFilePath()
	if err != nil {
		return err
	}
	jsonData, err := json.MarshalIndent(cfg, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, jsonData, 0o644)
}

func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, configFileName), nil
}
