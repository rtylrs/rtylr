package license

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	platformconfig "github.com/voxire/onion/internal/platform/config"
)

const (
	ModeHosted     = "hosted"
	ModeSelfHosted = "self_hosted"

	StatusHosted  = "hosted"
	StatusValid   = "valid"
	StatusMissing = "missing"
	StatusExpired = "expired"
	StatusInvalid = "invalid"
	StatusRevoked = "revoked"

	// #nosec G101 -- This is a public token format prefix, not a credential.
	tokenPrefix = "rtylr_license_v1"
)

var (
	// VendorPublicKey is injected into released binaries with -ldflags. It is
	// intentionally not read from customer-controlled runtime configuration.
	// #nosec G101 -- This is a public verification key placeholder populated at build time.
	VendorPublicKey string
	// BuildDeploymentMode is injected into released binaries with -ldflags so
	// deployment mode cannot be changed by flipping an environment variable.
	BuildDeploymentMode string

	ErrMissingLicense    = errors.New("license key is missing")
	ErrMissingPublicKey  = errors.New("license public key is missing")
	ErrMissingPrivateKey = errors.New("license signing private key is missing")
	ErrInvalidFormat     = errors.New("license key has an invalid format")
	ErrInvalidSignature  = errors.New("license signature is invalid")
	ErrExpired           = errors.New("license is expired")
)

type Entitlement struct {
	LicenseID  string            `json:"license_id"`
	OwnerEmail string            `json:"owner_email,omitempty"`
	CustomerID string            `json:"customer_id,omitempty"`
	Product    string            `json:"product"`
	Plan       string            `json:"plan,omitempty"`
	Issuer     string            `json:"issuer"`
	Features   []string          `json:"features,omitempty"`
	SeatLimit  int               `json:"seat_limit,omitempty"`
	IssuedAt   time.Time         `json:"issued_at,omitempty"`
	ExpiresAt  time.Time         `json:"expires_at,omitempty"`
	GraceUntil time.Time         `json:"grace_until,omitempty"`
	SelfHosted bool              `json:"self_hosted,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type StatusReport struct {
	Mode       string     `json:"mode"`
	Valid      bool       `json:"valid"`
	Status     string     `json:"status"`
	Reason     string     `json:"reason,omitempty"`
	LicenseID  string     `json:"license_id,omitempty"`
	OwnerEmail string     `json:"owner_email,omitempty"`
	CustomerID string     `json:"customer_id,omitempty"`
	Product    string     `json:"product,omitempty"`
	Plan       string     `json:"plan,omitempty"`
	Issuer     string     `json:"issuer,omitempty"`
	Features   []string   `json:"features,omitempty"`
	SeatLimit  int        `json:"seat_limit,omitempty"`
	IssuedAt   *time.Time `json:"issued_at,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	GraceUntil *time.Time `json:"grace_until,omitempty"`
	CheckedAt  time.Time  `json:"checked_at"`
}

func IsSelfHosted(cfg *platformconfig.AppConfig) bool {
	return isSelfHostedMode(deploymentMode(cfg))
}

func VerifyConfigured(cfg *platformconfig.AppConfig) StatusReport {
	return VerifyConfiguredAt(cfg, time.Now().UTC())
}

func VerifyConfiguredAt(cfg *platformconfig.AppConfig, now time.Time) StatusReport {
	key, err := configuredKey(cfg)
	if err != nil {
		return missingOrInvalidReportAt(cfg, err, now)
	}
	return VerifyKeyAt(cfg, key, now)
}

func VerifyKey(cfg *platformconfig.AppConfig, rawKey string) StatusReport {
	return VerifyKeyAt(cfg, rawKey, time.Now().UTC())
}

func VerifyKeyAt(cfg *platformconfig.AppConfig, rawKey string, now time.Time) StatusReport {
	mode := deploymentMode(cfg)
	rawKey = strings.TrimSpace(rawKey)

	if !IsSelfHosted(cfg) && rawKey == "" {
		return hostedReport(cfg, now)
	}

	entitlement, err := parseAndVerify(rawKey, publicKey(cfg), issuer(cfg), now, gracePeriod(cfg))
	if err != nil {
		return statusFromError(mode, err, now)
	}

	report := StatusReport{
		Mode:       mode,
		Valid:      true,
		Status:     StatusValid,
		LicenseID:  entitlement.LicenseID,
		OwnerEmail: entitlement.OwnerEmail,
		CustomerID: entitlement.CustomerID,
		Product:    entitlement.Product,
		Plan:       entitlement.Plan,
		Issuer:     entitlement.Issuer,
		Features:   entitlement.Features,
		SeatLimit:  entitlement.SeatLimit,
		CheckedAt:  now,
	}
	if !entitlement.IssuedAt.IsZero() {
		report.IssuedAt = &entitlement.IssuedAt
	}
	if !entitlement.ExpiresAt.IsZero() {
		report.ExpiresAt = &entitlement.ExpiresAt
	}
	if !entitlement.GraceUntil.IsZero() {
		report.GraceUntil = &entitlement.GraceUntil
	}
	return report
}

