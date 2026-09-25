//! Money / amount validation.
//!
//! POS dollars are tracked as `f64` on the wire today. These validators
//! catch the worst f64 hazards (NaN, infinity, negative amounts where
//! they're not allowed, absurdly large amounts that signal a bug upstream).
//!
//! Anywhere this module is used you should consider whether the upstream
//! type could be a fixed-point cents integer instead - but for receipt-
//! printing inputs that come from JS, f64 is what we get.

pub const MAX_RECEIPT_AMOUNT: f64 = 1_000_000.0;

/// A non-negative, finite, reasonably-bounded money amount.
pub fn validate_amount(label: &str, value: f64) -> Result<f64, String> {
    if value.is_nan() {
        return Err(format!("{} is NaN", label));
    }
    if !value.is_finite() {
        return Err(format!("{} is not finite", label));
    }
    if value < 0.0 {
        return Err(format!("{} cannot be negative ({})", label, value));
    }
    if value > MAX_RECEIPT_AMOUNT {
        return Err(format!(
            "{} ({}) exceeds the per-receipt maximum of {}",
            label, value, MAX_RECEIPT_AMOUNT
        ));
    }
    Ok(value)
}

/// A non-negative quantity. Allows fractional weight-priced items.
pub fn validate_quantity(value: f64) -> Result<f64, String> {
    if value.is_nan() || !value.is_finite() {
        return Err("Quantity must be a finite number".to_string());
    }
    if value <= 0.0 {
        return Err(format!("Quantity must be positive (got {})", value));
    }
    if value > 100_000.0 {
        return Err(format!("Quantity {} is unreasonably large", value));
    }
    Ok(value)
}

/// Returns true when `a` and `b` agree to within one cent. Cheap and
/// good enough for receipt-level reconciliation; do not use for financial
/// ledger math.
pub fn equal_within_cents(a: f64, b: f64) -> bool {
    (a - b).abs() < 0.005
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn rejects_nan() {
        assert!(validate_amount("subtotal", f64::NAN).is_err());
    }

    #[test]
    fn rejects_infinity() {
        assert!(validate_amount("subtotal", f64::INFINITY).is_err());
        assert!(validate_amount("subtotal", f64::NEG_INFINITY).is_err());
    }

    #[test]
    fn rejects_negative() {
        assert!(validate_amount("subtotal", -0.01).is_err());
    }

    #[test]
    fn accepts_zero_and_typical() {
        assert!(validate_amount("subtotal", 0.0).is_ok());
        assert!(validate_amount("subtotal", 19.99).is_ok());
    }

    #[test]
    fn rejects_absurd_amount() {
        assert!(validate_amount("subtotal", 999_999_999.0).is_err());
    }

    #[test]
    fn quantity_must_be_positive() {
        assert!(validate_quantity(0.0).is_err());
        assert!(validate_quantity(-1.0).is_err());
        assert!(validate_quantity(1.5).is_ok());
    }

    #[test]
    fn cents_equality() {
        assert!(equal_within_cents(9.03, 9.03));
        assert!(equal_within_cents(9.030001, 9.03));
        assert!(!equal_within_cents(9.03, 9.04));
    }
}
