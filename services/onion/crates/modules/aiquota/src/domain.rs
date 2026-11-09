//! Domain: allocation modes, entitlements, and quota views.
//! Mirrors `internal/domain`.

/// The per-tenant allocation policy for the monthly AI allowance.
#[derive(Debug, Clone, Copy, PartialEq)]
pub enum Mode {
    /// One allowance shared across the whole team (default).
    Pool,
    /// Splits the allowance evenly across seats.
    PerPerson,
}

impl Mode {
    pub fn as_str(self) -> &'static str {
        match self {
            Mode::Pool => "pool",
            Mode::PerPerson => "per_person",
        }
    }

    /// Go `Normalize()`: an already-valid Mode passes through.
    pub fn normalize(self) -> Self {
        self
    }

    pub fn parse(text: &str) -> Self {
        if text == Mode::PerPerson.as_str() {
            Mode::PerPerson
        } else {
            Mode::Pool
        }
    }
}

/// The plan-derived monthly AI allowance for a tenant, read from the
/// subscriptions + plans tables.
#[derive(Debug, Clone, PartialEq)]
pub struct Entitlement {
    pub plan_slug: String,
    pub ai_credits_per_month: i32,
    pub seat_count: i32,
    pub status: String,
}

/// The computed view of a tenant/bucket's current-month AI usage.
#[derive(Debug, Clone, PartialEq)]
pub struct QuotaStatus {
    pub plan_slug: String,
    pub mode: Mode,
    pub allowance_tokens: i64,
    pub used_tokens: i64,
    pub remaining_tokens: i64,
    pub overage_tokens: i64,
    pub allow_overage: bool,
    pub blocked: bool,
    /// Go `time.Time` (serialized RFC3339): the first day of next month.
    pub period_end: chrono::DateTime<chrono::Utc>,
}

impl Default for QuotaStatus {
    fn default() -> Self {
        Self {
            plan_slug: String::new(),
            mode: Mode::Pool,
            allowance_tokens: 0,
            used_tokens: 0,
            remaining_tokens: 0,
            overage_tokens: 0,
            allow_overage: false,
            blocked: false,
            period_end: chrono::DateTime::<chrono::Utc>::from_timestamp(0, 0)
                .unwrap_or(chrono::DateTime::<chrono::Utc>::MIN_UTC),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::Mode;

    #[test]
    fn mode_normalize() {
        assert_eq!(Mode::parse("per_person"), Mode::PerPerson);
        assert_eq!(Mode::parse("pool"), Mode::Pool);
        assert_eq!(Mode::parse("unknown"), Mode::Pool);
        assert_eq!(Mode::Pool.as_str(), "pool");
    }
}
