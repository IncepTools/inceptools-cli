package db_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"incepttools/src/db" // Import the package under test

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// openTestDB returns a GORM DB connection based on INCEPTOOLS_TEST_DIALECT env.
// Defaults to in-memory SQLite when the env is unset.
func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dialect := os.Getenv("INCEPTOOLS_TEST_DIALECT")
	dsn := os.Getenv("INCEPTOOLS_TEST_DSN")

	var dialector gorm.Dialector
	switch dialect {
	case "postgres":
		if dsn == "" {
			dsn = "host=localhost user=postgres password=postgres dbname=inceptools_test port=5432 sslmode=disable"
		}
		dialector = postgres.Open(dsn)
	case "mysql":
		if dsn == "" {
			dsn = "root:root@tcp(127.0.0.1:3306)/inceptools_test?charset=utf8mb4&parseTime=True&loc=Local"
		}
		dialector = mysql.Open(dsn)
	default:
		// SQLite in-memory — each call gets an isolated DB
		dialector = sqlite.Open(":memory:")
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open test database (%s): %v", dialect, err)
	}
	return db
}

// cleanTable drops the migration tracking table between tests.
func cleanTable(t *testing.T, db *gorm.DB, tableName string) {
	t.Helper()
	db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName))
}

// createTempMigrations writes migration SQL files into a temp directory and
// returns the directory path, which is cleaned up when the test finishes.
func createTempMigrations(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write migration file %s: %v", name, err)
		}
	}
	return dir
}

// testConfig returns a Config using a unique table name per test to avoid
// collisions when using a shared DB (Postgres/MySQL).
func testConfig(t *testing.T, migrationPath string) db.Config {
	t.Helper()
	return db.Config{
		TableName:     "test_migrations",
		MigrationPath: migrationPath,
	}
}

// countRecords returns the number of rows in the given table with the given status.
func countRecords(t *testing.T, gdb *gorm.DB, tableName, status string) int64 {
	t.Helper()
	var count int64
	gdb.Table(tableName).Where("status = ?", status).Count(&count)
	return count
}

// tableExists checks whether a table exists in the database.
func tableExists(db *gorm.DB, table string) bool {
	return db.Migrator().HasTable(table)
}

// ---------------------------------------------------------------------------
// LoadLocalMigrations — 4 tests
// ---------------------------------------------------------------------------

