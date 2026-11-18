package api

import (
	"encoding/json"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/user/internal/domain"
	hubuserv1 "github.com/voxire/voxie/gen/go/hub/user/v1"
)

func TestAvatarProtoRoundTrip(t *testing.T) {
	two := int32(2)
	cases := []domain.Avatar{
		{Kind: "photo", URL: "https://api.voxire.com/files/3f2b9c1e-7a4d"},
		{Kind: "illustrated", Style: "gaze", Seed: "k3x9a"},
		{Kind: "initials", Color: &two},
		{Kind: "emoji", Emoji: "🗿", Color: &two},
	}
	for _, in := range cases {
		got := avatarFromProto(avatarToProto(in))
		if got.Kind != in.Kind || got.URL != in.URL || got.Style != in.Style ||
			got.Seed != in.Seed || got.Emoji != in.Emoji ||
			(in.Color != nil) != (got.Color != nil) ||
			(in.Color != nil && *in.Color != *got.Color) {
			t.Fatalf("round trip of %+v gave %+v", in, got)
		}
	}
}

func TestAvatarOnTheWireLeavesUnusedFieldsOut(t *testing.T) {
	zero := int32(0)
	payload, err := json.Marshal(&hubuserv1.GetMeResponse{
		Id:     1,
		Avatar: avatarToProto(domain.Avatar{Kind: "initials", Color: &zero}),
	})
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		Avatar map[string]any `json:"avatar"`
	}
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}
	// Colour 0 is a real choice (the first palette colour) and must be sent.
	if decoded.Avatar["kind"] != "initials" || decoded.Avatar["color"] != float64(0) {
		t.Fatalf("avatar on the wire = %v", decoded.Avatar)
	}
	for _, field := range []string{"url", "style", "seed", "emoji"} {
		if _, present := decoded.Avatar[field]; present {
			t.Fatalf("unused field %q was sent: %v", field, decoded.Avatar)
		}
	}
}
