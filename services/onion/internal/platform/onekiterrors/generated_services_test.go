package onekiterrors_test

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

var oneKitServiceMethodPattern = regexp.MustCompile(`func \([^)]*\) [A-Z][A-Za-z0-9]+\([^)]*\) \(\*[^,]+, error\)`)

func TestGeneratedServiceImplementationsDoNotUseLegacyErrors(t *testing.T) {
	root := repoRoot(t)
	modulesDir := filepath.Join(root, "internal", "modules")

	var offenders []string
	err := filepath.WalkDir(modulesDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		if !strings.Contains(path, string(filepath.Separator)+"internal"+string(filepath.Separator)+"api"+string(filepath.Separator)) &&
			!strings.Contains(path, string(filepath.Separator)+"internal"+string(filepath.Separator)+"routes"+string(filepath.Separator)+"handlers"+string(filepath.Separator)) {
			return nil
		}

		contentBytes, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		content := string(contentBytes)
		if !oneKitServiceMethodPattern.MatchString(content) {
			return nil
		}
		if strings.Contains(content, "internal/platform/errors") ||
			strings.Contains(content, "errorModels.ServerError") ||
			strings.Contains(content, "PermissionDeniedServerError") ||
			strings.Contains(content, "errorModels.FromError") ||
			strings.Contains(content, "onekiterrors.FromError") {
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				rel = path
			}
			offenders = append(offenders, rel)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk generated service implementations: %v", err)
	}
	if len(offenders) > 0 {
		t.Fatalf("generated oneKit service implementations must use oneKit errors, found legacy errors in:\n%s", strings.Join(offenders, "\n"))
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current file")
	}

	root := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", ".."))
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("resolve repo root from %s: %v", currentFile, err)
	}
	return root
}
