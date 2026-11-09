//! CUA-S1-FORMS forward pass: byte collation, a tinyx transformer
//! encoder pair, and an option attention head. Mirrors
//! `internal/ai/classifier/model`.

use crate::checkpoint::{Config, Tensor};
use std::collections::BTreeMap;
use thiserror::Error;

#[derive(Debug, Error, PartialEq)]
pub enum ModelError {
    #[error("{0}")]
    Invalid(String),
}

fn invalid(message: impl Into<String>) -> ModelError {
    ModelError::Invalid(message.into())
}

/// One scoring decision: a form element's context plus the candidate
/// values extracted from the source document.
#[derive(Debug, Clone, PartialEq)]
pub struct Example {
    pub context: String,
    pub options: Vec<String>,
}

/// Go `maskedLogit = -float32(math.MaxFloat32)`: a finite negative,
/// deliberately not `-inf`, so logits on padded options keep raw-logit
/// parity with the Go reference.
const MASKED_LOGIT: f32 = -f32::MAX;

#[derive(Debug, Clone, Default)]
struct NormP {
    w: Vec<f32>,
    b: Vec<f32>,
}

#[derive(Debug, Clone, Default)]
struct AttnP {
    in_w: Vec<f32>,
    in_b: Vec<f32>,
    out_w: Vec<f32>,
    out_b: Vec<f32>,
}

#[derive(Debug, Clone, Default)]
struct EncLayerP {
    n1: NormP,
    n2: NormP,
    attn: AttnP,
    w1: Vec<f32>,
    b1: Vec<f32>,
    w2: Vec<f32>,
    b2: Vec<f32>,
}

#[derive(Debug, Clone, Default)]
struct HeadP {
    cn: NormP,
    on: NormP,
    q: Vec<f32>,
    k: Vec<f32>,
    v: Vec<f32>,
}

/// A wired checkpoint; safe for concurrent scoring.
#[derive(Debug, Clone)]
pub struct Model {
    cfg: Config,
    emb: Vec<f32>,
    pos: Vec<f32>,
    enc: Vec<EncLayerP>,
    oenc: EncLayerP,
    head: HeadP,
}

fn need(
    tensors: &mut BTreeMap<String, Tensor>,
    name: &str,
    shape: &[i64],
) -> Result<Vec<f32>, ModelError> {
    let (tensor_shape, data) = {
        let tensor = tensors
            .get(name)
            .ok_or_else(|| invalid(format!("s1: checkpoint is missing {name:?}")))?;
        if tensor.shape.len() != shape.len() {
            return Err(invalid(format!(
                "s1: {name:?} has rank {}, want {}",
                tensor.shape.len(),
                shape.len()
            )));
        }
        for (got, want) in tensor.shape.iter().zip(shape.iter()) {
            if got != want {
                return Err(invalid(format!(
                    "s1: {name:?} has shape {:?}, want {:?}",
                    tensor.shape, shape
                )));
            }
        }
        (tensor.shape.clone(), tensor.data.clone())
    };
    let _ = tensor_shape;
    tensors.remove(name);
    Ok(data)
}

fn need_norm(
    tensors: &mut BTreeMap<String, Tensor>,
    prefix: &str,
    width: i64,
) -> Result<NormP, ModelError> {
    Ok(NormP {
        w: need(tensors, &format!("{prefix}.weight"), &[width])?,
        b: need(tensors, &format!("{prefix}.bias"), &[width])?,
    })
}

fn need_attn(
    tensors: &mut BTreeMap<String, Tensor>,
    prefix: &str,
    width: i64,
) -> Result<AttnP, ModelError> {
    Ok(AttnP {
        in_w: need(
            tensors,
            &format!("{prefix}.in_proj_weight"),
            &[3 * width, width],
        )?,
        in_b: need(tensors, &format!("{prefix}.in_proj_bias"), &[3 * width])?,
        out_w: need(
            tensors,
            &format!("{prefix}.out_proj.weight"),
            &[width, width],
        )?,
        out_b: need(tensors, &format!("{prefix}.out_proj.bias"), &[width])?,
    })
}

