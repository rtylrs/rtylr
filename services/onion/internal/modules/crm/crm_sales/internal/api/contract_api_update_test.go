package api

import (
	"net/http"
	"testing"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/platform/money"
	crmsalesv1 "github.com/voxire/voxie/gen/go/crm/sales/v1"
)

func storedContract() domain.Contract {
	signedAt := "2026-01-02"
	closedBy := int64(42)
	return domain.Contract{
		ID:           7,
		BusinessID:   9,
		Plan:         "enterprise",
		MRR:          money.Amount(29900),
		Seats:        12,
		SignedAt:     &signedAt,
		SignedByName: "Ada Lovelace",
		ClosedBy:     &closedBy,
		Notes:        "original notes",
	}
}

// TestContractUpdateFromBody_PartialBodyKeepsStoredFields pins B23-R03: a body
// that only carries notes must not reset plan/mrr/seats/signed_at.
func TestContractUpdateFromBody_PartialBodyKeepsStoredFields(t *testing.T) {
	stored := storedContract()
	got, err := contractUpdateFromBody(&crmsalesv1.UpdateContractBody{Notes: "renewal terms"}, stored)
	if err != nil {
		t.Fatalf("contractUpdateFromBody: %v", err)
	}
	if got.Plan != stored.Plan {
		t.Errorf("Plan = %q, want stored %q", got.Plan, stored.Plan)
	}
	if got.MRR != stored.MRR {
		t.Errorf("MRR = %d, want stored %d", got.MRR, stored.MRR)
	}
	if got.Seats != stored.Seats {
		t.Errorf("Seats = %d, want stored %d", got.Seats, stored.Seats)
	}
	if got.SignedAt == nil || *got.SignedAt != *stored.SignedAt {
		t.Errorf("SignedAt = %v, want stored %v", got.SignedAt, stored.SignedAt)
	}
	if got.SignedByName != stored.SignedByName {
		t.Errorf("SignedByName = %q, want stored %q", got.SignedByName, stored.SignedByName)
	}
	if got.Notes != "renewal terms" {
		t.Errorf("Notes = %q, want renewal terms", got.Notes)
	}
	if got.ClosedBy == nil || *got.ClosedBy != *stored.ClosedBy {
		t.Errorf("ClosedBy = %v, want stored %v", got.ClosedBy, stored.ClosedBy)
	}
}

func TestContractUpdateFromBody_NilBodyKeepsEverything(t *testing.T) {
	stored := storedContract()
	got, err := contractUpdateFromBody(nil, stored)
	if err != nil {
		t.Fatalf("contractUpdateFromBody: %v", err)
	}
	if got.Plan != stored.Plan || got.MRR != stored.MRR || got.Seats != stored.Seats ||
		got.SignedAt != stored.SignedAt || got.SignedByName != stored.SignedByName || got.Notes != stored.Notes {
		t.Fatalf("nil body changed stored values: %+v", got)
	}
}

func TestContractUpdateFromBody_PresentFieldsReplace(t *testing.T) {
	stored := storedContract()
	signed := "2026-03-04"
	got, err := contractUpdateFromBody(&crmsalesv1.UpdateContractBody{
		Plan:         "starter",
		Mrr:          "0",
		Seats:        3,
		SignedAt:     &signed,
		SignedByName: "Grace Hopper",
		Notes:        "new notes",
	}, stored)
	if err != nil {
		t.Fatalf("contractUpdateFromBody: %v", err)
	}
	if got.Plan != "starter" {
		t.Errorf("Plan = %q, want starter", got.Plan)
	}
	if !got.MRR.IsZero() {
		t.Errorf("MRR = %s, want 0 (explicit zero)", got.MRR.String())
	}
	if got.Seats != 3 {
		t.Errorf("Seats = %d, want 3", got.Seats)
	}
	if got.SignedAt == nil || *got.SignedAt != signed {
		t.Errorf("SignedAt = %v, want %q", got.SignedAt, signed)
	}
	if got.SignedByName != "Grace Hopper" || got.Notes != "new notes" {
		t.Errorf("signer/notes not replaced: %+v", got)
	}
}

func TestContractUpdateFromBody_SignedAtEmptyClears(t *testing.T) {
	stored := storedContract()
	empty := ""
	got, err := contractUpdateFromBody(&crmsalesv1.UpdateContractBody{SignedAt: &empty}, stored)
	if err != nil {
		t.Fatalf("contractUpdateFromBody: %v", err)
	}
	if got.SignedAt != nil {
		t.Errorf("SignedAt = %v, want nil (explicit clear)", got.SignedAt)
	}
}

func TestContractUpdateFromBody_RejectsInvalidValues(t *testing.T) {
	stored := storedContract()
	cases := []struct {
		name string
		body *crmsalesv1.UpdateContractBody
	}{
		{"negative mrr", &crmsalesv1.UpdateContractBody{Mrr: "-1"}},
		{"malformed mrr", &crmsalesv1.UpdateContractBody{Mrr: "not-money"}},
		{"negative seats", &crmsalesv1.UpdateContractBody{Seats: -5}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := contractUpdateFromBody(tc.body, stored)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if got := crmPermissionStatus(err); got != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", got, http.StatusBadRequest)
			}
		})
	}
}

func TestContractUpdateFromBody_SeatsZeroKeepsStored(t *testing.T) {
	stored := storedContract()
	got, err := contractUpdateFromBody(&crmsalesv1.UpdateContractBody{}, stored)
	if err != nil {
		t.Fatalf("contractUpdateFromBody: %v", err)
	}
	if got.Seats != stored.Seats {
		t.Errorf("Seats = %d, want stored %d", got.Seats, stored.Seats)
	}
}

func TestContractUpdateFromBody_WhitespacePlanKeepsStored(t *testing.T) {
	stored := storedContract()
	got, err := contractUpdateFromBody(&crmsalesv1.UpdateContractBody{Plan: "   "}, stored)
	if err != nil {
		t.Fatalf("contractUpdateFromBody: %v", err)
	}
	if got.Plan != stored.Plan {
		t.Errorf("Plan = %q, want stored %q", got.Plan, stored.Plan)
	}
}
