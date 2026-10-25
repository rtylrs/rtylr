package infrastructure_test

import (
	"strings"
	"testing"
)

// hr_time_entries.work_type_id has no foreign key behind it, so a hard DELETE of
// a work type left every entry that used it pointing at an id that no longer
// exists, including timesheets already submitted and approved. Those entries
// feed approvals and Finance's billable pull, so the classification has to
// survive (#1069). The archived column exists for exactly this and was never
// written.
func TestDeletingAWorkTypeArchivesItInstead(t *testing.T) {
	body := namedBlock(t, readQuery(t, "../../sql/queries/worktime.sql"), "DeleteWorkType")

	if strings.Contains(body, "DELETE FROM") {
		t.Errorf("DeleteWorkType still hard deletes, which orphans every hr_time_entries.work_type_id that referenced it:\n%s", body)
	}
	if !strings.Contains(body, "archived = TRUE") {
		t.Errorf("DeleteWorkType does not set archived:\n%s", body)
	}
	// Archiving an already-archived type must affect no rows, so the caller is
	// told it was not found rather than that it was deleted twice.
	if !strings.Contains(body, "archived = FALSE") {
		t.Errorf("DeleteWorkType does not exclude already-archived rows, so a repeat delete reports success:\n%s", body)
	}
	// Work types belong to a business.
	if !strings.Contains(body, "business_id = ?") {
		t.Errorf("DeleteWorkType lost its business_id scope:\n%s", body)
	}
}

// The archive is only useful if the picker stops offering the type. ListWorkTypes
// already did this; it must keep doing it, or archiving becomes a no-op the user
// can see.
func TestListWorkTypesHidesArchivedUnlessAsked(t *testing.T) {
	body := namedBlock(t, readQuery(t, "../../sql/queries/worktime.sql"), "ListWorkTypes")

	if !strings.Contains(body, "archived = FALSE") {
		t.Errorf("ListWorkTypes no longer filters archived rows, so an archived work type still appears in Log Hours:\n%s", body)
	}
}
