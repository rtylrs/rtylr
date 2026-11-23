// Package domain holds the pager's tickets and the rules that depend on
// neither storage nor transport.
package domain

import (
	"errors"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	StatusWaiting   = "waiting"
	StatusReady     = "ready"
	StatusCollected = "collected"
	StatusExpired   = "expired"

	LabelNumber = "number"
	LabelName   = "name"

	// LabelModeEither takes a number when the entry looks like one and a name
	// otherwise. It is the default.
	LabelModeNumber = "number"
	LabelModeName   = "name"
	LabelModeEither = "either"
)

// OpenLifetime is how long a labelled ticket stays open. Receipt numbers repeat
// every day, so an older open ticket reads as expired and its number is free
// again.
const OpenLifetime = time.Hour

// UnlabelledLifetime is how long a scanned phone waits for the cashier to label
// it before its ticket expires and leaves the board.
const UnlabelledLifetime = 15 * time.Minute

// MaxOpenUnlabelledTickets bounds what anonymous scans can open at one venue.
const MaxOpenUnlabelledTickets = 200

// MaxScansPerClient is how many scans one client address gets at one venue
// per UnlabelledLifetime, so it can hold at most that many open unlabelled
// tickets and filling a venue's cap takes many addresses, not one loop.
const MaxScansPerClient = 20

// Match numbers are what a scanned phone shows the cashier: three digits while
// any is free, four once every three digit number is taken. The open
// unlabelled cap keeps a venue far below the 900 three digit numbers, so four
// digits only appear if the cap is raised.
const (
	MinMatchNumber         = 100
	MaxMatchNumber         = 999
	MinFallbackMatchNumber = 1000
	MaxFallbackMatchNumber = 9999
)

const (
	maxNumberLength = 12
	maxNameLength   = 40
)

var (
	ErrVenueNotFound      = errors.New("pager venue not found")
	ErrTicketNotFound     = errors.New("pager ticket not found")
	ErrInvalidLabel       = errors.New("pager label is invalid")
	ErrInvalidLabelMode   = errors.New("pager label mode is invalid")
	ErrInvalidTransition  = errors.New("pager ticket is not in a state that allows this")
	ErrTooManyOpenTickets = errors.New("pager venue has too many open tickets")
	// ErrLabelTaken means another open ticket already holds the number.
	ErrLabelTaken = errors.New("pager number already has an open ticket")
	// ErrCodeTaken is the storage layer reporting that another open
	// unlabelled ticket already shows the match number.
	ErrCodeTaken = errors.New("pager match number already in use")
)

// Ticket is one phone waiting for its order. Label and LabelKind are empty
// until the cashier labels it. Code is what the phone shows the cashier until
// then: a match number such as "427", or on a ticket opened before the numbers,
// its word key ("panda2") or two character code ("A7"). The time fields are zero until the moment
// they describe has happened.
type Ticket struct {
	ID             int64
	BusinessID     int64
	Label          string
	LabelKind      string
	Code           string
	Status         string
	PublicToken    string
	CreatedAt      time.Time
	ReadyAt        time.Time
	RemindedAt     time.Time
	AcknowledgedAt time.Time
}

func (t Ticket) Labelled() bool {
	return t.LabelKind != ""
}

// OpenSince is the oldest creation time a labelled ticket can have and still
// be open.
func OpenSince(now time.Time) time.Time {
	return now.Add(-OpenLifetime)
}

// UnlabelledSince is the same for a ticket nobody has labelled.
func UnlabelledSince(now time.Time) time.Time {
	return now.Add(-UnlabelledLifetime)
}

// StatusAt is the status as read at now: an open ticket past its lifetime is
// expired whatever the stored status says.
func (t Ticket) StatusAt(now time.Time) string {
	open := t.Status == StatusWaiting || t.Status == StatusReady
	since := OpenSince(now)
	if !t.Labelled() {
		since = UnlabelledSince(now)
	}
	if open && t.CreatedAt.Before(since) {
		return StatusExpired
	}
	return t.Status
}

// ShownCode is the code as the phone and the board show it. A match number is
// shown as it is. A ticket opened before the numbers keeps its old code in
// capitals ("PANDA2", "A7") until it is labelled or expires, so both screens
// still show the same characters.
func ShownCode(code string) string {
	return strings.ToUpper(code)
}

type Settings struct {
	// Enabled makes the venue reachable on the public pager routes.
	Enabled bool
	// LabelMode is what the cashier labels a ticket with.
	LabelMode string
}

// DefaultSettings is what a business that never saved any has.
func DefaultSettings() Settings {
	return Settings{LabelMode: LabelModeEither}
}

func ValidLabelMode(mode string) bool {
	return mode == LabelModeNumber || mode == LabelModeName || mode == LabelModeEither
}

// Venue is the business a pager link points at.
type Venue struct {
	BusinessID int64
	Name       string
	LogoURL    string
}

// ParseLabel turns what the cashier typed into a label and its kind under the
// venue's mode. An order number has at least one digit, so in either mode "42"
// and "A17" are numbers and "Sara" is a name.
func ParseLabel(raw, mode string) (label, kind string, err error) {
	switch mode {
	case LabelModeNumber:
		label, err = orderNumber(raw)
		return label, LabelNumber, err
	case LabelModeName:
		label, err = NormalizeName(raw)
		return label, LabelName, err
	}
	if number, err := orderNumber(raw); err == nil {
		return number, LabelNumber, nil
	}
	label, err = NormalizeName(raw)
	return label, LabelName, err
}

func orderNumber(raw string) (string, error) {
	number, err := NormalizeNumber(raw)
	if err != nil {
		return "", err
	}
	if !strings.ContainsFunc(number, unicode.IsDigit) {
		return "", ErrInvalidLabel
	}
	return number, nil
}

// NormalizeNumber turns a receipt number into the stored form: trimmed,
// Arabic-Indic and Persian digits as ASCII, letters in capitals. Anything other
// than 1 to 12 ASCII letters and digits is refused.
func NormalizeNumber(raw string) (string, error) {
	var out strings.Builder
	for _, r := range strings.TrimSpace(raw) {
		switch {
		case r >= '0' && r <= '9', r >= 'A' && r <= 'Z':
		case r >= 'a' && r <= 'z':
			r -= 'a' - 'A'
		case r >= '٠' && r <= '٩':
			r = '0' + (r - '٠')
		case r >= '۰' && r <= '۹':
			r = '0' + (r - '۰')
		default:
			return "", ErrInvalidLabel
		}
		if out.Len() == maxNumberLength {
			return "", ErrInvalidLabel
		}
		out.WriteRune(r)
	}
	if out.Len() == 0 {
		return "", ErrInvalidLabel
	}
	return out.String(), nil
}

// NormalizeName trims a customer's name and folds runs of spaces into one. It
// must be 1 to 40 characters with no control characters.
func NormalizeName(raw string) (string, error) {
	if strings.ContainsFunc(raw, func(r rune) bool { return unicode.IsControl(r) && !unicode.IsSpace(r) }) {
		return "", ErrInvalidLabel
	}
	name := strings.Join(strings.Fields(raw), " ")
	if name == "" || utf8.RuneCountInString(name) > maxNameLength || !utf8.ValidString(name) {
		return "", ErrInvalidLabel
	}
	return name, nil
}
