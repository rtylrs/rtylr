package model

import (
	"math"
	"testing"
)

func TestCollateUnicode(t *testing.T) {
	got := byteIDs("Aé中🧑", 64)
	want := []int{66, 196, 170, 229, 185, 174, 241, 160, 168, 146}
	if len(got) != len(want) {
		t.Fatalf("byteIDs len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("byteIDs[%d] = %d, want %d", i, got[i], want[i])
		}
	}
	truncated := byteIDs("Aé中🧑", 8)
	if len(truncated) != 8 || truncated[7] != 160 {
		t.Fatalf("truncation diverged: %v", truncated)
	}
	b, err := collate([]Example{{Context: "Aé中🧑", Options: []string{"ok", "skip"}}}, 32, 16)
	if err != nil {
		t.Fatalf("collate: %v", err)
	}
	if len(b.ctxIDs[0]) != 10 {
		t.Fatalf("context padded to %d ids, want 10", len(b.ctxIDs[0]))
	}
	for i, id := range want {
		if b.ctxIDs[0][i] != id {
			t.Fatalf("ctxIDs[0][%d] = %d, want %d", i, b.ctxIDs[0][i], id)
		}
	}
}

func TestSoftmaxEmpty(_ *testing.T) {
	softmaxInPlace(nil)
	softmaxInPlace([]float32{})
}

func TestSoftmaxAllNegativeInf(t *testing.T) {
	a := []float32{float32(math.Inf(-1)), float32(math.Inf(-1)), float32(math.Inf(-1))}
	softmaxInPlace(a)
	var sum float32
	for _, v := range a {
		if math.IsNaN(float64(v)) {
			t.Fatalf("softmax produced NaN: %v", a)
		}
		sum += v
	}
	if math.Abs(float64(sum-1)) > 1e-6 {
		t.Fatalf("softmax sum = %v, want 1", sum)
	}
}