fn need_enc_layer(
    tensors: &mut BTreeMap<String, Tensor>,
    prefix: &str,
    width: i64,
) -> Result<EncLayerP, ModelError> {
    Ok(EncLayerP {
        n1: need_norm(tensors, &format!("{prefix}.norm1"), width)?,
        n2: need_norm(tensors, &format!("{prefix}.norm2"), width)?,
        attn: need_attn(tensors, &format!("{prefix}.self_attn"), width)?,
        w1: need(
            tensors,
            &format!("{prefix}.linear1.weight"),
            &[4 * width, width],
        )?,
        b1: need(tensors, &format!("{prefix}.linear1.bias"), &[4 * width])?,
        w2: need(
            tensors,
            &format!("{prefix}.linear2.weight"),
            &[width, 4 * width],
        )?,
        b2: need(tensors, &format!("{prefix}.linear2.bias"), &[width])?,
    })
}

impl Model {
    /// Wire decoded weights into a model. Fails on missing or
    /// unexpected tensors.
    pub fn new(cfg: Config, tensors: BTreeMap<String, Tensor>) -> Result<Self, ModelError> {
        if tensors.is_empty() {
            return Err(invalid("s1: checkpoint contains no tensors"));
        }
        let mut tensors = tensors;
        let width = cfg.width;
        let emb = need(&mut tensors, "embedding.weight", &[257, width])?;
        let pos_len = cfg.context_tokens.max(cfg.option_tokens);
        let pos = need(&mut tensors, "position.weight", &[pos_len, width])?;
        let mut enc = Vec::with_capacity(cfg.layers as usize);
        for index in 0..cfg.layers {
            enc.push(need_enc_layer(
                &mut tensors,
                &format!("encoder.layers.{index}"),
                width,
            )?);
        }
        let oenc = need_enc_layer(&mut tensors, "option_encoder.layers.0", width)?;
        let cn = need_norm(&mut tensors, "head.context_norm", width)?;
        let on = need_norm(&mut tensors, "head.option_norm", width)?;
        let q = need(&mut tensors, "head.query.weight", &[cfg.rank, width])?;
        let k = need(&mut tensors, "head.key.weight", &[cfg.rank, width])?;
        let v = need(&mut tensors, "head.value.weight", &[cfg.rank, width])?;
        if let Some((name, _)) = tensors.iter().next() {
            return Err(invalid(format!(
                "s1: unexpected checkpoint tensor {name:?}"
            )));
        }
        Ok(Model {
            cfg,
            emb,
            pos,
            enc,
            oenc,
            head: HeadP { cn, on, q, k, v },
        })
    }

    fn embed(&self, ids: &[i64]) -> Vec<f32> {
        let width = self.cfg.width as usize;
        let mut out = vec![0.0f32; ids.len() * width];
        for (t, id) in ids.iter().enumerate() {
            let e = &self.emb[*id as usize * width..(*id as usize + 1) * width];
            let p = &self.pos[t * width..(t + 1) * width];
            let o = &mut out[t * width..(t + 1) * width];
            for d in 0..width {
                o[d] = e[d] + p[d];
            }
        }
        out
    }

    fn enc_layer(&self, x: &[f32], valid: &[bool], p: &EncLayerP) -> Vec<f32> {
        let (w, h) = (self.cfg.width as usize, self.cfg.heads as usize);
        let dh = w / h;
        let scale = 1.0f32 / (dh as f32).sqrt();
        let len = valid.len();
        let mut normed = vec![0.0f32; x.len()];
        for t in 0..len {
            layernorm(
                &mut normed[t * w..(t + 1) * w],
                &x[t * w..(t + 1) * w],
                &p.n1.w,
                &p.n1.b,
            );
        }
        let mut qkv = vec![0.0f32; len * 3 * w];
        for t in 0..len {
            linear(
                &mut qkv[t * 3 * w..(t + 1) * 3 * w],
                &p.attn.in_w,
                Some(&p.attn.in_b),
                &normed[t * w..(t + 1) * w],
            );
        }
        let mut attn_out = vec![0.0f32; len * w];
        let mut scores = vec![0.0f32; len];
        for head in 0..h {
            for i in 0..len {
                for j in 0..len {
                    if !valid[j] {
                        scores[j] = f32::NEG_INFINITY;
                        continue;
                    }
                    let mut acc = 0.0f32;
                    let qo = i * 3 * w + head * dh;
                    let ko = j * 3 * w + w + head * dh;
                    for d in 0..dh {
                        acc += qkv[qo + d] * qkv[ko + d];
                    }
                    scores[j] = acc * scale;
                }
                softmax_in_place(&mut scores);
                for d in 0..dh {
                    let mut acc = 0.0f32;
                    for j in 0..len {
                        acc += scores[j] * qkv[j * 3 * w + 2 * w + head * dh + d];
                    }
                    attn_out[i * w + head * dh + d] = acc;
                }
            }
        }
        let mut y = x.to_vec();
        let mut proj = vec![0.0f32; w];
        for t in 0..len {
            linear(
                &mut proj,
                &p.attn.out_w,
                Some(&p.attn.out_b),
                &attn_out[t * w..(t + 1) * w],
            );
            add_into(&mut y[t * w..(t + 1) * w], &proj);
        }
        let mut ff = vec![0.0f32; 4 * w];
        let mut back = vec![0.0f32; w];
        for t in 0..len {
            layernorm(
                &mut normed[t * w..(t + 1) * w],
                &y[t * w..(t + 1) * w],
                &p.n2.w,
                &p.n2.b,
            );
            linear(&mut ff, &p.w1, Some(&p.b1), &normed[t * w..(t + 1) * w]);
            relu_in_place(&mut ff);
            linear(&mut back, &p.w2, Some(&p.b2), &ff);
            add_into(&mut y[t * w..(t + 1) * w], &back);
        }
        y
    }

