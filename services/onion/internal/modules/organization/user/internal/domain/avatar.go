package domain

import (
	"encoding/json"
	"errors"
	"net/url"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Avatar is the avatar a user chose. A user without one shows the default,
// initials on a colour derived from their id, which needs nothing stored.
type Avatar struct {
	Kind  string `json:"kind"`
	URL   string `json:"url,omitempty"`
	Style string `json:"style,omitempty"`
	Seed  string `json:"seed,omitempty"`
	Color *int32 `json:"color,omitempty"`
	Emoji string `json:"emoji,omitempty"`
}

const (
	AvatarKindPhoto       = "photo"
	AvatarKindIllustrated = "illustrated"
	AvatarKindInitials    = "initials"
	AvatarKindEmoji       = "emoji"

	// AvatarPaletteSize matches the twelve colours in @voxire/ui avatarColor.
	AvatarPaletteSize = 12
)

// AvatarStyles are the generated styles the portal offers. The client renders
// them, so an unknown style would draw nothing; refusing it here keeps a stale
// or hand-made request from saving an avatar nobody can see.
var AvatarStyles = map[string]struct{}{
	"notionists": {}, "lorelei": {}, "personas": {}, "micah": {},
	"avataaars": {}, "adventurer": {}, "big-smile": {}, "thumbs": {},
	"croodles": {}, "glass": {}, "cameo": {}, "shadows": {}, "cutouts": {},
	"voxel-art": {}, "line-face": {}, "voxel-bot": {}, "slice": {},
	"gaze": {}, "stack": {}, "patchwork": {}, AvatarStyleMoods: {},
}

// AvatarStyleMoods is RTYLR's own set of twelve reaction faces, drawn for the
// portal. Its seed names the face rather than feeding a generator, so only
// these ids draw anything.
const AvatarStyleMoods = "rtylr-moods"

var avatarMoods = map[string]struct{}{
	"stone-face": {}, "this-is-fine": {}, "side-eye": {}, "galaxy-brain": {},
	"stonks": {}, "mind-blown": {}, "facepalm": {}, "deal-with-it": {},
	"shrug": {}, "smug": {}, "thinking": {}, "nervous-sweat": {},
}

var (
	ErrAvatarInvalid = errors.New("invalid avatar")
	cdnFilePath      = regexp.MustCompile(`^/files/[A-Za-z0-9-]{8,64}$`)
	avatarSeed       = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)
)

// Validate checks the avatar against its kind. photoBase is the public CDN
// origin: a photo must be a file this platform stored, never an arbitrary URL,
// so an avatar cannot be used to make every colleague's browser fetch a
// third-party tracker or an unvetted image.
func (a Avatar) Validate(photoBase string) error {
	switch a.Kind {
	case AvatarKindPhoto:
		if !isCDNFile(a.URL, photoBase) {
			return ErrAvatarInvalid
		}
	case AvatarKindIllustrated:
		if _, ok := AvatarStyles[a.Style]; !ok || !avatarSeed.MatchString(a.Seed) {
			return ErrAvatarInvalid
		}
		if _, isMood := avatarMoods[a.Seed]; a.Style == AvatarStyleMoods && !isMood {
			return ErrAvatarInvalid
		}
	case AvatarKindInitials:
		if !validColor(a.Color) {
			return ErrAvatarInvalid
		}
	case AvatarKindEmoji:
		if !validColor(a.Color) || !isEmoji(a.Emoji) {
			return ErrAvatarInvalid
		}
	default:
		return ErrAvatarInvalid
	}
	return nil
}

// Normalized drops the fields that do not belong to the kind, so what is
// stored is exactly what was validated.
func (a Avatar) Normalized() Avatar {
	switch a.Kind {
	case AvatarKindPhoto:
		return Avatar{Kind: a.Kind, URL: a.URL}
	case AvatarKindIllustrated:
		return Avatar{Kind: a.Kind, Style: a.Style, Seed: a.Seed}
	case AvatarKindInitials:
		return Avatar{Kind: a.Kind, Color: a.Color}
	default:
		return Avatar{Kind: a.Kind, Emoji: a.Emoji, Color: a.Color}
	}
}

// EncodeAvatar is the stored form.
func EncodeAvatar(a Avatar) (string, error) {
	b, err := json.Marshal(a.Normalized())
	return string(b), err
}

// DecodeAvatar reads a stored avatar. A value that no longer validates (a
// style that was retired, say) reads as no avatar, so the user falls back to
// the default instead of an empty circle.
func DecodeAvatar(stored, photoBase string) *Avatar {
	if strings.TrimSpace(stored) == "" {
		return nil
	}
	var a Avatar
	if err := json.Unmarshal([]byte(stored), &a); err != nil {
		return nil
	}
	if a.Validate(photoBase) != nil {
		return nil
	}
	n := a.Normalized()
	return &n
}

func validColor(c *int32) bool {
	return c != nil && *c >= 0 && *c < AvatarPaletteSize
}

func isCDNFile(raw, base string) bool {
	if raw == "" || base == "" || len(raw) > 1024 {
		return false
	}
	u, err := url.Parse(raw)
	if err != nil || u.RawQuery != "" || u.Fragment != "" || u.User != nil {
		return false
	}
	b, err := url.Parse(strings.TrimRight(base, "/"))
	if err != nil || b.Host == "" {
		return false
	}
	return u.Scheme == b.Scheme && u.Host == b.Host &&
		cdnFilePath.MatchString(strings.TrimPrefix(u.Path, strings.TrimRight(b.Path, "/")))
}

// isEmoji accepts one emoji: possibly several code points (skin tones, ZWJ
// families, flags), but no letters, digits, spaces or markup.
func isEmoji(s string) bool {
	if s == "" || len(s) > 32 || !utf8.ValidString(s) {
		return false
	}
	for _, r := range s {
		if r < 0x80 && r != '#' && r != '*' {
			if !unicode.IsDigit(r) {
				return false
			}
		}
		if unicode.IsLetter(r) || unicode.IsSpace(r) || unicode.IsControl(r) {
			return false
		}
	}
	hasPictograph := false
	for _, r := range s {
		if r >= 0x2000 {
			hasPictograph = true
			break
		}
	}
	return hasPictograph && utf8.RuneCountInString(s) <= 10
}
