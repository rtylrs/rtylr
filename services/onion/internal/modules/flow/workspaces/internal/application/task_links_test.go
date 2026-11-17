package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

type stubPreviewer struct{ calls int }

func (p *stubPreviewer) Preview(_ context.Context, link domain.ClassifiedLink) (string, string) {
	p.calls++
	if link.Kind == domain.LinkKindGitHubPR {
		return "Add the developer switch", "open"
	}
	return "", ""
}

func TestTaskLinksLifecycle(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	preview := &stubPreviewer{}
	svc.WithLinkPreviewer(preview)

	p := createV2Project(ctx, t, svc, "links", "", v2UserA)
	task, err := svc.CreateTask(ctx, domain.TaskInput{ProjectID: p.ID, BusinessID: v2Business, Title: "linked", ReporterID: v2UserA})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	taskID := task.ID

	pr, err := svc.CreateTaskLink(ctx, taskID, v2Business, v2UserA, " https://github.com/rtylrs/rtylr/pull/1615 ")
	if err != nil {
		t.Fatalf("CreateTaskLink: %v", err)
	}
	if pr.Kind != domain.LinkKindGitHubPR || pr.Ref != "rtylrs/rtylr#1615" || pr.Title != "Add the developer switch" || pr.State != "open" || pr.AddedBy != v2UserA {
		t.Errorf("link = %+v", pr)
	}

	again, err := svc.CreateTaskLink(ctx, taskID, v2Business, v2UserB, "https://github.com/rtylrs/rtylr/pull/1615")
	if err != nil || again.ID != pr.ID {
		t.Fatalf("re-adding the same URL: id=%d err=%v, want existing %d", again.ID, err, pr.ID)
	}
	if preview.calls != 1 {
		t.Errorf("previewer called %d times, want 1 (duplicates are not refetched)", preview.calls)
	}

	if _, err := svc.CreateTaskLink(ctx, taskID, v2Business, v2UserA, "https://example.com/spec"); err != nil {
		t.Fatalf("generic link: %v", err)
	}
	if _, err := svc.CreateTaskLink(ctx, taskID, v2Business, v2UserA, "javascript:alert(1)"); !errors.Is(err, domain.ErrInvalidTaskLink) {
		t.Errorf("javascript: link err = %v", err)
	}
	if _, err := svc.CreateTaskLink(ctx, taskID, v2Business+1, v2UserA, "https://example.com/other"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("link from another business err = %v, want ErrUnauthorized", err)
	}

	links, err := svc.ListTaskLinks(ctx, taskID, v2Business)
	if err != nil || len(links) != 2 {
		t.Fatalf("ListTaskLinks: %d links, err %v", len(links), err)
	}
	if other, _ := svc.ListTaskLinks(ctx, taskID, v2Business+1); len(other) != 0 {
		t.Errorf("another business sees %d links", len(other))
	}

	if err := svc.DeleteTaskLink(ctx, pr.ID, v2Business, v2UserB); err != nil {
		t.Fatalf("DeleteTaskLink: %v", err)
	}
	links, _ = svc.ListTaskLinks(ctx, taskID, v2Business)
	if len(links) != 1 {
		t.Errorf("after delete: %d links, want 1", len(links))
	}

	detail, err := svc.GetTaskDetail(ctx, taskID)
	if err != nil {
		t.Fatalf("GetTaskDetail: %v", err)
	}
	var added, removed int
	for _, a := range detail.Activity {
		switch a.Action {
		case "link_added":
			added++
		case "link_removed":
			removed++
		}
	}
	if added != 2 || removed != 1 {
		t.Errorf("activity: %d link_added, %d link_removed; want 2 and 1", added, removed)
	}
}
