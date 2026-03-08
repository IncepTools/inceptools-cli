package migration

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

// RunMigrations is the main entry point
func RunMigrations(db *gorm.DB, cfg Config) error {
	// 1. Apply Defaults
	if cfg.TableName == "" {
		cfg.TableName = "schema_migrations"
	}
	if cfg.MigrationPath == "" {
		cfg.MigrationPath = "./migrations"
	}

	// 2. Init Migration Table
	// We force the table name using Table() or by setting a distinct model
	if err := db.Table(cfg.TableName).AutoMigrate(&MigrationRecord{}); err != nil {
		return fmt.Errorf("failed to init migration table: %v", err)
	}

	// 3. Load Local Files
	localFiles, err := loadLocalMigrations(cfg.MigrationPath)
	if err != nil {
		return err
	}

	// 4. Load DB History (Group by Version to get latest status)
	var history []MigrationRecord
	if err := db.Table(cfg.TableName).Order("id asc").Find(&history).Error; err != nil {
		return err
	}

	// Map version to its LATEST record
	latestStatus := make(map[string]MigrationRecord)
	for _, h := range history {
		latestStatus[h.Version] = h
	}

	log.Println("--- Starting Migration Sync ---")

	// 5. PHASE A: Handle Missing Files (Auto-Rollback)
	// If it was SUCCESS in DB, but file is gone locally -> Rollback using DB script
	for version, record := range latestStatus {
		if record.Status == StatusSuccess {
			if _, exists := localFiles[version]; !exists {
				log.Printf("Create/Sync: File for version %s missing. Rolling back...", version)

				// Run stored DOWN script
				if err := db.Exec(record.DownScript).Error; err != nil {
					log.Printf("Error rolling back %s: %v", version, err)
					continue // Or return err if you want strict failure
				}

				// Mark as Rollback in DB (Update existing or insert new? Request implies marking)
				// We update the record to reflect current state
				record.Status = StatusRollback
				record.Message = "Auto-rollback: file missing"
				record.AppliedAt = time.Now()
				db.Table(cfg.TableName).Save(&record)

				// Update our local map so Phase B knows it's rolled back
				latestStatus[version] = record
			}
		}
	}

	// 6. PHASE B: Apply New or Pending Migrations
	// Iterate through sorted local files
	var versions []string
	for v := range localFiles {
		versions = append(versions, v)
	}
	sort.Strings(versions)

	for _, v := range versions {
		mig := localFiles[v]
		lastRecord, known := latestStatus[v]

		// Condition to run UP:
		// 1. Never run before (!known)
		// 2. Previously Rolled back (lastRecord.Status == StatusRollback)
		shouldRun := !known || lastRecord.Status == StatusRollback

		if shouldRun {
			log.Printf("Applying Migration: %s - %s", mig.Version, mig.Name)

			// Execute UP
			if err := db.Exec(mig.UpScript).Error; err != nil {
				return fmt.Errorf("failed to apply %s: %v", mig.Version, err)
			}

			// Create NEW Record (History)
			newRecord := MigrationRecord{
				Version:     mig.Version,
				Name:        mig.Name,
				UpScript:    mig.UpScript,
				DownScript:  mig.DownScript,
				Status:      StatusSuccess,
				AppliedAt:   time.Now(),
				Message:     "Applied successfully",
			}
			if err := db.Table(cfg.TableName).Create(&newRecord).Error; err != nil {
				return fmt.Errorf("failed to save migration log for %s: %v", mig.Version, err)
			}
		} else {
			// Already applied
			// log.Printf("Skipping %s (Already Applied)", mig.Version)
		}
	}

	log.Println("--- Migration Sync Complete ---")
	return nil
}

// RollbackMigrations runs down scripts in reverse order. Steps=0 means rollback all.
func RollbackMigrations(db *gorm.DB, cfg Config, steps int) error {
	if cfg.TableName == "" {
		cfg.TableName = "schema_migrations"
	}
	if cfg.MigrationPath == "" {
		cfg.MigrationPath = "./migrations"
	}

	// Load DB history
	var history []MigrationRecord
	if err := db.Table(cfg.TableName).Where("status = ?", StatusSuccess).Order("version desc").Find(&history).Error; err != nil {
		return fmt.Errorf("failed to load migration history: %v", err)
	}

	if len(history) == 0 {
		log.Println("No applied migrations to roll back.")
		return nil
	}

	// Deduplicate: keep only the latest record per version
	seen := make(map[string]bool)
	var toRollback []MigrationRecord
	for _, h := range history {
		if !seen[h.Version] {
			seen[h.Version] = true
			toRollback = append(toRollback, h)
		}
	}

	// Limit to N steps if specified
	if steps > 0 && steps < len(toRollback) {
		toRollback = toRollback[:steps]
	}

	log.Println("--- Starting Migration Rollback ---")

	for _, record := range toRollback {
		log.Printf("Rolling back: %s - %s", record.Version, record.Name)

		if err := db.Exec(record.DownScript).Error; err != nil {
			return fmt.Errorf("failed to rollback %s: %v", record.Version, err)
		}

		record.Status = StatusRollback
		record.Message = "Manual rollback"
		record.AppliedAt = time.Now()
		db.Table(cfg.TableName).Save(&record)
	}

	log.Println("--- Migration Rollback Complete ---")
	return nil
}

// Helper struct for local files
type localMigration struct {
	Version    string
	Name       string
	UpScript   string
	DownScript string
}

func loadLocalMigrations(path string) (map[string]localMigration, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("could not read migration directory: %v", err)
	}

	migs := make(map[string]*localMigration)

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		// Parse filename: 000001_topic.up.sql
		parts := strings.Split(f.Name(), "_")
		if len(parts) < 2 {
			continue
		}
		version := parts[0] // 000001

		// Determine type (up or down)
		isUp := strings.HasSuffix(f.Name(), ".up.sql")
		isDown := strings.HasSuffix(f.Name(), ".down.sql")

		if !isUp && !isDown {
			continue
		}

		if _, ok := migs[version]; !ok {
			// Extract name (everything between version and .up.sql)
			rawName := strings.TrimPrefix(f.Name(), version+"_")
			rawName = strings.TrimSuffix(rawName, ".up.sql")
			rawName = strings.TrimSuffix(rawName, ".down.sql")

			migs[version] = &localMigration{
				Version: version,
				Name:    rawName,
			}
		}

		content, err := os.ReadFile(filepath.Join(path, f.Name()))
		if err != nil {
			return nil, err
		}

		if isUp {
			migs[version].UpScript = string(content)
		} else {
			migs[version].DownScript = string(content)
		}
	}

	// Convert pointer map to value map
	result := make(map[string]localMigration)
	for k, v := range migs {
		result[k] = *v
	}

	return result, nil
}
