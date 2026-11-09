//! String/id utilities, mirroring `internal/platform/utils/*`
//! (`stringutils`, `times` constants, `uuid.NewString`).

/// True when `text` contains any of `patterns` (`ContainsOneOf`).
pub fn contains_one_of(text: &str, patterns: &[&str]) -> bool {
    patterns.iter().any(|pattern| text.contains(pattern))
}

/// A new `uuid` v4 string, as `uuid.NewString()` in Go.
pub fn new_id() -> String {
    uuid::Uuid::new_v4().to_string()
}

pub const FORMAT_WITH_DATE_TIME_ZONE: &str = "2006-01-02 15:04:05 -0700 MST";
pub const FORMAT_WITH_DATE_TIME: &str = "2006-01-02 15:04:05";
pub const FORMAT_WITH_DATE: &str = "2006-01-02";
pub const FORMAT_CARD_EXPIRY: &str = "01/2006";
pub const FORMAT_DDMMYYYY: &str = "01/02/2006";
pub const FORMAT_DDMMYYYYHHMMSS: &str = "02012006150405";
pub const FORMAT_FULL_DATE: &str = "January 02, 2006";

pub const MILLISECONDS_PER_SECOND: i64 = 1000;
pub const SECONDS_PER_MINUTE: i64 = 60;
pub const SECONDS_PER_DAY: i64 = 86400;
pub const MINUTES_PER_HOUR: i64 = 60;
pub const HOURS_PER_DAY: i64 = 24;

/// `time.Hour * 24` as seconds.
pub const DAY_SECS: i64 = HOURS_PER_DAY * 3600;

#[cfg(test)]
mod tests {
    use super::{contains_one_of, new_id};

    #[test]
    fn string_and_id_helpers() {
        assert!(contains_one_of("hello world", &["lo", "x"]));
        assert!(!contains_one_of("hello", &["x", "y"]));
        let id = new_id();
        assert_eq!(id.len(), 36);
        assert!(uuid::Uuid::parse_str(&id).is_ok());
    }
}
