//! Classifier checkpoint loading: a safetensors file plus a JSON document
//! with the model config and a SHA-256 state signature. Mirrors
//! `internal/ai/classifier/checkpoint`.

use serde_json::Value;
use sha2::Digest;
use std::collections::BTreeMap;
use std::path::PathBuf;
use thiserror::Error;

const CHECKPOINT_FORMAT: &str = "cua-s1";
const MAX_CONFIG_BYTES: usize = 1 << 20;
const MAX_WEIGHTS_BYTES: usize = 64 << 20;
const MAX_HEADER_BYTES: usize = 8 << 20;

/// Config mirrors the validated model fields.
#[derive(Debug, Clone, PartialEq)]
pub struct Config {
    pub encoder: String,
    pub width: i64,
    pub rank: i64,
    pub context_tokens: i64,
    pub option_tokens: i64,
    pub layers: i64,
    pub heads: i64,
}

/// One F32 safetensors entry with its decoded shape.
#[derive(Debug, Clone)]
pub struct Tensor {
    pub shape: Vec<i64>,
    pub data: Vec<f32>,
}

/// A verified checkpoint with decoded weights and hosting metadata.
#[derive(Debug, Clone)]
pub struct Checkpoint {
    pub config: Config,
    pub tensors: BTreeMap<String, Tensor>,
    pub params: i64,
    pub weights_bytes: i64,
}

#[derive(Debug, Error, PartialEq)]
pub enum CheckpointError {
    #[error("checkpoint state signature mismatch")]
    SignatureMismatch,
    #[error("unsupported checkpoint format: {0}")]
    UnsupportedFormat(String),
    #[error("checkpoint not found")]
    NotFound,
    #[error("checkpoint IO: {0}")]
    Io(String),
    #[error("checkpoint: {0}")]
    Invalid(String),
}

pub fn kind(error: &CheckpointError) -> &'static str {
    match error {
        CheckpointError::SignatureMismatch => "signature_mismatch",
        CheckpointError::UnsupportedFormat(_) => "unsupported_format",
        CheckpointError::NotFound => "not_found",
        _ => "load_error",
    }
}

pub fn load(path: &str) -> Result<Checkpoint, CheckpointError> {
    let (weights_path, config_path) = resolve_paths(path)?;
    let raw = read_limited(&config_path, MAX_CONFIG_BYTES)?;
    let doc: Value = serde_json::from_slice(&raw)
        .map_err(|error| CheckpointError::Invalid(format!("invalid checkpoint JSON: {error}")))?;
    let doc = doc
        .as_object()
        .ok_or_else(|| CheckpointError::Invalid("checkpoint JSON must contain an object".into()))?;
    let (config, raw_config, signature) = parse_config(doc)?;
    let TensorFile { tensors, dtypes } = read_safetensors(&weights_path)?;
    verify_signature(&raw_config, &tensors, &dtypes, &signature)?;
    let weights_bytes = std::fs::metadata(&weights_path)
        .map_err(|error| CheckpointError::Invalid(format!("stat weights: {error}")))?
        .len() as i64;
    let params: i64 = tensors
        .values()
        .map(|tensor| tensor.data.len() as i64)
        .sum();
    Ok(Checkpoint {
        config,
        tensors,
        params,
        weights_bytes,
    })
}

fn resolve_paths(path: &str) -> Result<(PathBuf, PathBuf), CheckpointError> {
    let p = PathBuf::from(path);
    match p
        .extension()
        .and_then(|ext| ext.to_str())
        .map(str::to_ascii_lowercase)
        .as_deref()
    {
        Some("pt" | "pth" | "bin" | "pkl" | "pickle") => Err(CheckpointError::UnsupportedFormat(
            format!("legacy pickle checkpoints are not supported: {path}"),
        )),
        Some("safetensors") => {
            let config = PathBuf::from(path.strip_suffix(".safetensors").unwrap_or(path))
                .with_extension("json");
            Ok((p, config))
        }
        Some("json") => {
            let weights = PathBuf::from(path.strip_suffix(".json").unwrap_or(path))
                .with_extension("safetensors");
            Ok((weights, p))
        }
        None => {
            let meta = std::fs::metadata(path).map_err(|error| {
                if error.kind() == std::io::ErrorKind::NotFound {
                    CheckpointError::NotFound
                } else {
                    CheckpointError::Invalid(format!("checkpoint dir: {error}"))
                }
            })?;
            if !meta.is_dir() {
                let _ = meta;
                return Err(CheckpointError::Invalid(format!(
                    "checkpoint path must be a directory, .safetensors or .json: {path}"
                )));
            }
            Ok((
                PathBuf::from(path).join("model.safetensors"),
                PathBuf::from(path).join("config.json"),
            ))
        }
        _ => Err(CheckpointError::Invalid(format!(
            "checkpoint path must be a directory, .safetensors or .json: {path}"
        ))),
    }
}

