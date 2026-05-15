package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// newMessageID mints a short, unique ID for a single stream message. We
// avoid UUIDs because the chat protocol IDs only need to be stable for the
// life of a single SSE response.
func newMessageID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		// Random source failure is essentially impossible on the platforms
		// we run on; fall back to time so we still mint *something* unique.
		return "msg-" + time.Now().UTC().Format("20060102150405.000000000")
	}
	return "msg-" + hex.EncodeToString(b[:])
}

// newBlockID is the same scheme for sub-message blocks.
func newBlockID(prefix string) string {
	var b [6]byte
	if _, err := rand.Read(b[:]); err != nil {
		return prefix + "-" + time.Now().UTC().Format("150405.000000000")
	}
	return prefix + "-" + hex.EncodeToString(b[:])
}
