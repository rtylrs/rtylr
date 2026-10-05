package application

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// UpdateTask used to write its activity rows before attempting the update,
// each as its own unbatched write with no transaction around it. When the
// update then failed the rows survived, so a task's history claimed changes
// that never happened. Every task-detail edit currently sends
// priority: "no_priority", which 500s, so in normal use a large share of
// attempted edits were writing false history while changing nothing (#1070).
//
// The fix is structural: entries are collected into a slice and there is now a
// single write site, after s.repo.UpdateTask has returned nil. This asserts
// that shape on the source, because the service holds a concrete
// *infrastructure.Repository rather than an interface, so no fake can be
// injected to observe call order at runtime.
func TestUpdateTaskWritesActivityOnlyAfterTheUpdateSucceeds(t *testing.T) {
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not locate the workspaces application package")
	}

	dir := filepath.Dir(sourceFile)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read application package: %v", err)
	}
	var builder strings.Builder
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		raw, readErr := os.ReadFile(filepath.Join(dir, entry.Name()))
		if readErr != nil {
			t.Fatalf("read %s: %v", entry.Name(), readErr)
		}
		builder.Write(raw)
		builder.WriteString("\n")
	}
	source := builder.String()

	const marker = "func (s *Service) UpdateTask("
	start := strings.Index(source, marker)
	if start < 0 {
		t.Fatal("UpdateTask not found; rename the test or restore the method")
	}
	// The method ends at the first line-start closing brace after it.
	body := source[start:]
	if end := strings.Index(body, "\n}\n"); end >= 0 {
		body = body[:end]
	}

	update := strings.Index(body, "s.repo.UpdateTask(ctx, input)")
	if update < 0 {
		t.Fatal("UpdateTask does not call s.repo.UpdateTask")
	}

	log := strings.Index(body, "s.repo.LogActivity(")
	if log < 0 {
		t.Fatal("UpdateTask no longer writes activity at all; it should still record what changed")
	}
	if log < update {
		t.Errorf("UpdateTask writes activity at offset %d, before the update at offset %d; "+
			"a failed update would leave history for changes that never happened", log, update)
	}

	if n := strings.Count(body, "s.repo.LogActivity("); n != 1 {
		t.Errorf("UpdateTask has %d activity write sites, want exactly 1 after the update; "+
			"more than one is how the ordering regressed the first time", n)
	}
}
