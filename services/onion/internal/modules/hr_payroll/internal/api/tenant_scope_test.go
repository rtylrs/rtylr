package api

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// Every payroll service call takes (businessID, ...) where the first argument is
// bound to hr_employees.business_id or hr_payroll_runs.business_id by the query
// underneath. Passing device.User.UserID there asks for the caller's user id as
// a business id, which matches no row for any provisioned tenant — users.id and
// businesses.id are independent sequences — and turns the request into a 403.
//
// Four handlers passed device.User.UserID twice. The correct source is
// tenant.ResolveBusinessID, which the trigger/export/list/config handlers
// already use. This guards the shape rather than the wiring: the handlers hold a
// concrete *application.Service, so a fake cannot be injected to observe the
// arguments at runtime.
func TestNoPayrollHandlerPassesTheCallerAsTheBusiness(t *testing.T) {
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not locate the payroll api package")
	}
	dir := filepath.Dir(sourceFile)

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read payroll api package: %v", err)
	}

	doubledCaller := regexp.MustCompile(`\(\s*ctx\s*,\s*device\.User\.UserID\s*,\s*device\.User\.UserID\b`)

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		source, readErr := os.ReadFile(filepath.Join(dir, name))
		if readErr != nil {
			t.Fatalf("read %s: %v", name, readErr)
		}
		if doubledCaller.Match(source) {
			t.Errorf("%s passes device.User.UserID as both the business and the caller; "+
				"resolve the business with tenant.ResolveBusinessID", name)
		}
		if strings.Contains(string(source), "GetEmployeeIDByUserID(ctx, device.User.UserID") {
			t.Errorf("%s passes the caller's user id as the business scope of GetEmployeeIDByUserID; "+
				"resolve the business with tenant.ResolveBusinessID", name)
		}
		if strings.Contains(string(source), "GetEmployeeIDByUserID(") &&
			!strings.Contains(string(source), "tenant.ResolveBusinessID(") {
			t.Errorf("%s calls GetEmployeeIDByUserID without resolving the business via "+
				"tenant.ResolveBusinessID", name)
		}
	}
}
