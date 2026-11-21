package api

import (
	"context"
	"encoding/json"
	"testing"
)

func TestCreateShiftRequest_UnmarshalCamelAndStringIDs(t *testing.T) {
	raw := `{"name":"Morning","locationId":"7","startTime":"08:00","endTime":"16:00","color":"#4CAF50"}`
	var req CreateShiftRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if req.LocationID != 7 || req.Name != "Morning" || req.StartTime != "08:00" || req.EndTime != "16:00" {
		t.Fatalf("unexpected decoded request: %+v", req)
	}
	h := &CreateShiftHandler{}
	if err := h.Validate(context.Background(), &req); err != nil {
		t.Fatalf("validate camelCase template: %v", err)
	}
}

func TestCreateShiftHandler_RejectsInvalidTimeStatusAndDate(t *testing.T) {
	h := &CreateShiftHandler{}
	cases := []CreateShiftRequest{
		{LocationID: 1, Name: "Morning", StartTime: "invalid", EndTime: "16:00"},
		{LocationID: 1, Name: "Morning", StartTime: "08:00", EndTime: "16:00", Status: "published"},
		{LocationID: 1, Name: "Morning", StartTime: "08:00", EndTime: "16:00", Color: "rgb(10, 20, 30)"},
		{StaffID: 5, Date: "Mon", StartTime: "08:00", EndTime: "16:00"},
	}
	for i, tc := range cases {
		req := tc
		if err := h.Validate(context.Background(), &req); err == nil {
			t.Errorf("case %d: expected validation error for %+v", i, req)
		}
	}
}
