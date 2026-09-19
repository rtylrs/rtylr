package classifier

import (
	"context"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestRealCheckpointParity(t *testing.T) {
	dir := os.Getenv("REAL_CHECKPOINT")
	if dir == "" {
		dir = "models/cua-s1-forms"
	}
	s, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "golden.json"))
	if err != nil {
		t.Fatalf("read golden (run testdata/gen_fixtures.py ignored --checkpoint %s): %v", dir, err)
	}
	var g goldenFile
	if err := json.Unmarshal(raw, &g); err != nil {
		t.Fatalf("parse golden: %v", err)
	}
	var maxDiff float64
	for i, ex := range g.Examples {
		logits, err := s.Logits(context.Background(), Example{Context: ex.Context, Options: ex.Options})
		if err != nil {
			t.Fatalf("example %d: %v", i, err)
		}
		for o := range logits {
			if d := math.Abs(float64(logits[o]) - ex.Logits[o]); d > 1e-4 {
				t.Fatalf("example %d option %d: logit %v, want %v (diff %v)", i, o, logits[o], ex.Logits[o], d)
			} else if d > maxDiff {
				maxDiff = d
			}
		}
		if got := argmax(logits); got != ex.ReferenceArgmax {
			t.Fatalf("example %d: argmax %d, want %d", i, got, ex.ReferenceArgmax)
		}
	}
	t.Logf("params=%d weights=%d bytes", s.Params(), s.WeightsBytes())
	t.Logf("max |go - torch| logit diff over %d real-weight examples: %g", len(g.Examples), maxDiff)
}