    fn encode_context(&self, ids: &[i64], valid: &[bool]) -> Vec<f32> {
        let mut x = self.embed(ids);
        let safe = safe_mask(valid);
        for layer in &self.enc {
            x = self.enc_layer(&x, &safe, layer);
        }
        x
    }

    fn encode_options(&self, ids: &[Vec<i64>], tok_mask: &[Vec<bool>]) -> Vec<f32> {
        let width = self.cfg.width as usize;
        let count = ids.len();
        let mut out = vec![0.0f32; count * width];
        for (o, option_ids) in ids.iter().enumerate() {
            let x = self.embed(option_ids);
            let x = self.enc_layer(&x, &safe_mask(&tok_mask[o]), &self.oenc);
            let mut tokens = 0.0f32;
            for t in 0..option_ids.len() {
                if tok_mask[o][t] {
                    tokens += 1.0;
                    add_into(
                        &mut out[o * width..(o + 1) * width],
                        &x[t * width..(t + 1) * width],
                    );
                }
            }
            if tokens < 1.0 {
                tokens = 1.0;
            }
            let scale = 1.0 / tokens;
            for d in 0..width {
                out[o * width + d] *= scale;
            }
        }
        out
    }

    fn score_head(
        &self,
        ctx: &[f32],
        ctx_valid: &[bool],
        opts: &[f32],
        opt_valid: &[bool],
    ) -> Vec<f32> {
        let (w, r) = (self.cfg.width as usize, self.cfg.rank as usize);
        let lc = ctx_valid.len();
        let no = opt_valid.len();
        let mut cn = vec![0.0f32; ctx.len()];
        for t in 0..lc {
            layernorm(
                &mut cn[t * w..(t + 1) * w],
                &ctx[t * w..(t + 1) * w],
                &self.head.cn.w,
                &self.head.cn.b,
            );
        }
        let mut on = vec![0.0f32; opts.len()];
        for o in 0..no {
            layernorm(
                &mut on[o * w..(o + 1) * w],
                &opts[o * w..(o + 1) * w],
                &self.head.on.w,
                &self.head.on.b,
            );
        }
        let scale = 1.0f32 / (r as f32).sqrt();
        let mut q = vec![0.0f32; no * r];
        let mut k = vec![0.0f32; lc * r];
        let mut v = vec![0.0f32; lc * r];
        for o in 0..no {
            linear(
                &mut q[o * r..(o + 1) * r],
                &self.head.q,
                None,
                &on[o * w..(o + 1) * w],
            );
        }
        for t in 0..lc {
            linear(
                &mut k[t * r..(t + 1) * r],
                &self.head.k,
                None,
                &cn[t * w..(t + 1) * w],
            );
            linear(
                &mut v[t * r..(t + 1) * r],
                &self.head.v,
                None,
                &cn[t * w..(t + 1) * w],
            );
        }
        let mut logits = vec![0.0f32; no];
        let mut attn = vec![0.0f32; r];
        let mut p = vec![0.0f32; lc];
        for o in 0..no {
            if !opt_valid[o] {
                logits[o] = MASKED_LOGIT;
                continue;
            }
            for t in 0..lc {
                if !ctx_valid[t] {
                    p[t] = MASKED_LOGIT;
                    continue;
                }
                p[t] = dot(&q[o * r..(o + 1) * r], &k[t * r..(t + 1) * r]) * scale;
            }
            softmax_in_place(&mut p);
            for value in attn.iter_mut() {
                *value = 0.0;
            }
            for t in 0..lc {
                for d in 0..r {
                    attn[d] += p[t] * v[t * r + d];
                }
            }
            logits[o] = dot(&q[o * r..(o + 1) * r], &attn) * scale;
        }
        logits
    }

