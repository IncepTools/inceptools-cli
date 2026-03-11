package core_test

import (
	"incepttools/src/core"
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
