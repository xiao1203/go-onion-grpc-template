package testhelper

import (
    "os"
    "path/filepath"
)

// RepoRoot tries to find the repository root by walking up from the current
// working directory until it finds a go.mod. Falls back to CWD on failure.
func RepoRoot() string {
    dir, _ := os.Getwd()
    for i := 0; i < 10 && dir != "/" && dir != "." && dir != ""; i++ {
        if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
            return dir
        }
        dir = filepath.Dir(dir)
    }
    // fallback
    if cwd, err := os.Getwd(); err == nil {
        return cwd
    }
    return "."
}

// FixturePath returns an absolute path from the repo root joined with rel.
func FixturePath(rel string) string {
    return filepath.Join(RepoRoot(), rel)
}

