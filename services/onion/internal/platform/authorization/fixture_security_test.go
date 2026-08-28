package authorization

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"
)

func TestConsoleAdminFixtureDoesNotContainUsableCredentials(t *testing.T) {
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not locate the fixture security test")
	}

	fixturePath := filepath.Join(
		filepath.Dir(sourceFile),
		"..", "..", "..", "scripts", "fixtures", "99_console_admin.sql",
	)
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
