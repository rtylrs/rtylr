//! Authentication foundation: JWT, authorization, tenant, origins,
//! secure keys and email encryption, mirroring
//! `internal/platform/{auth,authorization,origins,securekey,emailcrypto}`
//! and `internal/platform/utils/jwtutils`.

pub mod authorization;
pub mod emailcrypto;
pub mod jwt;
pub mod origins;
pub mod platform;
pub mod roles;
pub mod securekey;
pub mod tenant;

/// Workspace crate version, surfaced in CI and version checks.
pub const CRATE_VERSION: &str = env!("CARGO_PKG_VERSION");

#[cfg(test)]
mod tests {
    use super::CRATE_VERSION;

    #[test]
    fn crate_version_is_nonempty() {
        assert!(!CRATE_VERSION.is_empty());
    }
}