func GenerateSignedKey(rawPrivateKey string, entitlement Entitlement) (string, error) {
	priv, err := decodePrivateKey(rawPrivateKey)
	if err != nil {
		return "", err
	}
	if entitlement.IssuedAt.IsZero() {
		entitlement.IssuedAt = time.Now().UTC()
	}
	if entitlement.Product == "" {
		entitlement.Product = "rtylr"
	}

	payload, err := json.Marshal(entitlement)
	if err != nil {
		return "", err
	}
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	message := tokenPrefix + "." + encodedPayload
	signature := ed25519.Sign(priv, []byte(message))

	return message + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func PublicKeyFromPrivateKey(rawPrivateKey string) (string, error) {
	priv, err := decodePrivateKey(rawPrivateKey)
	if err != nil {
		return "", err
	}
	pub, ok := priv.Public().(ed25519.PublicKey)
	if !ok || len(pub) != ed25519.PublicKeySize {
		return "", ErrInvalidFormat
	}
	return base64.RawURLEncoding.EncodeToString(pub), nil
}

func HashKey(rawKey string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(rawKey)))
	return fmt.Sprintf("%x", sum[:])
}

func ParseAndVerify(rawKey, rawPublicKey, expectedIssuer string, now time.Time) (*Entitlement, error) {
	return parseAndVerify(rawKey, rawPublicKey, expectedIssuer, now, 0)
}

func parseAndVerify(rawKey, rawPublicKey, expectedIssuer string, now time.Time, configuredGracePeriod time.Duration) (*Entitlement, error) {
	rawKey = strings.TrimSpace(rawKey)
	if rawKey == "" {
		return nil, ErrMissingLicense
	}

	parts := strings.Split(rawKey, ".")
	if len(parts) != 3 || parts[0] != tokenPrefix {
		return nil, ErrInvalidFormat
	}

	pub, err := decodePublicKey(rawPublicKey)
	if err != nil {
		return nil, err
	}

	payloadBytes, err := decodeBase64(parts[1])
	if err != nil {
		return nil, fmt.Errorf("%w: payload", ErrInvalidFormat)
	}
	signature, err := decodeBase64(parts[2])
	if err != nil {
		return nil, fmt.Errorf("%w: signature", ErrInvalidFormat)
	}

	message := []byte(parts[0] + "." + parts[1])
	if !ed25519.Verify(pub, message, signature) {
		return nil, ErrInvalidSignature
	}

	var entitlement Entitlement
	if err := json.Unmarshal(payloadBytes, &entitlement); err != nil {
		return nil, fmt.Errorf("%w: payload json", ErrInvalidFormat)
	}
	if entitlement.Issuer == "" {
		entitlement.Issuer = expectedIssuer
	}
	if expectedIssuer != "" && entitlement.Issuer != expectedIssuer {
		return nil, ErrInvalidSignature
	}
	if entitlement.Product == "" {
		entitlement.Product = "rtylr"
	}
	if !entitlement.ExpiresAt.IsZero() && now.After(entitlement.ExpiresAt) {
		graceUntil := entitlement.GraceUntil
		if graceUntil.IsZero() && configuredGracePeriod > 0 {
			graceUntil = entitlement.ExpiresAt.Add(configuredGracePeriod)
			entitlement.GraceUntil = graceUntil
		}
		if graceUntil.IsZero() || now.After(graceUntil) {
			return nil, ErrExpired
		}
	}

	return &entitlement, nil
}

func HasFeature(report StatusReport, feature string) bool {
	if !report.Valid {
		return false
	}
	feature = strings.TrimSpace(feature)
	for _, enabled := range report.Features {
		if enabled == feature {
			return true
		}
	}
	return false
}

