package testhelper

import (
    "fmt"
    "os"
    "testing"

    gmysql "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

// OpenGormTestDB opens a GORM *gorm.DB using TEST_DB_* environment variables.
// If variables are not set, it falls back to docker-compose's default host mapping
// (127.0.0.1:23306, app/apppass, app_test).
func OpenGormTestDB(t *testing.T) *gorm.DB {
    t.Helper()
    host := getenv("TEST_DB_HOST", "127.0.0.1")
    port := getenv("TEST_DB_PORT", "23306")
    user := getenv("TEST_DB_USER", "app")
    pass := getenv("TEST_DB_PASS", "apppass")
    name := getenv("TEST_DB_NAME", "app_test")

    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_0900_ai_ci&loc=Local", user, pass, host, port, name)
    db, err := gorm.Open(gmysql.Open(dsn), &gorm.Config{})
    if err != nil {
        t.Fatalf("open gorm (test db): %v", err)
    }
    if sqldb, err := db.DB(); err == nil {
        // Best-effort ping
        _ = sqldb.Ping()
    }
    return db
}

func getenv(k, def string) string {
    if v := os.Getenv(k); v != "" { return v }
    return def
}

