package cmd

import (
	"encoding/json"
	"incepttools/src/core"
	"incepttools/src/ui"
	"os"
)

func HandleInit() {
	ui.Heading("Initializing inceptools project...")

	// 1. Create migrations directory
	dir := "./migrations"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, 0755); err != nil {
			ui.Error("Failed to create migrations directory: %v", err)
			return
		}
		ui.Success("Created migrations directory: ./migrations")
	} else {
		ui.Info("Migrations directory already exists")
	}

	// 2. Create icpt.json
	configFile := "icpt.json"
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		cfg := core.ICPTConfig{
			EnvVariable:   "DATABASE_URL",
			MigrationPath: "./migrations",
			Dialect:       "postgres",
		}
		data, _ := json.MarshalIndent(cfg, "", "  ")
		if err := os.WriteFile(configFile, data, 0644); err != nil {
			ui.Error("Failed to create icpt.json: %v", err)
			return
		}
		ui.Success("Created configuration file: icpt.json")
	} else {
		ui.Info("Configuration file icpt.json already exists")
	}

	ui.Finished("inceptools is ready! You can now use 'inceptools create' to start migrating.")
}
