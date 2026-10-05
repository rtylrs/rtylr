package emailqueue

import "github.com/voxire/onion/internal/platform/emailcrypto"

// Keep the queue package's small internal API stable while sharing the
// encryption implementation with the legacy email enqueue path. Both paths
// write to email_jobs and must apply the same at-rest seal.
const encryptedPrefix = emailcrypto.EncryptedPrefix

func sealHTML(html string) string { return emailcrypto.SealHTML(html) }

func openHTML(stored string) string { return emailcrypto.OpenHTML(stored) }

func hexDecode(s string) ([]byte, error) { return emailcrypto.DecodeHex(s) }
