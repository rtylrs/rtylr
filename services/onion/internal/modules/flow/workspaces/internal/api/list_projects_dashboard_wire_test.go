package api

import (
	"encoding/json"
	"testing"

	commonpaginationv1 "github.com/voxire/voxie/gen/go/common/pagination/v1"
	flowdashboardv1 "github.com/voxire/voxie/gen/go/flow/dashboard/v1"
	flowprojectv2 "github.com/voxire/voxie/gen/go/flow/project/v2"
	flowtaskv1 "github.com/voxire/voxie/gen/go/flow/task/v1"
)

// The workspace-removal proposal calls byte-for-byte regression coverage on
// these two responses required, and none existed. They are the two payloads the
// removal is most likely to disturb: ListProjects because its scoping rule is
// being rewritten from workspace membership to business plus curation, and
// GetFlowDashboard because it aggregates across that same scope.
//
// These assert the exact serialised bytes rather than field-by-field values, so
// a renamed field, a dropped field, a changed omitempty, or an int64 that stops
// travelling as a string all fail here instead of silently reaching a client
// that reads the old name and gets undefined.
//
// No database: the point is the wire contract, not the query.

func TestListProjectsResponseWireShapeIsStable(t *testing.T) {
	archivedAt := "2026-08-01T00:00:00Z"
	purgeAt := "2026-09-01T00:00:00Z"

	res := flowprojectv2.ListProjectsResponse{
		Projects: []*flowprojectv2.Project{{
			Id:             12,
			OwnerId:        34,
			Name:           "Atlas",
			Description:    "desc",
			Icon:           "rocket",
			Color:          "#64C5B9",
			Status:         "active",
			Visibility:     "private",
			IsPersonal:     true,
			StartDate:      "2026-01-01",
			DueDate:        "2026-12-31",
			CreatedBy:      56,
			CreatedAt:      "2026-01-01T00:00:00Z",
			UpdatedAt:      "2026-02-01T00:00:00Z",
			TaskCount:      7,
			CompletedCount: 3,
			MemberCount:    2,
			Code:           "ATL",
			ArchivedAt:     &archivedAt,
			PurgeAt:        &purgeAt,
		}},
		Pagination: &commonpaginationv1.PaginationResponse{Page: 2, Limit: 50, Total: 137},
	}

	got, err := json.Marshal(&res)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	// id, owner_id and created_by are int64 on the wire as STRINGS. That is not
	// incidental: the generated clients parse them as strings, so switching them
	// to numbers breaks every consumer at once.
	const want = `{"projects":[{"name":"Atlas","description":"desc","icon":"rocket","color":"#64C5B9","status":"active","visibility":"private","is_personal":true,"start_date":"2026-01-01","due_date":"2026-12-31","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-02-01T00:00:00Z","task_count":7,"completed_count":3,"member_count":2,"code":"ATL","archived_at":"2026-08-01T00:00:00Z","purge_at":"2026-09-01T00:00:00Z","id":"12","owner_id":"34","created_by":"56"}],"pagination":{"page":2,"limit":50,"total":137}}`

	if string(got) != want {
		t.Fatalf("ListProjectsResponse wire shape changed.\n got: %s\nwant: %s", got, want)
	}
}

// An empty page must still be an object with the pagination envelope, not null
// and not a bare list. A client that renders a count reads pagination.total.
func TestListProjectsResponseEmptyPageWireShape(t *testing.T) {
	res := flowprojectv2.ListProjectsResponse{
		Pagination: &commonpaginationv1.PaginationResponse{Page: 1, Limit: 50, Total: 0},
	}

	got, err := json.Marshal(&res)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	const want = `{"pagination":{"page":1,"limit":50}}`
	if string(got) != want {
		t.Fatalf("empty ListProjectsResponse wire shape changed.\n got: %s\nwant: %s", got, want)
	}
}

func TestGetFlowDashboardResponseWireShapeIsStable(t *testing.T) {
	res := flowdashboardv1.GetFlowDashboardResponse{
		OpenTasks:         11,
		InProgressTasks:   5,
		CompletedThisWeek: 9,
		OverdueTasks:      2,
		MyTasks: []*flowtaskv1.Task{{
			Id:    77,
			Title: "Ship the backfill",
		}},
	}

	got, err := json.Marshal(&res)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	// Two things this pins that are easy to get wrong.
	//
	// Every int64 travels as a string, and a ZERO int64 travels as "0" rather
	// than being omitted. So an unset task field arrives as the string "0", and
	// in JavaScript Boolean("0") is true: a client guarding with
	// `if (task.workspaceId)` takes the truthy branch on an absent value.
	//
	// workspace_id is still on this payload because the Task contract still
	// declares it. When Phase 4 removes it, this test fails, and that failure is
	// the intended signal rather than a surprise.
	const want = `{"my_tasks":[{"title":"Ship the backfill","id":"77","project_id":"0","workspace_id":"0","owner_id":"0","reporter_id":"0","sort_order":"0"}],"open_tasks":"11","in_progress_tasks":"5","completed_this_week":"9","overdue_tasks":"2"}`
	if string(got) != want {
		t.Fatalf("GetFlowDashboardResponse wire shape changed.\n got: %s\nwant: %s", got, want)
	}
}

// Zero counts stay on the payload as "0" rather than vanishing into undefined,
// so a client rendering "0 overdue" has a key to read. Worth pinning: the
// omitempty tag on the int64 field suggests these would disappear, and the
// custom marshaller is the reason they do not.
func TestGetFlowDashboardResponseZeroCountsWireShape(t *testing.T) {
	res := flowdashboardv1.GetFlowDashboardResponse{}

	got, err := json.Marshal(&res)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	const want = `{"open_tasks":"0","in_progress_tasks":"0","completed_this_week":"0","overdue_tasks":"0"}`
	if string(got) != want {
		t.Fatalf("empty GetFlowDashboardResponse wire shape changed.\n got: %s\nwant: %s", got, want)
	}
}
