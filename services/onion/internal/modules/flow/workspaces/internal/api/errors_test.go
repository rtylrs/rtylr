package api

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestWorkspaceHandlersDoNotExposeRawErrors(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}

	entries, err := os.ReadDir(filepath.Dir(currentFile))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		contents, readErr := os.ReadFile(filepath.Join(filepath.Dir(currentFile), entry.Name()))
		if readErr != nil {
			t.Fatalf("read %s: %v", entry.Name(), readErr)
		}
		for lineNumber, line := range strings.Split(string(contents), "\n") {
			if strings.Contains(line, "onekiterrors.New") && strings.Contains(line, "err.Error()") {
				t.Errorf("%s:%d passes a raw error to an API response", entry.Name(), lineNumber+1)
			}
		}
	}
}
