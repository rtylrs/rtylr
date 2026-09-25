package cdn

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// sniffMIME reads up to the first 512 bytes of r to detect the content type by
// magic bytes, then returns the detected type alongside a reader that replays
// the entire original stream. The detected type never includes a charset
// parameter.
//
// A seekable body is rewound rather than wrapped. Wrapping it in an
// io.MultiReader silently downgraded an io.ReadSeeker to a plain io.Reader, and
// the handler's body is a multipart.File, which is seekable. s3.PutObject needs
// to seek back to the start to compute the payload hash whenever the endpoint
// is not TLS, so every upload failed with 500 against a plaintext object store:
// the shape onion ships for self-host and documents for local development.
// Production was unaffected because the SDK signs a streaming body as
// UNSIGNED-PAYLOAD over TLS, which is why this stayed invisible.
func sniffMIME(r io.Reader) (string, io.Reader, error) {
	head := make([]byte, 512)
	n, err := io.ReadFull(r, head)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return "", nil, err
	}
	head = head[:n]

	detected := http.DetectContentType(head)
	if i := strings.IndexByte(detected, ';'); i >= 0 {
		detected = strings.TrimSpace(detected[:i])
	}

	detected = strings.ToLower(detected)

	// Rewinding preserves seekability and buffers nothing. The MultiReader
	// below stays as the fallback for a body that cannot seek, which reads
	// correctly and is only unusable for signing.
	if seeker, ok := r.(io.ReadSeeker); ok {
		if _, seekErr := seeker.Seek(0, io.SeekStart); seekErr == nil {
			return detected, seeker, nil
		}
	}

	return detected, io.MultiReader(bytes.NewReader(head), r), nil
}

// reconcileMIME decides the authoritative content type for an upload from the
// client-declared type and the type sniffed from the bytes. The sniffed type
// wins when it is itself an allowed media type — this blocks an HTML/JS/SVG
// payload smuggled in under an image/* Content-Type. Some allowed formats
// (avif, tiff, quicktime) are not detectable by http.DetectContentType, which
// reports application/octet-stream for them; only then do we fall back to the
// declared type. Anything else is rejected.
func reconcileMIME(declared, sniffed string) (string, error) {
	declared = strings.ToLower(strings.TrimSpace(declared))
	sniffed = strings.ToLower(strings.TrimSpace(sniffed))

	if sniffed == "text/plain" && declared == "text/csv" {
		return declared, nil
	}
	if isAllowedMIMEType(sniffed) {
		return sniffed, nil
	}
	if sniffed == "application/octet-stream" && isAllowedMIMEType(declared) {
		return declared, nil
	}
	if sniffed == "application/zip" && isAllowedMIMEType(declared) && strings.Contains(declared, "officedocument") {
		return declared, nil
	}
	return "", fmt.Errorf("content type %q is not an allowed media type", sniffed)
}
