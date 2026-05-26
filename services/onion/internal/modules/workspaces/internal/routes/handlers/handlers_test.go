package handlers

import (
	"context"
	"testing"
)

func TestListColumnsHandler_Validate_RequiresProjectID(t *testing.T) {
	h := &ListColumnsHandler{}
	req := &ListColumnsRequest{ProjectID: JSONID(0)}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when ProjectID is zero")
	}
}

func TestListColumnsHandler_Validate_HappyPath(t *testing.T) {
	h := &ListColumnsHandler{}
	req := &ListColumnsRequest{ProjectID: JSONID(5)}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreateColumnHandler_Validate_RequiresProjectID(t *testing.T) {
	h := &CreateColumnHandler{}
	req := &CreateColumnRequest{ProjectID: JSONID(0), Name: "To Do"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when ProjectID is zero")
	}
}

func TestCreateColumnHandler_Validate_RequiresName(t *testing.T) {
	h := &CreateColumnHandler{}
	req := &CreateColumnRequest{ProjectID: JSONID(3), Name: ""}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when name is empty")
	}
}

func TestCreateColumnHandler_Validate_HappyPath(t *testing.T) {
	h := &CreateColumnHandler{}
	req := &CreateColumnRequest{ProjectID: JSONID(3), Name: "In Progress"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreateColumnHandler_Validate_BothZero(t *testing.T) {
	h := &CreateColumnHandler{}
	req := &CreateColumnRequest{}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when both ProjectID and Name are empty")
	}
}

func TestCreateColumnHandler_Validate_WithColor(t *testing.T) {
	h := &CreateColumnHandler{}
	req := &CreateColumnRequest{ProjectID: JSONID(1), Name: "Done", Color: "#22c55e"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateColumnHandler_Validate_RequiresID(t *testing.T) {
	h := &UpdateColumnHandler{}
	req := &UpdateColumnRequest{ID: JSONID(0), Name: "Updated"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when ID is zero")
	}
}

func TestUpdateColumnHandler_Validate_RequiresName(t *testing.T) {
	h := &UpdateColumnHandler{}
	req := &UpdateColumnRequest{ID: JSONID(5), Name: ""}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when name is empty")
	}
}

func TestUpdateColumnHandler_Validate_HappyPath(t *testing.T) {
	h := &UpdateColumnHandler{}
	req := &UpdateColumnRequest{ID: JSONID(5), Name: "Done"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDeleteColumnHandler_Validate_RequiresID(t *testing.T) {
	h := &DeleteColumnHandler{}
	req := &DeleteColumnRequest{ID: JSONID(0)}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when ID is zero")
	}
}

func TestDeleteColumnHandler_Validate_HappyPath(t *testing.T) {
	h := &DeleteColumnHandler{}
	req := &DeleteColumnRequest{ID: JSONID(7)}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestMoveTaskToColumnHandler_Validate_RequiresTaskID(t *testing.T) {
	h := &MoveTaskToColumnHandler{}
	req := &MoveTaskToColumnRequest{TaskID: JSONID(0), ColumnID: JSONID(3)}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when TaskID is zero")
	}
}

func TestMoveTaskToColumnHandler_Validate_HappyPath(t *testing.T) {
	h := &MoveTaskToColumnHandler{}
	req := &MoveTaskToColumnRequest{TaskID: JSONID(10), ColumnID: JSONID(3)}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestJSONID_IsZero(t *testing.T) {
	var zero JSONID
	if !zero.IsZero() {
		t.Error("expected zero JSONID to be zero")
	}
	nonzero := JSONID(5)
	if nonzero.IsZero() {
		t.Error("expected non-zero JSONID to not be zero")
	}
}

func TestJSONID_Int64(t *testing.T) {
	j := JSONID(42)
	if j.Int64() != 42 {
		t.Errorf("expected 42, got %d", j.Int64())
	}
}

func TestListColumnsHandler_Validate_LargeProjectID(t *testing.T) {
	h := &ListColumnsHandler{}
	req := &ListColumnsRequest{ProjectID: JSONID(999999)}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error for large project id: %v", err)
	}
}
