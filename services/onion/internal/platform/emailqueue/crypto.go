package emailqueue

import platformEmail "github.com/voxire/onion/internal/platform/email"

// Keep the queue package's small internal API stable while sharing the
// encryption implementation with the legacy email enqueue path. Both paths
// write to email_jobs and must apply the same at-rest seal.
const encryptedPrefix = "enc:v1:"

func sealHTML(html string) string { return platformEmail.SealHTML(html) }

func openHTML(stored string) string { return platformEmail.OpenHTML(stored) }

// hexDecode preserves the queue package's historical test/helper surface
// while delegating the actual parsing to the shared email implementation.
func hexDecode(s string) ([]byte, error) { return platformEmail.DecodeHex(s) }