    fn forward(&self, batch: &Batch) -> Vec<Vec<f32>> {
        let mut logits = Vec::with_capacity(batch.ctx_ids.len());
        for i in 0..batch.ctx_ids.len() {
            let context_vec = self.encode_context(&batch.ctx_ids[i], &batch.ctx_mask[i]);
            let opts = self.encode_options(&batch.opt_ids[i], &batch.opt_tok[i]);
            logits.push(self.score_head(
                &context_vec,
                &batch.ctx_mask[i],
                &opts,
                &batch.opt_mask[i],
            ));
        }
        logits
    }

    /// Score one example and return raw per-option logits.
    pub fn logits(&self, ex: &Example) -> Result<Vec<f32>, ModelError> {
        let rows = self.logits_batch(std::slice::from_ref(ex))?;
        Ok(rows.into_iter().next().unwrap_or_default())
    }

    /// Score one logit row per example in a single pass.
    pub fn logits_batch(&self, exs: &[Example]) -> Result<Vec<Vec<f32>>, ModelError> {
        if exs.is_empty() {
            return Err(invalid("s1: cannot score an empty batch"));
        }
        let batch = collate(exs, self.cfg.context_tokens, self.cfg.option_tokens)?;
        Ok(self.forward(&batch))
    }

    /// Return normalized probabilities over the live options and the
    /// winning index.
    pub fn score(&self, ex: &Example) -> Result<(Vec<f64>, usize), ModelError> {
        let logits = self.logits(ex)?;
        let mut tmp = logits.clone();
        softmax_in_place(&mut tmp);
        let mut best = 0;
        for (i, p) in tmp.iter().enumerate() {
            if *p > tmp[best] {
                best = i;
            }
        }
        Ok((tmp.into_iter().map(f64::from).collect(), best))
    }
}

fn safe_mask(valid: &[bool]) -> Vec<bool> {
    let mut out = valid.to_vec();
    if !out.is_empty() {
        out[0] = true;
    }
    out
}

fn linear(y: &mut [f32], w: &[f32], b: Option<&[f32]>, x: &[f32]) {
    let cols = x.len();
    let rows = y.len();
    for r in 0..rows {
        let mut acc = 0.0f32;
        let base = r * cols;
        for c in 0..cols {
            acc += w[base + c] * x[c];
        }
        if let Some(bias) = b {
            acc += bias[r];
        }
        y[r] = acc;
    }
}

fn dot(a: &[f32], b: &[f32]) -> f32 {
    let mut acc = 0.0f32;
    for (i, value) in a.iter().enumerate() {
        acc += value * b[i];
    }
    acc
}

fn add_into(dst: &mut [f32], src: &[f32]) {
    for (i, value) in src.iter().enumerate() {
        dst[i] += value;
    }
}

fn layernorm(y: &mut [f32], x: &[f32], w: &[f32], b: &[f32]) {
    const EPS: f32 = 1e-5;
    let n = x.len() as f32;
    let mut mean = 0.0f32;
    for value in x {
        mean += value;
    }
    mean /= n;
    let mut variance = 0.0f32;
    for value in x {
        let d = value - mean;
        variance += d * d;
    }
    variance /= n;
    let inv = 1.0f32 / (variance + EPS).sqrt();
    for i in 0..x.len() {
        y[i] = (x[i] - mean) * inv * w[i] + b[i];
    }
}

fn softmax_in_place(a: &mut [f32]) {
    if a.is_empty() {
        return;
    }
    let mut max = a[0];
    for value in a.iter().skip(1) {
        if *value > max {
            max = *value;
        }
    }
    let mut sum = 0.0f32;
    for value in a.iter_mut() {
        let mut d = *value - max;
        if d.is_nan() {
            d = 0.0;
        }
        let e = d.exp();
        *value = e;
        sum += e;
    }
    if sum == 0.0 {
        let uniform = 1.0f32 / a.len() as f32;
        for value in a.iter_mut() {
            *value = uniform;
        }
        return;
    }
    let inv = 1.0 / sum;
    for value in a.iter_mut() {
        *value *= inv;
    }
}

