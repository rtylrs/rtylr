package infrastructure_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// sqlc rewrites sqlc.arg into a placeholder in a WHERE expression and leaves it
// alone inside a JOIN ON clause, so the literal reaches MySQL and the statement
// fails to parse (#1046). Nothing in the toolchain reports this: sqlc generates
// happily and the code compiles.
//
// This runs without a database, which is the point. CI has no MySQL, so the
// runtime test beside this one is skipped there, and this is what would have
// caught the defect.
func TestNoSqlcArgInsideAJoinClause(t *testing.T) {
	dir := "../../sql/queries"
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read query directory: %v", err)
	}

	// A JOIN's ON clause runs to the SET, WHERE, GROUP BY, ORDER BY or the end
	// of the statement, whichever comes first.
	joinClause := regexp.MustCompile(`(?is)\bJOIN\b.*?\bON\b(.*?)(?:\bSET\b|\bWHERE\b|\bGROUP\s+BY\b|\bORDER\s+BY\b|;)`)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		raw, readErr := os.ReadFile(filepath.Join(dir, entry.Name()))
		if readErr != nil {
			t.Fatalf("read %s: %v", entry.Name(), readErr)
		}

		// Comments are stripped first. They document this very defect, so
		// scanning them finds the words rather than the code.
		for _, match := range joinClause.FindAllStringSubmatch(withoutComments(string(raw)), -1) {
			if strings.Contains(match[1], "sqlc.arg") || strings.Contains(match[1], "sqlc.narg") {
				t.Errorf("%s has sqlc.arg inside a JOIN ON clause, which reaches MySQL verbatim "+
					"and fails with \"FUNCTION sqlc.arg does not exist\"; move the condition into WHERE:\n%s",
					entry.Name(), strings.TrimSpace(match[1]))
			}
		}
	}
}

// withoutComments removes SQL line comments, keeping line structure so the
// remaining statements still parse the way the regex expects.
func withoutComments(sql string) string {
	lines := strings.Split(sql, "\n")
	for i, line := range lines {
		if at := strings.Index(line, "--"); at >= 0 {
			lines[i] = line[:at]
		}
	}
	return strings.Join(lines, "\n")
}
