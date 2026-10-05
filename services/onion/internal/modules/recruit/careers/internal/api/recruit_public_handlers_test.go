package api

import (
	"testing"

	"github.com/voxire/onion/internal/platform/storage"
)

func TestIsOwnedResumeURL(t *testing.T) {
	tests := []struct {
		name      string
		resumeURL string
		want      bool
	}{
		{name: "empty resume is optional", resumeURL: "", want: true},
		{name: "legacy external URL", resumeURL: "https://files.example.com/resume.pdf", want: true},
		{name: "matching generated key", resumeURL: "resumes/21/34/uuid-resume.pdf", want: true},
		{name: "different owner key", resumeURL: "resumes/22/34/uuid-resume.pdf", want: false},
		{name: "different career key", resumeURL: "resumes/21/35/uuid-resume.pdf", want: false},
		{name: "unscoped key", resumeURL: "private/resume.pdf", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isOwnedResumeURL(test.resumeURL, 21, 34); got != test.want {
				t.Fatalf("isOwnedResumeURL(%q) = %t, want %t", test.resumeURL, got, test.want)
			}
		})
	}
}

func TestValidateResumeMetadata(t *testing.T) {
	tests := []struct {
		name     string
		metadata storage.ObjectMetadata
		wantErr  bool
	}{
		{name: "allowed PDF", metadata: storage.ObjectMetadata{ContentType: "application/pdf", Size: storage.MaxResumeBytes}},
		{name: "unsupported content type", metadata: storage.ObjectMetadata{ContentType: "image/png", Size: 1}, wantErr: true},
		{name: "oversized resume", metadata: storage.ObjectMetadata{ContentType: "application/pdf", Size: storage.MaxResumeBytes + 1}, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := validateResumeMetadata(test.metadata); (err != nil) != test.wantErr {
				t.Fatalf("validateResumeMetadata() error = %v, wantErr %t", err, test.wantErr)
			}
		})
	}
}
