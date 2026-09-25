package license

import (
	"testing"
	"time"
)

func validReport() StatusReport {
	return StatusReport{Mode: ModeSelfHosted, Valid: true, Status: StatusValid}
}

func TestApplyRevocationNilRecordUnchanged(t *testing.T) {
	report := validReport()
	got := ApplyRevocation(report, nil, time.Now().UTC())
	if !got.Valid || got.Status != StatusValid {
		t.Fatalf("expected unchanged valid report, got %#v", got)
	}
}

func TestApplyRevocationRevokedInvalidates(t *testing.T) {
	got := ApplyRevocation(validReport(), &LicenseRecord{Status: LicenseStatusRevoked}, time.Now().UTC())
	if got.Valid {
		t.Fatalf("expected revoked license to be invalid, got %#v", got)
	}
	if got.Status != StatusRevoked {
		t.Fatalf("expected status %q, got %q", StatusRevoked, got.Status)
	}
}

func TestApplyRevocationActiveUnchanged(t *testing.T) {
	got := ApplyRevocation(validReport(), &LicenseRecord{Status: LicenseStatusActive}, time.Now().UTC())
	if !got.Valid || got.Status != StatusValid {
		t.Fatalf("expected active license to stay valid, got %#v", got)
	}
}

func TestApplyRevocationExpiredPastGrace(t *testing.T) {
	now := time.Now().UTC()
	expired := now.Add(-48 * time.Hour)
	grace := now.Add(-24 * time.Hour)
	got := ApplyRevocation(validReport(), &LicenseRecord{
		Status:     LicenseStatusActive,
		ExpiresAt:  &expired,
		GraceUntil: &grace,
	}, now)
	if got.Valid {
		t.Fatalf("expected expired-past-grace license to be invalid, got %#v", got)
	}
	if got.Status != StatusExpired {
		t.Fatalf("expected status %q, got %q", StatusExpired, got.Status)
	}
}

func TestApplyRevocationExpiredWithinGraceStaysValid(t *testing.T) {
	now := time.Now().UTC()
	expired := now.Add(-1 * time.Hour)
	grace := now.Add(24 * time.Hour)
	got := ApplyRevocation(validReport(), &LicenseRecord{
		Status:     LicenseStatusActive,
		ExpiresAt:  &expired,
		GraceUntil: &grace,
	}, now)
	if !got.Valid || got.Status != StatusValid {
		t.Fatalf("expected within-grace license to stay valid, got %#v", got)
	}
}
