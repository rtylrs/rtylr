package cdn

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

// s3.PutObject seeks back to the start of the body to compute the payload hash
// whenever the endpoint is not TLS. sniffMIME used to wrap its input in an
// io.MultiReader, which has no Seek, so the handler's multipart.File was
// silently downgraded and every upload against a plaintext object store failed
// with 500: the shape onion ships for self-host and documents for local
// development (#1077).
func TestSniffMIMEKeepsASeekableBodySeekable(t *testing.T) {
	png := "\x89PNG\r\n\x1a\n" + strings.Repeat("\x00", 64)

	_, body, err := sniffMIME(strings.NewReader(png))
	if err != nil {
		t.Fatalf("sniffMIME: %v", err)
	}

	seeker, ok := body.(io.ReadSeeker)
	if !ok {
		t.Fatalf("sniffMIME returned %T for a seekable input; PutObject cannot sign it without TLS", body)
	}

	// Seekable is only useful if the stream really does start at byte zero.
	got, err := io.ReadAll(seeker)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(got) != png {
		t.Fatalf("replayed %d bytes, want the original %d", len(got), len(png))
	}

	// And it must be repeatable, which is what signing then sending needs.
	if _, err := seeker.Seek(0, io.SeekStart); err != nil {
		t.Fatalf("second seek: %v", err)
	}
	again, err := io.ReadAll(seeker)
	if err != nil {
		t.Fatalf("second read: %v", err)
	}
	if string(again) != png {
		t.Fatalf("second replay differs: %d bytes, want %d", len(again), len(png))
	}
}

// unseekableReader is a plain reader with no Seek, standing in for a body that
// genuinely cannot be rewound.
type unseekableReader struct{ r io.Reader }

func (u unseekableReader) Read(p []byte) (int, error) { return u.r.Read(p) }

// A body that cannot seek must still be replayed in full. That is the case the
// io.MultiReader fallback exists for, and dropping it would lose the sniffed
// prefix.
func TestSniffMIMEStillReplaysAnUnseekableBody(t *testing.T) {
	payload := "\x89PNG\r\n\x1a\n" + strings.Repeat("A", 700)

	mime, body, err := sniffMIME(unseekableReader{r: strings.NewReader(payload)})
	if err != nil {
		t.Fatalf("sniffMIME: %v", err)
	}
	if mime != "image/png" {
		t.Errorf("detected %q, want image/png", mime)
	}
	got, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(got) != payload {
		t.Fatalf("replayed %d bytes, want %d; the sniffed prefix was lost", len(got), len(payload))
	}
}

// failingSeeker reports itself as seekable and then refuses. sniffMIME must not
// hand back a reader positioned past the sniffed prefix.
type failingSeeker struct{ r io.Reader }

func (f *failingSeeker) Read(p []byte) (int, error) { return f.r.Read(p) }
func (f *failingSeeker) Seek(int64, int) (int64, error) {
	return 0, errors.New("seek refused")
}

func TestSniffMIMEFallsBackWhenSeekFails(t *testing.T) {
	payload := "\x89PNG\r\n\x1a\n" + strings.Repeat("B", 700)

	_, body, err := sniffMIME(&failingSeeker{r: bytes.NewReader([]byte(payload))})
	if err != nil {
		t.Fatalf("sniffMIME: %v", err)
	}
	got, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(got) != payload {
		t.Fatalf("replayed %d bytes, want %d; a refused seek lost the prefix", len(got), len(payload))
	}
}
