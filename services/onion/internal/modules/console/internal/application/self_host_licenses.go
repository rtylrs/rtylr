package application

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/console/internal/infrastructure"
	platformconfig "github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/license"
)

const defaultSelfHostPlan = "self_hosted"

var (
	ErrSelfHostLicenseNotFound          = errors.New("self-host license not found")
	ErrSelfHostLicenseSigningKeyMissing = errors.New("license signing private key is not configured")
	ErrInvalidSelfHostLicenseInput      = errors.New("invalid self-host license input")

	defaultSelfHostFeatures = []string{"dash", "pos", "erp", "hr", "crm", "finance", "flow", "insights"}
)

type SelfHostLicenseDTO struct {
	ID            int64    `json:"id"`
	OwnerEmail    string   `json:"owner_email"`
	CustomerName  string   `json:"customer_name"`
	Plan          string   `json:"plan"`
	Features      []string `json:"features"`
	SeatLimit     int      `json:"seat_limit"`
	ExpiresAt     *string  `json:"expires_at,omitempty"`
	GraceUntil    *string  `json:"grace_until,omitempty"`
	Status        string   `json:"status"`
	KeyPreview    string   `json:"key_preview"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
	RegeneratedAt *string  `json:"regenerated_at,omitempty"`
	RevokedAt     *string  `json:"revoked_at,omitempty"`
}

type ListSelfHostLicensesResult struct {
	Licenses []SelfHostLicenseDTO `json:"licenses"`
	Total    int64                `json:"total"`
	Limit    int                  `json:"limit"`
	Offset   int                  `json:"offset"`
}

type SelfHostLicenseKeyResult struct {
	License    SelfHostLicenseDTO `json:"license"`
	LicenseKey string             `json:"license_key"`
	PublicKey  string             `json:"public_key"`
}

type CreateSelfHostLicenseInput struct {
	OwnerEmail   string   `json:"owner_email"`
	CustomerName string   `json:"customer_name"`
	Plan         string   `json:"plan"`
	Features     []string `json:"features"`
	SeatLimit    int      `json:"seat_limit"`
	ExpiresAt    string   `json:"expires_at"`
	GraceUntil   string   `json:"grace_until"`
}

func (s *Service) ListSelfHostLicenses(ctx context.Context, limit, offset int, email string) (*ListSelfHostLicensesResult, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	rows, total, err := s.repo.ListSelfHostLicenses(ctx, limit, offset, email)
	if err != nil {
		return nil, err
	}

	licenses := make([]SelfHostLicenseDTO, 0, len(rows))
	for _, row := range rows {
		licenses = append(licenses, toSelfHostLicenseDTO(row))
	}

	return &ListSelfHostLicensesResult{
		Licenses: licenses,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	}, nil
}

func (s *Service) CreateSelfHostLicense(ctx context.Context, input CreateSelfHostLicenseInput, adminID int64) (*SelfHostLicenseKeyResult, error) {
	ownerEmail, err := normalizeOwnerEmail(input.OwnerEmail)
	if err != nil {
		return nil, err
	}
	plan := normalizeSelfHostPlan(input.Plan)
	features := normalizeSelfHostFeatures(input.Features)
	expiresAt, err := parseOptionalLicenseTime(input.ExpiresAt)
	if err != nil {
		return nil, err
	}
	graceUntil, err := parseOptionalLicenseTime(input.GraceUntil)
	if err != nil {
		return nil, err
	}

	if existing, err := s.repo.GetSelfHostLicenseByOwnerEmail(ctx, ownerEmail); err == nil {
		publicKey, err := currentLicensePublicKey()
		if err != nil {
			return nil, err
		}
		return &SelfHostLicenseKeyResult{
			License: toSelfHostLicenseDTO(*existing),
			// Keys are shown only on the initial successful issuance. Existing
			// records are never allowed to re-disclose their secret.
			LicenseKey: "",
			PublicKey:  publicKey,
		}, nil
	} else if !errors.Is(err, infrastructure.ErrSelfHostLicenseNotFound) {
		return nil, err
	}

	licenseKey, publicKey, err := generateSelfHostLicenseKey(ownerEmail, plan, features, input.SeatLimit, expiresAt, graceUntil)
	if err != nil {
		return nil, err
	}
	featuresJSON, err := json.Marshal(features)
	if err != nil {
		return nil, err
	}

	id, err := s.repo.CreateSelfHostLicense(ctx, infrastructure.CreateSelfHostLicenseParams{
		OwnerEmail:   ownerEmail,
		CustomerName: strings.TrimSpace(input.CustomerName),
		Plan:         plan,
		FeaturesJSON: string(featuresJSON),
		SeatLimit:    input.SeatLimit,
		ExpiresAt:    expiresAt,
		GraceUntil:   graceUntil,
		// Keep only the hash at rest. The signed key is returned once to the
		// administrator and can be regenerated if it is lost.
		LicenseKey: "",
		KeyHash:    license.HashKey(licenseKey),
		CreatedBy:  adminID,
	})
	if err != nil {
		return nil, err
	}

	row, err := s.repo.GetSelfHostLicenseByID(ctx, id)
	if err != nil {
		return nil, err
	}

	defer invalidateConsoleScope(ctx)
	return &SelfHostLicenseKeyResult{
		License:    toSelfHostLicenseDTO(*row),
		LicenseKey: licenseKey,
		PublicKey:  publicKey,
	}, nil
}

func (s *Service) RegenerateSelfHostLicenseKey(ctx context.Context, id, adminID int64) (*SelfHostLicenseKeyResult, error) {
	row, err := s.repo.GetSelfHostLicenseByID(ctx, id)
	if err != nil {
		if errors.Is(err, infrastructure.ErrSelfHostLicenseNotFound) {
			return nil, ErrSelfHostLicenseNotFound
		}
		return nil, err
	}

	features := featuresFromRow(*row)
	licenseKey, publicKey, err := generateSelfHostLicenseKey(
		row.OwnerEmail,
		normalizeSelfHostPlan(row.Plan),
		features,
		int(row.SeatLimit.Int64),
		nullTimePtr(row.ExpiresAt),
		nullTimePtr(row.GraceUntil),
	)
	if err != nil {
		return nil, err
	}

	if err := s.repo.RegenerateSelfHostLicenseKey(ctx, id, row.OwnerEmail, "", license.HashKey(licenseKey), adminID); err != nil {
		if errors.Is(err, infrastructure.ErrSelfHostLicenseNotFound) {
			return nil, ErrSelfHostLicenseNotFound
		}
		return nil, err
	}
	updated, err := s.repo.GetSelfHostLicenseByID(ctx, id)
	if err != nil {
		return nil, err
	}

	defer invalidateConsoleScope(ctx)
	return &SelfHostLicenseKeyResult{
		License:    toSelfHostLicenseDTO(*updated),
		LicenseKey: licenseKey,
		PublicKey:  publicKey,
	}, nil
}

func (s *Service) RevokeSelfHostLicense(ctx context.Context, id, adminID int64) (*SelfHostLicenseDTO, error) {
	row, err := s.repo.GetSelfHostLicenseByID(ctx, id)
	if err != nil {
		if errors.Is(err, infrastructure.ErrSelfHostLicenseNotFound) {
			return nil, ErrSelfHostLicenseNotFound
		}
		return nil, err
	}
	if err := s.repo.RevokeSelfHostLicense(ctx, id, row.OwnerEmail, adminID); err != nil {
		if errors.Is(err, infrastructure.ErrSelfHostLicenseNotFound) {
			return nil, ErrSelfHostLicenseNotFound
		}
		return nil, err
	}
	row, err = s.repo.GetSelfHostLicenseByID(ctx, id)
	if err != nil {
		return nil, err
	}
	dto := toSelfHostLicenseDTO(*row)
	defer invalidateConsoleScope(ctx)
	return &dto, nil
}

func generateSelfHostLicenseKey(ownerEmail, plan string, features []string, seatLimit int, expiresAt, graceUntil *time.Time) (string, string, error) {
	privateKey := licenseSigningPrivateKey()
	if strings.TrimSpace(privateKey) == "" {
		return "", "", ErrSelfHostLicenseSigningKeyMissing
	}

	now := time.Now().UTC()
	entitlement := license.Entitlement{
		LicenseID:  randomLicenseID(),
		OwnerEmail: ownerEmail,
		CustomerID: ownerEmail,
		Product:    "rtylr",
		Plan:       plan,
		Issuer:     licenseIssuer(),
		Features:   features,
		SeatLimit:  seatLimit,
		IssuedAt:   now,
		SelfHosted: true,
		Metadata: map[string]string{
			"sales_issued": "true",
		},
	}
	if expiresAt != nil {
		entitlement.ExpiresAt = expiresAt.UTC()
	}
	if graceUntil != nil {
		entitlement.GraceUntil = graceUntil.UTC()
	}

	licenseKey, err := license.GenerateSignedKey(privateKey, entitlement)
	if err != nil {
		return "", "", err
	}
	publicKey, err := license.PublicKeyFromPrivateKey(privateKey)
	if err != nil {
		return "", "", err
	}
	return licenseKey, publicKey, nil
}

func currentLicensePublicKey() (string, error) {
	privateKey := licenseSigningPrivateKey()
	if strings.TrimSpace(privateKey) != "" {
		return license.PublicKeyFromPrivateKey(privateKey)
	}
	if publicKey := strings.TrimSpace(licensePublicKey()); publicKey != "" {
		return publicKey, nil
	}
	return "", ErrSelfHostLicenseSigningKeyMissing
}

func licenseSigningPrivateKey() string {
	return platformconfig.GetLicenseSigningPrivateKey()
}

func licensePublicKey() string {
	return platformconfig.GetLicensePublicKey()
}

func licenseIssuer() string {
	return platformconfig.GetLicenseIssuer()
}

func normalizeOwnerEmail(raw string) (string, error) {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return "", fmt.Errorf("%w: owner email is required", ErrInvalidSelfHostLicenseInput)
	}
	addr, err := mail.ParseAddress(raw)
	if err != nil || addr.Address != raw {
		return "", fmt.Errorf("%w: owner email is invalid", ErrInvalidSelfHostLicenseInput)
	}
	return addr.Address, nil
}

func normalizeSelfHostPlan(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultSelfHostPlan
	}
	return raw
}

func normalizeSelfHostFeatures(raw []string) []string {
	if len(raw) == 0 {
		return append([]string(nil), defaultSelfHostFeatures...)
	}
	seen := make(map[string]struct{}, len(raw))
	result := make([]string, 0, len(raw))
	for _, value := range raw {
		feature := strings.ToLower(strings.TrimSpace(value))
		if feature == "" {
			continue
		}
		if _, ok := seen[feature]; ok {
			continue
		}
		seen[feature] = struct{}{}
		result = append(result, feature)
	}
	if len(result) == 0 {
		return append([]string(nil), defaultSelfHostFeatures...)
	}
	return result
}

func parseOptionalLicenseTime(raw string) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04",
		"2006-01-02",
	}
	var lastErr error
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, raw)
		if err == nil {
			utc := parsed.UTC()
			return &utc, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("%w: invalid time %q: %w", ErrInvalidSelfHostLicenseInput, raw, lastErr)
}

func randomLicenseID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("selfhost_%d", time.Now().UnixNano())
	}
	return "selfhost_" + hex.EncodeToString(b[:])
}

func toSelfHostLicenseDTO(row infrastructure.SelfHostLicenseRow) SelfHostLicenseDTO {
	dto := SelfHostLicenseDTO{
		ID:         row.ID,
		OwnerEmail: row.OwnerEmail,
		Plan:       row.Plan,
		Features:   featuresFromRow(row),
		Status:     row.Status,
		// License material is never returned from list/detail responses.
		KeyPreview: "",
		CreatedAt:  formatNullTime(row.CreatedAt),
		UpdatedAt:  formatNullTime(row.UpdatedAt),
	}
	if row.CustomerName.Valid {
		dto.CustomerName = row.CustomerName.String
	}
	if row.SeatLimit.Valid {
		dto.SeatLimit = int(row.SeatLimit.Int64)
	}
	dto.ExpiresAt = formatNullTimePtr(row.ExpiresAt)
	dto.GraceUntil = formatNullTimePtr(row.GraceUntil)
	dto.RegeneratedAt = formatNullTimePtr(row.RegeneratedAt)
	dto.RevokedAt = formatNullTimePtr(row.RevokedAt)
	return dto
}

func featuresFromRow(row infrastructure.SelfHostLicenseRow) []string {
	if !row.Features.Valid || strings.TrimSpace(row.Features.String) == "" {
		return append([]string(nil), defaultSelfHostFeatures...)
	}
	var features []string
	if err := json.Unmarshal([]byte(row.Features.String), &features); err != nil {
		return append([]string(nil), defaultSelfHostFeatures...)
	}
	return normalizeSelfHostFeatures(features)
}

func nullTimePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	utc := value.Time.UTC()
	return &utc
}

func formatNullTime(value sql.NullTime) string {
	if !value.Valid {
		return ""
	}
	return value.Time.UTC().Format(time.RFC3339)
}

func formatNullTimePtr(value sql.NullTime) *string {
	if !value.Valid {
		return nil
	}
	formatted := value.Time.UTC().Format(time.RFC3339)
	return &formatted
}
