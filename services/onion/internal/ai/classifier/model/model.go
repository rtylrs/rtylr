// Package model implements the ported CUA-S1-FORMS forward pass: byte
// collation, a tinyx transformer encoder pair, and an option attention head.
package model

import (
	"context"
	"fmt"
	"math"

	"github.com/voxire/onion/internal/ai/classifier/checkpoint"
)

var maskedLogit = -float32(math.MaxFloat32)

type normP struct{ w, b []float32 }

type attnP struct {
	inW, inB   []float32
	outW, outB []float32
}

type encLayerP struct {
	n1, n2 normP
	attn   attnP
	w1, b1 []float32
	w2, b2 []float32
}

type headP struct {
	cn, on  normP
	q, k, v []float32
}

// Model holds one wired checkpoint and is safe for concurrent scoring.
type Model struct {
	cfg  checkpoint.Config
	emb  []float32
	pos  []float32
	enc  []encLayerP
	oenc encLayerP
	head headP
}

// New wires decoded weights into a model and fails on missing or unexpected
// tensors.
func New(cfg checkpoint.Config, tensors map[string]checkpoint.Tensor) (*Model, error) {
	if len(tensors) == 0 {
		return nil, fmt.Errorf("s1: checkpoint contains no tensors")
	}
	remaining := make(map[string]checkpoint.Tensor, len(tensors))
	for name, t := range tensors {
		remaining[name] = t
	}
	return assemble(cfg, remaining)
}

func need(t map[string]checkpoint.Tensor, name string, shape ...int) ([]float32, error) {
	got, ok := t[name]
	if !ok {
		return nil, fmt.Errorf("s1: checkpoint is missing %q", name)
	}
	if len(got.Shape) != len(shape) {
		return nil, fmt.Errorf("s1: %q has rank %d, want %d", name, len(got.Shape), len(shape))
	}
	for i, d := range shape {
		if got.Shape[i] != d {
			return nil, fmt.Errorf("s1: %q has shape %v, want %v", name, got.Shape, shape)
		}
	}
	return got.Data, nil
}

func needNorm(t map[string]checkpoint.Tensor, prefix string, w int) (normP, error) {
	var n normP
	var err error
	if n.w, err = need(t, prefix+".weight", w); err != nil {
		return n, err
	}
	if n.b, err = need(t, prefix+".bias", w); err != nil {
		return n, err
	}
	return n, nil
}

func needAttn(t map[string]checkpoint.Tensor, prefix string, w int) (attnP, error) {
	var a attnP
	var err error
	if a.inW, err = need(t, prefix+".in_proj_weight", 3*w, w); err != nil {
		return a, err
	}
	if a.inB, err = need(t, prefix+".in_proj_bias", 3*w); err != nil {
		return a, err
	}
	if a.outW, err = need(t, prefix+".out_proj.weight", w, w); err != nil {
		return a, err
	}
	if a.outB, err = need(t, prefix+".out_proj.bias", w); err != nil {
		return a, err
	}
	return a, nil
}

func needEncLayer(t map[string]checkpoint.Tensor, prefix string, w int) (encLayerP, error) {
	var l encLayerP
	var err error
	if l.n1, err = needNorm(t, prefix+".norm1", w); err != nil {
		return l, err
	}
	if l.n2, err = needNorm(t, prefix+".norm2", w); err != nil {
		return l, err
	}
	if l.attn, err = needAttn(t, prefix+".self_attn", w); err != nil {
		return l, err
	}
	if l.w1, err = need(t, prefix+".linear1.weight", 4*w, w); err != nil {
		return l, err
	}
	if l.b1, err = need(t, prefix+".linear1.bias", 4*w); err != nil {
		return l, err
	}
	if l.w2, err = need(t, prefix+".linear2.weight", w, 4*w); err != nil {
		return l, err
	}
	if l.b2, err = need(t, prefix+".linear2.bias", w); err != nil {
		return l, err
	}
	return l, nil
}

