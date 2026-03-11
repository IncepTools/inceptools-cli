package cmd

import (
	"incepttools/src/core"
	"incepttools/src/db"
	"incepttools/src/ui"
	"os"

	"gorm.io/gorm"
)

func HandleMigrate(cliURL string) {
	// 1. Load config from icpt.json
	cfg, err := core.LoadConfig()
	if err != nil {
		ui.Error("Failed to load configuration: %v", err)
		return
	}

	// 2. Resolve DSN: CLI arg > env variable from config
	dsn := cliURL
	if dsn == "" {
		dsn = os.Getenv(cfg.EnvVariable)
	}
	if dsn == "" {
		ui.Error("No database URL provided.\n  Set %s env var or pass URL as argument.\n  Usage: migrate <database_url>", cfg.EnvVariable)
		return
	}

	// 3. Open DB connection using dialect from config
	dialector, err := db.GetDialector(cfg.Dialect, dsn)
	if err != nil {
		ui.Error("%v", err)
		return
	}

	gdb, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		ui.Error("Failed to connect database: %v", err)
		return
	}

	// 4. Define Migration Configuration
	migCfg := db.Config{
		TableName:     "schema_migrations",
		MigrationPath: cfg.MigrationPath,
	}

	// 5. Run the Logic
	if err := db.RunMigrations(gdb, migCfg); err != nil {
		ui.Error("Migration failed: %v", err)
		return
	}

	ui.Finished("Migration process finished.")
}

func HandleDown(cliURL string, steps int) {
	cfg, err := core.LoadConfig()
	if err != nil {
		ui.Error("Failed to load configuration: %v", err)
		return
	}

	dsn := cliURL
	if dsn == "" {
		dsn = os.Getenv(cfg.EnvVariable)
	}
	if dsn == "" {
		ui.Error("No database URL provided.\n  Set %s env var or pass URL as argument.\n  Usage: down <database_url>", cfg.EnvVariable)
		return
	}

	dialector, err := db.GetDialector(cfg.Dialect, dsn)
	if err != nil {
		ui.Error("%v", err)
		return
	}

	gdb, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		ui.Error("Failed to connect database: %v", err)
		return
	}

	migCfg := db.Config{
		TableName:     "schema_migrations",
		MigrationPath: cfg.MigrationPath,
	}

	if err := db.RollbackMigrations(gdb, migCfg, steps); err != nil {
		ui.Error("Rollback failed: %v", err)
		return
	}

	ui.Finished("Rollback process finished.")
}
