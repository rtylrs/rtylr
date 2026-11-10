//! Budget tier classification. Mirrors `internal/application/budget.go`:
//! the tier is decided by the largest figure in the string, in thousands
//! of dollars. `20K`/`5k`/`15k+` and `10,000`/`25,000` count; bare digit
//! runs and a `k` that starts a word do not.

use regex::Regex;
use std::sync::OnceLock;

pub const TIER_HOT: &str = "HOT";
pub const TIER_WARM: &str = "WARM";
pub const TIER_NEW: &str = "NEW";

const HOT_FROM_THOUSANDS: f64 = 15.0;
const WARM_FROM_THOUSANDS: f64 = 5.0;

fn budget_figure() -> &'static Regex {
    static RE: OnceLock<Regex> = OnceLock::new();
    RE.get_or_init(|| {
        Regex::new(r"(?i)(\d{1,3}(?:,\d{3})+)|(\d+)\s*(k)\b")
            .unwrap_or_else(|_| Regex::new(r"(\d+)\s*(k)\b").expect("fallback budget regex"))
    })
}

/// Classify a free-text budget string.
pub fn budget_tier(budget: &str) -> &'static str {
    let mut max_thousands = 0.0f64;
    for captures in budget_figure().captures_iter(budget) {
        if let Some(comma_group) = captures.get(1) {
            // "10,000" is thousands-separated dollars.
            let raw = comma_group.as_str().replace(',', "");
            if let Ok(value) = raw.parse::<f64>() {
                max_thousands = max_thousands.max(value / 1000.0);
            }
            continue;
        }
        if captures.get(3).is_some() {
            // "20K" is twenty thousand.
            if let Some(digits) = captures.get(2) {
                if let Ok(value) = digits.as_str().parse::<f64>() {
                    max_thousands = max_thousands.max(value);
                }
            }
        }
        // Bare digits with no K and no comma separator are ignored.
    }
    if max_thousands >= HOT_FROM_THOUSANDS {
        TIER_HOT
    } else if max_thousands >= WARM_FROM_THOUSANDS {
        TIER_WARM
    } else {
        TIER_NEW
    }
}

#[cfg(test)]
mod tests {
    use super::{budget_tier, TIER_HOT, TIER_NEW, TIER_WARM};

    #[test]
    fn tiers() {
        assert_eq!(budget_tier("$8K - $20K"), TIER_HOT);
        assert_eq!(budget_tier("$5k-$15k"), TIER_HOT);
        assert_eq!(budget_tier("15k+"), TIER_HOT);
        assert_eq!(budget_tier("10,000"), TIER_WARM);
        assert_eq!(budget_tier("25,000"), TIER_HOT);
        assert_eq!(budget_tier("5k"), TIER_WARM);
        assert_eq!(budget_tier("20260115"), TIER_NEW);
        assert_eq!(budget_tier("within 2 weeks"), TIER_NEW);
        assert_eq!(budget_tier("15 key pages"), TIER_NEW);
        assert_eq!(budget_tier("8 keywords"), TIER_NEW);
        assert_eq!(budget_tier(""), TIER_NEW);
    }
}
