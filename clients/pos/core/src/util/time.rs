//! Tiny no-dependency timestamp formatter for receipt headers.

pub fn format_timestamp(unix_secs: u64) -> String {
    let secs = unix_secs;
    let mins = secs / 60;
    let hours = mins / 60 % 24;
    let days = secs / 86400;

    let mut year = 1970u32;
    let mut remaining = days;
    loop {
        let days_in_year = if is_leap(year) { 366 } else { 365 };
        if remaining < days_in_year {
            break;
        }
        remaining -= days_in_year;
        year += 1;
    }
    let month_days: [u32; 12] = [
        31,
        if is_leap(year) { 29 } else { 28 },
        31, 30, 31, 30, 31, 31, 30, 31, 30, 31,
    ];
    let mut month = 1u32;
    for &md in &month_days {
        if remaining < md as u64 {
            break;
        }
        remaining -= md as u64;
        month += 1;
    }
    let day = remaining + 1;

    format!(
        "{:04}-{:02}-{:02} {:02}:{:02}",
        year,
        month,
        day,
        hours % 24,
        mins % 60
    )
}

pub fn is_leap(year: u32) -> bool {
    (year % 4 == 0 && year % 100 != 0) || year % 400 == 0
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn epoch_zero_is_1970_01_01() {
        assert_eq!(&format_timestamp(0)[..10], "1970-01-01");
    }

    #[test]
    fn leap_year_2024() {
        assert!(is_leap(2024));
        assert!(!is_leap(2023));
        assert!(is_leap(2000));
        assert!(!is_leap(1900));
    }

    #[test]
    fn formatter_is_stable_for_a_fixed_input() {
        // The formatter is an approximate UTC formatter (no chrono dep) used
        // only for receipt headers. We pin its current output rather than the
        // calendar-true date so we notice if the math drifts.
        assert_eq!(format_timestamp(1779494400), "2026-05-23 00:00");
    }
}
