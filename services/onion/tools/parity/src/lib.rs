//! HTTP replay parity skeleton.
//!
//! A corpus is JSONL, one request vector per line:
//! `{"id":"...","method":"GET","path":"/v1/things?limit=10","body":null}`.
//! A recorded run is JSONL, one response per vector:
//! `{"id":"...","status":200,"body":{...}}`.
//!
//! Live replay (`parity replay --go http://... --rust http://...`) seeds
//! the throwaway MySQL from `scripts/test-db.sh` (container
//! `onion-env-db-1`, capped with
//! `docker update --memory 1g --memory-swap 1g --cpus 1 onion-env-db-1`),
//! runs both servers against identical data, replays the corpus against
//! each, and diffs the normalized responses. Per-module replay wires up
//! in Phase 3; the whole-app diff is the Phase 4 gate.

use std::collections::BTreeMap;

use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct Vector {
    pub id: String,
    pub method: String,
    pub path: String,
    pub body: Option<serde_json::Value>,
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct Record {
    pub id: String,
    pub status: u16,
    pub body: Option<serde_json::Value>,
}

#[derive(Debug, Clone, PartialEq)]
pub struct CorpusError {
    pub line: usize,
    pub message: String,
}

impl std::fmt::Display for CorpusError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "line {}: {}", self.line, self.message)
    }
}

impl std::error::Error for CorpusError {}

fn parse_jsonl<T: for<'a> Deserialize<'a>>(text: &str, kind: &str) -> Result<Vec<T>, CorpusError> {
    let mut out = Vec::new();
    for (index, line) in text.lines().enumerate() {
        if line.trim().is_empty() {
            continue;
        }
        let value: T = serde_json::from_str(line).map_err(|e| CorpusError {
            line: index + 1,
            message: format!("invalid {kind}: {e}"),
        })?;
        out.push(value);
    }
    Ok(out)
}

pub fn parse_corpus(text: &str) -> Result<Vec<Vector>, CorpusError> {
    parse_jsonl(text, "vector")
}

pub fn parse_run(text: &str) -> Result<Vec<Record>, CorpusError> {
    parse_jsonl(text, "record")
}

/// Canonicalize a JSON value for diffing: sort object keys recursively.
pub fn canonicalize(value: &serde_json::Value) -> serde_json::Value {
    match value {
        serde_json::Value::Object(map) => {
            let mut sorted = serde_json::Map::new();
            let mut keys: Vec<&String> = map.keys().collect();
            keys.sort();
            for key in keys {
                sorted.insert(key.clone(), canonicalize(&map[key]));
            }
            serde_json::Value::Object(sorted)
        }
        serde_json::Value::Array(items) => {
            serde_json::Value::Array(items.iter().map(canonicalize).collect())
        }
        other => other.clone(),
    }
}

#[derive(Debug, Clone, PartialEq)]
pub struct Diff {
    pub id: String,
    pub go_status: u16,
    pub rust_status: u16,
    pub bodies_equal: bool,
}

/// Diff two runs by vector id. Missing ids count as unequal.
pub fn diff_runs(go: &[Record], rust: &[Record]) -> Vec<Diff> {
    let rust_by_id: BTreeMap<&str, &Record> = rust.iter().map(|r| (r.id.as_str(), r)).collect();
    let mut diffs = Vec::new();
    for record in go {
        let (rust_status, bodies_equal) = match rust_by_id.get(record.id.as_str()) {
            Some(other) => {
                let go_body = record.body.as_ref().map(canonicalize);
                let rust_body = other.body.as_ref().map(canonicalize);
                (other.status, go_body == rust_body)
            }
            None => (0, false),
        };
        if record.status != rust_status || !bodies_equal {
            diffs.push(Diff {
                id: record.id.clone(),
                go_status: record.status,
                rust_status,
                bodies_equal,
            });
        }
    }
    // Include ids only in `rust` as status-mismatched diffs.
    for other in rust {
        if !go.iter().any(|r| r.id == other.id) {
            diffs.push(Diff {
                id: other.id.clone(),
                go_status: 0,
                rust_status: other.status,
                bodies_equal: false,
            });
        }
    }
    diffs
}

#[cfg(test)]
mod tests {
    use super::{canonicalize, diff_runs, parse_corpus, parse_run};

    #[test]
    fn parses_corpus_and_run() {
        let corpus =
            parse_corpus("{\"id\":\"a\",\"method\":\"GET\",\"path\":\"/x\",\"body\":null}\n")
                .unwrap();
        assert_eq!(corpus.len(), 1);
        assert_eq!(corpus[0].id, "a");
        let run = parse_run("{\"id\":\"a\",\"status\":200,\"body\":{\"b\":1,\"a\":2}}\n").unwrap();
        assert_eq!(run[0].status, 200);
    }

    #[test]
    fn canonicalize_sorts_keys() {
        let value: serde_json::Value =
            serde_json::from_str("{\"b\":1,\"a\":{\"d\":4,\"c\":3}}").unwrap();
        let out = canonicalize(&value);
        assert_eq!(
            serde_json::to_string(&out).unwrap(),
            "{\"a\":{\"c\":3,\"d\":4},\"b\":1}"
        );
    }

    #[test]
    fn diff_catches_status_and_body_mismatch() {
        let go = parse_run(
            "{\"id\":\"a\",\"status\":200,\"body\":{\"x\":1}}\n{\"id\":\"b\",\"status\":200,\"body\":null}\n",
        )
        .unwrap();
        let rust = parse_run(
            "{\"id\":\"a\",\"status\":200,\"body\":{\"x\":2}}\n{\"id\":\"b\",\"status\":404,\"body\":null}\n",
        )
        .unwrap();
        let diffs = diff_runs(&go, &rust);
        assert_eq!(diffs.len(), 2);
        assert!(diffs.iter().any(|d| d.id == "a" && !d.bodies_equal));
        assert!(diffs
            .iter()
            .any(|d| d.id == "b" && d.go_status != d.rust_status));
    }

    #[test]
    fn diff_clean_when_equal() {
        let go = parse_run("{\"id\":\"a\",\"status\":200,\"body\":{\"b\":1,\"a\":2}}\n").unwrap();
        let rust = parse_run("{\"id\":\"a\",\"status\":200,\"body\":{\"a\":2,\"b\":1}}\n").unwrap();
        assert!(diff_runs(&go, &rust).is_empty());
    }
}