fn relu_in_place(a: &mut [f32]) {
    for value in a.iter_mut() {
        if *value < 0.0 {
            *value = 0.0;
        }
    }
}

#[derive(Debug)]
struct Batch {
    ctx_ids: Vec<Vec<i64>>,
    ctx_mask: Vec<Vec<bool>>,
    opt_ids: Vec<Vec<Vec<i64>>>,
    opt_mask: Vec<Vec<bool>>,
    opt_tok: Vec<Vec<Vec<bool>>>,
}

fn byte_ids(text: &str, length: i64) -> Vec<i64> {
    let raw = text.as_bytes();
    let take = (length as usize).min(raw.len());
    raw[..take].iter().map(|byte| *byte as i64 + 1).collect()
}

fn collate(exs: &[Example], ctx_len: i64, opt_len: i64) -> Result<Batch, ModelError> {
    struct Row {
        ctx: Vec<i64>,
        opts: Vec<Vec<i64>>,
    }
    let mut rows: Vec<Row> = Vec::with_capacity(exs.len());
    let mut max_ctx = 1usize;
    let mut max_opts = 0usize;
    let mut max_tok = 1usize;
    for (i, ex) in exs.iter().enumerate() {
        if ex.context.is_empty() {
            return Err(invalid(format!("s1: example {i} has an empty context")));
        }
        if ex.options.is_empty() {
            return Err(invalid(format!("s1: example {i} has no options")));
        }
        let mut row = Row {
            ctx: byte_ids(&ex.context, ctx_len),
            opts: Vec::with_capacity(ex.options.len()),
        };
        for (j, option) in ex.options.iter().enumerate() {
            if option.is_empty() {
                return Err(invalid(format!("s1: example {i} option {j} is empty")));
            }
            row.opts.push(byte_ids(option, opt_len));
        }
        max_ctx = max_ctx.max(row.ctx.len());
        max_opts = max_opts.max(row.opts.len());
        for ids in &row.opts {
            max_tok = max_tok.max(ids.len());
        }
        rows.push(row);
    }
    let mut batch = Batch {
        ctx_ids: Vec::with_capacity(exs.len()),
        ctx_mask: Vec::with_capacity(exs.len()),
        opt_ids: Vec::with_capacity(exs.len()),
        opt_mask: Vec::with_capacity(exs.len()),
        opt_tok: Vec::with_capacity(exs.len()),
    };
    for row in rows {
        let mut ctx_ids = vec![0i64; max_ctx];
        let mut ctx_mask = vec![false; max_ctx];
        ctx_ids[..row.ctx.len()].copy_from_slice(&row.ctx);
        for mask in ctx_mask.iter_mut().take(row.ctx.len()) {
            *mask = true;
        }
        batch.ctx_ids.push(ctx_ids);
        batch.ctx_mask.push(ctx_mask);
        let mut opt_ids = vec![vec![0i64; max_tok]; max_opts];
        let mut opt_mask = vec![false; max_opts];
        let mut opt_tok = vec![vec![false; max_tok]; max_opts];
        for (o, ids) in row.opts.iter().enumerate() {
            opt_ids[o][..ids.len()].copy_from_slice(ids);
            opt_mask[o] = true;
            for mask in opt_tok[o].iter_mut().take(ids.len()) {
                *mask = true;
            }
        }
        batch.opt_ids.push(opt_ids);
        batch.opt_mask.push(opt_mask);
        batch.opt_tok.push(opt_tok);
    }
    Ok(batch)
}

#[cfg(test)]
mod tests {
    use super::{byte_ids, softmax_in_place};

    #[test]
    fn byte_ids_shift_by_one() {
        assert_eq!(byte_ids("ab", 8), vec![98, 99]);
        assert_eq!(byte_ids("abcdef", 3), vec![98, 99, 100]);
    }

    #[test]
    fn softmax_normalizes_and_handles_masked() {
        let mut v = vec![f32::NEG_INFINITY, 0.0];
        softmax_in_place(&mut v);
        assert!(v[0].abs() < 1e-6);
        assert!((v.into_iter().sum::<f32>() - 1.0).abs() < 1e-6);
    }

    #[test]
    fn softmax_uniform_when_all_equal() {
        let mut v = vec![1.0, 1.0, 1.0];
        softmax_in_place(&mut v);
        assert!(v.iter().all(|p| (*p - 1.0 / 3.0).abs() < 1e-6));
    }
}
