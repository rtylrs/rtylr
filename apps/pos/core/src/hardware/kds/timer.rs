//! Cook-time bookkeeping for KDS items.
//!
//! Two pure-function concerns live here:
//!
//! 1. **Visual tier** - Given a target cook time and how long an item
//!    has been queued/cooking, return a colored "tier" (`Fresh /
//!    Warning / Overdue`). This is the single visual cue cooks
//!    actually use; we keep the calculation pure so it can be unit-
//!    tested without wall-clock injection.
//!
//! 2. **Quote times** - An exponential-weighted moving average of
//!    recent completed cook times per station, used by the POS to
//!    show "≈ N min" when adding a new item. EWMA is preferred over
//!    a sliding window because it adapts to load shifts without
//!    storing a long history.

use serde::{Deserialize, Serialize};

#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum CookTier {
    /// Below 80% of target - green.
    Fresh,
    /// Between 80% and 100% - yellow.
    Warning,
    /// Over 100% of target - red.
    Overdue,
}

/// Classify how late a cooking item is against its target.
///
/// `elapsed_seconds` and `target_seconds` are u64 to avoid f64 drift
/// over long shifts. A `target_seconds` of 0 always returns Overdue
/// (defensive - a misconfigured target shouldn't silently report green).
pub fn cook_tier(elapsed_seconds: u64, target_seconds: u64) -> CookTier {
    if target_seconds == 0 {
        return CookTier::Overdue;
    }
    if elapsed_seconds > target_seconds {
        return CookTier::Overdue;
    }
    // Integer floor of (elapsed * 100) / target. Below 80 → fresh,
    // 80..=100 → warning. saturating_mul guards u64 overflow.
    let pct = elapsed_seconds.saturating_mul(100) / target_seconds;
    if pct < 80 {
        CookTier::Fresh
    } else {
        CookTier::Warning
    }
}

/// Exponential-weighted moving average of cook-time samples (seconds).
///
/// `alpha` is the smoothing factor in [0,1]; higher = more reactive to
/// recent samples. POS feedback latency makes 0.2–0.4 a good default.
#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct CookTimeEwma {
    /// Smoothing factor in [0,1].
    pub alpha: f64,
    /// Current estimate in seconds. None until the first sample.
    pub estimate_seconds: Option<f64>,
    /// Number of samples observed (for telemetry / confidence display).
    pub samples: u64,
}

impl CookTimeEwma {
    pub fn new(alpha: f64) -> Self {
        let alpha = alpha.clamp(0.0, 1.0);
        Self { alpha, estimate_seconds: None, samples: 0 }
    }

    /// Record a completed cook-time sample. Non-finite samples (NaN,
    /// infinity) and negative samples are ignored so a clock glitch
    /// can't poison the estimate.
    pub fn observe(&mut self, sample_seconds: f64) {
        if !sample_seconds.is_finite() || sample_seconds < 0.0 {
            return;
        }
        self.estimate_seconds = Some(match self.estimate_seconds {
            None => sample_seconds,
            Some(prev) => self.alpha * sample_seconds + (1.0 - self.alpha) * prev,
        });
        self.samples = self.samples.saturating_add(1);
    }

    /// Quote time in whole seconds, clamped to [floor, ceil] to keep
    /// frontends from showing silly "0 min" or "60 min" estimates.
    pub fn quote_seconds(&self, floor_secs: u64, ceil_secs: u64) -> Option<u64> {
        let est = self.estimate_seconds?;
        let rounded = est.round().max(0.0) as u64;
        Some(rounded.clamp(floor_secs, ceil_secs))
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    // ── Cook tier ─────────────────────────────────────────────────────────────

    #[test]
    fn fresh_below_80_percent() {
        assert_eq!(cook_tier(0, 600), CookTier::Fresh);
        assert_eq!(cook_tier(479, 600), CookTier::Fresh); // 79.83%
    }

    #[test]
    fn warning_between_80_and_100_percent() {
        assert_eq!(cook_tier(480, 600), CookTier::Warning); // 80%
        assert_eq!(cook_tier(600, 600), CookTier::Warning); // 100%
    }

    #[test]
    fn overdue_above_100_percent() {
        assert_eq!(cook_tier(601, 600), CookTier::Overdue);
        assert_eq!(cook_tier(1_000_000, 600), CookTier::Overdue);
    }

    #[test]
    fn zero_target_is_always_overdue() {
        // Misconfigured target - never silently report green.
        assert_eq!(cook_tier(0, 0), CookTier::Overdue);
        assert_eq!(cook_tier(100, 0), CookTier::Overdue);
    }

    #[test]
    fn cook_tier_does_not_overflow_on_huge_elapsed() {
        // saturating_mul guards against u64 overflow during the % calc.
        let _ = cook_tier(u64::MAX, 600);
    }

    // ── EWMA ──────────────────────────────────────────────────────────────────

    #[test]
    fn ewma_starts_unset_and_first_sample_seeds() {
        let mut e = CookTimeEwma::new(0.3);
        assert!(e.estimate_seconds.is_none());
        e.observe(300.0);
        assert_eq!(e.estimate_seconds, Some(300.0));
        assert_eq!(e.samples, 1);
    }

    #[test]
    fn ewma_smooths_toward_new_samples() {
        let mut e = CookTimeEwma::new(0.5);
        e.observe(100.0);
        e.observe(200.0);
        // 0.5 * 200 + 0.5 * 100 = 150
        assert_eq!(e.estimate_seconds, Some(150.0));
    }

    #[test]
    fn ewma_alpha_one_immediately_follows_latest() {
        let mut e = CookTimeEwma::new(1.0);
        e.observe(100.0);
        e.observe(500.0);
        assert_eq!(e.estimate_seconds, Some(500.0));
    }

    #[test]
    fn ewma_alpha_zero_freezes_first_sample() {
        let mut e = CookTimeEwma::new(0.0);
        e.observe(100.0);
        e.observe(500.0);
        assert_eq!(e.estimate_seconds, Some(100.0));
    }

    #[test]
    fn ewma_alpha_is_clamped_to_unit_interval() {
        let high = CookTimeEwma::new(2.5);
        assert_eq!(high.alpha, 1.0);
        let low = CookTimeEwma::new(-0.5);
        assert_eq!(low.alpha, 0.0);
    }

    #[test]
    fn ewma_ignores_nan_and_negative_samples() {
        let mut e = CookTimeEwma::new(0.5);
        e.observe(200.0);
        e.observe(f64::NAN);
        e.observe(-100.0);
        e.observe(f64::INFINITY);
        // Only the first sample stuck
        assert_eq!(e.estimate_seconds, Some(200.0));
        assert_eq!(e.samples, 1);
    }

    #[test]
    fn quote_seconds_clamps_to_configured_bounds() {
        let mut e = CookTimeEwma::new(1.0);
        e.observe(5.0);
        assert_eq!(e.quote_seconds(60, 600), Some(60));
        e.observe(9999.0);
        assert_eq!(e.quote_seconds(60, 600), Some(600));
        e.observe(300.0);
        assert_eq!(e.quote_seconds(60, 600), Some(300));
    }

    #[test]
    fn quote_seconds_is_none_before_first_sample() {
        let e = CookTimeEwma::new(0.3);
        assert_eq!(e.quote_seconds(60, 600), None);
    }
}
