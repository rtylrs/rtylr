package api

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_documents/internal/domain"
)

func TestLegacyDocumentsResponseMatchesRESTClientShape(t *testing.T) {
	expiry := time.Date(2027, time.January, 1, 0, 0, 0, 0, time.UTC)
	created := time.Date(2026, time.September, 20, 12, 30, 0, 0, time.UTC)
	documents := legacyDocumentsResponse([]domain.EmployeeDocumentRecord{{
		ID:           42,
		EmployeeID:   4101,
		EmployeeName: "Alice Worker",
		DocumentType: "id_card",
		Title:        "Passport",
		StorageMode:  "linked",
		ExternalURL:  "https://example.com/passport.pdf",
		ExpiryDate:   &expiry,
		CreatedAt:    created,
	}})
	if len(documents) != 1 {
		t.Fatalf("expected one document, got %d", len(documents))
	}

	encoded, err := json.Marshal(documents)
	if err != nil {
		t.Fatalf("marshal legacy document list: %v", err)
	}
	var rows []map[string]any
	if err := json.Unmarshal(encoded, &rows); err != nil {
		t.Fatalf("legacy endpoint must encode a JSON array: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected one encoded row, got %d", len(rows))
	}
	row := rows[0]
	for _, field := range []string{"id", "staff_id", "staff_name", "name", "type", "uploaded_at", "expires_at", "size_kb", "url"} {
		if _, ok := row[field]; !ok {
			t.Errorf("legacy response is missing %q: %s", field, encoded)
		}
	}
	if row["staff_name"] != "Alice Worker" || row["type"] != "I-9" || row["expires_at"] != "2027-01-01" {
		t.Fatalf("legacy response values do not match the live schema: %s", encoded)
	}
}
