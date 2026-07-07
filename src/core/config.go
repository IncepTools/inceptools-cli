package core

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// DBConfig describes a single database entry inside the "databases" map.
type DBConfig struct {
	EnvVariable   string `json:"env_variable"`
	MigrationPath string `json:"migration_path"`
	Dialect       string `json:"dialect,omitempty"`
}

// ICPTConfig represents the icpt.json configuration file.
//
// Two shapes are supported:
//
// Single database (legacy, still valid):
//
//	{
//	  "env_variable": "DATABASE_URL",
//	  "migration_path": "./migrations",
//	  "dialect": "postgres"
//	}
//
// Multiple databases:
//
//	{
//	  "dialect": "postgres",
//	  "databases": {
//	    "global": {"env_variable": "DATABASE_URL", "migration_path": "./migrations/global"},
//	    "noida":  {"env_variable": "NOIDA_DATABASE_URL", "migration_path": "./migrations/noida"}
//	  }
//	}
//
// Per-database "dialect" overrides the top-level one when set.
type ICPTConfig struct {
	EnvVariable   string              `json:"env_variable,omitempty"`
	MigrationPath string              `json:"migration_path,omitempty"`
	Dialect       string              `json:"dialect,omitempty"`
	Databases     map[string]DBConfig `json:"databases,omitempty"`
}

// Target is a fully resolved database the CLI can operate on.
type Target struct {
	Name          string
	EnvVariable   string
	MigrationPath string
	Dialect       string
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

	// Defaults for the single-database shape. Only applied when no
	// "databases" map is present, so a multi-db config never gains a
	// phantom default target.
	if len(cfg.Databases) == 0 {
		if cfg.EnvVariable == "" {
			cfg.EnvVariable = "DATABASE_URL"
		}
		if cfg.MigrationPath == "" {
			cfg.MigrationPath = "./migrations"
		}
	}
	if cfg.Dialect == "" {
		cfg.Dialect = "postgres"
	}

	return cfg, nil
}

// Targets resolves the configured databases into a deterministic
// (name-sorted) list. A legacy single-database config yields one target
// named "default".
func (c ICPTConfig) Targets() []Target {
	if len(c.Databases) == 0 {
		return []Target{{
			Name:          "default",
			EnvVariable:   c.EnvVariable,
			MigrationPath: c.MigrationPath,
			Dialect:       c.Dialect,
		}}
	}

	names := make([]string, 0, len(c.Databases))
	for name := range c.Databases {
		names = append(names, name)
	}
	sort.Strings(names)

	targets := make([]Target, 0, len(names))
	for _, name := range names {
		db := c.Databases[name]
		dialect := db.Dialect
		if dialect == "" {
			dialect = c.Dialect
		}
		migrationPath := db.MigrationPath
		if migrationPath == "" {
			migrationPath = "./migrations/" + name
		}
		targets = append(targets, Target{
			Name:          name,
			EnvVariable:   db.EnvVariable,
			MigrationPath: migrationPath,
			Dialect:       dialect,
		})
	}
	return targets
}

// SelectTargets filters Targets() by name. An empty name returns all
// targets; an unknown name returns an error listing the valid choices.
func (c ICPTConfig) SelectTargets(name string) ([]Target, error) {
	targets := c.Targets()
	if name == "" {
		return targets, nil
	}
	for _, t := range targets {
		if t.Name == name {
			return []Target{t}, nil
		}
	}
	names := make([]string, 0, len(targets))
	for _, t := range targets {
		names = append(names, t.Name)
	}
	return nil, fmt.Errorf("unknown database %q (configured: %v)", name, names)
}