func configuredKey(cfg *platformconfig.AppConfig) (string, error) {
	if cfg == nil {
		if file := strings.TrimSpace(os.Getenv("LICENSE_FILE")); file != "" {
			return readLicenseFile(file)
		}
		if key := strings.TrimSpace(os.Getenv("LICENSE_KEY")); key != "" {
			return key, nil
		}
		return "", ErrMissingLicense
	}
	if file := strings.TrimSpace(cfg.LicenseFile); file != "" {
		return readLicenseFile(file)
	}
	if key := strings.TrimSpace(cfg.LicenseKey); key != "" {
		return key, nil
	}
	return "", ErrMissingLicense
}

func readLicenseFile(path string) (string, error) {
	// #nosec G304 G703 -- this path is deployment configuration, never request data; hosted builds do not accept it from clients.
	contents, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(contents)), nil
}

func decodePublicKey(raw string) (ed25519.PublicKey, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, ErrMissingPublicKey
	}
	decoded, err := decodeBase64(raw)
	if err != nil || len(decoded) != ed25519.PublicKeySize {
		return nil, ErrMissingPublicKey
	}
	return ed25519.PublicKey(decoded), nil
}

func decodePrivateKey(raw string) (ed25519.PrivateKey, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, ErrMissingPrivateKey
	}
	decoded, err := decodeBase64(raw)
	if err != nil {
		return nil, fmt.Errorf("%w: private key", ErrInvalidFormat)
	}
	switch len(decoded) {
	case ed25519.PrivateKeySize:
		return ed25519.PrivateKey(decoded), nil
	case ed25519.SeedSize:
		return ed25519.NewKeyFromSeed(decoded), nil
	default:
		return nil, fmt.Errorf("%w: private key", ErrInvalidFormat)
	}
}

func decodeBase64(raw string) ([]byte, error) {
	encodings := []*base64.Encoding{
		base64.RawURLEncoding,
		base64.URLEncoding,
		base64.RawStdEncoding,
		base64.StdEncoding,
	}
	var lastErr error
	for _, encoding := range encodings {
		decoded, err := encoding.DecodeString(raw)
		if err == nil {
			return decoded, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func missingOrInvalidReportAt(cfg *platformconfig.AppConfig, err error, now time.Time) StatusReport {
	if !IsSelfHosted(cfg) {
		return hostedReport(cfg, now)
	}
	return statusFromError(deploymentMode(cfg), err, now)
}

func hostedReport(cfg *platformconfig.AppConfig, now time.Time) StatusReport {
	return StatusReport{
		Mode:      deploymentMode(cfg),
		Valid:     true,
		Status:    StatusHosted,
		Product:   "rtylr",
		Plan:      ModeHosted,
		Issuer:    issuer(cfg),
		CheckedAt: now,
	}
}

func statusFromError(mode string, err error, now time.Time) StatusReport {
	status := StatusInvalid
	switch {
	case errors.Is(err, ErrMissingLicense):
		status = StatusMissing
	case errors.Is(err, ErrExpired):
		status = StatusExpired
	}
	return StatusReport{
		Mode:      mode,
		Valid:     false,
		Status:    status,
		Reason:    err.Error(),
		CheckedAt: now,
	}
}

func deploymentMode(cfg *platformconfig.AppConfig) string {
	if BuildDeploymentMode != "" {
		return normalizeMode(BuildDeploymentMode)
	}
	if cfg == nil {
		if mode := strings.TrimSpace(os.Getenv("DEPLOYMENT_MODE")); mode != "" {
			return normalizeMode(mode)
		}
		if isTruthy(os.Getenv("SELF_HOSTED")) {
			return ModeSelfHosted
		}
		return ModeHosted
	}
	if cfg.DeploymentMode != "" {
		return normalizeMode(cfg.DeploymentMode)
	}
	return ModeHosted
}

func issuer(cfg *platformconfig.AppConfig) string {
	if cfg != nil && cfg.LicenseIssuer != "" {
		return cfg.LicenseIssuer
	}
	if v := strings.TrimSpace(os.Getenv("LICENSE_ISSUER")); v != "" {
		return v
	}
	return "voxire"
}

func publicKey(_ *platformconfig.AppConfig) string {
	return VendorPublicKey
}

func gracePeriod(cfg *platformconfig.AppConfig) time.Duration {
	if cfg != nil {
		return cfg.LicenseGracePeriod
	}
	return 0
}

func normalizeMode(mode string) string {
	mode = strings.ToLower(strings.TrimSpace(mode))
	mode = strings.ReplaceAll(mode, "-", "_")
	switch mode {
	case "selfhosted":
		return ModeSelfHosted
	case "":
		return ModeHosted
	default:
		return mode
	}
}

func isSelfHostedMode(mode string) bool {
	return normalizeMode(mode) == ModeSelfHosted
}

func isTruthy(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}
