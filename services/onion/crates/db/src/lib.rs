//! Database foundation: sqlx MySQL pools, DSN and safety parity with
//! `internal/platform/database/dbclient`, tenant scoping parity with
//! `internal/platform/tenant`.
//!
//! DSN parsing rules mirror the Go driver boundary: credentials may contain
//! `@`, `?`, `/` or even `tcp(...)`, so host/db extraction anchors on the
//! last `@tcp(` of a native DSN instead of a naive delimiter search.

pub mod pool;
pub mod safety;
pub mod tenant;

/// Workspace crate version, surfaced in CI and version checks.
pub const CRATE_VERSION: &str = env!("CARGO_PKG_VERSION");

#[derive(Debug, Clone, PartialEq)]
pub struct DsnMetadata {
    pub host: String,
    pub db_name: String,
}

/// Extract `(host[:port], db_name)` from a native or `mysql://` DSN.
/// Returns `None` only when the DSN cannot be parsed.
pub fn parse_dsn_metadata(dsn: &str) -> Option<DsnMetadata> {
    let trimmed = dsn.trim();
    if trimmed.to_ascii_lowercase().starts_with("mysql://") {
        let rest = &trimmed["mysql://".len()..];
        let (authority, path) = match rest.split_once('/') {
            Some((authority, path)) => (authority, path),
            None => (rest, ""),
        };
        let host = authority.rsplit('@').next().unwrap_or(authority).trim();
        if host.is_empty() {
            return None;
        }
        let db_name = path.split('?').next().unwrap_or("").trim_matches('/');
        return Some(DsnMetadata {
            host: host.to_string(),
            db_name: db_name.to_string(),
        });
    }
    let tcp = trimmed.rfind("@tcp(")?;
    let after_tcp = &trimmed[tcp + "@tcp(".len()..];
    let close = after_tcp.find(')')?;
    let host = after_tcp[..close].trim();
    if host.is_empty() {
        return None;
    }
    let tail = &after_tcp[close + 1..];
    let db_name = tail
        .strip_prefix('/')
        .unwrap_or(tail)
        .split('?')
        .next()
        .unwrap_or("")
        .trim()
        .to_string();
    Some(DsnMetadata {
        host: host.to_string(),
        db_name,
    })
}

pub fn host_from_dsn(dsn: &str) -> String {
    parse_dsn_metadata(dsn).map(|m| m.host).unwrap_or_default()
}

pub fn db_name_from_dsn(dsn: &str) -> String {
    parse_dsn_metadata(dsn)
        .map(|m| m.db_name)
        .unwrap_or_default()
}

#[derive(Debug, Clone, PartialEq)]
pub struct DsnParts {
    pub host: String,
    pub db_name: String,
    pub user: String,
    pub password: String,
}

pub fn parse_dsn_parts(dsn: &str) -> Option<DsnParts> {
    let metadata = parse_dsn_metadata(dsn)?;
    let (user, password) = if dsn.trim().to_ascii_lowercase().starts_with("mysql://") {
        let rest = dsn.trim()["mysql://".len()..].to_string();
        let authority = rest.split('/').next().unwrap_or("");
        match authority.rsplit_once('@') {
            Some((credentials, _)) => match credentials.split_once(':') {
                Some((user, password)) => (user.to_string(), password.to_string()),
                None => (credentials.to_string(), String::new()),
            },
            None => (String::new(), String::new()),
        }
    } else {
        let before_tcp = dsn.rfind("@tcp(")?;
        let credentials = &dsn[..before_tcp];
        match credentials.split_once(':') {
            Some((user, password)) => (user.to_string(), password.to_string()),
            None => (credentials.to_string(), String::new()),
        }
    };
    Some(DsnParts {
        host: metadata.host,
        db_name: metadata.db_name,
        user,
        password,
    })
}

/// sqlx URL for the parsed DSN, pinning Go's session params.
pub fn to_sqlx_url(dsn: &str) -> Option<String> {
    let parts = parse_dsn_parts(dsn)?;
    let authority = if parts.user.is_empty() {
        parts.host.clone()
    } else if parts.password.is_empty() {
        format!("{}@{}", parts.user, parts.host)
    } else {
        format!("{}:{}@{}", parts.user, parts.password, parts.host)
    };
    Some(format!(
        "mysql://{authority}/{}?charset=utf8mb4&collation=utf8mb4_unicode_ci&timezone=UTC",
        parts.db_name
    ))
}

#[cfg(test)]
mod tests {
    use super::{db_name_from_dsn, host_from_dsn, parse_dsn_metadata};

    #[test]
    fn host_and_db_from_dsns() {
        let cases = [
            (
                "bolted:pw@tcp(voxire-db.rds.amazonaws.com:3306)/voxire?parseTime=true",
                "voxire-db.rds.amazonaws.com:3306",
                "voxire",
            ),
            (
                "user:pw@tcp(localhost:3306)/onion_dev",
                "localhost:3306",
                "onion_dev",
            ),
            (
                "mysql://bolted:pw@voxire-db.rds.amazonaws.com:3306/voxire?tls=true",
                "voxire-db.rds.amazonaws.com:3306",
                "voxire",
            ),
            ("mysql://host.internal:3306/db", "host.internal:3306", "db"),
            ("not-a-dsn", "", ""),
        ];
        for (dsn, want_host, want_db) in cases {
            assert_eq!(host_from_dsn(dsn), want_host, "dsn={dsn}");
            assert_eq!(db_name_from_dsn(dsn), want_db, "dsn={dsn}");
        }
    }

    #[test]
    fn credentials_punctuation_is_not_endpoint() {
        let cases = [
            "u:tcp(foo)@tcp(host.example:3306)/db",
            "u:p?ss@tcp(host.example:3306)/db?parseTime=true",
            "u:p@ss@tcp(host.example:3306)/db",
            "u:p/ss@tcp(host.example:3306)/db",
            "u:mysql://ss@tcp(host.example:3306)/db",
        ];
        for dsn in cases {
            assert_eq!(host_from_dsn(dsn), "host.example:3306", "dsn={dsn}");
            assert_eq!(db_name_from_dsn(dsn), "db", "dsn={dsn}");
        }
    }

    #[test]
    fn metadata_is_none_on_garbage() {
        assert_eq!(parse_dsn_metadata(""), None);
    }
}
