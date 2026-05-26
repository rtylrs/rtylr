package handlers

import (
	"context"
	"testing"
)

func TestListViewsHandler_Validate_RequiresWorkspaceID(t *testing.T) {
	h := &ListViewsHandler{}
	req := &ListViewsRequest{WorkspaceID: 0}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when WorkspaceID is zero")
	}
}

func TestListViewsHandler_Validate_HappyPath(t *testing.T) {
	h := &ListViewsHandler{}
	req := &ListViewsRequest{WorkspaceID: 5}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestListViewsHandler_Validate_LargeWorkspaceID(t *testing.T) {
	h := &ListViewsHandler{}
	req := &ListViewsRequest{WorkspaceID: 99999}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreateViewHandler_Validate_RequiresName(t *testing.T) {
	h := &CreateViewHandler{}
	req := &CreateViewRequest{Name: "", WorkspaceID: 3}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when name is empty")
	}
}

func TestCreateViewHandler_Validate_RequiresWorkspaceID(t *testing.T) {
	h := &CreateViewHandler{}
	req := &CreateViewRequest{Name: "My View", WorkspaceID: 0}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when WorkspaceID is zero")
	}
}

func TestCreateViewHandler_Validate_HappyPath(t *testing.T) {
	h := &CreateViewHandler{}
	req := &CreateViewRequest{Name: "Active Tasks", WorkspaceID: 5}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreateViewHandler_Validate_BothEmpty(t *testing.T) {
	h := &CreateViewHandler{}
	req := &CreateViewRequest{}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when all fields are empty")
	}
}

func TestCreateViewHandler_Validate_WithOptionalFields(t *testing.T) {
	h := &CreateViewHandler{}
	req := &CreateViewRequest{
		Name:        "Sprint View",
		WorkspaceID: 2,
		SortBy:      "due_date",
		GroupBy:     "status",
		IsDefault:   true,
	}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error with optional fields: %v", err)
	}
}

func TestUpdateViewHandler_Validate_RequiresName(t *testing.T) {
	h := &UpdateViewHandler{}
	req := &UpdateViewRequest{Name: ""}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when name is empty")
	}
}

func TestUpdateViewHandler_Validate_HappyPath(t *testing.T) {
	h := &UpdateViewHandler{}
	req := &UpdateViewRequest{Name: "Updated View"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateViewHandler_Validate_WithSortBy(t *testing.T) {
	h := &UpdateViewHandler{}
	req := &UpdateViewRequest{Name: "Sorted View", SortBy: "priority"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateViewHandler_Validate_WithGroupBy(t *testing.T) {
	h := &UpdateViewHandler{}
	req := &UpdateViewRequest{Name: "Grouped View", GroupBy: "assignee"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreateViewHandler_Validate_DefaultFlag(t *testing.T) {
	h := &CreateViewHandler{}
	req := &CreateViewRequest{Name: "Default View", WorkspaceID: 1, IsDefault: true}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestListViewsHandler_Validate_OneID(t *testing.T) {
	h := &ListViewsHandler{}
	req := &ListViewsRequest{WorkspaceID: 1}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
