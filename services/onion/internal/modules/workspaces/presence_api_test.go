package workspaces

import "testing"

func TestPresenceEntriesToProtoUsesContractStringIDs(t *testing.T) {
	entries := presenceEntriesToProto([]PresenceEntry{{
		EmployeeID: 42,
		UserID:     84,
		FirstName:  "Ada",
		LastName:   "Lovelace",
		Status:     "remote",
	}})

	if len(entries) != 1 {
		t.Fatalf("presence entry count = %d, want 1", len(entries))
	}
	entry := entries[0]
	if entry.EmployeeId != "42" || entry.UserId != "84" {
		t.Fatalf("contract ids = (%q, %q), want (42, 84)", entry.EmployeeId, entry.UserId)
	}
	if entry.FirstName == nil || *entry.FirstName != "Ada" {
		t.Fatalf("first_name = %v, want Ada", entry.FirstName)
	}
	if entry.LastName == nil || *entry.LastName != "Lovelace" {
		t.Fatalf("last_name = %v, want Lovelace", entry.LastName)
	}
	if entry.Status != "remote" {
		t.Fatalf("status = %q, want remote", entry.Status)
	}
}
