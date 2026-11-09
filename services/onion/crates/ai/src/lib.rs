//! Form-element classifier: loads a cua-s1 checkpoint and scores form
//! elements against candidate document values. Mirrors
//! `internal/ai/classifier`.

pub mod checkpoint;
pub mod model;
pub mod planner;

use checkpoint::{CheckpointError, Config};
use model::Model;
use thiserror::Error;

pub use model::Example;
pub use planner::{Decision, Entity, Field, Plan, Request, Scorer};

#[derive(Debug, Error, PartialEq)]
pub enum LoadError {
    #[error("{0}")]
    Checkpoint(#[from] CheckpointError),
    #[error("{0}")]
    Model(#[from] model::ModelError),
}

/// A loaded checkpoint ready to score form elements.
#[derive(Debug, Clone)]
pub struct ScorerLoaded {
    checkpoint: crate::checkpoint::Checkpoint,
    model: Model,
}

impl ScorerLoaded {
    /// Read a cua-s1 checkpoint from a directory, `.safetensors`, or
    /// `.json` path, verify it, and wire the weights into memory.
    pub fn load(path: &str) -> Result<Self, LoadError> {
        let checkpoint = checkpoint::load(path)?;
        let model = Model::new(checkpoint.config.clone(), checkpoint.tensors.clone())?;
        Ok(ScorerLoaded { checkpoint, model })
    }

    pub fn config(&self) -> &Config {
        &self.checkpoint.config
    }

    pub fn params(&self) -> i64 {
        self.checkpoint.params
    }

    pub fn weights_bytes(&self) -> i64 {
        self.checkpoint.weights_bytes
    }

    /// Score one example and return raw per-option logits.
    pub fn logits(&self, ex: &Example) -> Result<Vec<f32>, LoadError> {
        self.model.logits(ex).map_err(LoadError::Model)
    }

    /// Score a whole form in one pass.
    pub fn logits_batch(&self, exs: &[Example]) -> Result<Vec<Vec<f32>>, LoadError> {
        self.model.logits_batch(exs).map_err(LoadError::Model)
    }

    /// Return normalized probabilities over the live options and the
    /// winning index.
    pub fn score(&self, ex: &Example) -> Result<(Vec<f64>, usize), LoadError> {
        self.model.score(ex).map_err(LoadError::Model)
    }
}

impl Scorer for ScorerLoaded {
    fn logits_batch(
        &self,
        exs: &[Example],
    ) -> Option<Result<Vec<Vec<f32>>, planner::PlannerError>> {
        Some(self.logits_batch(exs).map_err(|error| match error {
            LoadError::Checkpoint(error) => planner::PlannerError::Invalid(error.to_string()),
            LoadError::Model(error) => planner::PlannerError::Model(error),
        }))
    }

    fn score(&self, ex: &Example) -> Result<(Vec<f64>, usize), planner::PlannerError> {
        self.score(ex).map_err(|error| match error {
            LoadError::Checkpoint(error) => planner::PlannerError::Invalid(error.to_string()),
            LoadError::Model(error) => planner::PlannerError::Model(error),
        })
    }
}

/// Classify a checkpoint load failure without exposing filesystem paths.
pub fn checkpoint_error_kind(error: &LoadError) -> &'static str {
    match error {
        LoadError::Checkpoint(error) => checkpoint::kind(error),
        LoadError::Model(_) => "load_error",
    }
}

/// Workspace crate version, surfaced in CI and version checks.
pub const CRATE_VERSION: &str = env!("CARGO_PKG_VERSION");

#[cfg(test)]
mod tests {
    use super::CRATE_VERSION;

    #[test]
    fn crate_version_is_nonempty() {
        assert!(!CRATE_VERSION.is_empty());
    }

    fn fixture_dir() -> std::path::PathBuf {
        std::path::PathBuf::from(env!("CARGO_MANIFEST_DIR"))
            .join("../../internal/ai/classifier/testdata/forms")
    }

    #[derive(Debug, serde::Deserialize)]
    struct GoldenExample {
        context: String,
        options: Vec<String>,
        logits: Vec<f64>,
        reference_argmax: usize,
    }

    #[derive(Debug, serde::Deserialize)]
    struct GoldenFile {
        examples: Vec<GoldenExample>,
    }

    fn argmax(logits: &[f32]) -> usize {
        let mut best = 0;
        for (i, value) in logits.iter().enumerate() {
            if *value > logits[best] {
                best = i;
            }
        }
        best
    }

    #[test]
    fn golden_parity() {
        let s = super::ScorerLoaded::load(&fixture_dir().to_string_lossy())
            .expect("Load fixture checkpoint");
        assert_eq!(s.params(), 14_720);
        let raw = std::fs::read(fixture_dir().join("golden.json")).expect("read golden");
        let golden: GoldenFile = serde_json::from_slice(&raw).expect("parse golden");
        assert!(!golden.examples.is_empty(), "golden has no examples");
        for (i, example) in golden.examples.iter().enumerate() {
            let logits = s
                .logits(&super::Example {
                    context: example.context.clone(),
                    options: example.options.clone(),
                })
                .expect("logits");
            assert_eq!(logits.len(), example.logits.len(), "example {i} logits len");
            for (o, (got, want)) in logits.iter().zip(example.logits.iter()).enumerate() {
                let diff = (f64::from(*got) - want).abs();
                assert!(
                    diff < 1e-4,
                    "example {i} option {o}: logit {got}, want {want} (diff {diff})"
                );
            }
            assert_eq!(
                argmax(&logits),
                example.reference_argmax,
                "example {i} argmax"
            );
        }
    }

    #[test]
    fn batch_parity() {
        let s = super::ScorerLoaded::load(&fixture_dir().to_string_lossy())
            .expect("Load fixture checkpoint");
        let raw = std::fs::read(fixture_dir().join("golden.json")).expect("read golden");
        let golden: GoldenFile = serde_json::from_slice(&raw).expect("parse golden");
        assert!(!golden.examples.is_empty());
        let exs: Vec<super::Example> = golden
            .examples
            .iter()
            .map(|ex| super::Example {
                context: ex.context.clone(),
                options: ex.options.clone(),
            })
            .collect();
        let rows = s.logits_batch(&exs).expect("LogitsBatch");
        assert_eq!(rows.len(), golden.examples.len());
        for (i, row) in rows.iter().enumerate() {
            assert_eq!(
                argmax(row),
                golden.examples[i].reference_argmax,
                "example {i} batch argmax"
            );
        }
    }

    #[test]
    fn rejects_empty_and_pickle() {
        let s = super::ScorerLoaded::load(&fixture_dir().to_string_lossy())
            .unwrap_or_else(|error| panic!("Load: {error}"));
        assert!(s
            .logits(&super::Example {
                context: String::new(),
                options: vec!["click".into(), "skip".into()],
            })
            .is_err());
        assert!(s
            .logits(&super::Example {
                context: "x".into(),
                options: Vec::new(),
            })
            .is_err());
        assert!(
            super::ScorerLoaded::load(&fixture_dir().join("model.pt").to_string_lossy()).is_err()
        );
    }
}