fn read_limited(path: &PathBuf, limit: usize) -> Result<Vec<u8>, CheckpointError> {
    let display = path.display();
    let metadata = std::fs::metadata(path).map_err(|error| {
        if error.kind() == std::io::ErrorKind::NotFound {
            CheckpointError::NotFound
        } else {
            CheckpointError::Io(format!("read config: {display}"))
        }
    })?;
    if metadata.len() > limit as u64 {
        return Err(CheckpointError::Invalid(format!(
            "file exceeds {limit} bytes: {display}"
        )));
    }
    let bytes = std::fs::read(path)
        .map_err(|error| CheckpointError::Io(format!("read: {display}: {error}")))?;
    if bytes.len() > limit {
        return Err(CheckpointError::Invalid(format!(
            "file exceeds {limit} bytes: {display}"
        )));
    }
    Ok(bytes)
}

fn parse_config(
    doc: &serde_json::Map<String, Value>,
) -> Result<(Config, Value, String), CheckpointError> {
    if doc.get("format").and_then(Value::as_str) != Some(CHECKPOINT_FORMAT) {
        return Err(CheckpointError::UnsupportedFormat(format!(
            "format {:?}",
            doc.get("format")
        )));
    }
    if doc.get("format_version").and_then(Value::as_i64) != Some(1) {
        return Err(CheckpointError::UnsupportedFormat(format!(
            "version {:?}",
            doc.get("format_version")
        )));
    }
    let raw_config = doc
        .get("config")
        .filter(|value| value.is_object())
        .cloned()
        .ok_or_else(|| {
            CheckpointError::Invalid("checkpoint JSON field 'config' must be an object".into())
        })?;
    let signature = doc
        .get("state_signature")
        .and_then(Value::as_str)
        .filter(|sig| sig.len() == 64)
        .ok_or_else(|| {
            CheckpointError::Invalid(
                "checkpoint JSON field 'state_signature' must be a SHA-256 digest".into(),
            )
        })?
        .to_string();
    let raw = raw_config
        .as_object()
        .unwrap_or(&serde_json::Map::new())
        .clone();
    if raw.get("encoder").and_then(Value::as_str) != Some("tinyx") {
        return Err(CheckpointError::UnsupportedFormat(format!(
            "encoder {:?} (want tinyx)",
            raw.get("encoder")
        )));
    }
    let config = Config {
        encoder: raw
            .get("encoder")
            .and_then(Value::as_str)
            .unwrap_or_default()
            .to_string(),
        width: positive_int(&raw, "width")?,
        rank: positive_int(&raw, "rank")?,
        context_tokens: positive_int(&raw, "context_tokens")?,
        option_tokens: positive_int(&raw, "option_tokens")?,
        layers: positive_int(&raw, "layers")?,
        heads: positive_int(&raw, "heads")?,
    };
    if config.width % config.heads != 0 {
        return Err(CheckpointError::Invalid(
            "model width must be divisible by attention heads".into(),
        ));
    }
    Ok((config, raw_config, signature))
}

fn positive_int(map: &serde_json::Map<String, Value>, key: &str) -> Result<i64, CheckpointError> {
    let number = map.get(key).and_then(Value::as_i64).ok_or_else(|| {
        CheckpointError::Invalid(format!("config field {key:?} must be a positive integer"))
    })?;
    if number <= 0 {
        return Err(CheckpointError::Invalid(format!(
            "config field {key:?} must be a positive integer"
        )));
    }
    Ok(number)
}

