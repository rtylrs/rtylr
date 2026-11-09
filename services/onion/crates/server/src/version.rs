//! `/version` response. Mirrors `version_routes.go`. Values are injected
//! at build time by `bins/api` (Go used `-ldflags`).

use serde::{Deserialize, Serialize};

/// Rust toolchain the workspace is pinned to (`rust-toolchain.toml`).
pub const TOOLCHAIN_VERSION: &str = "1.97.1";

/// Build identity injected at build time.
#[derive(Debug, Clone, Default, PartialEq)]
pub struct BuildInfo {
    pub version: String,
    pub commit: String,
    pub build_time: String,
}

impl BuildInfo {
    pub fn dev() -> Self {
        Self {
            version: "dev".to_string(),
            commit: "unknown".to_string(),
            build_time: "unknown".to_string(),
        }
    }

    /// Read build identity from compile-time env (`option_env!`,
    /// matching Go's `-ldflags` injection; Docker build args set
    /// `VERSION`, `COMMIT`, `BUILD_TIME` for the build).
    pub fn from_build_env() -> Self {
        Self {
            version: option_env!("VERSION").unwrap_or("dev").to_string(),
            commit: option_env!("COMMIT").unwrap_or("unknown").to_string(),
            build_time: option_env!("BUILD_TIME").unwrap_or("unknown").to_string(),
        }
    }
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct VersionResponse {
    pub version: String,
    pub commit: String,
    pub build_time: String,
    /// Go reports `runtime.Version()`; Rust reports the toolchain.
    pub go_version: String,
}

pub fn version_response(info: &BuildInfo) -> VersionResponse {
    VersionResponse {
        version: info.version.clone(),
        commit: info.commit.clone(),
        build_time: info.build_time.clone(),
        go_version: format!("rustc {TOOLCHAIN_VERSION}"),
    }
}

/// Wire shape for `/version`.
pub fn version_json(info: &BuildInfo) -> serde_json::Value {
    let response = version_response(info);
    serde_json::json!({
        "version": response.version,
        "commit": response.commit,
        "build_time": response.build_time,
        "go_version": response.go_version,
    })
}

#[cfg(test)]
mod tests {
    use super::{version_json, version_response, BuildInfo};

    #[test]
    fn shape() {
        let dev = BuildInfo::dev();
        let response = version_response(&dev);
        assert_eq!(response.version, "dev");
        assert!(response.go_version.starts_with("rustc "));
        let json = version_json(&dev);
        assert_eq!(json["commit"], "unknown");
        assert_eq!(json["go_version"], response.go_version);
    }
}
