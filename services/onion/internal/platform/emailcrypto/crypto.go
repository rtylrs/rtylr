package emailcrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/voxire/onion/internal/platform/logger"
)

// EncryptedPrefix marks a stored email body as AES-256-GCM ciphertext. Rows
// written before a key was configured carry no prefix and are sent as-is, so
// rolling out the key never breaks queued mail.
const (
	EncryptedPrefix      = "enc:v1:"
	invalidEncryptedBody = EncryptedPrefix + "invalid-key"
	// Keep the comma-separated old keys here until their queued jobs have drained.
	previousEncryptionKeys = "EMAIL_ENCRYPTION_PREVIOUS_KEYS"
)

// The key is read from the environment on every use so rotating or adding it
// takes effect without a restart; the "unset" warning fires only once.
var (
	warnPlaintextOnce  sync.Once
	warnInvalidKeyOnce sync.Once
)

// encryptionKeys returns the current key followed by explicitly retained
// fallback keys. The current key is always the first key used for new bodies;
// older keys are only tried while opening existing rows during a rotation.
func encryptionKeys() (keys [][]byte, configured bool, invalidCurrent bool) {
	raw := strings.TrimSpace(os.Getenv("EMAIL_ENCRYPTION_KEY"))
	previous := os.Getenv(previousEncryptionKeys)
	if raw == "" && strings.TrimSpace(previous) == "" {
		warnPlaintextOnce.Do(func() {
			logger.Global().Warn("EMAIL_ENCRYPTION_KEY not set; email_jobs bodies stored in PLAINTEXT — reset/invite tokens readable from DB dumps")
		})
		return nil, false, false
	}

	keys = make([][]byte, 0, 2)
	seen := make(map[string]struct{})
	if raw == "" {
		invalidCurrent = true
		warnInvalidKeyOnce.Do(func() {
			logger.Global().Error("EMAIL_ENCRYPTION_KEY is required when EMAIL_ENCRYPTION_PREVIOUS_KEYS is set; refusing to store sensitive mail in plaintext")
		})
	} else if current, err := parseEncryptionKey(raw); err != nil {
		invalidCurrent = true
		warnInvalidKeyOnce.Do(func() {
			logger.Global().Error("EMAIL_ENCRYPTION_KEY invalid (want 32 bytes hex or base64); refusing to store sensitive mail in plaintext")
		})
	} else {
		keys = append(keys, current)
		seen[string(current)] = struct{}{}
	}

	for _, part := range strings.Split(previous, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		key, keyErr := parseEncryptionKey(part)
		if keyErr != nil {
			logger.Global().Error("EMAIL_ENCRYPTION_PREVIOUS_KEYS contains an invalid key; ignoring it")
			continue
		}
		if _, duplicate := seen[string(key)]; duplicate {
			continue
		}
		seen[string(key)] = struct{}{}
		keys = append(keys, key)
	}
	return keys, true, invalidCurrent
}

func parseEncryptionKey(raw string) ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(raw)
	if err == nil && len(key) == 32 {
		return key, nil
	}
	key, hexErr := DecodeHex(raw)
	if hexErr == nil && len(key) == 32 {
		return key, nil
	}
	return nil, errors.New("invalid encryption key")
}

// DecodeHex decodes the hexadecimal key representation accepted by
// EMAIL_ENCRYPTION_KEY. It is exported for the queue package's compatibility
// tests and keeps key parsing in one place.
func DecodeHex(s string) ([]byte, error) {
	if len(s)%2 != 0 {
		return nil, errors.New("odd-length hex")
	}
	out := make([]byte, len(s)/2)
	for i := 0; i < len(out); i++ {
		hi, ok1 := hexVal(s[i*2])
		lo, ok2 := hexVal(s[i*2+1])
		if !ok1 || !ok2 {
			return nil, errors.New("bad hex")
		}
		out[i] = hi<<4 | lo
	}
	return out, nil
}

func hexVal(c byte) (byte, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, true
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, true
	}
	return 0, false
}

// SealHTML encrypts an email body at rest when a key is configured. It is
// idempotent, so callers may safely pass a body that was already sealed.
func SealHTML(html string) string {
	if html == "" || strings.HasPrefix(html, EncryptedPrefix) {
		return html
	}

	keys, configured, invalid := encryptionKeys()
	if !configured {
		return html
	}
	if invalid || len(keys) == 0 {
		return invalidEncryptedBody
	}

	block, err := aes.NewCipher(keys[0])
	if err != nil {
		logger.Global().Error("email encryption unavailable; quarantining body", logger.Error(err))
		return invalidEncryptedBody
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		logger.Global().Error("email encryption unavailable; quarantining body", logger.Error(err))
		return invalidEncryptedBody
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		logger.Global().Error("email nonce generation failed; quarantining body", logger.Error(err))
		return invalidEncryptedBody
	}

	sealed := gcm.Seal(nonce, nonce, []byte(html), nil)
	return EncryptedPrefix + base64.StdEncoding.EncodeToString(sealed)
}

// OpenHTML reverses SealHTML. Unprefixed rows are legacy plaintext and pass
// through unchanged; undecryptable rows return empty rather than sending
// garbage to a recipient.
func OpenHTML(stored string) string {
	if !strings.HasPrefix(stored, EncryptedPrefix) {
		return stored
	}
	if stored == invalidEncryptedBody {
		logger.Global().Error("email row was quarantined because EMAIL_ENCRYPTION_KEY was invalid; skipping body")
		return ""
	}

	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(stored, EncryptedPrefix))
	if err != nil {
		logger.Global().Error("email row decryption failed (bad encoding); skipping body", logger.Error(err))
		return ""
	}

	keys, configured, _ := encryptionKeys()
	if !configured || len(keys) == 0 {
		logger.Global().Error("encrypted email row found but no usable EMAIL_ENCRYPTION_KEY is configured; skipping body")
		return ""
	}
	for _, key := range keys {
		if plaintext, ok := openHTMLWithKey(raw, key); ok {
			return plaintext
		}
	}

	logger.Global().Error("email row decryption failed with every configured key; skipping body")
	return ""
}

func openHTMLWithKey(raw, key []byte) (string, bool) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", false
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", false
	}
	if len(raw) < gcm.NonceSize() {
		return "", false
	}

	nonce, ciphertext := raw[:gcm.NonceSize()], raw[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", false
	}
	return string(plaintext), true
}