struct TensorFile {
    tensors: BTreeMap<String, Tensor>,
    dtypes: BTreeMap<String, String>,
}

fn read_safetensors(path: &PathBuf) -> Result<TensorFile, CheckpointError> {
    let display = path.display();
    let raw = read_limited(path, MAX_WEIGHTS_BYTES)?;
    if raw.len() < 8 {
        return Err(CheckpointError::Invalid(format!(
            "truncated safetensors header: {display}"
        )));
    }
    let header_len = u64::from_le_bytes(raw[..8].try_into().unwrap_or_default()) as usize;
    if header_len > MAX_HEADER_BYTES {
        return Err(CheckpointError::Invalid(format!(
            "safetensors header exceeds {MAX_HEADER_BYTES} bytes: {display}"
        )));
    }
    if raw.len() - 8 < header_len {
        return Err(CheckpointError::Invalid(format!(
            "truncated safetensors header: {display}"
        )));
    }
    let header_raw = &raw[8..8 + header_len];
    let data = &raw[8 + header_len..];
    let header: Value = serde_json::from_slice(header_raw).map_err(|error| {
        CheckpointError::Invalid(format!("invalid safetensors header: {error}"))
    })?;
    let header = header.as_object().ok_or_else(|| {
        CheckpointError::Invalid(format!("invalid safetensors header in {display}"))
    })?;
    let mut tensors = BTreeMap::new();
    let mut dtypes = BTreeMap::new();
    for (name, entry) in header.iter() {
        if name == "__metadata__" {
            continue;
        }
        let entry = entry
            .as_object()
            .ok_or_else(|| CheckpointError::Invalid(format!("invalid header entry {name:?}")))?;
        if entry.get("dtype").and_then(Value::as_str) != Some("F32") {
            return Err(CheckpointError::Invalid(format!(
                "unsupported dtype {:?} for {name:?} (reads F32 only)",
                entry.get("dtype")
            )));
        }
        let shape_raw = entry
            .get("shape")
            .and_then(Value::as_array)
            .ok_or_else(|| CheckpointError::Invalid(format!("invalid shape for {name:?}")))?;
        let mut shape = Vec::with_capacity(shape_raw.len());
        let mut count: i64 = 1;
        for dim in shape_raw {
            let dim = dim
                .as_i64()
                .ok_or_else(|| CheckpointError::Invalid(format!("invalid shape for {name:?}")))?;
            if dim < 0 {
                return Err(CheckpointError::Invalid(format!(
                    "invalid shape for {name:?}"
                )));
            }
            if dim != 0 && count > i64::MAX / dim {
                return Err(CheckpointError::Invalid(format!(
                    "invalid shape for {name:?}"
                )));
            }
            count *= dim;
            shape.push(dim);
        }
        let offsets = entry
            .get("data_offsets")
            .and_then(Value::as_array)
            .ok_or_else(|| {
                CheckpointError::Invalid(format!("invalid data offsets for {name:?}"))
            })?;
        if offsets.len() != 2 {
            return Err(CheckpointError::Invalid(format!(
                "invalid data offsets for {name:?}"
            )));
        }
        let start = offsets[0].as_i64().ok_or_else(|| {
            CheckpointError::Invalid(format!("invalid data offsets for {name:?}"))
        })?;
        let end = offsets[1].as_i64().ok_or_else(|| {
            CheckpointError::Invalid(format!("invalid data offsets for {name:?}"))
        })?;
        if start < 0 || end < start || end > data.len() as i64 {
            return Err(CheckpointError::Invalid(format!(
                "data offsets out of range for {name:?}"
            )));
        }
        if count < 0 || count > (data.len() / 4) as i64 {
            return Err(CheckpointError::Invalid(format!(
                "data offsets out of range for {name:?}"
            )));
        }
        if end - start != count * 4 {
            return Err(CheckpointError::Invalid(format!(
                "size mismatch for {name:?}"
            )));
        }
        let mut values = Vec::with_capacity(count as usize);
        for i in 0..count as usize {
            let bytes: [u8; 4] = data[start as usize + 4 * i..start as usize + 4 * i + 4]
                .try_into()
                .unwrap_or_default();
            values.push(f32::from_le_bytes(bytes));
        }
        tensors.insert(
            name.clone(),
            Tensor {
                shape: shape.clone(),
                data: values,
            },
        );
        dtypes.insert(name.clone(), "torch.float32".to_string());
        let _ = shape;
    }
    let _ = display;
    Ok(TensorFile { tensors, dtypes })
}

