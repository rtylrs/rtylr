package authorization

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"
)

func TestConsoleAdminFixtureDoesNotContainUsableCredentials(t *testing.T) {
	fixturePath := filepath.Join(findRepositoryRoot(t), "internal", "testutil", "fixtures", "99_console_admin.sql")
	fixture, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read console admin fixture: %v", err)
	}

	plaintextCredentialComment := regexp.MustCompile(`(?mi)^\s*--\s*` + "pass" + `word\s*:`)
	if plaintextCredentialComment.Match(fixture) {
		t.Fatal("console admin fixture contains a plaintext credential comment")
	}

	bcryptMaterial := regexp.MustCompile(`\$2[aby]\$[0-9]{2}\$[./A-Za-z0-9]{53}`)
	if bcryptMaterial.Match(fixture) {
		t.Fatal("console admin fixture contains usable bcrypt material")
	}

	if !regexp.MustCompile(`(?m)fixture-auth-disabled`).Match(fixture) {
		t.Fatal("console admin fixture is missing the authentication-disabled marker")
	}
}

// findRepositoryRoot works in both normal and -trimpath test builds. With
// -trimpath, runtime.Caller returns an import path rather than a filesystem
// path, so the test working directory is the reliable first candidate.
func findRepositoryRoot(t *testing.T) string {
	t.Helper()

	candidates := make([]string, 0, 2)
	if workingDir, err := os.Getwd(); err == nil {
		candidates = append(candidates, workingDir)
	}
	if _, sourceFile, _, ok := runtime.Caller(0); ok && filepath.IsAbs(sourceFile) {
		candidates = append(candidates, filepath.Dir(sourceFile))
	}

	for _, candidate := range candidates {
		dir, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		for {
			if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
				return dir
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	t.Fatal("could not locate the repository root from the test working directory")
	return ""
}
