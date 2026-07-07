package cmd

import (
	"os"

	"github.com/IncepTools/inceptools-cli/src/core"
	"github.com/IncepTools/inceptools-cli/src/db"
	"github.com/IncepTools/inceptools-cli/src/ui"

	"gorm.io/gorm"
)

// openTarget resolves a target's DSN and opens the GORM connection.
// cliURL (when non-empty) overrides the env variable.
func openTarget(target core.Target, cliURL string) (*gorm.DB, error) {
	dsn := cliURL
	if dsn == "" {
		dsn = os.Getenv(target.EnvVariable)
	}
	if dsn == "" {
		ui.Error("No database URL provided for %q.\n  Set %s env var or pass URL as argument.", target.Name, target.EnvVariable)
		return nil, os.ErrNotExist
	}

	dialector, err := db.GetDialector(target.Dialect, dsn)
	if err != nil {
		ui.Error("%v", err)
		return nil, err
	}

	gdb, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		ui.Error("Failed to connect database %q: %v", target.Name, err)
		return nil, err
	}
	return gdb, nil
}

// HandleMigrate runs pending migrations. dbName selects a single configured
// database; when empty, every configured database is migrated in turn.
func HandleMigrate(cliURL string, dbName string) {
	cfg, err := core.LoadConfig()
	if err != nil {
		ui.Error("Failed to load configuration: %v", err)
		return
	}

	targets, err := cfg.SelectTargets(dbName)
	if err != nil {
		ui.Error("%v", err)
		return
	}

	if cliURL != "" && len(targets) > 1 {
		ui.Error("A database URL argument is ambiguous with multiple databases configured.\n  Select one with: migrate -db <name> <database_url>")
		return
	}

	for _, target := range targets {
		if len(targets) > 1 || target.Name != "default" {
			ui.Heading("Migrating database: %s", target.Name)
		}

		gdb, err := openTarget(target, cliURL)
		if err != nil {
			return
		}

		migCfg := db.Config{
			TableName:     "schema_migrations",
			MigrationPath: target.MigrationPath,
		}

		if err := db.RunMigrations(gdb, migCfg); err != nil {
			ui.Error("Migration failed for %q: %v", target.Name, err)
			return
		}
	}

	ui.Finished("Migration process finished.")
}

// HandleDown rolls back migrations. With multiple databases configured,
// dbName is required so a rollback never fans out across databases by
// accident.
func HandleDown(cliURL string, dbName string, steps int) {
	cfg, err := core.LoadConfig()
	if err != nil {
		ui.Error("Failed to load configuration: %v", err)
		return
	}

	if dbName == "" && len(cfg.Databases) > 1 {
		ui.Error("Multiple databases configured; rollback requires an explicit target.\n  Usage: down -db <name> [-steps N]")
		return
	}

	targets, err := cfg.SelectTargets(dbName)
	if err != nil {
		ui.Error("%v", err)
		return
	}
	target := targets[0]

	if target.Name != "default" {
		ui.Heading("Rolling back database: %s", target.Name)
	}

	gdb, err := openTarget(target, cliURL)
	if err != nil {
		return
	}

	migCfg := db.Config{
		TableName:     "schema_migrations",
		MigrationPath: target.MigrationPath,
	}

	if err := db.RollbackMigrations(gdb, migCfg, steps); err != nil {
		ui.Error("Rollback failed for %q: %v", target.Name, err)
		return
	}

	ui.Finished("Rollback process finished.")
}
