package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"incepttools/src/db" // Import your local package

	"github.com/glebarez/sqlite"
	"gorm.io/driver/clickhouse"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

// ICPTConfig represents the icpt.json configuration file
type ICPTConfig struct {
	EnvVariable   string `json:"env_variable"`
	MigrationPath string `json:"migration_path"`
	Dialect       string `json:"dialect"`
}

func loadConfig() ICPTConfig {
	data, err := os.ReadFile("icpt.json")
	if err != nil {
		log.Fatal("Failed to read icpt.json: ", err)
	}

	var cfg ICPTConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Fatal("Failed to parse icpt.json: ", err)
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

	return cfg
}

func main() {
	// Flags
	createCmd := flag.NewFlagSet("create", flag.ExitOnError)

	if len(os.Args) < 2 {
		fmt.Println("expected 'create', 'migrate', or 'down' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "create":
		createCmd.Parse(os.Args[2:])
		handleCreate()
	case "migrate":
		dbURL := ""
		if len(os.Args) >= 3 {
			dbURL = os.Args[2]
		}
		handleMigrate(dbURL)
	case "down":
		downCmd := flag.NewFlagSet("down", flag.ExitOnError)
		steps := downCmd.Int("steps", 0, "Number of migrations to roll back (0 = all)")
		downCmd.Parse(os.Args[2:])
		dbURL := ""
		if downCmd.NArg() >= 1 {
			dbURL = downCmd.Arg(0)
		}
		handleDown(dbURL, *steps)
	default:
		fmt.Println("expected 'create', 'migrate', or 'down' subcommands")
		os.Exit(1)
	}
}

// ---------------- CLI: Create New Migration ----------------
func handleCreate() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter migration topic (e.g., add_email_json): ")
	topic, _ := reader.ReadString('\n')
	topic = strings.TrimSpace(topic)
	if topic == "" {
		log.Fatal("Topic cannot be empty")
	}

	dir := "./migrations"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, 0755)
	}

	// Calculate next version
	nextVersion := 1
	files, _ := os.ReadDir(dir)
	for _, f := range files {
		parts := strings.Split(f.Name(), "_")
		if len(parts) > 0 {
			if v, err := strconv.Atoi(parts[0]); err == nil {
				if v >= nextVersion {
					nextVersion = v + 1
				}
			}
		}
	}

	// Pad with zeros (000001)
	verStr := fmt.Sprintf("%06d", nextVersion)

	// File Names
	upName := fmt.Sprintf("%s_%s.up.sql", verStr, topic)
	downName := fmt.Sprintf("%s_%s.down.sql", verStr, topic)

	// Create Files
	createFile(filepath.Join(dir, upName), "-- SQL Up Migration\n")
	createFile(filepath.Join(dir, downName), "-- SQL Down Migration\n")

	fmt.Printf("Created:\n %s\n %s\n", upName, downName)
}

func createFile(path, content string) {
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		log.Fatal(err)
	}
}

// ---------------- CLI: Run Migration ----------------
func handleMigrate(cliURL string) {
	// 1. Load config from icpt.json
	cfg := loadConfig()

	// 2. Resolve DSN: CLI arg > env variable from config
	dsn := cliURL
	if dsn == "" {
		dsn = os.Getenv(cfg.EnvVariable)
	}
	if dsn == "" {
		log.Fatalf("No database URL provided.\n"+
			"  Set %s env var or pass URL as argument.\n"+
			"  Usage: migrate <database_url>", cfg.EnvVariable)
	}

	// 3. Open DB connection using dialect from config
	var dialector gorm.Dialector
	switch strings.ToLower(cfg.Dialect) {
	case "postgres", "pg", "mysql", "sqlite", "sqlserver", "mssql", "clickhouse":
		dialector = getDialector(cfg.Dialect, dsn)
	default:
		log.Fatalf("Unsupported dialect %q in icpt.json.\n"+
			"Supported: postgres, pg, mysql, sqlite, sqlserver, mssql, clickhouse", cfg.Dialect)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect database: ", err)
	}

	// 4. Define Migration Configuration
	migCfg := migration.Config{
		TableName:      "schema_migrations",
		MigrationPath:  cfg.MigrationPath,
	}

	// 5. Run the Logic
	if err := migration.RunMigrations(db, migCfg); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("Migration process finished.")
}

// ---------------- CLI: Run Down Migration ----------------
func handleDown(cliURL string, steps int) {
	cfg := loadConfig()

	dsn := cliURL
	if dsn == "" {
		dsn = os.Getenv(cfg.EnvVariable)
	}
	if dsn == "" {
		log.Fatalf("No database URL provided.\n"+
			"  Set %s env var or pass URL as argument.\n"+
			"  Usage: down <database_url>", cfg.EnvVariable)
	}

	var dialector gorm.Dialector
	switch strings.ToLower(cfg.Dialect) {
	case "postgres", "pg", "mysql", "sqlite", "sqlserver", "mssql", "clickhouse":
		dialector = getDialector(cfg.Dialect, dsn)
	default:
		log.Fatalf("Unsupported dialect %q in icpt.json.", cfg.Dialect)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect database: ", err)
	}

	migCfg := migration.Config{
		TableName:     "schema_migrations",
		MigrationPath: cfg.MigrationPath,
	}

	if err := migration.RollbackMigrations(db, migCfg, steps); err != nil {
		log.Fatalf("Rollback failed: %v", err)
	}

	fmt.Println("Rollback process finished.")
}

// getDialector returns the appropriate GORM dialector based on the dialect string
func getDialector(dialect, dsn string) gorm.Dialector {
	switch strings.ToLower(dialect) {
	case "postgres", "pg":
		return postgres.Open(dsn)
	case "mysql":
		return mysql.Open(dsn)
	case "sqlite":
		return sqlite.Open(dsn)
	case "sqlserver", "mssql":
		return sqlserver.Open(dsn)
	case "clickhouse":
		return clickhouse.Open(dsn)
	default:
		log.Fatalf("Unsupported dialect: %s", dialect)
		return nil
	}
}
