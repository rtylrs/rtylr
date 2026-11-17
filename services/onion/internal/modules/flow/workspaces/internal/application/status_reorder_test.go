package application

import (
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

func statusIDs(statuses []domain.Status) []int64 {
	ids := make([]int64, 0, len(statuses))
	for _, s := range statuses {
		ids = append(ids, s.ID)
	}
	return ids
}

func TestReorderProjectStatuses_WritesTheWholeOrder(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	p := createV2Project(ctx, t, svc, "status-reorder", "", v2UserA)

	before, err := svc.ListProjectStatuses(ctx, p.ID, v2UserA, v2Business)
	if err != nil {
		t.Fatalf("ListProjectStatuses: %v", err)
	}
	if len(before) < 3 {
		t.Fatalf("seeded %d statuses, want at least 3", len(before))
	}

	reversed := statusIDs(before)
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	got, err := svc.ReorderProjectStatuses(ctx, p.ID, v2UserA, v2Business, reversed)
	if err != nil {
		t.Fatalf("ReorderProjectStatuses: %v", err)
	}

	after, err := svc.ListProjectStatuses(ctx, p.ID, v2UserA, v2Business)
	if err != nil {
		t.Fatalf("ListProjectStatuses after: %v", err)
	}
	for i, s := range after {
		if s.ID != reversed[i] || s.SortOrder != int32(i*domain.StatusOrderSpacing) {
			t.Errorf("position %d = id %d at %d, want id %d at %d",
				i, s.ID, s.SortOrder, reversed[i], i*domain.StatusOrderSpacing)
		}
	}
	if len(got) != len(after) || got[0].ID != reversed[0] {
		t.Errorf("response order %v does not match stored order %v", statusIDs(got), statusIDs(after))
	}
}

func TestReorderProjectStatuses_RejectsAnythingButTheExactSet(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	p := createV2Project(ctx, t, svc, "status-reorder-set", "", v2UserA)
	other := createV2Project(ctx, t, svc, "status-reorder-other", "", v2UserA)

	before, err := svc.ListProjectStatuses(ctx, p.ID, v2UserA, v2Business)
	if err != nil {
		t.Fatalf("ListProjectStatuses: %v", err)
	}
	foreign, err := svc.ListProjectStatuses(ctx, other.ID, v2UserA, v2Business)
	if err != nil {
		t.Fatalf("ListProjectStatuses other: %v", err)
	}
	ids := statusIDs(before)

	cases := map[string][]int64{
		"missing":   ids[1:],
		"extra":     append(append([]int64{}, ids...), foreign[0].ID),
		"duplicate": append(append([]int64{}, ids[:len(ids)-1]...), ids[0]),
		"foreign":   append(append([]int64{}, ids[:len(ids)-1]...), foreign[0].ID),
		"empty":     {},
	}
	for name, order := range cases {
		if _, err := svc.ReorderProjectStatuses(ctx, p.ID, v2UserA, v2Business, order); !errors.Is(err, domain.ErrStatusOrderMismatch) {
			t.Errorf("%s: err = %v, want ErrStatusOrderMismatch", name, err)
		}
	}

	after, err := svc.ListProjectStatuses(ctx, p.ID, v2UserA, v2Business)
	if err != nil {
		t.Fatalf("ListProjectStatuses after: %v", err)
	}
	for i := range before {
		if after[i].ID != before[i].ID || after[i].SortOrder != before[i].SortOrder {
			t.Fatalf("a rejected reorder changed position %d", i)
		}
	}
}

func TestReorderProjectStatuses_RequiresProjectAdmin(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	p := createV2Project(ctx, t, svc, "status-reorder-role", "", v2UserA)
	if err := svc.AddProjectMember(ctx, domain.AddProjectMemberInput{
		ProjectID: p.ID, UserID: v2UserB, Role: domain.FlowRoleMember,
	}); err != nil {
		t.Fatalf("AddProjectMember: %v", err)
	}

	before, err := svc.ListProjectStatuses(ctx, p.ID, v2UserA, v2Business)
	if err != nil {
		t.Fatalf("ListProjectStatuses: %v", err)
	}
	ids := statusIDs(before)
	ids[0], ids[1] = ids[1], ids[0]

	if _, err := svc.ReorderProjectStatuses(ctx, p.ID, v2UserB, v2Business, ids); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("member reorder err = %v, want ErrUnauthorized", err)
	}
	after, err := svc.ListProjectStatuses(ctx, p.ID, v2UserA, v2Business)
	if err != nil {
		t.Fatalf("ListProjectStatuses after: %v", err)
	}
	if after[0].ID != before[0].ID {
		t.Errorf("a rejected member reorder changed the order")
	}
}
