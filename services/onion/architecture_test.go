package onion_test

import (
	"encoding/csv"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

const onionImport = "github.com/voxire/onion/"

type architectureFinding struct {
	rule, file, detail string
}

type architectureScan struct {
	findings []architectureFinding
	edges    map[string]map[string]bool
}

func (s *architectureScan) add(rule, file, detail string) {
	s.findings = append(s.findings, architectureFinding{rule, file, detail})
}

// The first directory under modules is a bounded context; its child packages
// are internal capabilities, not independent owners of the same state.
func architectureOwner(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, onionImport), "/")
	if len(parts) >= 3 && parts[0] == "internal" {
		switch parts[1] {
		case "modules", "contracts":
			return parts[2]
		case "platform":
			return "platform/" + parts[2]
		}
	}
	return ""
}

func (s *architectureScan) source(path string, source []byte) error {
	f, err := parser.ParseFile(token.NewFileSet(), path, source, parser.ParseComments)
	if err != nil {
		return err
	}
	if ast.IsGenerated(f) {
		return nil
	}
	owner := architectureOwner(path)
	protected := strings.Contains(path, "/domain/") || strings.HasPrefix(path, "internal/contracts/")
	application := strings.Contains(path, "/application/")
	imports := make(map[string]string)
	for _, imp := range f.Imports {
		target, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			return err
		}
		name := filepath.Base(target)
		if imp.Name != nil {
			name = imp.Name.Name
		}
		imports[name] = target
		foreign := architectureOwner(target)
		businessEdge := foreign != "" && foreign != owner && !strings.HasPrefix(foreign, "platform/")
		if businessEdge {
			if strings.HasPrefix(owner, "platform/") {
				s.add("platform-import", path, target)
			} else if owner != "" {
				if s.edges == nil {
					s.edges = make(map[string]map[string]bool)
				}
				if s.edges[owner] == nil {
					s.edges[owner] = make(map[string]bool)
				}
				s.edges[owner][foreign] = true
			}
			if strings.Contains(target, "/internal/") && strings.Contains(strings.TrimPrefix(target, onionImport+"internal/"), "/internal/") || strings.Contains(target, "/persistence/") || strings.Contains(target, "/sqlc") {
				s.add("foreign-persistence-import", path, target)
			}
		}
		if protected && (target == "database/sql" || target == "net/http" || strings.Contains(target, "/infrastructure") || strings.Contains(target, "/adapters") || strings.Contains(target, "/dbclient") || strings.Contains(target, "/gen/") || target == "github.com/Masterminds/squirrel") {
			s.add("domain-import", path, target)
		}
		if application && (strings.Contains(target, "/internal/infrastructure") || strings.Contains(target, "/internal/api") || strings.Contains(target, "/adapters/")) {
			s.add("application-adapter-import", path, target)
		}
		if (owner == "aiquery" || strings.Contains(path, "/read/")) && (strings.Contains(target, "/writesql") || strings.Contains(target, "/write/")) {
			s.add("read-authority", path, target)
		}
	}
	if protected {
		badTags := make(map[string]bool)
		ast.Inspect(f, func(n ast.Node) bool {
			field, ok := n.(*ast.Field)
			if !ok || field.Tag == nil {
				return true
			}
			value, err := strconv.Unquote(field.Tag.Value)
			if err != nil {
				s.add("invalid-tag", path, field.Tag.Value)
				return true
			}
			for _, key := range []string{"db", "gorm", "sql", "json", "xml", "form", "query", "uri"} {
				if _, exists := reflect.StructTag(value).Lookup(key); exists {
					badTags[key] = true
				}
			}
			return true
		})
		for _, key := range sortedArchitectureKeys(badTags) {
			rule := "transport-tag"
			if key == "db" || key == "gorm" || key == "sql" {
				rule = "persistence-tag"
			}
			s.add(rule, path, key)
		}
	}
	for _, decl := range f.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				if typ, ok := spec.(*ast.TypeSpec); ok && typ.Name.IsExported() {
					s.generatedSurface(path, typ.Type, imports)
				}
			}
		case *ast.FuncDecl:
			if decl.Name.IsExported() {
				s.generatedSurface(path, decl.Type, imports)
			}
			s.transactionCalls(path, owner, decl, imports)
		}
	}
	return nil
}

