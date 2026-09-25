package securekey

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

const DefaultRandomBytes = 32

type Secret struct {
	Plaintext     string
	Hash          string
	DisplayPrefix string
}

type Options struct {
	Prefix           string
	RandomBytes      int
	DisplayPrefixLen int
}

func Generate(opts Options) (Secret, error) {
	randomBytes := opts.RandomBytes
	if randomBytes <= 0 {
		randomBytes = DefaultRandomBytes
	}

	raw := make([]byte, randomBytes)
	if _, err := rand.Read(raw); err != nil {
		return Secret{}, fmt.Errorf("generate key: %w", err)
	}

	plaintext := opts.Prefix + hex.EncodeToString(raw)
	hash := Hash(plaintext)

	displayLen := opts.DisplayPrefixLen
	if displayLen <= 0 || displayLen > len(plaintext) {
		displayLen = min(len(plaintext), 12)
	}

	return Secret{
		Plaintext:     plaintext,
		Hash:          hash,
		DisplayPrefix: plaintext[:displayLen],
	}, nil
}

func Hash(secret string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(secret)))
	return hex.EncodeToString(sum[:])
}
