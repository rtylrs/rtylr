package api

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/hr_equipment/internal/domain"
)

// An asset that was never assigned has no assignment date to report. Reporting
// purchase_date as assigned_date invents an assignment in the asset register.
func TestAssetToProtoDoesNotReportPurchaseDateAsAssignedDate(t *testing.T) {
	payload := assetToProto(domain.AssetRecord{
		ID:           1,
		OwnerID:      2,
		Name:         "MacBook",
		PurchaseDate: "2025-01-15",
		Status:       "available",
		Condition:    "good",
		CreatedAt:    time.Unix(0, 0),
		UpdatedAt:    time.Unix(0, 0),
	})

	if payload.AssignedDate != "" {
		t.Fatalf("AssignedDate = %q, want unset", payload.AssignedDate)
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if strings.Contains(string(encoded), "assigned_date") {
		t.Fatalf("assigned_date present in payload: %s", encoded)
	}
}
