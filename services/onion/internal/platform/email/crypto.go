package email

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

// encryptedPrefix marks a stored email body as AES-256-GCM ciphertext. Rows
// written before a key was configured carry no prefix and are sent as-is, so
// rolling out the key never breaks queued mail.
const encryptedPrefix = "enc:v1:"

// The key is read from the environment on every use so rotating or adding it
// takes effect without a restart; the "unset" warning fires only once.
var warnPlaintextOnce sync.Once

func encryptionKey() []byte {
	raw := strings.TrimSpace(os.Getenv("EMAIL_ENCRYPTION_KEY"))
	if raw == "" {
		warnPlaintextOnce.Do(func() {
			logger.Global().Warn("EMAIL_ENCRYPTION_KEY not set; email_jobs bodies stored in PLAINTEXT — reset/invite tokens readable from DB dumps")
		})
		return nil
	}
	key, err := base64.StdEncoding.DecodeString(raw)
	if err != nil || len(key) != 32 {
		keyB, hexErr := DecodeHex(raw)
		if hexErr != nil || len(keyB) != 32 {
			logger.Global().Error("EMAIL_ENCRYPTION_KEY invalid (want 32 bytes hex or base64); storing plaintext")
			return nil
		}
		key = keyB
	}
	return key
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
	key := encryptionKey()
	if key == nil || html == "" || strings.HasPrefix(html, encryptedPrefix) {
		return html
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		logger.Global().Error("email encryption unavailable; storing plaintext", logger.Error(err))
		return html
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		logger.Global().Error("email encryption unavailable; storing plaintext", logger.Error(err))
		return html
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		logger.Global().Error("email nonce generation failed; storing plaintext", logger.Error(err))
		return html
	}

	sealed := gcm.Seal(nonce, nonce, []byte(html), nil)
	return encryptedPrefix + base64.StdEncoding.EncodeToString(sealed)
}

// OpenHTML reverses SealHTML. Unprefixed rows are legacy plaintext and pass
// through unchanged; undecryptable rows return empty rather than sending
// garbage to a recipient.
func OpenHTML(stored string) string {
	if !strings.HasPrefix(stored, encryptedPrefix) {
		return stored
	}

	key := encryptionKey()
	if key == nil {
		logger.Global().Error("encrypted email row found but EMAIL_ENCRYPTION_KEY is not configured; skipping body")
		return ""
	}

	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(stored, encryptedPrefix))
	if err != nil {
		logger.Global().Error("email row decryption failed (bad encoding); skipping body", logger.Error(err))
		return ""
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		logger.Global().Error("email decryption unavailable; skipping body", logger.Error(err))
		return ""
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		logger.Global().Error("email decryption unavailable; skipping body", logger.Error(err))
		return ""
	}
	if len(raw) < gcm.NonceSize() {
		logger.Global().Error("email row too short to decrypt; skipping body")
		return ""
	}

	nonce, ciphertext := raw[:gcm.NonceSize()], raw[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		logger.Global().Error("email row decryption failed; skipping body", logger.Error(err))
		return ""
	}
	return string(plaintext)
}
