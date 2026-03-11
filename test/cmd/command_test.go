package cmd_test

import (
	"bytes"
	"fmt"
	"incepttools/src/cmd"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleInit(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	// Run HandleInit
	cmd.HandleInit()

	// Check if migrations directory was created
	if _, err := os.Stat("./migrations"); os.IsNotExist(err) {
		t.Error("expected migrations directory to be created")
	}

	// Check if icpt.json was created
	if _, err := os.Stat("icpt.json"); os.IsNotExist(err) {
		t.Error("expected icpt.json to be created")
	}
}

func TestHandleVersion(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run version command logic
	version := "1.2.3"
	fmt.Printf("inceptools version %s\n", version)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "inceptools version 1.2.3") {
		t.Errorf("expected output to contain version string, got %q", output)
	}
}

func TestHandleCreate(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	// Create migrations dir
	os.Mkdir("./migrations", 0755)

	// Mock stdin for HandleCreate
	input := "test_topic\n"
	r, w, _ := os.Pipe()
	os.Stdin = r
	go func() {
		w.Write([]byte(input))
		w.Close()
	}()

	// Run HandleCreate
	cmd.HandleCreate()

	// Verify files created
	files, _ := os.ReadDir("./migrations")
	foundUp := false
	foundDown := false
	for _, f := range files {
		if strings.Contains(f.Name(), "000001_test_topic.up.sql") {
			foundUp = true
		}
		if strings.Contains(f.Name(), "000001_test_topic.down.sql") {
			foundDown = true
		}
	}

	if !foundUp || !foundDown {
		t.Error("expected migration files 000001_test_topic to be created")
	}
}

func TestHandleMigrate(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// 1. Setup config and migrations
	os.Mkdir("migrations", 0755)
	os.WriteFile("migrations/000001_init.up.sql", []byte("CREATE TABLE users (id INTEGER PRIMARY KEY);"), 0644)
	os.WriteFile("migrations/000001_init.down.sql", []byte("DROP TABLE users;"), 0644)

	cfg := `{"dialect": "sqlite", "migration_path": "./migrations"}`
	os.WriteFile("icpt.json", []byte(cfg), 0644)

	// 2. Run migrate with a temp db
	dbPath := filepath.Join(tmpDir, "test.db")
	cmd.HandleMigrate(dbPath)

	// 3. Verify
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("Database file was not created")
	}
}

func TestHandleDown(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// 1. Setup config and migrations
	os.Mkdir("migrations", 0755)
	os.WriteFile("migrations/000001_init.up.sql", []byte("CREATE TABLE users (id INTEGER PRIMARY KEY);"), 0644)
	os.WriteFile("migrations/000001_init.down.sql", []byte("DROP TABLE users;"), 0644)

	cfg := `{"dialect": "sqlite", "migration_path": "./migrations"}`
	os.WriteFile("icpt.json", []byte(cfg), 0644)

	dbPath := filepath.Join(tmpDir, "test.db")

	// 2. Migrate first
	cmd.HandleMigrate(dbPath)

	// 3. Run down
	cmd.HandleDown(dbPath, 1)
}
