package classifier

import (
	"context"

	"github.com/voxire/onion/internal/ai/classifier/checkpoint"
	"github.com/voxire/onion/internal/ai/classifier/model"
)

// Config is the validated checkpoint configuration.
type Config = checkpoint.Config

// Example is one scoring decision: a form element and its candidate options.
type Example = model.Example

// Scorer is a loaded checkpoint ready to score form elements.
type Scorer struct {
	file *checkpoint.File
	m    *model.Model
}

// CheckpointErrorKind classifies a checkpoint load failure without exposing
// filesystem paths.
func CheckpointErrorKind(err error) string { return checkpoint.Kind(err) }

// Load reads a cua-s1 checkpoint from a directory, .safetensors, or .json
// path, verifies it, and wires the weights into memory.
func Load(path string) (*Scorer, error) {
	f, err := checkpoint.Load(path)
	if err != nil {
		return nil, err
	}
	m, err := model.New(f.Config, f.Tensors)
	if err != nil {
		return nil, err
	}
	return &Scorer{file: f, m: m}, nil
}

// Config returns the validated checkpoint configuration.
func (s *Scorer) Config() Config { return s.file.Config }

// Params returns the trainable parameter count.
func (s *Scorer) Params() int { return s.file.Params }

// WeightsBytes returns the on-disk size of model.safetensors.
func (s *Scorer) WeightsBytes() int64 { return s.file.WeightsBytes }

// Logits scores one example and returns raw per-option logits.
func (s *Scorer) Logits(ctx context.Context, ex Example) ([]float32, error) {
	return s.m.Logits(ctx, ex)
}

// LogitsBatch scores a whole form in one pass.
func (s *Scorer) LogitsBatch(ctx context.Context, exs []Example) ([][]float32, error) {
	return s.m.LogitsBatch(ctx, exs)
}

// Score returns normalized probabilities over the live options and the
// winning index.
func (s *Scorer) Score(ctx context.Context, ex Example) (probs []float64, best int, err error) {
	return s.m.Score(ctx, ex)
}
