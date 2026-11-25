package application

import (
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

func ghPullRequestBody(action, state string, merged bool, sha string) string {
	return fmt.Sprintf(`{"action":%q,"installation":{"id":9001},"repository":{"id":501,"full_name":"acme/web"},
"pull_request":{"number":7,"title":"Login page","state":%q,"draft":false,"merged":%t,
"html_url":"https://github.com/acme/web/pull/7","head":{"sha":%q,"ref":"fix/t-1-login"}}}`, action, state, merged, sha)
}

func ghCheckRunBody(id int64, name, status, conclusion, sha string) string {
	return fmt.Sprintf(`{"action":"completed","installation":{"id":9001},"repository":{"id":501,"full_name":"acme/web"},
"check_run":{"id":%d,"name":%q,"status":%q,"conclusion":%q,"head_sha":%q,"details_url":"https://ci.example/run/%d"}}`,
		id, name, status, conclusion, sha, id)
}

func ghStatusBody(state, sha string) string {
	return fmt.Sprintf(`{"installation":{"id":9001},"repository":{"id":501,"full_name":"acme/web"},
"sha":%q,"context":"ci/lint","state":%q,"target_url":"https://lint.example"}`, sha, state)
}

func TestGitHubChecksRollUpOnLinkedTasks(t *testing.T) {
	svc, ctx, _, _, _ := ghLinkSvc(t)
	p := createV2Project(ctx, t, svc, "checks", domain.VisibilityPrivate, v2UserA)
	task, err := svc.CreateTask(ctx, domain.TaskInput{ProjectID: p.ID, BusinessID: v2Business, OwnerID: v2UserA, ReporterID: v2UserA, Title: "Login"})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if _, err := svc.CreateTaskLink(ctx, task.ID, v2Business, v2UserA, "https://github.com/acme/web/pull/7"); err != nil {
		t.Fatalf("CreateTaskLink: %v", err)
	}
	stateOf := func() string {
		tasks := []domain.Task{{ID: task.ID, BusinessID: v2Business}}
		if err := svc.attachTaskChecksStates(ctx, tasks); err != nil {
			t.Fatalf("attachTaskChecksStates: %v", err)
		}
		return tasks[0].ChecksState
	}
	n := 0
	deliver := func(event, body string) {
		n++
		ghWebhook(ctx, t, svc, fmt.Sprintf("checks-%d-%d", time.Now().UnixNano(), n), event, body)
	}

	deliver("pull_request", ghPullRequestBody("opened", "open", false, "abc"))
	if got := stateOf(); got != "" {
		t.Fatalf("no checks yet: state = %q", got)
	}
	deliver("check_run", ghCheckRunBody(11, "build", "in_progress", "", "abc"))
	if got := stateOf(); got != domain.ChecksRunning {
		t.Fatalf("after build started: %q", got)
	}
	deliver("check_run", ghCheckRunBody(11, "build", "completed", "failure", "abc"))
	deliver("check_run", ghCheckRunBody(11, "build", "in_progress", "", "abc")) // late delivery
	if got := stateOf(); got != domain.ChecksFailing {
		t.Fatalf("after build failed (and a late in_progress): %q", got)
	}
	time.Sleep(1100 * time.Millisecond)                                              // updated_at has second precision
	deliver("check_run", ghCheckRunBody(12, "build", "completed", "success", "abc")) // re-run
	deliver("status", ghStatusBody("pending", "abc"))
	if got := stateOf(); got != domain.ChecksRunning {
		t.Fatalf("after the re-run passed and lint started: %q", got)
	}
	deliver("status", ghStatusBody("success", "abc"))
	if got := stateOf(); got != domain.ChecksPassing {
		t.Fatalf("after lint passed: %q", got)
	}

	prs, err := svc.ListTaskPullRequests(ctx, v2UserA, v2Business, task.ID)
	if err != nil || len(prs) != 1 {
		t.Fatalf("ListTaskPullRequests = %+v, %v", prs, err)
	}
	pr := prs[0]
	if !pr.Known || pr.Title != "Login page" || pr.State != "open" || pr.ChecksState != domain.ChecksPassing || len(pr.Checks) != 2 {
		t.Fatalf("pull request = %+v", pr)
	}

	// A new push moves the head: the old commit's checks no longer count.
	deliver("pull_request", ghPullRequestBody("synchronize", "open", false, "def"))
	if got := stateOf(); got != "" {
		t.Fatalf("after a new push with no checks yet: %q", got)
	}
	deliver("pull_request", ghPullRequestBody("closed", "closed", true, "def"))
	prs, _ = svc.ListTaskPullRequests(ctx, v2UserA, v2Business, task.ID)
	if prs[0].State != "merged" {
		t.Fatalf("merged PR state = %q", prs[0].State)
	}
	links, err := svc.ListTaskLinks(ctx, task.ID, v2Business)
	if err != nil || len(links) != 1 || links[0].State != "merged" {
		t.Fatalf("link state after merge = %+v, %v", links, err)
	}

	// Another business with a link to the same URL sees nothing of acme's checks.
	other := []domain.Task{{ID: task.ID, BusinessID: ghOtherBusiness}}
	if err := svc.attachTaskChecksStates(ctx, other); err != nil || other[0].ChecksState != "" {
		t.Fatalf("another business saw %q, %v", other[0].ChecksState, err)
	}
}
