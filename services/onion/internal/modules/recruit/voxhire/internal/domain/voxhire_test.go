package domain

import "testing"

func TestSubmissionCVStatus(t *testing.T) {
	if got := SubmissionCVStatus(false); got != CVStatusPending {
		t.Fatalf("SubmissionCVStatus(false) = %q, want %q", got, CVStatusPending)
	}
	if got := SubmissionCVStatus(true); got != CVStatusUploadFailed {
		t.Fatalf("SubmissionCVStatus(true) = %q, want %q", got, CVStatusUploadFailed)
	}
	// The status column is VARCHAR(32).
	if len(CVStatusUploadFailed) > 32 {
		t.Fatalf("%q does not fit cv_processing_status", CVStatusUploadFailed)
	}
}
