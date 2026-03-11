package core

import (
	"encoding/json"
	"os"
)

// ICPTConfig represents the icpt.json configuration file
type ICPTConfig struct {
	EnvVariable   string `json:"env_variable"`
	MigrationPath string `json:"migration_path"`
	Dialect       string `json:"dialect"`
}

func LoadConfig() (ICPTConfig, error) {
	var cfg ICPTConfig
	data, err := os.ReadFile("icpt.json")
	if err != nil {
		// Return defaults if config file is missing
		return ICPTConfig{
			EnvVariable:   "DATABASE_URL",
			MigrationPath: "./migrations",
			Dialect:       "postgres",
		}, nil
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	// Defaults
	if cfg.EnvVariable == "" {
		cfg.EnvVariable = "DATABASE_URL"
	}
	if cfg.MigrationPath == "" {
		cfg.MigrationPath = "./migrations"
	}
	if cfg.Dialect == "" {
		cfg.Dialect = "postgres"
	}

	return cfg, nil
}