func assemble(cfg checkpoint.Config, t map[string]checkpoint.Tensor) (*Model, error) {
	w := cfg.Width
	m := &Model{cfg: cfg}
	var err error
	if m.emb, err = need(t, "embedding.weight", 257, w); err != nil {
		return nil, err
	}
	posLen := cfg.ContextTokens
	if cfg.OptionTokens > posLen {
		posLen = cfg.OptionTokens
	}
	if m.pos, err = need(t, "position.weight", posLen, w); err != nil {
		return nil, err
	}
	for i := 0; i < cfg.Layers; i++ {
		l, err := needEncLayer(t, fmt.Sprintf("encoder.layers.%d", i), w)
		if err != nil {
			return nil, err
		}
		m.enc = append(m.enc, l)
	}
	if m.oenc, err = needEncLayer(t, "option_encoder.layers.0", w); err != nil {
		return nil, err
	}
	for _, n := range []string{"head.context_norm", "head.option_norm"} {
		np, err := needNorm(t, n, w)
		if err != nil {
			return nil, err
		}
		if n == "head.context_norm" {
			m.head.cn = np
		} else {
			m.head.on = np
		}
	}
	for _, n := range []string{"query", "key", "value"} {
		v, err := need(t, "head."+n+".weight", cfg.Rank, w)
		if err != nil {
			return nil, err
		}
		switch n {
		case "query":
			m.head.q = v
		case "key":
			m.head.k = v
		default:
			m.head.v = v
		}
	}
	delete(t, "embedding.weight")
	delete(t, "position.weight")
	for i := 0; i < cfg.Layers; i++ {
		for _, s := range []string{
			"norm1.weight", "norm1.bias", "norm2.weight", "norm2.bias",
			"self_attn.in_proj_weight", "self_attn.in_proj_bias",
			"self_attn.out_proj.weight", "self_attn.out_proj.bias",
			"linear1.weight", "linear1.bias", "linear2.weight", "linear2.bias",
		} {
			delete(t, fmt.Sprintf("encoder.layers.%d.%s", i, s))
		}
	}
	for _, s := range []string{
		"norm1.weight", "norm1.bias", "norm2.weight", "norm2.bias",
		"self_attn.in_proj_weight", "self_attn.in_proj_bias",
		"self_attn.out_proj.weight", "self_attn.out_proj.bias",
		"linear1.weight", "linear1.bias", "linear2.weight", "linear2.bias",
	} {
		delete(t, "option_encoder.layers.0."+s)
	}
	for _, s := range []string{
		"context_norm.weight", "context_norm.bias",
		"option_norm.weight", "option_norm.bias",
		"query.weight", "key.weight", "value.weight",
	} {
		delete(t, "head."+s)
	}
	for name := range t {
		return nil, fmt.Errorf("s1: unexpected checkpoint tensor %q", name)
	}
	return m, nil
}

func (m *Model) embed(ids []int) []float32 {
	w := m.cfg.Width
	out := make([]float32, len(ids)*w)
	for t, id := range ids {
		e := m.emb[id*w : (id+1)*w]
		p := m.pos[t*w : (t+1)*w]
		o := out[t*w : (t+1)*w]
		for d := 0; d < w; d++ {
			o[d] = e[d] + p[d]
		}
	}
	return out
}

func (m *Model) encLayer(ctx context.Context, x []float32, valid []bool, l int, p *encLayerP) ([]float32, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	w, h := m.cfg.Width, m.cfg.Heads
	dh := w / h
	scale := float32(1 / math.Sqrt(float64(dh)))

	normed := make([]float32, len(x))
	for t := 0; t < l; t++ {
		layernorm(normed[t*w:(t+1)*w], x[t*w:(t+1)*w], p.n1.w, p.n1.b, 1e-5)
	}
	qkv := make([]float32, l*3*w)
	for t := 0; t < l; t++ {
		linear(qkv[t*3*w:(t+1)*3*w], p.attn.inW, p.attn.inB, normed[t*w:(t+1)*w], 3*w, w)
	}
	attnOut := make([]float32, l*w)
	scores := make([]float32, l)
	for head := 0; head < h; head++ {
		for i := 0; i < l; i++ {
			for j := 0; j < l; j++ {
				if !valid[j] {
					scores[j] = float32(math.Inf(-1))
					continue
				}
				var acc float32
				qo := (i * 3 * w) + head*dh
				ko := (j * 3 * w) + w + head*dh
				for d := 0; d < dh; d++ {
					acc += qkv[qo+d] * qkv[ko+d]
				}
				scores[j] = acc * scale
			}
			softmaxInPlace(scores)
			for d := 0; d < dh; d++ {
				var acc float32
				for j := 0; j < l; j++ {
					acc += scores[j] * qkv[(j*3*w)+2*w+head*dh+d]
				}
				attnOut[i*w+head*dh+d] = acc
			}
		}
	}
	y := make([]float32, len(x))
	copy(y, x)
	proj := make([]float32, w)
	for t := 0; t < l; t++ {
		linear(proj, p.attn.outW, p.attn.outB, attnOut[t*w:(t+1)*w], w, w)
		addInto(y[t*w:(t+1)*w], proj)
	}
	ff := make([]float32, 4*w)
	back := make([]float32, w)
	for t := 0; t < l; t++ {
		layernorm(normed[t*w:(t+1)*w], y[t*w:(t+1)*w], p.n2.w, p.n2.b, 1e-5)
		linear(ff, p.w1, p.b1, normed[t*w:(t+1)*w], 4*w, w)
		reluInPlace(ff)
		linear(back, p.w2, p.b2, ff, w, 4*w)
		addInto(y[t*w:(t+1)*w], back)
	}
	return y, nil
}

func safeMask(valid []bool) []bool {
	out := make([]bool, len(valid))
	copy(out, valid)
	if len(out) > 0 {
		out[0] = true
	}
	return out
}

func (m *Model) encodeContext(ctx context.Context, ids []int, valid []bool) ([]float32, error) {
	x := m.embed(ids)
	safe := safeMask(valid)
	for i := range m.enc {
		var err error
		x, err = m.encLayer(ctx, x, safe, len(ids), &m.enc[i])
		if err != nil {
			return nil, err
		}
	}
	return x, nil
}

