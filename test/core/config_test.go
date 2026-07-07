package core_test

import (
	"github.com/IncepTools/inceptools-cli/src/core"
	"os"
	"testing"
)

func TestLoadConfig_ReturnsDefaultsOnMissingFile(t *testing.T) {
	// Ensure file does not exist
	os.Remove("icpt.json")

	cfg, err := core.LoadConfig()
	if err != nil {
		t.Fatalf("Expected no error on missing file, got %v", err)
	}

	if cfg.EnvVariable != "DATABASE_URL" {
		t.Errorf("Expected default EnvVariable 'DATABASE_URL', got '%s'", cfg.EnvVariable)
	}
	if cfg.MigrationPath != "./migrations" {
		t.Errorf("Expected default MigrationPath './migrations', got '%s'", cfg.MigrationPath)
	}
	if cfg.Dialect != "postgres" {
		t.Errorf("Expected default Dialect 'postgres', got '%s'", cfg.Dialect)
	}
}

func TestLoadConfig_LoadsCorrectly(t *testing.T) {
	content := `{
		"env_variable": "MY_DB_URL",
		"migration_path": "./custom_migrations",
		"dialect": "mysql"
	}`
	err := os.WriteFile("icpt.json", []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("icpt.json")

	cfg, err := core.LoadConfig()
	if err != nil {
		t.Fatalf("Unexpected error loading config: %v", err)
	}

	if cfg.EnvVariable != "MY_DB_URL" {
		t.Errorf("Expected EnvVariable 'MY_DB_URL', got '%s'", cfg.EnvVariable)
	}
	if cfg.MigrationPath != "./custom_migrations" {
		t.Errorf("Expected MigrationPath './custom_migrations', got '%s'", cfg.MigrationPath)
	}
	if cfg.Dialect != "mysql" {
		t.Errorf("Expected Dialect 'mysql', got '%s'", cfg.Dialect)
	}
}

func TestLoadConfig_FillsPartialDefaults(t *testing.T) {
	content := `{"dialect": "sqlite"}`
	err := os.WriteFile("icpt.json", []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("icpt.json")

	cfg, err := core.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Dialect != "sqlite" {
		t.Errorf("Expected Dialect 'sqlite', got '%s'", cfg.Dialect)
	}
	if cfg.EnvVariable != "DATABASE_URL" {
		t.Errorf("Expected default EnvVariable 'DATABASE_URL', got '%s'", cfg.EnvVariable)
	}
}

func TestLoadConfig_MultiDatabase(t *testing.T) {
	content := `{
		"dialect": "postgres",
		"databases": {
			"global": {"env_variable": "DATABASE_URL", "migration_path": "./migrations/global"},
			"noida":  {"env_variable": "NOIDA_DATABASE_URL", "dialect": "mysql"}
		}
	}`
	if err := os.WriteFile("icpt.json", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("icpt.json")

	cfg, err := core.LoadConfig()
	if err != nil {
		t.Fatalf("Unexpected error loading config: %v", err)
	}

	targets := cfg.Targets()
	if len(targets) != 2 {
		t.Fatalf("Expected 2 targets, got %d", len(targets))
	}

	// Targets are sorted by name: global, noida
	if targets[0].Name != "global" || targets[0].EnvVariable != "DATABASE_URL" {
		t.Errorf("Unexpected first target: %+v", targets[0])
	}
	if targets[0].Dialect != "postgres" {
		t.Errorf("Expected global to inherit top-level dialect, got '%s'", targets[0].Dialect)
	}
	if targets[1].Name != "noida" || targets[1].Dialect != "mysql" {
		t.Errorf("Expected noida to override dialect with mysql, got '%s'", targets[1].Dialect)
	}
	// Missing migration_path defaults to ./migrations/<name>
	if targets[1].MigrationPath != "./migrations/noida" {
		t.Errorf("Expected default MigrationPath './migrations/noida', got '%s'", targets[1].MigrationPath)
	}

	// Selection by name
	selected, err := cfg.SelectTargets("noida")
	if err != nil || len(selected) != 1 || selected[0].Name != "noida" {
		t.Errorf("SelectTargets('noida') failed: %v, %+v", err, selected)
	}
	if _, err := cfg.SelectTargets("mumbai"); err == nil {
		t.Error("Expected error selecting unknown database, got nil")
	}
}

func TestTargets_LegacySingleConfig(t *testing.T) {
	cfg := core.ICPTConfig{
		EnvVariable:   "MY_DB_URL",
		MigrationPath: "./migrations",
		Dialect:       "postgres",
	}
	targets := cfg.Targets()
	if len(targets) != 1 {
		t.Fatalf("Expected 1 target, got %d", len(targets))
	}
	if targets[0].Name != "default" || targets[0].EnvVariable != "MY_DB_URL" {
		t.Errorf("Unexpected legacy target: %+v", targets[0])
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	err := os.WriteFile("icpt.json", []byte("{invalid json"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("icpt.json")

	_, err = core.LoadConfig()
	if err == nil {
		t.Error("Expected error on invalid JSON, got nil")
	}
}
