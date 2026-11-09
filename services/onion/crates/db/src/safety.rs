//! DSN safety guards, mirroring `internal/platform/database/dbclient/safety.go`.
//!
//! Fail-closed: unknown endpoints, or development-labelled runs pointing
//! off-machine, refuse to connect unless an operator opts in.

use crate::parse_dsn_metadata;
use thiserror::Error;

pub const KNOWN_PRODUCTION_HOSTS: [&str; 2] = [
    "voxire-db.rds.amazonaws.com",
    "voxire-db.c32oko2gqonb.eu-west-1.rds.amazonaws.com",
];

#[derive(Debug, Error, PartialEq)]
pub enum SafetyError {
    #[error("refusing to connect: could not safely parse {0} DSN; refusing an unknown database endpoint")]
    Unparseable(String),
    #[error(
        "refusing to connect: {0} host {1:?} is the production database, but this run has not declared that it means to use it (APP_ENV={2:?}). If this is a real production deploy, set APP_ENV=production. If this is another tier (staging, release candidate) that intentionally targets the production database, set ALLOW_PRODUCTION_DB=true. If this is a local/dev run, point DATABASE_URL/DB_HOST at a local database (see .env.example)"
    )]
    Production(String, String, String),
    #[error(
        "refusing to connect: APP_ENV={0:?} declares development, but {1} host {2:?} is not on this machine. If you meant to run against the local container, run `make local-env-up` and copy .env.example (DB_HOST=localhost, DB_NAME=onion_dev). If this tier is not development, set APP_ENV to the tier it is. If a development run is genuinely meant to reach this database, set ALLOW_REMOTE_DB=true"
    )]
    RemoteDevelopment(String, String, String),
}

fn host_without_port(host: &str) -> String {
    let trimmed = host.trim();
    if let Some(stripped) = trimmed.strip_prefix('[') {
        if let Some(end) = stripped.find(']') {
            return stripped[..end].to_string();
        }
    }
    match trimmed.rfind(':') {
        Some(colon) if trimmed[colon + 1..].chars().all(|c| c.is_ascii_digit()) => {
            trimmed[..colon].to_string()
        }
        _ => trimmed.to_string(),
    }
}

pub fn is_production_host(host: &str) -> bool {
    let normalized = host_without_port(host).trim().to_ascii_lowercase();
    KNOWN_PRODUCTION_HOSTS
        .iter()
        .any(|known| *known == normalized)
}

pub fn is_local_db_host(host: &str) -> bool {
    let normalized = host_without_port(host).trim().to_ascii_lowercase();
    if normalized.is_empty() {
        return false;
    }
    if normalized == "localhost" {
        return true;
    }
    if let Ok(ip) = normalized.parse::<std::net::IpAddr>() {
        return ip.is_loopback() || ip.is_unspecified();
    }
    !normalized.contains('.')
}

pub fn is_development_env(app_env_normalized: &str) -> bool {
    matches!(app_env_normalized, "development" | "dev")
}

fn parse_host(role: &str, dsn: &str) -> Result<String, SafetyError> {
    let metadata =
        parse_dsn_metadata(dsn).ok_or_else(|| SafetyError::Unparseable(role.to_string()))?;
    if metadata.host.trim().is_empty() {
        return Err(SafetyError::Unparseable(role.to_string()));
    }
    Ok(metadata.host)
}

pub fn check_production_safety(
    role: &str,
    dsn: &str,
    declared_production: bool,
    allow_production_db: bool,
) -> Result<(), SafetyError> {
    let host = parse_host(role, dsn)?;
    if !is_production_host(&host) {
        return Ok(());
    }
    if declared_production || allow_production_db {
        return Ok(());
    }
    Err(SafetyError::Production(
        role.to_string(),
        host,
        String::new(),
    ))
}

pub fn check_development_safety(
    role: &str,
    dsn: &str,
    app_env_normalized: &str,
    allow_remote_db: bool,
) -> Result<(), SafetyError> {
    if !is_development_env(app_env_normalized) {
        return Ok(());
    }
    let host = parse_host(role, dsn)?;
    if is_local_db_host(&host) || allow_remote_db {
        return Ok(());
    }
    Err(SafetyError::RemoteDevelopment(
        app_env_normalized.to_string(),
        role.to_string(),
        host,
    ))
}

pub fn check_dsn_safety(
    role: &str,
    dsn: &str,
    declared_production: bool,
    allow_production_db: bool,
    app_env_normalized: &str,
    allow_remote_db: bool,
) -> Result<(), SafetyError> {
    check_production_safety(role, dsn, declared_production, allow_production_db)?;
    check_development_safety(role, dsn, app_env_normalized, allow_remote_db)
}

#[cfg(test)]
mod tests {
    use super::{
        check_development_safety, check_production_safety, is_local_db_host, is_production_host,
        SafetyError,
    };

    const LOCAL: &str = "user:pw@tcp(localhost:3306)/onion_dev";

    #[test]
    fn production_hosts_exact() {
        assert!(is_production_host("voxire-db.rds.amazonaws.com"));
        assert!(is_production_host(
            "voxire-db.c32oko2gqonb.eu-west-1.rds.amazonaws.com:3306"
        ));
        assert!(!is_production_host("other.rds.amazonaws.com"));
        assert!(!is_production_host("voxire-db.rds.amazonaws.com.evil.com"));
    }

    #[test]
    fn local_hosts() {
        assert!(is_local_db_host("localhost"));
        assert!(is_local_db_host("127.0.0.1:3306"));
        assert!(is_local_db_host("db"));
        assert!(is_local_db_host("mysql"));
        assert!(!is_local_db_host("db.internal"));
        assert!(!is_local_db_host("10.0.0.5"));
        assert!(!is_local_db_host(""));
    }

    #[test]
    fn guards_ported() {
        let prod = "bolted:pw@tcp(voxire-db.rds.amazonaws.com:3306)/voxire?parseTime=true";
        assert_eq!(
            check_production_safety("database", prod, false, false).unwrap_err(),
            SafetyError::Production(
                "database".to_string(),
                "voxire-db.rds.amazonaws.com:3306".to_string(),
                String::new()
            )
        );
        assert!(check_production_safety("database", prod, true, false).is_ok());
        assert!(check_production_safety("database", prod, false, true).is_ok());
        assert!(check_production_safety("database", LOCAL, false, false).is_ok());
        assert!(check_development_safety("database", LOCAL, "development", false).is_ok());
        assert!(check_development_safety(
            "database",
            "user:pw@tcp(db.internal:3306)/db",
            "development",
            false
        )
        .is_err());
        assert!(check_development_safety(
            "database",
            "user:pw@tcp(db.internal:3306)/db",
            "development",
            true
        )
        .is_ok());
        assert!(check_development_safety("database", LOCAL, "", false).is_ok());
    }
}
