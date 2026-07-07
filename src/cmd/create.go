package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/IncepTools/inceptools-cli/src/core"
	"github.com/IncepTools/inceptools-cli/src/ui"
)

// HandleCreate creates a new pair of up/down migration files. dbName selects
// which configured database's migration directory to use; when empty and
// multiple databases are configured, the user is prompted to pick one.
func HandleCreate(dbName string) {
	cfg, err := core.LoadConfig()
	if err != nil {
		ui.Error("Failed to load configuration: %v", err)
		return
	}

	reader := bufio.NewReader(os.Stdin)

	targets, err := cfg.SelectTargets(dbName)
	if err != nil {
		ui.Error("%v", err)
		return
	}

	target := targets[0]
	if len(targets) > 1 {
		fmt.Println("Multiple databases configured:")
		for i, t := range targets {
			fmt.Printf("  %d) %s (%s)\n", i+1, t.Name, t.MigrationPath)
		}
		fmt.Print("Select database [1]: ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)
		if choice != "" {
			idx, err := strconv.Atoi(choice)
			if err != nil || idx < 1 || idx > len(targets) {
				ui.Error("Invalid selection: %s", choice)
				return
			}
			target = targets[idx-1]
		}
	}

	fmt.Print("Enter migration topic (e.g., add_email_json): ")
	topic, _ := reader.ReadString('\n')
	topic = strings.TrimSpace(topic)
	if topic == "" {
		ui.Error("Topic cannot be empty")
		return
	}

	dir := target.MigrationPath
	if dir == "" {
		dir = "./migrations"
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			ui.Error("Failed to create migrations directory %s: %v", dir, err)
			return
		}
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
	if err := os.WriteFile(filepath.Join(dir, upName), []byte("-- SQL Up Migration\n"), 0644); err != nil {
		ui.Error("Failed to create %s: %v", upName, err)
		return
	}
	if err := os.WriteFile(filepath.Join(dir, downName), []byte("-- SQL Down Migration\n"), 0644); err != nil {
		ui.Error("Failed to create %s: %v", downName, err)
		return
	}

	ui.Success("Created migration files in %s:", dir)
	fmt.Printf("  %s\n  %s\n", upName, downName)
}
