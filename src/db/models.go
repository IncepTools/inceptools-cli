package db

import "time"

// Config holds migration configuration
type Config struct {
	TableName     string
	MigrationPath string
}

// MigrationRecord represents a migration record in the database
type MigrationRecord struct {
	ID          uint      `gorm:"primaryKey"`
	Version     string    `gorm:"index"`
	Name        string
	UpScript    string    `gorm:"type:text"`
	DownScript  string    `gorm:"type:text"`
	Status      string    `gorm:"index"`
	AppliedAt   time.Time
	Message     string
}

// Status constants
const (
	StatusSuccess  = "success"
	StatusRollback = "rollback"
	StatusFailed   = "failed"
)
