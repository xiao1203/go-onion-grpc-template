package mysql

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	gmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// OpenFromEnv opens a *sql.DB using environment variables.
// Use prefix "TEST_" to read TEST_DB_* for test database.
// Without prefix, reads DB_* for dev database.
func OpenFromEnv(prefix string) (*sql.DB, error) {
	host := getenv(prefix+"DB_HOST", "127.0.0.1")
	port := getenv(prefix+"DB_PORT", "3306")
	user := getenv(prefix+"DB_USER", "root")
	pass := getenv(prefix+"DB_PASS", "")
	name := getenv(prefix+"DB_NAME", "app_dev")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_0900_ai_ci&loc=Local", user, pass, host, port, name)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

// OpenGormFromEnv opens a *gorm.DB using environment variables.
// Uses same env keys as OpenFromEnv.
func OpenGormFromEnv(prefix string) (*gorm.DB, error) {
	host := getenv(prefix+"DB_HOST", "127.0.0.1")
	port := getenv(prefix+"DB_PORT", "3306")
	user := getenv(prefix+"DB_USER", "root")
	pass := getenv(prefix+"DB_PASS", "")
	name := getenv(prefix+"DB_NAME", "app_dev")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_0900_ai_ci&loc=Local", user, pass, host, port, name)
	db, err := gorm.Open(gmysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	// configure connection pool on underlying sql.DB
	if sqldb, err := db.DB(); err == nil {
		sqldb.SetConnMaxLifetime(5 * time.Minute)
		sqldb.SetMaxOpenConns(25)
		sqldb.SetMaxIdleConns(25)
		if err := sqldb.Ping(); err != nil {
			_ = sqldb.Close()
			return nil, err
		}
	}
	return db, nil
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
