package testhelper

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-testfixtures/testfixtures/v3"
)

// LoadTestFixtures loads fixture YAMLs into the test DB.
// Example:
//
//	testhelper.LoadTestFixtures(t, testfixtures.Directory("testdata/fixture/sample"))
//
// Accepts one or more directories. Paths may be relative; they are cleaned.
func LoadTestFixtures(t *testing.T, opts ...func(*testfixtures.Loader) error) {
	t.Helper()

	// Ensure environment defaults when running outside Docker.
	EnsureTestDBEnv(t)

	host := getenv("TEST_DB_HOST", "127.0.0.1")
	port := getenv("TEST_DB_PORT", "23306")
	user := getenv("TEST_DB_USER", "app")
	pass := getenv("TEST_DB_PASS", "apppass")
	name := getenv("TEST_DB_NAME", "app_test")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true&charset=utf8mb4&collation=utf8mb4_0900_ai_ci&loc=Local",
		user, pass, host, port, name,
	)

	sqldb, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("open sql db for fixtures: %v", err)
	}
	t.Cleanup(func() { _ = sqldb.Close() })

	if err := sqldb.Ping(); err != nil {
		t.Fatalf("ping sql db: %v", err)
	}

	// Build options: base DB + dialect + caller-provided sources/options
	args := []func(*testfixtures.Loader) error{
		testfixtures.Database(sqldb),
		testfixtures.Dialect("mysql"),
	}
	args = append(args, opts...)

	fx, err := testfixtures.New(args...)
	if err != nil {
		t.Fatalf("testfixtures.New: %v", err)
	}
	if err := fx.Load(); err != nil {
		t.Fatalf("fixtures.Load: %v", err)
	}
}