func (m *Model) encodeOptions(ctx context.Context, ids [][]int, tokMask [][]bool) ([]float32, error) {
	w := m.cfg.Width
	n := len(ids)
	out := make([]float32, n*w)
	for o := 0; o < n; o++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		x := m.embed(ids[o])
		var err error
		x, err = m.encLayer(ctx, x, safeMask(tokMask[o]), len(ids[o]), &m.oenc)
		if err != nil {
			return nil, err
		}
		var count float32
		for t := range ids[o] {
			if tokMask[o][t] {
				count++
				addInto(out[o*w:(o+1)*w], x[t*w:(t+1)*w])
			}
		}
		if count < 1 {
			count = 1
		}
		scale := 1 / count
		for d := 0; d < w; d++ {
			out[o*w+d] *= scale
		}
	}
	return out, nil
}

func (m *Model) scoreHead(cctx context.Context, ctx []float32, ctxValid []bool, opts []float32, optValid []bool) ([]float32, error) {
	if err := cctx.Err(); err != nil {
		return nil, err
	}
	w, r := m.cfg.Width, m.cfg.Rank
	lc := len(ctxValid)
	no := len(optValid)
	cn := make([]float32, len(ctx))
	for t := 0; t < lc; t++ {
		layernorm(cn[t*w:(t+1)*w], ctx[t*w:(t+1)*w], m.head.cn.w, m.head.cn.b, 1e-5)
	}
	on := make([]float32, len(opts))
	for o := 0; o < no; o++ {
		layernorm(on[o*w:(o+1)*w], opts[o*w:(o+1)*w], m.head.on.w, m.head.on.b, 1e-5)
	}
	scale := float32(1 / math.Sqrt(float64(r)))
	q := make([]float32, no*r)
	k := make([]float32, lc*r)
	v := make([]float32, lc*r)
	for o := 0; o < no; o++ {
		linear(q[o*r:(o+1)*r], m.head.q, nil, on[o*w:(o+1)*w], r, w)
	}
	for t := 0; t < lc; t++ {
		linear(k[t*r:(t+1)*r], m.head.k, nil, cn[t*w:(t+1)*w], r, w)
		linear(v[t*r:(t+1)*r], m.head.v, nil, cn[t*w:(t+1)*w], r, w)
	}
	att := make([]float32, r)
	p := make([]float32, lc)
	logits := make([]float32, no)
	for o := 0; o < no; o++ {
		if err := cctx.Err(); err != nil {
			return nil, err
		}
		if !optValid[o] {
			logits[o] = maskedLogit
			continue
		}
		for t := 0; t < lc; t++ {
			if !ctxValid[t] {
				p[t] = maskedLogit
				continue
			}
			p[t] = dot(q[o*r:(o+1)*r], k[t*r:(t+1)*r]) * scale
		}
		softmaxInPlace(p)
		for d := 0; d < r; d++ {
			att[d] = 0
		}
		for t := 0; t < lc; t++ {
			for d := 0; d < r; d++ {
				att[d] += p[t] * v[t*r+d]
			}
		}
		logits[o] = dot(q[o*r:(o+1)*r], att) * scale
	}
	return logits, nil
}

func (m *Model) forward(ctx context.Context, b batch) ([][]float32, error) {
	logits := make([][]float32, len(b.ctxIDs))
	for i := range b.ctxIDs {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		contextVec, err := m.encodeContext(ctx, b.ctxIDs[i], b.ctxMask[i])
		if err != nil {
			return nil, err
		}
		opts, err := m.encodeOptions(ctx, b.optIDs[i], b.optTok[i])
		if err != nil {
			return nil, err
		}
		logits[i], err = m.scoreHead(ctx, contextVec, b.ctxMask[i], opts, b.optMask[i])
		if err != nil {
			return nil, err
		}
	}
	return logits, nil
}

// Logits scores one example and returns raw per-option logits.
func (m *Model) Logits(ctx context.Context, ex Example) ([]float32, error) {
	rows, err := m.LogitsBatch(ctx, []Example{ex})
	if err != nil {
		return nil, err
	}
	return rows[0], nil
}

// LogitsBatch scores one logit row per example in a single pass.
func (m *Model) LogitsBatch(ctx context.Context, exs []Example) ([][]float32, error) {
	if len(exs) == 0 {
		return nil, fmt.Errorf("s1: cannot score an empty batch")
	}
	b, err := collate(exs, m.cfg.ContextTokens, m.cfg.OptionTokens)
	if err != nil {
		return nil, err
	}
	return m.forward(ctx, b)
}

// Score returns normalized probabilities over the live options and the
// winning index.
func (m *Model) Score(ctx context.Context, ex Example) (probs []float64, best int, err error) {
	logits, err := m.Logits(ctx, ex)
	if err != nil {
		return nil, 0, err
	}
	tmp := make([]float32, len(logits))
	copy(tmp, logits)
	softmaxInPlace(tmp)
	probs = make([]float64, len(tmp))
	best = 0
	for i, p := range tmp {
		probs[i] = float64(p)
		if p > tmp[best] {
			best = i
		}
	}
	return probs, best, nil
}
