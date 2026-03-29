package tests

import (
	"path/filepath"
	"runtime"

	"github.com/google/uuid"
)

// nonExistentUUID returns a valid UUID that is guaranteed not to exist in the database.
func nonExistentUUID() uuid.UUID {
	return uuid.MustParse("00000000-0000-0000-0000-000000000001")
}

// repoRoot returns the absolute path to the repository root,
// regardless of the working directory when tests are run.
func repoRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..")
}

// fixture returns the absolute path to a file under examples/.
func fixture(name string) string {
	return filepath.Join(repoRoot(), "examples", name)
}
