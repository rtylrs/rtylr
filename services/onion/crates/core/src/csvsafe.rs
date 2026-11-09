//! CSV formula-injection guard, mirroring `internal/platform/csvsafe`.

/// Prefix-quote values a spreadsheet would evaluate as a formula.
pub fn cell(value: &str) -> String {
    match value.chars().next() {
        Some('=') | Some('+') | Some('-') | Some('@') | Some('\t') | Some('\r') => {
            format!("'{value}")
        }
        _ => value.to_string(),
    }
}

/// Apply [`cell`] to a row.
pub fn row(values: &[&str]) -> Vec<String> {
    values.iter().map(|value| cell(value)).collect()
}

#[cfg(test)]
mod tests {
    use super::{cell, row};

    #[test]
    fn guards_formulas() {
        assert_eq!(cell(""), "");
        assert_eq!(cell("hello"), "hello");
        assert_eq!(cell("=SUM(A1:A2)"), "'=SUM(A1:A2)");
        assert_eq!(cell("+123"), "'+123");
        assert_eq!(cell("-123"), "'-123");
        assert_eq!(cell("@handle"), "'@handle");
        assert_eq!(row(&["a", "=b"]), vec!["a", "'=b"]);
    }
}
