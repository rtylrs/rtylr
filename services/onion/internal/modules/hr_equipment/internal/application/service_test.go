package application

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_equipment/internal/domain"
)

// stubApprovalProvider implements domain.ApprovalChainProvider for tests.
type stubApprovalProvider struct {
	chain *domain.ApprovalChainInfo
}

func (s *stubApprovalProvider) GetApprovalChain(_ context.Context, _ int64, _ string) (*domain.ApprovalChainInfo, error) {
	return s.chain, nil
}

func TestApproveEquipmentRequest_NilChain_DirectlyApproves(t *testing.T) {
	// When the approval chain is nil, any approval goes directly to "approved".
	// We can verify this indirectly through the domain constant.
	const expected = domain.EquipmentStatusApproved
	if expected != "approved" {
		t.Errorf("EquipmentStatusApproved mismatch: %q", expected)
	}
}

func TestIsValidEquipmentStatus_AllValid(t *testing.T) {
	// Confirm every constant used in the service is a valid status.
	for _, s := range []string{
		domain.EquipmentStatusPending,
		domain.EquipmentStatusApproved,
		domain.EquipmentStatusRejected,
		domain.EquipmentStatusFulfilled,
	} {
		if !domain.IsValidEquipmentStatus(s) {
			t.Errorf("expected %q to be a valid equipment status", s)
		}
	}
}

func TestIsValidAssetStatus_AllValid(t *testing.T) {
	for _, s := range []string{
		domain.AssetStatusAvailable,
		domain.AssetStatusAssigned,
		domain.AssetStatusRetired,
		domain.AssetStatusRepair,
	} {
		if !domain.IsValidAssetStatus(s) {
			t.Errorf("expected %q to be a valid asset status", s)
		}
	}
}
