package api

import (
	"testing"

	"github.com/voxire/onion/internal/modules/recruit/careers/internal/domain"
)

func TestOGCardETagChangesWithMutedTextColor(t *testing.T) {
	before := domain.Customization{MutedTextColor: "#6b7280"}
	after := domain.Customization{MutedTextColor: "#ef4444"}

	if got, want := ogCardETag(after), ogCardETag(before); got == want {
		t.Fatalf("OG card ETag did not change after muted_text_color changed: %s", got)
	}
}