func (s *architectureScan) generatedSurface(path string, node ast.Node, imports map[string]string) {
	if strings.Contains(path, "/internal/") || !strings.HasPrefix(path, "internal/modules/") {
		return
	}
	ast.Inspect(node, func(n ast.Node) bool {
		if structure, ok := n.(*ast.StructType); ok {
			for _, field := range structure.Fields.List {
				if len(field.Names) == 0 || field.Names[0].IsExported() {
					s.generatedSurface(path, field.Type, imports)
				}
			}
			return false
		}
		selector, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		name, ok := selector.X.(*ast.Ident)
		if ok && (strings.Contains(imports[name.Name], "/sqlc") || strings.Contains(imports[name.Name], "/readsql") || strings.Contains(imports[name.Name], "/writesql")) {
			s.add("generated-public-api", path, name.Name+"."+selector.Sel.Name)
		}
		return true
	})
}

func (s *architectureScan) transactionCalls(path, owner string, fn *ast.FuncDecl, imports map[string]string) {
	txs := make(map[string]bool)
	if fn.Type.Params != nil {
		for _, param := range fn.Type.Params.List {
			ast.Inspect(param.Type, func(n ast.Node) bool {
				selector, ok := n.(*ast.SelectorExpr)
				if !ok || selector.Sel.Name != "Tx" {
					return true
				}
				if pkg, ok := selector.X.(*ast.Ident); ok && imports[pkg.Name] == "database/sql" {
					for _, name := range param.Names {
						txs[name.Name] = true
					}
				}
				return true
			})
		}
	}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		assignment, ok := n.(*ast.AssignStmt)
		if !ok || len(assignment.Lhs) == 0 || len(assignment.Rhs) != 1 {
			return true
		}
		call, ok := assignment.Rhs[0].(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if ok && (selector.Sel.Name == "BeginTx" || selector.Sel.Name == "Begin") {
			if name, ok := assignment.Lhs[0].(*ast.Ident); ok {
				txs[name.Name] = true
			}
		}
		return true
	})
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		pkg, ok := selector.X.(*ast.Ident)
		if !ok {
			return true
		}
		foreign := architectureOwner(imports[pkg.Name])
		if foreign == "" || foreign == owner {
			return true
		}
		for _, arg := range call.Args {
			if name, ok := arg.(*ast.Ident); ok && txs[name.Name] {
				s.add("cross-module-transaction", path, fn.Name.Name+" passes "+name.Name+" to "+pkg.Name+"."+selector.Sel.Name)
			}
		}
		return true
	})
}

func sortedArchitectureKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (s *architectureScan) cycles() {
	state := make(map[string]int)
	var visit func(string, []string)
	visit = func(owner string, stack []string) {
		if state[owner] == 1 {
			for i, member := range stack {
				if member == owner {
					s.add("module-cycle", "internal/modules", strings.Join(append(stack[i:], owner), " -> "))
					break
				}
			}
			return
		}
		if state[owner] == 2 {
			return
		}
		state[owner] = 1
		for _, target := range sortedArchitectureKeys(s.edges[owner]) {
			visit(target, append(stack, owner))
		}
		state[owner] = 2
	}
	for _, owner := range sortedArchitectureKeys(s.edges) {
		visit(owner, nil)
	}
}

