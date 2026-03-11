package db_test

import (
	"incepttools/src/db"
	"testing"
)

func TestGetDialector_Success(t *testing.T) {
	tests := []struct {
		dialect string
		dsn     string
		valid   bool
	}{
		{"postgres", "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai", true},
		{"pg", "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai", true},
		{"mysql", "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local", true},
		{"sqlite", "test.db", true},
		{"sqlserver", "sqlserver://username:password@localhost:1433?database=dbname", true},
		{"mssql", "sqlserver://username:password@localhost:1433?database=dbname", true},
		{"clickhouse", "clickhouse://localhost:9000/default", true},
		{"invalid", "dsn", false},
	}

	for _, tt := range tests {
		t.Run(tt.dialect, func(t *testing.T) {
			_, err := db.GetDialector(tt.dialect, tt.dsn)
			if tt.valid && err != nil {
				t.Fatalf("Expected valid dialector for %s, got error: %v", tt.dialect, err)
			}
			if !tt.valid && err == nil {
				t.Fatalf("Expected error for invalid dialect %s, got nil", tt.dialect)
			}
		})
	}
}