fn verify_signature(
    raw_config: &Value,
    tensors: &BTreeMap<String, Tensor>,
    dtypes: &BTreeMap<String, String>,
    expected: &str,
) -> Result<(), CheckpointError> {
    let mut digest = sha2::Sha256::new();
    let canonical_config = canonical(raw_config).map_err(CheckpointError::Invalid)?;
    digest.update(canonical_config.as_bytes());
    for (name, tensor) in tensors.iter() {
        let shape: Vec<Value> = tensor.shape.iter().map(|dim| Value::from(*dim)).collect();
        let row = canonical(&Value::from(vec![
            Value::from(name.as_str()),
            Value::from(
                dtypes
                    .get(name)
                    .map(String::as_str)
                    .unwrap_or("torch.float32"),
            ),
            Value::from(shape),
        ]))
        .map_err(CheckpointError::Invalid)?;
        digest.update(row.as_bytes());
        for value in &tensor.data {
            digest.update(value.to_le_bytes());
        }
    }
    let got = hex::encode(digest.finalize());
    if got != expected {
        return Err(CheckpointError::SignatureMismatch);
    }
    Ok(())
}

/// Python `json.dumps` compatible canonical JSON, mirroring Go
/// `canonical` in `checkpoint.go`.
pub fn canonical(value: &Value) -> Result<String, String> {
    match value {
        Value::Null => Ok("null".into()),
        Value::Bool(true) => Ok("true".into()),
        Value::Bool(false) => Ok("false".into()),
        Value::Number(number) => {
            if let Some(int) = number.as_i64() {
                Ok(int.to_string())
            } else if let Some(float) = number.as_f64() {
                Ok(py_float(float))
            } else {
                Ok(number.to_string())
            }
        }
        Value::String(text) => Ok(canonical_string(text)),
        Value::Array(items) => {
            let mut out = String::from("[");
            for (i, item) in items.iter().enumerate() {
                if i > 0 {
                    out.push(',');
                }
                out.push_str(&canonical(item)?);
            }
            out.push(']');
            Ok(out)
        }
        Value::Object(map) => {
            let mut keys: Vec<&String> = map.keys().collect();
            keys.sort();
            let mut out = String::from("{");
            for (i, key) in keys.iter().enumerate() {
                if i > 0 {
                    out.push(',');
                }
                out.push_str(&canonical_string(key));
                out.push(':');
                out.push_str(&canonical(&map[*key])?);
            }
            out.push('}');
            Ok(out)
        }
    }
}

fn canonical_string(text: &str) -> String {
    let mut out = String::from("\"");
    for ch in text.chars() {
        match ch {
            '"' => out.push_str("\\\""),
            '\\' => out.push_str("\\\\"),
            '\n' => out.push_str("\\n"),
            '\r' => out.push_str("\\r"),
            '\t' => out.push_str("\\t"),
            '\u{8}' => out.push_str("\\b"),
            '\u{c}' => out.push_str("\\f"),
            c if (c as u32) < 0x20 => out.push_str(&format!("\\u{:04x}", c as u32)),
            c if (c as u32) < 0x7f => out.push(c),
            c if (c as u32) <= 0xffff => out.push_str(&format!("\\u{:04x}", c as u32)),
            c => {
                let cp = c as u32 - 0x10000;
                out.push_str(&format!(
                    "\\u{:04x}\\u{:04x}",
                    0xd800 + (cp >> 10),
                    0xdc00 + (cp & 0x3ff)
                ));
            }
        }
    }
    out.push('"');
    out
}

