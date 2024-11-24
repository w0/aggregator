package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const configFile = ".gatorconfig.json"

type Config struct {
	DbURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() (Config, error) {
	configPath, err := testPath()

	if err != nil {
		return Config{}, err
	}

	// open file for reading
	f, err := os.Open(configPath)

	if err != nil {
		return Config{}, err
	}

	//close file
	defer f.Close()

	// new decoder from open file
	d := json.NewDecoder(f)

	var cfg Config

	// decode json
	err = d.Decode(&cfg)

	if err != nil {
		return Config{}, err
	}

	return cfg, nil

}

func (cfg *Config) SetUser(username string) error {
	cfg.CurrentUserName = username
	return setContent(*cfg)
}

func testPath() (string, error) {
	userHome, err := os.UserHomeDir()

	if err != nil {
		return "", err
	}

	gatorConfig := filepath.Join(userHome, configFile)

	if _, err := os.Stat(gatorConfig); errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	return gatorConfig, nil

}

func setContent(cfg Config) error {
	configPath, err := testPath()

	if err != nil {
		return err
	}

	b, err := json.Marshal(cfg)

	if err != nil {
		return err
	}

	err = os.WriteFile(configPath, b, 0644)

	if err != nil {
		return err
	}

	return nil

}
