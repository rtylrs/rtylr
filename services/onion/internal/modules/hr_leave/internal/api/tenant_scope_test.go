package api

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// Every leave service call takes (ownerID, callerUserID). The first argument is
// named ownerID but is bound to hr_employees.business_id by the query underneath
// GetEmployeeIDByUserID, so passing the caller's user id there asks for
// business_id = <a user id>, matches nothing, and turns every read into a 403.
//
// Three handlers passed device.User.UserID twice. The correct source is
// tenant.ResolveBusinessID, which approve_leave_handler.go and
// create_leave_request_handler.go already use. This guards the shape rather than
// the wiring: the handlers hold a concrete *application.Service, so a fake
// cannot be injected to observe the arguments at runtime.
func TestNoLeaveHandlerPassesTheCallerAsTheBusiness(t *testing.T) {
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not locate the leave api package")
	}
	dir := filepath.Dir(sourceFile)

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read leave api package: %v", err)
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
	}
}
