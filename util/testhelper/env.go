package testhelper

import (
	"os"
	"testing"
)

// EnsureTestDBEnv sets sane TEST_DB_* defaults for integration tests when they
// are executed outside Docker. When running inside the api container, these
// variables are already provided by docker-compose and this function becomes a no-op.
//
// Defaults chosen to match docker-compose.yml mappings:
//
//	TEST_DB_HOST=127.0.0.1
//	TEST_DB_PORT=23306
//	TEST_DB_USER=app
//	TEST_DB_PASS=apppass
//	TEST_DB_NAME=app_test
func EnsureTestDBEnv(t *testing.T) {
	t.Helper()
	if os.Getenv("TEST_DB_HOST") == "" {
		_ = os.Setenv("TEST_DB_HOST", "127.0.0.1")
	}
	if os.Getenv("TEST_DB_PORT") == "" {
		_ = os.Setenv("TEST_DB_PORT", "23306")
	}
	if os.Getenv("TEST_DB_USER") == "" {
		_ = os.Setenv("TEST_DB_USER", "app")
	}
	if os.Getenv("TEST_DB_PASS") == "" {
		_ = os.Setenv("TEST_DB_PASS", "apppass")
	}
	if os.Getenv("TEST_DB_NAME") == "" {
		_ = os.Setenv("TEST_DB_NAME", "app_test")
	}
}
