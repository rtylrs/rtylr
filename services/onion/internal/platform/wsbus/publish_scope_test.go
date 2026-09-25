package wsbus

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// authorizerScopes mirrors the production AuthorizeFn switch wired in
// cmd/api/main.go. A scope published without a case here means no client can
// ever subscribe to the topic (default-deny), so realtime events silently die.
var authorizerScopes = map[string]struct{}{
	"user":      {},
	"owner":     {},
	"business":  {},
	"workspace": {},
	"project":   {},
	"task":      {},
}

// helperTopicScopes maps local helpers whose return value hides the literal
// prefix from simple extraction, to their actual scope.
var helperTopicScopes = map[string]string{
	"ownerTopic(": "owner",
}

var (
	publishCallRe = regexp.MustCompile(`(?s)wsbus\.Publish\([^)]*?Topic:\s*(.{0,120}?)\)`)
	stringScopeRe = regexp.MustCompile(`"([a-z_]+):"`)
	sprintfRe     = regexp.MustCompile(`fmt\.Sprintf\("([a-z_]+):`)
)

// TestPublishedTopicsHaveAuthorizers walks the repository for every
// wsbus.Publish call site and asserts each topic's scope scheme is one the
// production subscription authorizer handles (#764).
func TestPublishedTopicsHaveAuthorizers(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}

	published := map[string][]string{}
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "voxie" || name == "node_modules" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		src := string(data)
		for _, match := range publishCallRe.FindAllStringSubmatch(src, -1) {
			expr := match[1]
			rel, _ := filepath.Rel(root, path)
			if sc := sprintfRe.FindStringSubmatch(expr); sc != nil {
				published[sc[1]] = append(published[sc[1]], rel)
				continue
			}
			if sc := stringScopeRe.FindStringSubmatch(expr); sc != nil {
				published[sc[1]] = append(published[sc[1]], rel)
				continue
			}
			for helper, scope := range helperTopicScopes {
				if strings.Contains(expr, helper) {
					published[scope] = append(published[scope], rel)
					break
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}

	if len(published) == 0 {
		t.Fatal("no wsbus.Publish topics found; the walk is broken")
	}

	var unknown []string
	var missingCase []string
	scopes := make([]string, 0, len(published))
	for scope := range published {
		scopes = append(scopes, scope)
		if _, ok := authorizerScopes[scope]; !ok {
			unknown = append(unknown, scope)
		}
	}
	sort.Strings(scopes)

	apiDir := filepath.Join(root, "cmd", "api")
	entries, readErr := os.ReadDir(apiDir)
	if readErr != nil {
		t.Fatalf("read cmd/api: %v", readErr)
	}
	var apiSource strings.Builder
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		raw, fileErr := os.ReadFile(filepath.Join(apiDir, entry.Name()))
		if fileErr != nil {
			t.Fatalf("read %s: %v", entry.Name(), fileErr)
		}
		apiSource.Write(raw)
		apiSource.WriteString("\n")
	}
	for _, scope := range scopes {
		if !strings.Contains(apiSource.String(), `case "`+scope+`":`) {
			missingCase = append(missingCase, scope)
		}
	}

	if len(unknown) > 0 {
		t.Errorf("scopes published without an authorizer entry (default-deny kills these topics): %v", unknown)
	}
	if len(missingCase) > 0 {
		t.Errorf("scopes published but not handled by cmd/api/main.go switch: %v", missingCase)
	}
	t.Logf("published scopes verified against authorizer: %v", scopes)
}