func (s *architectureScan) persistence(t *testing.T, root string) {
	t.Helper()
	out := t.TempDir()
	for _, script := range []string{"sweep-queries.mjs", "sweep-tables.mjs"} {
		cmd := exec.Command("node", filepath.Join("scripts", "cross-app", script))
		cmd.Dir = root
		cmd.Env = append(os.Environ(), "CROSS_APP_ONION="+root, "CROSS_APP_OUT="+out)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("architecture inventory %s: %v\n%s", script, err, output)
		}
	}
	for _, row := range architectureCSV(t, filepath.Join(out, "backend_tables.csv")) {
		owner := row["owner_schema_domain"]
		if owner == "" || strings.Contains(owner, "|") {
			s.add("table-owner", "internal", row["table"]+": owner "+strconv.Quote(owner))
			continue
		}
		for _, access := range []string{"reading", "writing"} {
			for _, consumer := range strings.Fields(row["modules_"+access]) {
				if consumer != owner {
					s.add("foreign-table-access", "internal", consumer+" "+access+" "+row["table"]+" owned by "+owner)
				}
			}
		}
	}
	for _, row := range architectureCSV(t, filepath.Join(out, "backend_queries.csv")) {
		if row["module"] == "aiquery" && row["tables_written"] != "" {
			s.add("read-authority", row["file"], row["query_name"]+" writes "+row["tables_written"])
		}
	}
	data, err := os.ReadFile(filepath.Join(root, "internal/platform/database/schema_registry.tsv"))
	if err != nil {
		t.Fatal(err)
	}
	identities := make(map[string]bool)
	sources := make(map[string]bool)
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		columns := strings.Split(line, "\t")
		if len(columns) != 2 || identities[columns[0]] {
			t.Fatalf("invalid or duplicate migration identity: %s", line)
		}
		identities[columns[0]] = true
		for _, source := range strings.Split(columns[1], "|") {
			if architectureOwner(source) == "" || strings.Contains(source, "..") || sources[source] {
				s.add("migration-owner", "internal/platform/database/schema_registry.tsv", source)
			}
			sources[source] = true
			if _, err := os.Stat(filepath.Join(root, source)); err != nil {
				t.Fatalf("registered migration %s: %v", source, err)
			}
		}
	}
	t.Logf("checked %d stable migration identities and %d owned sources", len(identities), len(sources))
}

func architectureCSV(t *testing.T, path string) []map[string]string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil || len(rows) < 2 {
		t.Fatalf("missing or invalid architecture inventory %s: %v", path, err)
	}
	result := make([]map[string]string, 0, len(rows)-1)
	for _, cells := range rows[1:] {
		row := make(map[string]string)
		for i, name := range rows[0] {
			row[name] = cells[i]
		}
		result = append(result, row)
	}
	return result
}

type architectureException struct {
	ID, Rule, Owner, Axiom, Reason, Alternatives, Containment, ExitPlan, ReviewBy string
	Files                                                                         []string
	Details                                                                       []string
}

func TestArchitecture(t *testing.T) {
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	scan := &architectureScan{}
	err = filepath.WalkDir(filepath.Join(root, "internal"), func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		return scan.source(filepath.ToSlash(rel), data)
	})
	if err != nil {
		t.Fatal(err)
	}
	scan.cycles()
	scan.persistence(t, root)
	data, err := os.ReadFile(filepath.Join(root, "docs/architecture-exceptions.json"))
	if err != nil {
		t.Fatal(err)
	}
	var exceptions []architectureException
	if err := json.Unmarshal(data, &exceptions); err != nil {
		t.Fatal(err)
	}
	used := make(map[string]bool)
	for _, exception := range exceptions {
		review, err := time.Parse(time.DateOnly, exception.ReviewBy)
		if err != nil || review.Before(time.Now().UTC().Truncate(24*time.Hour)) {
			t.Fatalf("architecture exception %s has an invalid or expired review date", exception.ID)
		}
		if exception.ID == "" || exception.Rule == "" || exception.Owner == "" || exception.Axiom == "" || exception.Reason == "" || exception.Alternatives == "" || exception.Containment == "" || exception.ExitPlan == "" || len(exception.Files) == 0 || len(exception.Details) == 0 {
			t.Fatalf("incomplete architecture exception %s", exception.ID)
		}
		for _, path := range exception.Files {
			if strings.ContainsAny(path, "*?") || strings.Contains(path, "..") {
				t.Fatalf("exception %s must name exact files", exception.ID)
			}
			if _, err := os.Stat(filepath.Join(root, path)); err != nil {
				t.Fatalf("exception %s names missing file %s", exception.ID, path)
			}
		}
	}
	counts := make(map[string]int)
	seen := make(map[string]bool)
	for _, finding := range scan.findings {
		exempt := false
		for _, exception := range exceptions {
			for _, path := range exception.Files {
				if exception.Rule == finding.rule && path == finding.file && containsArchitectureDetail(exception.Details, finding.detail) {
					used[exception.ID] = true
					exempt = true
				}
			}
		}
		if exempt {
			continue
		}
		key := finding.rule + ":" + finding.file + ":" + finding.detail
		if seen[key] {
			continue
		}
		seen[key] = true
		counts[finding.rule]++
		if len(seen) <= 100 {
			t.Errorf("%s: %s: %s", finding.rule, finding.file, finding.detail)
		}
	}
	for _, rule := range sortedArchitectureKeys(counts) {
		t.Logf("%s: %d violations", rule, counts[rule])
	}
	if len(seen) > 100 {
		t.Errorf("%d total violations; first 100 shown", len(seen))
	}
	for _, exception := range exceptions {
		if !used[exception.ID] {
			t.Errorf("remove unused architecture exception %s", exception.ID)
		}
	}
}

