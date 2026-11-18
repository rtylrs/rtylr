package domain

import (
	"strings"
	"testing"
)

const cdn = "https://api.voxire.com"

func color(c int32) *int32 { return &c }

func TestAvatarValidate(t *testing.T) {
	cases := []struct {
		name   string
		avatar Avatar
		ok     bool
	}{
		{"photo on our CDN", Avatar{Kind: "photo", URL: cdn + "/files/3f2b9c1e-7a4d-4b8e-9f00-1a2b3c4d5e6f"}, true},
		{"photo on another host", Avatar{Kind: "photo", URL: "https://evil.example/files/3f2b9c1e-7a4d"}, false},
		{"photo over http when the CDN is https", Avatar{Kind: "photo", URL: "http://api.voxire.com/files/3f2b9c1e-7a4d"}, false},
		{"photo outside /files", Avatar{Kind: "photo", URL: cdn + "/hub/user/v1/users/me"}, false},
		{"photo with a query string", Avatar{Kind: "photo", URL: cdn + "/files/3f2b9c1e-7a4d?x=1"}, false},
		{"photo with a traversal", Avatar{Kind: "photo", URL: cdn + "/files/../users"}, false},
		{"illustrated, offered style", Avatar{Kind: "illustrated", Style: "notionists", Seed: "k3x9a"}, true},
		{"illustrated, unknown style", Avatar{Kind: "illustrated", Style: "pepe", Seed: "k3x9a"}, false},
		{"illustrated, seed with markup", Avatar{Kind: "illustrated", Style: "gaze", Seed: "<svg>"}, false},
		{"initials, palette colour", Avatar{Kind: "initials", Color: color(11)}, true},
		{"initials, colour off the palette", Avatar{Kind: "initials", Color: color(12)}, false},
		{"initials, no colour", Avatar{Kind: "initials"}, false},
		{"emoji", Avatar{Kind: "emoji", Emoji: "🫠", Color: color(3)}, true},
		{"emoji with a skin tone", Avatar{Kind: "emoji", Emoji: "👍🏽", Color: color(0)}, true},
		{"emoji ZWJ sequence", Avatar{Kind: "emoji", Emoji: "👩‍💻", Color: color(0)}, true},
		{"emoji that is text", Avatar{Kind: "emoji", Emoji: "hello", Color: color(0)}, false},
		{"emoji that is a digit", Avatar{Kind: "emoji", Emoji: "5", Color: color(0)}, false},
		{"emoji that is too long", Avatar{Kind: "emoji", Emoji: strings.Repeat("😀", 11), Color: color(0)}, false},
		{"unknown kind", Avatar{Kind: "meme"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.avatar.Validate(cdn)
			if (err == nil) != tc.ok {
				t.Fatalf("Validate() error = %v, want ok = %v", err, tc.ok)
			}
		})
	}
}

func TestPhotoNeedsAConfiguredCDN(t *testing.T) {
	a := Avatar{Kind: "photo", URL: cdn + "/files/3f2b9c1e-7a4d"}
	if a.Validate("") == nil {
		t.Fatal("a photo validated with no CDN origin configured")
	}
}

func TestEncodeStoresOnlyTheKindsFields(t *testing.T) {
	stored, err := EncodeAvatar(Avatar{Kind: "initials", Color: color(4), URL: "https://x", Emoji: "😀"})
	if err != nil {
		t.Fatal(err)
	}
	if stored != `{"kind":"initials","color":4}` {
		t.Fatalf("stored %s", stored)
	}
}

func TestDecodeFallsBackToTheDefault(t *testing.T) {
	if DecodeAvatar("", cdn) != nil {
		t.Fatal("empty decoded to an avatar")
	}
	if DecodeAvatar("{not json", cdn) != nil {
		t.Fatal("corrupt JSON decoded to an avatar")
	}
	// A style that was retired since it was saved reads as the default rather
	// than an avatar the client cannot draw.
	if DecodeAvatar(`{"kind":"illustrated","style":"retired","seed":"a"}`, cdn) != nil {
		t.Fatal("a retired style decoded to an avatar")
	}
	got := DecodeAvatar(`{"kind":"emoji","emoji":"🗿","color":2}`, cdn)
	if got == nil || got.Emoji != "🗿" || got.Color == nil || *got.Color != 2 {
		t.Fatalf("decoded %+v", got)
	}
}
