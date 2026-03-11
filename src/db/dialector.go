package db

import (
	"fmt"
	"strings"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/clickhouse"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

// GetDialector returns the appropriate GORM dialector based on the dialect string
func GetDialector(dialect, dsn string) (gorm.Dialector, error) {
	switch strings.ToLower(dialect) {
	case "postgres", "pg":
		return postgres.Open(dsn), nil
	case "mysql":
		return mysql.Open(dsn), nil
	case "sqlite":
		return sqlite.Open(dsn), nil
	case "sqlserver", "mssql":
		return sqlserver.Open(dsn), nil
	case "clickhouse":
		return clickhouse.Open(dsn), nil
	default:
		return nil, fmt.Errorf("unsupported dialect: %s", dialect)
	}
}