/// Formats `f` like Python's `json.dumps`: shortest round-trip digits,
/// fixed notation for `-4 < decpt <= 16`, scientific otherwise.
fn py_float(value: f64) -> String {
    if value.is_nan() {
        return "NaN".into();
    }
    if value.is_infinite() {
        return if value.is_sign_positive() {
            "Infinity".into()
        } else {
            "-Infinity".into()
        };
    }
    let (sign, f) = if value.is_sign_negative() && value != 0.0 {
        ("-", -value)
    } else if value == 0.0 {
        let negative_zero = value.is_sign_negative();
        return if negative_zero {
            "-0.0".into()
        } else {
            "0.0".into()
        };
    } else {
        ("", value)
    };
    // Shortest scientific form `d.ddd e exp`.
    let raw = format!("{f:e}");
    let (mantissa, exp_str) = raw.split_once('e').unwrap_or((raw.as_str(), "0"));
    let exp: i32 = exp_str.parse().unwrap_or(0);
    let digits: String = mantissa.replace('.', "").trim_end_matches('0').to_string();
    let digits = if digits.is_empty() {
        "0".to_string()
    } else {
        digits
    };
    let decpt = exp + 1;
    if decpt > -4 && decpt <= 16 {
        let mut out = String::from(sign);
        if decpt <= 0 {
            out.push_str("0.");
            out.push_str(&"0".repeat((-decpt) as usize));
            out.push_str(&digits);
        } else if decpt as usize >= digits.len() {
            out.push_str(&digits);
            out.push_str(&"0".repeat(decpt as usize - digits.len()));
            out.push_str(".0");
        } else {
            out.push_str(&digits[..decpt as usize]);
            out.push('.');
            out.push_str(&digits[decpt as usize..]);
        }
        out
    } else {
        let mut out = String::from(sign);
        out.push(digits.as_bytes()[0] as char);
        if digits.len() > 1 {
            out.push('.');
            out.push_str(&digits[1..]);
        }
        out.push('e');
        if exp < 0 {
            out.push('-');
        } else {
            out.push('+');
        }
        let magnitude = exp.abs();
        if magnitude < 10 {
            out.push('0');
        }
        out.push_str(&magnitude.to_string());
        out
    }
}

mod hex {
    pub fn encode(bytes: impl AsRef<[u8]>) -> String {
        bytes
            .as_ref()
            .iter()
            .map(|byte| format!("{byte:02x}"))
            .collect()
    }
}

#[cfg(test)]
mod tests {
    use super::{canonical, canonical_string, py_float, CheckpointError, Config};

    #[test]
    fn canonical_json_matches_go() {
        let value: serde_json::Value =
            serde_json::from_str("{\"b\":1,\"a\":[true,null]}").unwrap_or_default();
        assert_eq!(
            canonical(&value).unwrap_or_default(),
            "{\"a\":[true,null],\"b\":1}"
        );
        assert_eq!(canonical_string("héllo"), "\"h\\u00e9llo\"");
        assert_eq!(canonical_string("a\nb"), "\"a\\nb\"");
    }

    #[test]
    fn py_float_matches_python() {
        assert_eq!(py_float(0.5), "0.5");
        assert_eq!(py_float(2.0), "2.0");
        assert_eq!(py_float(0.0001), "0.0001");
        assert_eq!(py_float(1e16), "1e+16");
        assert_eq!(py_float(1e15), "1000000000000000.0");
        assert_eq!(py_float(-0.01), "-0.01");
    }

    #[test]
    fn config_and_errors() {
        let cfg = Config {
            encoder: "tinyx".into(),
            width: 16,
            rank: 8,
            context_tokens: 32,
            option_tokens: 16,
            layers: 2,
            heads: 4,
        };
        assert_eq!(cfg.width % cfg.heads, 0);
        assert_eq!(
            super::kind(&CheckpointError::SignatureMismatch),
            "signature_mismatch"
        );
        assert_eq!(
            super::kind(&CheckpointError::UnsupportedFormat("x".into())),
            "unsupported_format"
        );
    }
}
