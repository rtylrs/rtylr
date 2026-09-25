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
	root := findRepositoryRoot(t)
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

// findRepositoryRoot works in both normal and -trimpath test builds. Under
// -trimpath runtime.Caller returns an import path, not a filesystem path, so
// walk upward from the test working directory before using that fallback.
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
