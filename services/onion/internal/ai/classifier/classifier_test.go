package classifier

import (
	"context"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

const fixtureDir = "testdata/forms"

type goldenExample struct {
	Context         string    `json:"context"`
	Options         []string  `json:"options"`
	Label           int       `json:"label"`
	Logits          []float64 `json:"logits"`
	ReferenceArgmax int       `json:"reference_argmax"`
}

type goldenFile struct {
	Examples []goldenExample `json:"examples"`
}

type tb interface {
	Helper()
	Fatalf(format string, args ...any)
}

func loadGolden(t tb) (*Scorer, []goldenExample) {
	t.Helper()
	s, err := Load(fixtureDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(fixtureDir, "golden.json"))
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	var g goldenFile
	if err := json.Unmarshal(raw, &g); err != nil {
		t.Fatalf("parse golden: %v", err)
	}
	return s, g.Examples
}

func argmax(logits []float32) int {
	best := 0
	for i, v := range logits {
		if v > logits[best] {
			best = i
		}
	}
	return best
}

func TestGoldenParity(t *testing.T) {
	s, examples := loadGolden(t)
	if got := s.Params(); got != 14720 {
		t.Fatalf("Params = %d, want 14720 (reference parameter_count)", got)
	}
	var maxDiff float64
	for i, ex := range examples {
		logits, err := s.Logits(context.Background(), Example{Context: ex.Context, Options: ex.Options})
		if err != nil {
			t.Fatalf("example %d: %v", i, err)
		}
		if len(logits) != len(ex.Logits) {
			t.Fatalf("example %d: %d logits, want %d", i, len(logits), len(ex.Logits))
		}
		for o := range logits {
			if d := math.Abs(float64(logits[o]) - ex.Logits[o]); d > 1e-4 {
				t.Fatalf("example %d option %d: logit %v, want %v (diff %v)",
					i, o, logits[o], ex.Logits[o], d)
			} else if d > maxDiff {
				maxDiff = d
			}
		}
		if got := argmax(logits); got != ex.ReferenceArgmax {
			t.Fatalf("example %d: argmax %d, want %d", i, got, ex.ReferenceArgmax)
		}
	}
	t.Logf("max |go - torch| logit diff over %d examples: %g", len(examples), maxDiff)
}

func TestBatchParity(t *testing.T) {
	s, examples := loadGolden(t)
	exs := make([]Example, len(examples))
	for i, ex := range examples {
		exs[i] = Example{Context: ex.Context, Options: ex.Options}
	}
	rows, err := s.LogitsBatch(context.Background(), exs)
	if err != nil {
		t.Fatalf("LogitsBatch: %v", err)
	}
	for i, row := range rows {
		if got := argmax(row); got != examples[i].ReferenceArgmax {
			t.Fatalf("example %d: batch argmax %d, want %d", i, got, examples[i].ReferenceArgmax)
		}
	}
}

func TestRejectsEmpty(t *testing.T) {
	s, _ := loadGolden(t)
	if _, err := s.Logits(context.Background(), Example{Context: "", Options: []string{"click", "skip"}}); err == nil {
		t.Fatal("expected an error for empty context (reference yields NaN)")
	}
	if _, err := s.Logits(context.Background(), Example{Context: "x", Options: nil}); err == nil {
		t.Fatal("expected an error for missing options")
	}
	if _, err := Load(filepath.Join(fixtureDir, "model.pt")); err == nil {
		t.Fatal("expected an error for a pickle path")
	}
}

func BenchmarkScore(b *testing.B) {
	s, examples := loadGolden(b)
	exs := make([]Example, len(examples))
	for i, ex := range examples {
		exs[i] = Example{Context: ex.Context, Options: ex.Options}
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := s.LogitsBatch(context.Background(), exs); err != nil {
			b.Fatal(err)
		}
	}
}

func TestConcurrentScoring(t *testing.T) {
	t.Parallel()
	s, examples := loadGolden(t)
	exs := make([]Example, len(examples))
	for i, ex := range examples {
		exs[i] = Example{Context: ex.Context, Options: ex.Options}
	}
	var wg sync.WaitGroup
	for w := 0; w < 8; w++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for i := 0; i < 25; i++ {
				if _, err := s.LogitsBatch(context.Background(), exs); err != nil {
					t.Error(err)
					return
				}
				if _, _, err := s.Score(context.Background(), exs[(worker+i)%len(exs)]); err != nil {
					t.Error(err)
					return
				}
			}
		}(w)
	}
	wg.Wait()
}