func TestLoadLocalMigrations_ValidFiles(t *testing.T) {
	dir := createTempMigrations(t, map[string]string{
		"000001_create_users.up.sql":   "CREATE TABLE users (id INTEGER PRIMARY KEY);",
		"000001_create_users.down.sql": "DROP TABLE IF EXISTS users;",
		"000002_add_email.up.sql":      "ALTER TABLE users ADD COLUMN email TEXT;",
		"000002_add_email.down.sql":    "ALTER TABLE users DROP COLUMN email;",
	})

	migs, err := db.LoadLocalMigrations(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(migs) != 2 {
		t.Fatalf("expected 2 migrations, got %d", len(migs))
	}
	if migs["000001"].Name != "create_users" {
		t.Errorf("expected name 'create_users', got %q", migs["000001"].Name)
	}
	if migs["000002"].UpScript == "" {
		t.Error("expected up script for 000002 to be populated")
	}
}

func TestLoadLocalMigrations_SkipsNonSQL(t *testing.T) {
	dir := createTempMigrations(t, map[string]string{
		"000001_users.up.sql":   "CREATE TABLE users (id INTEGER PRIMARY KEY);",
		"000001_users.down.sql": "DROP TABLE IF EXISTS users;",
		"README.md":             "# Migrations",
		"notes.txt":             "just some notes",
	})
	os.Mkdir(filepath.Join(dir, "subdir"), 0755)

	migs, err := db.LoadLocalMigrations(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(migs) != 1 {
		t.Fatalf("expected 1 migration (non-SQL files skipped), got %d", len(migs))
	}
}

func TestLoadLocalMigrations_MissingDir(t *testing.T) {
	_, err := db.LoadLocalMigrations("/tmp/nonexistent_migration_dir_xyz_123")
	if err == nil {
		t.Fatal("expected error for missing directory, got nil")
	}
}

func TestLoadLocalMigrations_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	migs, err := db.LoadLocalMigrations(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(migs) != 0 {
		t.Fatalf("expected 0 migrations for empty dir, got %d", len(migs))
	}
}

// ---------------------------------------------------------------------------
// RunMigrations — 10 tests
// ---------------------------------------------------------------------------

func TestRunMigrations_TableCreation(t *testing.T) {
	gdb := openTestDB(t)
	cfg := testConfig(t, t.TempDir())
	cleanTable(t, gdb, cfg.TableName)

	if err := db.RunMigrations(gdb, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !tableExists(gdb, cfg.TableName) {
		t.Error("migration table was not created")
	}
	cleanTable(t, gdb, cfg.TableName)
}

func TestRunMigrations_AppliesSingle(t *testing.T) {
	gdb := openTestDB(t)
	dir := createTempMigrations(t, map[string]string{
		"000001_create_posts.up.sql":   "CREATE TABLE posts (id INTEGER PRIMARY KEY, title TEXT);",
		"000001_create_posts.down.sql": "DROP TABLE IF EXISTS posts;",
	})
	cfg := testConfig(t, dir)
	cleanTable(t, gdb, cfg.TableName)

	if err := db.RunMigrations(gdb, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !tableExists(gdb, "posts") {
		t.Error("expected 'posts' table to exist after migration")
	}
	if n := countRecords(t, gdb, cfg.TableName, db.StatusSuccess); n != 1 {
		t.Errorf("expected 1 success record, got %d", n)
	}

	gdb.Exec("DROP TABLE IF EXISTS posts")
	cleanTable(t, gdb, cfg.TableName)
}

func TestRunMigrations_AppliesMultipleInOrder(t *testing.T) {
	gdb := openTestDB(t)
	dir := createTempMigrations(t, map[string]string{
		"000001_create_a.up.sql":   "CREATE TABLE table_a (id INTEGER PRIMARY KEY);",
		"000001_create_a.down.sql": "DROP TABLE IF EXISTS table_a;",
		"000002_create_b.up.sql":   "CREATE TABLE table_b (id INTEGER PRIMARY KEY);",
		"000002_create_b.down.sql": "DROP TABLE IF EXISTS table_b;",
		"000003_create_c.up.sql":   "CREATE TABLE table_c (id INTEGER PRIMARY KEY);",
		"000003_create_c.down.sql": "DROP TABLE IF EXISTS table_c;",
	})
	cfg := testConfig(t, dir)
	cleanTable(t, gdb, cfg.TableName)

	if err := db.RunMigrations(gdb, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, tbl := range []string{"table_a", "table_b", "table_c"} {
		if !tableExists(gdb, tbl) {
			t.Errorf("expected table %q to exist", tbl)
		}
	}
	if n := countRecords(t, gdb, cfg.TableName, db.StatusSuccess); n != 3 {
		t.Errorf("expected 3 success records, got %d", n)
	}

	for _, tbl := range []string{"table_a", "table_b", "table_c"} {
		gdb.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", tbl))
	}
	cleanTable(t, gdb, cfg.TableName)
}

func TestRunMigrations_SkipsAlreadyApplied(t *testing.T) {
	gdb := openTestDB(t)
	dir := createTempMigrations(t, map[string]string{
		"000001_create_items.up.sql":   "CREATE TABLE items (id INTEGER PRIMARY KEY);",
		"000001_create_items.down.sql": "DROP TABLE IF EXISTS items;",
	})
	cfg := testConfig(t, dir)
	cleanTable(t, gdb, cfg.TableName)

	if err := db.RunMigrations(gdb, cfg); err != nil {
		t.Fatalf("first run failed: %v", err)
	}
	if err := db.RunMigrations(gdb, cfg); err != nil {
		t.Fatalf("second run failed: %v", err)
	}

	if n := countRecords(t, gdb, cfg.TableName, db.StatusSuccess); n != 1 {
		t.Errorf("expected exactly 1 success record after idempotent run, got %d", n)
	}

	gdb.Exec("DROP TABLE IF EXISTS items")
	cleanTable(t, gdb, cfg.TableName)
}

func TestRunMigrations_ReappliesAfterRollback(t *testing.T) {
	gdb := openTestDB(t)
	dir := createTempMigrations(t, map[string]string{
		"000001_create_tags.up.sql":   "CREATE TABLE tags (id INTEGER PRIMARY KEY, name TEXT);",
		"000001_create_tags.down.sql": "DROP TABLE IF EXISTS tags;",
	})
	cfg := testConfig(t, dir)
	cleanTable(t, gdb, cfg.TableName)

	if err := db.RunMigrations(gdb, cfg); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}
	if err := db.RollbackMigrations(gdb, cfg, 0); err != nil {
		t.Fatalf("rollback failed: %v", err)
	}
	if tableExists(gdb, "tags") {
		t.Error("tags table should not exist after rollback")
	}
	if err := db.RunMigrations(gdb, cfg); err != nil {
		t.Fatalf("re-migrate failed: %v", err)
	}
	if !tableExists(gdb, "tags") {
		t.Error("tags table should exist after re-apply")
	}

	gdb.Exec("DROP TABLE IF EXISTS tags")
	cleanTable(t, gdb, cfg.TableName)
}

func TestRunMigrations_AutoRollbackMissingFile(t *testing.T) {
	gdb := openTestDB(t)
	dir := createTempMigrations(t, map[string]string{
		"000001_create_alpha.up.sql":   "CREATE TABLE alpha (id INTEGER PRIMARY KEY);",
		"000001_create_alpha.down.sql": "DROP TABLE IF EXISTS alpha;",
		"000002_create_beta.up.sql":    "CREATE TABLE beta (id INTEGER PRIMARY KEY);",
		"000002_create_beta.down.sql":  "DROP TABLE IF EXISTS beta;",
	})
	cfg := testConfig(t, dir)
	cleanTable(t, gdb, cfg.TableName)

	if err := db.RunMigrations(gdb, cfg); err != nil {
		t.Fatalf("initial migrate failed: %v", err)
	}
	if !tableExists(gdb, "alpha") || !tableExists(gdb, "beta") {
		t.Fatal("both tables should exist after initial migrate")
	}

	os.Remove(filepath.Join(dir, "000002_create_beta.up.sql"))
	os.Remove(filepath.Join(dir, "000002_create_beta.down.sql"))

	if err := db.RunMigrations(gdb, cfg); err != nil {
		t.Fatalf("sync migrate failed: %v", err)
	}
	if tableExists(gdb, "beta") {
		t.Error("beta table should have been auto-rolled back (file removed)")
	}
	if !tableExists(gdb, "alpha") {
		t.Error("alpha table should still exist")
	}

	gdb.Exec("DROP TABLE IF EXISTS alpha")
	gdb.Exec("DROP TABLE IF EXISTS beta")
	cleanTable(t, gdb, cfg.TableName)
}

func TestRunMigrations_StoresScriptsInDB(t *testing.T) {
	gdb := openTestDB(t)
	upSQL := "CREATE TABLE scripts_test (id INTEGER PRIMARY KEY);"
	downSQL := "DROP TABLE IF EXISTS scripts_test;"
	dir := createTempMigrations(t, map[string]string{
		"000001_scripts_test.up.sql":   upSQL,
		"000001_scripts_test.down.sql": downSQL,
	})
	cfg := testConfig(t, dir)
	cleanTable(t, gdb, cfg.TableName)

	if err := db.RunMigrations(gdb, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var record db.MigrationRecord
	gdb.Table(cfg.TableName).Where("version = ?", "000001").First(&record)
	if record.UpScript != upSQL {
		t.Errorf("expected UpScript=%q, got %q", upSQL, record.UpScript)
	}
	if record.DownScript != downSQL {
		t.Errorf("expected DownScript=%q, got %q", downSQL, record.DownScript)
	}

	gdb.Exec("DROP TABLE IF EXISTS scripts_test")
	cleanTable(t, gdb, cfg.TableName)
}

func TestRunMigrations_DefaultConfig(t *testing.T) {
	gdb := openTestDB(t)
	dir := createTempMigrations(t, map[string]string{
		"000001_default_cfg.up.sql":   "CREATE TABLE default_cfg (id INTEGER PRIMARY KEY);",
		"000001_default_cfg.down.sql": "DROP TABLE IF EXISTS default_cfg;",
	})

	cfg := db.Config{MigrationPath: dir}
	cleanTable(t, gdb, "schema_migrations")

	if err := db.RunMigrations(gdb, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !tableExists(gdb, "schema_migrations") {
		t.Error("expected default table 'schema_migrations' to exist")
	}

	gdb.Exec("DROP TABLE IF EXISTS default_cfg")
	cleanTable(t, gdb, "schema_migrations")
}

func TestRunMigrations_InvalidSQL(t *testing.T) {
	gdb := openTestDB(t)
	dir := createTempMigrations(t, map[string]string{
		"000001_bad_sql.up.sql":   "THIS IS NOT VALID SQL AT ALL;",
		"000001_bad_sql.down.sql": "DROP TABLE IF EXISTS nonexistent;",
	})
	cfg := testConfig(t, dir)
	cleanTable(t, gdb, cfg.TableName)

	err := db.RunMigrations(gdb, cfg)
	if err == nil {
		t.Fatal("expected error for invalid SQL, got nil")
	}

	cleanTable(t, gdb, cfg.TableName)
}

func TestRunMigrations_EmptyDirectory(t *testing.T) {
	gdb := openTestDB(t)
	cfg := testConfig(t, t.TempDir())
	cleanTable(t, gdb, cfg.TableName)

	if err := db.RunMigrations(gdb, cfg); err != nil {
		t.Fatalf("expected no error for empty dir, got: %v", err)
	}
	if n := countRecords(t, gdb, cfg.TableName, db.StatusSuccess); n != 0 {
		t.Errorf("expected 0 records for empty dir, got %d", n)
	}

	cleanTable(t, gdb, cfg.TableName)
}

// ---------------------------------------------------------------------------
// RollbackMigrations — 6 tests
// ---------------------------------------------------------------------------

func TestRollbackMigrations_RollsBackAll(t *testing.T) {
	gdb := openTestDB(t)
	dir := createTempMigrations(t, map[string]string{
		"000001_rb_all_a.up.sql":   "CREATE TABLE rb_all_a (id INTEGER PRIMARY KEY);",
		"000001_rb_all_a.down.sql": "DROP TABLE IF EXISTS rb_all_a;",
		"000002_rb_all_b.up.sql":   "CREATE TABLE rb_all_b (id INTEGER PRIMARY KEY);",
		"000002_rb_all_b.down.sql": "DROP TABLE IF EXISTS rb_all_b;",
	})
	cfg := testConfig(t, dir)
	cleanTable(t, gdb, cfg.TableName)

	db.RunMigrations(gdb, cfg)
	if err := db.RollbackMigrations(gdb, cfg, 0); err != nil {
		t.Fatalf("rollback failed: %v", err)
	}
	for _, tbl := range []string{"rb_all_a", "rb_all_b"} {
		if tableExists(gdb, tbl) {
			t.Errorf("table %q should not exist after rollback all", tbl)
		}
	}

	gdb.Exec("DROP TABLE IF EXISTS rb_all_a")
	gdb.Exec("DROP TABLE IF EXISTS rb_all_b")
	cleanTable(t, gdb, cfg.TableName)
}

func TestRollbackMigrations_RollsBackNSteps(t *testing.T) {
	gdb := openTestDB(t)
	dir := createTempMigrations(t, map[string]string{
		"000001_rb_n_a.up.sql":   "CREATE TABLE rb_n_a (id INTEGER PRIMARY KEY);",
		"000001_rb_n_a.down.sql": "DROP TABLE IF EXISTS rb_n_a;",
		"000002_rb_n_b.up.sql":   "CREATE TABLE rb_n_b (id INTEGER PRIMARY KEY);",
		"000002_rb_n_b.down.sql": "DROP TABLE IF EXISTS rb_n_b;",
		"000003_rb_n_c.up.sql":   "CREATE TABLE rb_n_c (id INTEGER PRIMARY KEY);",
		"000003_rb_n_c.down.sql": "DROP TABLE IF EXISTS rb_n_c;",
	})
	cfg := testConfig(t, dir)
	cleanTable(t, gdb, cfg.TableName)

	db.RunMigrations(gdb, cfg)
	if err := db.RollbackMigrations(gdb, cfg, 1); err != nil {
		t.Fatalf("rollback failed: %v", err)
	}
	if tableExists(gdb, "rb_n_c") {
		t.Error("rb_n_c should have been rolled back")
	}
	if !tableExists(gdb, "rb_n_a") || !tableExists(gdb, "rb_n_b") {
		t.Error("rb_n_a and rb_n_b should still exist")
	}

	for _, tbl := range []string{"rb_n_a", "rb_n_b", "rb_n_c"} {
		gdb.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", tbl))
	}
	cleanTable(t, gdb, cfg.TableName)
}

func TestRollbackMigrations_NoApplied(t *testing.T) {
	gdb := openTestDB(t)
	cfg := testConfig(t, t.TempDir())
	cleanTable(t, gdb, cfg.TableName)

	gdb.Table(cfg.TableName).AutoMigrate(&db.MigrationRecord{})

	if err := db.RollbackMigrations(gdb, cfg, 0); err != nil {
		t.Fatalf("expected nil error for no applied migrations, got: %v", err)
	}

	cleanTable(t, gdb, cfg.TableName)
}

func TestRollbackMigrations_SetsStatusRollback(t *testing.T) {
	gdb := openTestDB(t)
	dir := createTempMigrations(t, map[string]string{
		"000001_status_check.up.sql":   "CREATE TABLE status_check (id INTEGER PRIMARY KEY);",
		"000001_status_check.down.sql": "DROP TABLE IF EXISTS status_check;",
	})
	cfg := testConfig(t, dir)
	cleanTable(t, gdb, cfg.TableName)

	db.RunMigrations(gdb, cfg)
	db.RollbackMigrations(gdb, cfg, 0)

	var record db.MigrationRecord
	gdb.Table(cfg.TableName).Where("version = ?", "000001").Order("id desc").First(&record)
	if record.Status != db.StatusRollback {
		t.Errorf("expected status %q, got %q", db.StatusRollback, record.Status)
	}
	if record.Message != "Manual rollback" {
		t.Errorf("expected message 'Manual rollback', got %q", record.Message)
	}

	gdb.Exec("DROP TABLE IF EXISTS status_check")
	cleanTable(t, gdb, cfg.TableName)
}

func TestRollbackMigrations_DropsTable(t *testing.T) {
	gdb := openTestDB(t)
	dir := createTempMigrations(t, map[string]string{
		"000001_drop_verify.up.sql":   "CREATE TABLE drop_verify (id INTEGER PRIMARY KEY, data TEXT);",
		"000001_drop_verify.down.sql": "DROP TABLE IF EXISTS drop_verify;",
	})
	cfg := testConfig(t, dir)
	cleanTable(t, gdb, cfg.TableName)

	db.RunMigrations(gdb, cfg)
	if !tableExists(gdb, "drop_verify") {
		t.Fatal("drop_verify should exist before rollback")
	}

	db.RollbackMigrations(gdb, cfg, 0)
	if tableExists(gdb, "drop_verify") {
		t.Error("drop_verify should NOT exist after rollback")
	}

	cleanTable(t, gdb, cfg.TableName)
}

func TestRollbackMigrations_FullCycle(t *testing.T) {
	gdb := openTestDB(t)
	dir := createTempMigrations(t, map[string]string{
		"000001_cycle_a.up.sql":   "CREATE TABLE cycle_a (id INTEGER PRIMARY KEY);",
		"000001_cycle_a.down.sql": "DROP TABLE IF EXISTS cycle_a;",
		"000002_cycle_b.up.sql":   "CREATE TABLE cycle_b (id INTEGER PRIMARY KEY);",
		"000002_cycle_b.down.sql": "DROP TABLE IF EXISTS cycle_b;",
	})
	cfg := testConfig(t, dir)
	cleanTable(t, gdb, cfg.TableName)

	if err := db.RunMigrations(gdb, cfg); err != nil {
		t.Fatalf("step 1 migrate failed: %v", err)
	}
	if !tableExists(gdb, "cycle_a") || !tableExists(gdb, "cycle_b") {
		t.Fatal("both tables should exist after step 1")
	}

	if err := db.RollbackMigrations(gdb, cfg, 0); err != nil {
		t.Fatalf("step 2 rollback failed: %v", err)
	}
	if tableExists(gdb, "cycle_a") || tableExists(gdb, "cycle_b") {
		t.Fatal("no tables should exist after step 2 rollback")
	}

	if err := db.RunMigrations(gdb, cfg); err != nil {
		t.Fatalf("step 3 re-migrate failed: %v", err)
	}
	if !tableExists(gdb, "cycle_a") || !tableExists(gdb, "cycle_b") {
		t.Fatal("both tables should exist after step 3 re-migrate")
	}

	if err := db.RollbackMigrations(gdb, cfg, 1); err != nil {
		t.Fatalf("step 4 partial rollback failed: %v", err)
	}
	if !tableExists(gdb, "cycle_a") {
		t.Error("cycle_a should still exist after partial rollback")
	}
	if tableExists(gdb, "cycle_b") {
		t.Error("cycle_b should be gone after partial rollback")
	}

	if err := db.RunMigrations(gdb, cfg); err != nil {
		t.Fatalf("step 5 re-migrate failed: %v", err)
	}
	if !tableExists(gdb, "cycle_a") || !tableExists(gdb, "cycle_b") {
		t.Fatal("both tables should exist after step 5")
	}

	gdb.Exec("DROP TABLE IF EXISTS cycle_a")
	gdb.Exec("DROP TABLE IF EXISTS cycle_b")
	cleanTable(t, gdb, cfg.TableName)
}
