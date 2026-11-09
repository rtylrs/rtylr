//! Integer-cents money, mirroring `internal/platform/money/money.go`.
//!
//! JSON uses the Go decimal shape (`1.20`, unquoted) and deserializes from a
//! quoted string, a bare number or `null` (zero).

use serde::{Deserialize, Deserializer, Serialize, Serializer};
use thiserror::Error;

#[derive(Debug, Error, PartialEq)]
#[error("invalid monetary amount")]
pub struct InvalidAmount;

pub type AmountResult = Result<Amount, InvalidAmount>;

/// Dollars from an external contract, rounded to the nearest cent.
pub fn dollars_to_cents(dollars: f64) -> i64 {
    (dollars * 100.0).round() as i64
}

/// External-contract dollars view of an internal cents amount.
pub fn cents_to_dollars(cents: i64) -> f64 {
    cents as f64 / 100.0
}

fn is_decimal_digits(value: &str) -> bool {
    !value.is_empty() && value.bytes().all(|byte| byte.is_ascii_digit())
}

/// Smallest currency unit (normally cents), as `Amount` in Go.
#[derive(Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord)]
pub struct Amount(pub i64);

impl Amount {
    pub fn zero() -> Self {
        Self(0)
    }

    pub fn is_zero(&self) -> bool {
        self.0 == 0
    }

    pub fn to_float(&self) -> f64 {
        cents_to_dollars(self.0)
    }

    pub fn from_string(value: &str) -> AmountResult {
        let text = value.trim();
        if text.is_empty() || text.contains(['e', 'E']) {
            return Err(InvalidAmount);
        }
        let (sign, unsigned) = if let Some(rest) = text.strip_prefix('-') {
            (-1_i64, rest)
        } else if let Some(rest) = text.strip_prefix('+') {
            (1_i64, rest)
        } else {
            (1_i64, text)
        };
        if unsigned.is_empty() {
            return Err(InvalidAmount);
        }
        let (whole, fraction, has_decimal) = match unsigned.split_once('.') {
            Some((whole, fraction)) => {
                if fraction.contains('.') || fraction.len() > 2 || fraction.is_empty() {
                    return Err(InvalidAmount);
                }
                (whole, fraction, true)
            }
            None => (unsigned, "", false),
        };
        let whole = if whole.is_empty() { "0" } else { whole };
        if !is_decimal_digits(whole) {
            return Err(InvalidAmount);
        }
        let fraction = if !has_decimal {
            "00"
        } else if fraction.len() == 1 {
            // Pad a single digit by appending '0' without allocating.
            return cents_from_parts(sign, whole, fraction, true);
        } else {
            fraction
        };
        if !is_decimal_digits(fraction) {
            return Err(InvalidAmount);
        }
        cents_from_parts(sign, whole, fraction, false)
    }

    pub fn from_float64(value: f64) -> AmountResult {
        if !value.is_finite() {
            return Err(InvalidAmount);
        }
        // Go formats with 'f', -1 precision, then parses the decimal.
        Self::from_string(&format!("{value:.15}").trim_end_matches('0').trim_end_matches('.'))
    }

    pub fn to_string(&self) -> String {
        let value = self.0;
        if value < 0 {
            format!("-{}", format_cents(-(value as i128) as i64))
        } else {
            format_cents(value)
        }
    }
}

fn cents_from_parts(sign: i64, whole: &str, fraction: &str, pad: bool) -> AmountResult {
    let whole_value: i64 = match whole.parse() {
        Ok(value) if value >= 0 => value,
        _ => return Err(InvalidAmount),
    };
    if whole_value > i64::MAX / 100 {
        return Err(InvalidAmount);
    }
    let mut fraction_text = fraction.to_string();
    if pad {
        fraction_text.push('0');
    } else if fraction_text.is_empty() {
        fraction_text = "00".to_string();
    }
    let fraction_value: i64 = match fraction_text.parse() {
        Ok(value) => value,
        Err(_) => return Err(InvalidAmount),
    };
    let cents = match whole_value.checked_mul(100).and_then(|v| v.checked_add(fraction_value)) {
        Some(value) => value,
        None => return Err(InvalidAmount),
    };
    if cents < 0 {
        return Err(InvalidAmount);
    }
    Ok(Amount(sign * cents))
}

fn format_cents(value: i64) -> String {
    format!("{}.{:02}", value / 100, value % 100)
}

impl Serialize for Amount {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: Serializer,
    {
        match serde_json::value::RawValue::from_string(self.to_string()) {
            Ok(raw) => raw.serialize(serializer),
            Err(_) => serializer.serialize_str("0"),
        }
    }
}

impl<'de> Deserialize<'de> for Amount {
    fn deserialize<D>(deserializer: D) -> Result<Self, D::Error>
    where
        D: Deserializer<'de>,
    {
        let value = serde_json::Value::deserialize(deserializer)?;
        match &value {
            serde_json::Value::Null => Ok(Amount::zero()),
            serde_json::Value::String(text) => Amount::from_string(text).map_err(serde::de::Error::custom),
            other => Amount::from_string(&other.to_string()).map_err(serde::de::Error::custom),
        }
    }
}

impl std::fmt::Display for Amount {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.write_str(&self.to_string())
    }
}

#[cfg(test)]
mod tests {
    use super::{dollars_to_cents, Amount};

    #[test]
    fn round_trips_amounts() {
        assert_eq!(Amount::from_string("1.20"), Ok(Amount(120)));
        assert_eq!(Amount::from_string("-0.05"), Ok(Amount(-5)));
        assert_eq!(Amount::from_string("10"), Ok(Amount(1000)));
        assert_eq!(Amount::from_string("0.5"), Ok(Amount(50)));
        assert_eq!(Amount::from_string(".07"), Ok(Amount(7)));
        assert_eq!(Amount(120).to_string(), "1.20");
        assert_eq!(Amount(-5).to_string(), "-0.05");
        assert_eq!(Amount(0).to_string(), "0.00");
    }

    #[test]
    fn rejects_bad_amounts() {
        for bad in ["", "1e3", "1.234", "10.", "a.b", "99999999999999999999.00"] {
            assert_eq!(
                Amount::from_string(bad),
                Err(super::InvalidAmount),
                "value={bad:?}"
            );
        }
    }

    #[test]
    fn json_matches_go_shape() {
        // Go MarshalJSON emits the bare decimal (e.g. 1.20, unquoted).
        assert_eq!(serde_json::to_string(&Amount(120)).unwrap(), "1.20");
        assert_eq!(
            serde_json::from_str::<Amount>("\"1.20\"").unwrap(),
            Amount(120)
        );
        assert_eq!(
            serde_json::from_str::<Amount>("1.20").unwrap(),
            Amount(120)
        );
        assert_eq!(serde_json::from_str::<Amount>("null").unwrap(), Amount(0));
    }

    #[test]
    fn conversions() {
        assert_eq!(dollars_to_cents(1.20), 120);
        assert_eq!(dollars_to_cents(0.005), 1);
        assert_eq!(super::cents_to_dollars(120), 1.2);
        assert!(Amount::from_float64(f64::NAN).is_err());
    }
}
