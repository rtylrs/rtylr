//! Normalize Go legacy route dumps for parity diffing.

use std::collections::BTreeSet;

#[derive(Debug, Clone, PartialEq, Eq, PartialOrd, Ord)]
pub struct Route {
    pub kind: String,
    pub method: String,
    pub pattern: String,
}

#[derive(Debug, Clone, PartialEq)]
pub struct DumpError {
    pub line: usize,
    pub message: String,
}

impl std::fmt::Display for DumpError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "line {}: {}", self.line, self.message)
    }
}

impl std::error::Error for DumpError {}

/// Parse the `kind<TAB>METHOD /pattern` dump written by `ROUTE_DUMP`
/// (`cmd/api/route_dump_test.go`).
pub fn parse_dump(text: &str) -> Result<BTreeSet<Route>, DumpError> {
    let mut routes = BTreeSet::new();
    for (index, line) in text.lines().enumerate() {
        let line_number = index + 1;
        if line.trim().is_empty() {
            continue;
        }
        let (kind, rest) = line.split_once('\t').ok_or(DumpError {
            line: line_number,
            message: "expected kind<TAB>METHOD /pattern".to_string(),
        })?;
        let (method, pattern) = rest.split_once(' ').ok_or(DumpError {
            line: line_number,
            message: "expected METHOD /pattern".to_string(),
        })?;
        if kind.is_empty() || method.is_empty() || !pattern.starts_with('/') {
            return Err(DumpError {
                line: line_number,
                message: "route fields must be non-empty with a /-pattern".to_string(),
            });
        }
        routes.insert(Route {
            kind: kind.to_string(),
            method: method.to_string(),
            pattern: pattern.to_string(),
        });
    }
    Ok(routes)
}

/// Canonical table: `kind<TAB>METHOD /pattern` sorted by (method, pattern,
/// kind), deduplicated.
pub fn render_table(routes: &BTreeSet<Route>) -> String {
    let mut rows: Vec<&Route> = routes.iter().collect();
    rows.sort_by(|a, b| {
        a.method
            .cmp(&b.method)
            .then_with(|| a.pattern.cmp(&b.pattern))
            .then_with(|| a.kind.cmp(&b.kind))
    });
    let mut out = String::new();
    for route in rows {
        out.push_str(&route.kind);
        out.push('\t');
        out.push_str(&route.method);
        out.push(' ');
        out.push_str(&route.pattern);
        out.push('\n');
    }
    out
}

/// `(routes only in `a`, routes only in `b`)` for a parity diff.
pub fn diff(a: &BTreeSet<Route>, b: &BTreeSet<Route>) -> (Vec<Route>, Vec<Route>) {
    let only_in_a: Vec<Route> = a.difference(b).cloned().collect();
    let only_in_b: Vec<Route> = b.difference(a).cloned().collect();
    (only_in_a, only_in_b)
}

#[cfg(test)]
mod tests {
    use super::{diff, parse_dump, render_table};

    #[test]
    fn normalizes_and_sorts() {
        let text = "raw\tGET /zebra\njson\tPOST /b\njson\tGET /a\njson\tPOST /b\n";
        let routes = parse_dump(text).unwrap();
        assert_eq!(routes.len(), 3);
        assert_eq!(
            render_table(&routes),
            "json\tGET /a\nraw\tGET /zebra\njson\tPOST /b\n"
        );
    }

    #[test]
    fn rejects_bad_line() {
        let err = parse_dump("json\tGET no-slash\n").unwrap_err();
        assert_eq!(err.line, 1);
    }

    #[test]
    fn diff_reports_only_rows() {
        let a = parse_dump("json\tGET /a\njson\tGET /b\n").unwrap();
        let b = parse_dump("json\tGET /a\njson\tGET /c\n").unwrap();
        let (only_a, only_b) = diff(&a, &b);
        assert_eq!(only_a.len(), 1);
        assert_eq!(only_a[0].pattern, "/b");
        assert_eq!(only_b.len(), 1);
        assert_eq!(only_b[0].pattern, "/c");
    }
}
