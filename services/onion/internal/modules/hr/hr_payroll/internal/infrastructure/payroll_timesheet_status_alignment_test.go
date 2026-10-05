package infrastructure_test

import (
	"os"
	"regexp"
	"testing"
)

// These helpers parse the two SQL files that own the cross-module status
// contract between hr_time and hr_payroll. The test below is deliberately
// source-level: the defect it guards (B11-R05) was two SQL files disagreeing
// about which status means "approved time", and no DB fixture can represent
// that disagreement - every fixture has to pick one status spelling.

var (
	b11SQLStatementHeader = regexp.MustCompile(`(?m)^-- name: (\w+)`)
	b11SetStatus          = regexp.MustCompile(`SET status = '([^']+)'`)
	b11StatusIn           = regexp.MustCompile(`status IN \(([^)]*)\)`)
	b11QuotedValue        = regexp.MustCompile(`'([^']+)'`)
)

func b11ReadSQLFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
}

// b11StatementBlock returns the SQL text of one named sqlc statement.
func b11StatementBlock(t *testing.T, fileContent, name string) string {
	t.Helper()
	headers := b11SQLStatementHeader.FindAllStringSubmatchIndex(fileContent, -1)
	for i, header := range headers {
		if fileContent[header[2]:header[3]] != name {
			continue
		}
		end := len(fileContent)
		if i+1 < len(headers) {
			end = headers[i+1][0]
		}
		return fileContent[header[0]:end]
	}
	t.Fatalf("statement %q not found", name)
	return ""
}

func b11StatementStatus(t *testing.T, fileContent, statement string) string {
	t.Helper()
	match := b11SetStatus.FindStringSubmatch(b11StatementBlock(t, fileContent, statement))
	if match == nil {
		t.Fatalf("statement %q does not set a status", statement)
	}
	return match[1]
}

// TestPayrollSummaryAcceptsTimesheetStatusWrittenByFinalApproval is the pure
// unit half of the B11-R05 regression: hr_payroll's timesheet summary must
// accept the status hr_time writes when a timesheet is finally approved
// (LockTimesheet), must keep accepting the legacy ApproveTimesheet status, and
// must not accept the intermediate submitted status. A DB-backed test would
// catch a drift only if it happened to seed the new final status; this reads
// both SQL files and fails on the contract change itself.
func TestPayrollSummaryAcceptsTimesheetStatusWrittenByFinalApproval(t *testing.T) {
	payrollSQL := b11ReadSQLFile(t, "../../sql/queries/payroll_cross_module.sql")
	hrTimeSQL := b11ReadSQLFile(t, "../../../hr_time/sql/queries/timesheets.sql")

	summary := b11StatementBlock(t, payrollSQL, "ListApprovedTimesheetSummaries")
	inClause := b11StatusIn.FindStringSubmatch(summary)
	if inClause == nil {
		t.Fatal("ListApprovedTimesheetSummaries has no `status IN (...)` filter; payroll would pay every timesheet regardless of approval")
	}
	accepted := map[string]bool{}
	for _, quoted := range b11QuotedValue.FindAllStringSubmatch(inClause[1], -1) {
		accepted[quoted[1]] = true
	}
	if len(accepted) == 0 {
		t.Fatalf("could not parse any status out of %q", inClause[0])
	}

	finalStatus := b11StatementStatus(t, hrTimeSQL, "LockTimesheet")
	if !accepted[finalStatus] {
		t.Errorf("payroll summary accepts %v, but hr_time's final approval (LockTimesheet) writes status %q; approved timesheet hours would be zero (B11-R05)", accepted, finalStatus)
	}

	if legacyStatus := b11StatementStatus(t, hrTimeSQL, "ApproveTimesheet"); !accepted[legacyStatus] {
		t.Errorf("payroll summary accepts %v, missing the legacy timesheet status %q", accepted, legacyStatus)
	}

	submittedStatus := b11StatementStatus(t, hrTimeSQL, "SubmitTimesheet")
	if accepted[submittedStatus] {
		t.Errorf("payroll summary accepts intermediate status %q; unapproved timesheets would be paid", submittedStatus)
	}
}