func containsArchitectureDetail(details []string, value string) bool {
	for _, detail := range details {
		if detail == value {
			return true
		}
	}
	return false
}

func TestArchitectureScanner(t *testing.T) {
	tests := []struct {
		name, path, source, rule string
	}{
		{"single import", "internal/modules/consumer/example.go", `package pos; import "github.com/voxire/onion/internal/modules/owner/products/internal/infrastructure/persistence/sqlc"`, "foreign-persistence-import"},
		{"generated alias", "internal/modules/consumer/example.go", `package pos; import q "github.com/voxire/onion/internal/modules/consumer/orders/internal/persistence/sqlc"; type Result = q.Order`, "generated-public-api"},
		{"exported result", "internal/modules/consumer/example.go", `package pos; import q "github.com/voxire/onion/internal/modules/consumer/orders/internal/persistence/sqlc"; func Get() q.Order { panic("fixture") }`, "generated-public-api"},
		{"exported parameter", "internal/modules/consumer/example.go", `package pos; import q "github.com/voxire/onion/internal/modules/consumer/orders/internal/persistence/sqlc"; func Save(order q.Order) {}`, "generated-public-api"},
		{"transport tag", "internal/modules/consumer/orders/internal/domain/order.go", "package domain; type Order struct { ID int `json:\"id\"` }", "transport-tag"},
		{"domain database", "internal/modules/consumer/orders/internal/domain/order.go", `package domain; import "database/sql"; type Order struct { ID sql.NullInt64 }`, "domain-import"},
		{"transaction parameter", "internal/modules/consumer/orders/order.go", `package orders; import ("context"; "database/sql"; stock "github.com/voxire/onion/internal/modules/owner/stock"); func Pay(ctx context.Context, tx *sql.Tx) { stock.Apply(ctx, tx) }`, "cross-module-transaction"},
		{"transaction local", "internal/modules/consumer/orders/order.go", `package orders; import stock "github.com/voxire/onion/internal/modules/owner/stock"; func Pay() { tx, _ := pool.BeginTx(ctx, nil); stock.Apply(ctx, tx) }`, "cross-module-transaction"},
		{"private generated detail", "internal/modules/consumer/example.go", `package pos; import q "github.com/voxire/onion/internal/modules/consumer/orders/internal/persistence/sqlc"; type Module struct { queries *q.Queries }`, ""},
		{"local transaction", "internal/modules/owner/orders/order.go", `package orders; import ("context"; "database/sql"; stock "github.com/voxire/onion/internal/modules/owner/stock"); func Pay(ctx context.Context, tx *sql.Tx) { stock.Apply(ctx, tx) }`, ""},
		{"foreign transaction into platform", "internal/modules/consumer/orders/order.go", `package orders; import ("context"; "database/sql"; "github.com/voxire/onion/internal/platform/eventbus"); func Pay(ctx context.Context, tx *sql.Tx) { eventbus.Publish(ctx, tx, event) }`, "cross-module-transaction"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scan := &architectureScan{}
			if err := scan.source(test.path, []byte(test.source)); err != nil {
				t.Fatal(err)
			}
			if test.rule == "" {
				if len(scan.findings) != 0 {
					t.Fatalf("unexpected violations: %+v", scan.findings)
				}
				return
			}
			for _, finding := range scan.findings {
				if finding.rule == test.rule {
					return
				}
			}
			t.Fatalf("did not detect %s: %+v", test.rule, scan.findings)
		})
	}
	t.Run("module cycle across packages", func(t *testing.T) {
		scan := &architectureScan{edges: map[string]map[string]bool{"alpha": {"beta": true}, "beta": {"alpha": true}}}
		scan.cycles()
		if len(scan.findings) != 1 || scan.findings[0].rule != "module-cycle" {
			t.Fatalf("did not detect module cycle: %+v", scan.findings)
		}
	})
}
