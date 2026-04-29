package cmd

import (
	"bufio"
	"fmt"
	"github.com/IncepTools/inceptools-cli/src/ui"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func HandleCreate() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter migration topic (e.g., add_email_json): ")
	topic, _ := reader.ReadString('\n')
	topic = strings.TrimSpace(topic)
	if topic == "" {
		ui.Error("Topic cannot be empty")
		return
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
	if err := os.WriteFile(filepath.Join(dir, upName), []byte("-- SQL Up Migration\n"), 0644); err != nil {
		ui.Error("Failed to create %s: %v", upName, err)
		return
	}
	if err := os.WriteFile(filepath.Join(dir, downName), []byte("-- SQL Down Migration\n"), 0644); err != nil {
		ui.Error("Failed to create %s: %v", downName, err)
		return
	}

	ui.Success("Created migration files:")
	fmt.Printf("  %s\n  %s\n", upName, downName)
}
