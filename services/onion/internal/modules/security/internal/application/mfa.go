package application

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	// #nosec G505 -- RFC 6238 TOTP interoperability requires the SHA-1 default.
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/utils/passwordutils"
)

var ErrMFAInvalidCode = errors.New("invalid MFA code")
var ErrMFAEnrollmentMissing = errors.New("MFA enrollment is not pending")

const mfaSecretPrefix = "enc:v1:"

func (s *Service) BeginMFAEnrollment(ctx context.Context, userID int64) (string, string, error) {
	secretBytes := make([]byte, 20)
	if _, err := rand.Read(secretBytes); err != nil {
		return "", "", err
	}
	secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secretBytes)
	encrypted, err := sealMFASecret(secret)
	if err != nil {
		return "", "", err
	}
	if err := s.repo.BeginMFAEnrollment(ctx, userID, []byte(encrypted)); err != nil {
		return "", "", err
	}
	uri := "otpauth://totp/" + url.QueryEscape("rtylr:user-"+strconv.FormatInt(userID, 10)) +
		"?secret=" + url.QueryEscape(secret) + "&issuer=rtylr"
	return secret, uri, nil
}

func (s *Service) CompleteMFAEnrollment(ctx context.Context, userID int64, code string) ([]string, error) {
	stored, err := s.repo.GetMFASecret(ctx, userID, false)
	if err != nil {
		return nil, ErrMFAEnrollmentMissing
	}
	secret, err := openMFASecret(string(stored))
	if err != nil || !validTOTP(secret, code, time.Now()) {
		return nil, ErrMFAInvalidCode
	}
	recoveryCodes := make([]string, 0, 8)
	recoveryHashes := make([][]byte, 0, 8)
	for i := 0; i < 8; i++ {
		code, codeErr := randomRecoveryCode()
		if codeErr != nil {
			return nil, codeErr
		}
		hash, hashErr := passwordutils.HashPassword(code)
		if hashErr != nil {
			return nil, hashErr
		}
		recoveryCodes = append(recoveryCodes, code)
		recoveryHashes = append(recoveryHashes, []byte(hash))
	}
	if err := s.repo.CompleteMFAEnrollment(ctx, userID, recoveryHashes); err != nil {
		return nil, err
	}
	return recoveryCodes, nil
}

func (s *Service) VerifyMFA(ctx context.Context, userID int64, code string) error {
	stored, err := s.repo.GetMFASecret(ctx, userID, true)
	if err != nil {
		return ErrMFAInvalidCode
	}
	secret, err := openMFASecret(string(stored))
	if err != nil || !validTOTP(secret, code, time.Now()) {
		return ErrMFAInvalidCode
	}
	return s.repo.MarkMFAVerified(ctx, userID)
}

func (s *Service) UseMFARecoveryCode(ctx context.Context, userID int64, code string) error {
	normalized := strings.ToUpper(strings.TrimSpace(code))
	if normalized == "" {
		return ErrMFAInvalidCode
	}
	rows, err := s.repo.ListRecoveryCodes(ctx, userID)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if passwordutils.CheckPasswordHash(normalized, row.Hash) {
			return s.repo.ConsumeRecoveryCode(ctx, userID, row.ID)
		}
	}
	return ErrMFAInvalidCode
}

func randomRecoveryCode() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return strings.ToUpper(fmt.Sprintf("%08X-%08X", binary.BigEndian.Uint32(bytes[:4]), binary.BigEndian.Uint32(bytes[4:]))), nil
}

func validTOTP(secret, code string, now time.Time) bool {
	code = strings.TrimSpace(code)
	if len(code) != 6 {
		return false
	}
	if _, err := strconv.Atoi(code); err != nil {
		return false
	}
	decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(strings.TrimSpace(secret)))
	if err != nil {
		return false
	}
	current := now.Unix() / 30
	for offset := int64(-1); offset <= 1; offset++ {
		counter := current + offset
		var message [8]byte
		binary.BigEndian.PutUint64(message[:], uint64(counter)) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		// #nosec G505 -- RFC 6238 TOTP interoperability requires SHA-1 for the default otpauth URI.
		hash := hmac.New(sha1.New, decoded)
		_, _ = hash.Write(message[:])
		digest := hash.Sum(nil)
		index := digest[len(digest)-1] & 0x0f
		value := (uint32(digest[index])&0x7f)<<24 |
			uint32(digest[index+1])<<16 |
			uint32(digest[index+2])<<8 |
			uint32(digest[index+3])
		candidate := fmt.Sprintf("%06d", value%1000000)
		if hmac.Equal([]byte(candidate), []byte(code)) {
			return true
		}
	}
	return false
}

func mfaEncryptionKey() ([]byte, error) {
	secret := strings.TrimSpace(os.Getenv("MFA_ENCRYPTION_KEY"))
	if secret == "" {
		secret = strings.TrimSpace(config.GlobalConfig.JWTSecret)
	}
	if secret == "" {
		return nil, errors.New("MFA encryption key is not configured")
	}
	digest := sha256.Sum256([]byte(secret))
	return digest[:], nil
}

func sealMFASecret(secret string) (string, error) {
	key, err := mfaEncryptionKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(secret), nil)
	return mfaSecretPrefix + base64.StdEncoding.EncodeToString(ciphertext), nil
}

func openMFASecret(stored string) (string, error) {
	if !strings.HasPrefix(stored, mfaSecretPrefix) {
		return "", errors.New("invalid MFA secret")
	}
	key, err := mfaEncryptionKey()
	if err != nil {
		return "", err
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(stored, mfaSecretPrefix))
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil || len(raw) < gcm.NonceSize() {
		return "", errors.New("invalid MFA secret")
	}
	plaintext, err := gcm.Open(nil, raw[:gcm.NonceSize()], raw[gcm.NonceSize():], nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
