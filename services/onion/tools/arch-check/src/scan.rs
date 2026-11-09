//! Rust-source scanner: a line-oriented port of `architecture_test.go`'s
//! rules. It intentionally does the same substring/edge checks instead
//! of reaching for a tree-sitter dependency.
//!
//! Ownership mirrors Go: the first directory under `crates/modules/` is
//! the owner; crates under `crates/` top level (plus `crates/contracts`,
//! `crates/media-engine`) are platform. Module edges come from real Cargo
//! manifests (`[dependencies]` path deps between module crates) plus
//! `use` tokens, so `module-cycle` can fire (DFS over the dep graph).

use std::collections::{HashMap, HashSet};

#[derive(Debug, Clone, PartialEq, Eq, Hash)]
pub struct Finding {
    pub rule: &'static str,
    pub file: String,
    pub detail: String,
}

impl Finding {
    pub fn new(rule: &'static str, file: &str, detail: String) -> Self {
        Self {
            rule,
            file: file.to_string(),
            detail,
        }
    }
}

/// Crate directory (`crates/<name>` or `crates/modules/<m>`) → set of
/// crate directories it depends on. Filled from Cargo manifests and `use`
/// lines by `main.rs`.
pub type EdgeGraph = HashMap<String, HashSet<String>>;

pub fn find_cycles(edges: &EdgeGraph) -> Vec<Finding> {
    let mut sorted: Vec<String> = edges.keys().cloned().collect();
    sorted.sort();
    let mut state: HashMap<String, u8> = HashMap::new();
    let mut trail: Vec<String> = Vec::new();
    let mut findings = Vec::new();
    for owner in &sorted {
        visit(owner, edges, &mut state, &mut trail, &mut findings);
    }
    findings
}

fn visit(
    owner: &str,
    edges: &EdgeGraph,
    state: &mut HashMap<String, u8>,
    trail: &mut Vec<String>,
    findings: &mut Vec<Finding>,
) {
    match state.get(owner).copied() {
        Some(1) => {
            if let Some(pos) = trail.iter().position(|member| member == owner) {
                let mut cycle: Vec<String> = trail[pos..].to_vec();
                cycle.push(owner.to_string());
                findings.push(Finding::new(
                    "module-cycle",
                    "crates/modules",
                    cycle.join(" -> "),
                ));
            }
            return;
        }
        Some(2) => return,
        _ => {}
    }
    state.insert(owner.to_string(), 1);
    trail.push(owner.to_string());
    if let Some(targets) = edges.get(owner) {
        let mut iter: Vec<String> = targets.iter().cloned().collect();
        iter.sort();
        for target in iter {
            visit(&target, edges, state, trail, findings);
        }
    }
    trail.pop();
    state.insert(owner.to_string(), 2);
}

#[derive(Debug, Clone, PartialEq)]
pub struct LayeredSource {
    pub path: String,
    pub source: String,
}

pub struct ScanOutput {
    pub findings: Vec<Finding>,
}

fn is_platform_crate(name: &str) -> bool {
    matches!(
        name,
        "core"
            | "contracts"
            | "db"
            | "auth"
            | "events"
            | "realtime"
            | "integrations"
            | "ai"
            | "server"
            | "media-engine"
    )
}

pub fn crate_dir_of(path: &str) -> Option<String> {
    let rest = path.strip_prefix("crates/")?;
    if let Some(modules) = rest.strip_prefix("modules/") {
        let module = modules.split('/').next()?;
        Some(format!("crates/modules/{module}"))
    } else {
        let name = rest.split('/').next()?;
        Some(format!("crates/{name}"))
    }
}

fn crate_owner(crate_dir: &str) -> String {
    if let Some(module) = crate_dir.strip_prefix("crates/modules/") {
        module.to_string()
    } else {
        crate_dir
            .strip_prefix("crates/")
            .unwrap_or(crate_dir)
            .to_string()
    }
}

struct FileInfo {
    path: String,
    owner: String,
    layer: Option<String>,
    module_crate: bool,
    source: String,
}

fn file_info(path: &str, source: String) -> FileInfo {
    let crate_dir = crate_dir_of(path).unwrap_or_else(|| "crates".to_string());
    let module_crate = crate_dir.starts_with("crates/modules/");
    let layer = if module_crate || crate_dir == "crates/contracts" {
        if path.ends_with("/api.rs") || path.contains("/api/") {
            Some("api".to_string())
        } else {
            path.split('/')
                .find(|part| {
                    ["domain", "application", "infrastructure", "jobs", "queries"].contains(part)
                })
                .map(str::to_string)
        }
    } else {
        None
    };
    FileInfo {
        path: path.to_string(),
        owner: crate_owner(&crate_dir),
        layer,
        module_crate,
        source,
    }
}

/// Scan in-memory fixtures (one `SourceFile` per path). `dep_edges` are
/// Cargo-derived module edges used for `module-cycle` detection.
pub fn scan_sources(sources: Vec<LayeredSource>, dep_edges: &EdgeGraph) -> ScanOutput {
    let infos: Vec<FileInfo> = sources
        .into_iter()
        .map(|source| file_info(&source.path, source.source))
        .collect();
    scan_infos(&infos, dep_edges)
}

fn scan_infos(infos: &[FileInfo], dep_edges: &EdgeGraph) -> ScanOutput {
    let mut findings = Vec::new();

    for info in infos {
        if crate_dir_of(&info.path).is_none() {
            continue;
        }
        let is_platform = is_platform_crate(&info.owner) && !info.module_crate;

        // platform-import: a platform crate depending on a module crate.
        if is_platform {
            for line in info.source.lines() {
                if line.contains("modules::") || line.contains("::infrastructure") {
                    findings.push(Finding::new(
                        "platform-import",
                        &info.path,
                        line.trim().to_string(),
                    ));
                }
            }
        }

        // persistence-tag / invalid-tag / transport-tag: attribute tags on
        // protected (domain or contracts) types. Go gates on `/domain/` or
        // `internal/contracts/`; our equivalents are module domain layers
        // and the contracts crate.
        let protected = info.layer.as_deref() == Some("domain") || info.owner == "contracts";
        if protected {
            for line in info.source.lines() {
                let trimmed = line.trim();
                if trimmed.starts_with("#[sqlx(") {
                    findings.push(Finding::new(
                        "persistence-tag",
                        &info.path,
                        "persistence attribute in protected surface".to_string(),
                    ));
                }
                if trimmed.starts_with("#[serde(") && !trimmed.ends_with(")]") {
                    findings.push(Finding::new("invalid-tag", &info.path, trimmed.to_string()));
                }
            }
            // domain-import: framework crates inside domain.
            for forbidden in [
                "sqlx::",
                "axum::",
                "serde::",
                "serde_json::",
                "jsonwebtoken::",
                "reqwest::",
            ] {
                if info.source.contains(forbidden) {
                    findings.push(Finding::new(
                        "domain-import",
                        &info.path,
                        forbidden.trim_end_matches("::").to_string(),
                    ));
                }
            }
            // transport-tag: serde/sqlx derives on domain types.
            if info.source.contains("#[derive(")
                && (info.source.contains("Serialize")
                    || info.source.contains("Deserialize")
                    || info.source.contains("FromRow"))
            {
                findings.push(Finding::new(
                    "transport-tag",
                    &info.path,
                    "wire derive in domain".to_string(),
                ));
            }
        }

        if info.module_crate {
            match info.layer.as_deref() {
                Some("application") => {
                    // application-adapter-import: application reaching for
                    // concrete adapters.
                    for needle in ["::infrastructure::", "::queries::", "api::"] {
                        if info.source.contains(needle) {
                            findings.push(Finding::new(
                                "application-adapter-import",
                                &info.path,
                                needle.to_string(),
                            ));
                        }
                    }
                }
                Some("infrastructure") | Some("queries") => {
                    // foreign-persistence-import: cross-module `use` of
                    // persistence surfaces.
                    for line in info.source.lines() {
                        if line.contains("use ")
                            && (line.contains("::infrastructure") || line.contains("::queries"))
                        {
                            if let Some(use_pos) = line.find("use ") {
                                let rest = &line[use_pos + 4..];
                                let first = rest.split("::").next().unwrap_or("").trim();
                                if first != "crate" && !rest.starts_with("crate::") {
                                    findings.push(Finding::new(
                                        "foreign-persistence-import",
                                        &info.path,
                                        line.trim().to_string(),
                                    ));
                                }
                            }
                        }
                    }
                }
                Some("api") => {
                    // generated-public-api: pub items in the api surface
                    // naming queries/persistence generated types.
                    for line in info.source.lines() {
                        let trimmed = line.trim();
                        let is_pub = trimmed.starts_with("pub ") || trimmed.starts_with("pub fn ");
                        if is_pub
                            && (line.contains("queries::")
                                || line.contains("sqlc::")
                                || line.contains("SqlxPool"))
                        {
                            findings.push(Finding::new(
                                "generated-public-api",
                                &info.path,
                                trimmed.to_string(),
                            ));
                        }
                    }
                }
                _ => {}
            }

            // read-authority: aiquery and read paths must not reference
            // write surfaces.
            if (info.owner == "aiquery" || info.path.contains("/read/"))
                && (info.source.contains("::write::") || info.source.contains("writesql"))
            {
                findings.push(Finding::new(
                    "read-authority",
                    &info.path,
                    "write surface".to_string(),
                ));
            }

            // cross-module-transaction: a public fn that takes a transaction
            // and hands it to a different owner's pub fn in the same file.
            for line in info.source.lines() {
                if line.contains("Transaction")
                    && (line.starts_with("pub fn") || line.contains("pub async fn"))
                {
                    let is_foreign_call = info.source.lines().any(|other| {
                        other.contains("::apply(")
                            || other.contains("::save(")
                            || other.contains("::persist(")
                    });
                    if is_foreign_call {
                        findings.push(Finding::new(
                            "cross-module-transaction",
                            &info.path,
                            line.trim().to_string(),
                        ));
                    }
                }
            }
        }
    }

    findings.extend(find_cycles(dep_edges));

    ScanOutput { findings }
}

#[cfg(test)]
mod tests {
    use super::{scan_sources, EdgeGraph, LayeredSource};
    use std::collections::{HashMap, HashSet};

    fn source(path: &str, source: &str) -> LayeredSource {
        LayeredSource {
            path: path.to_string(),
            source: source.to_string(),
        }
    }

    fn has_rule(output: &super::ScanOutput, rule: &str) -> bool {
        output.findings.iter().any(|finding| finding.rule == rule)
    }

    fn empty_edges() -> EdgeGraph {
        HashMap::new()
    }

    #[test]
    fn foreign_persistence_import_detected() {
        let out = scan_sources(
            vec![source(
                "crates/modules/consumer/src/infrastructure/repo.rs",
                "use other::infrastructure::types;",
            )],
            &empty_edges(),
        );
        assert!(has_rule(&out, "foreign-persistence-import"));
    }

    #[test]
    fn generated_public_api_detected() {
        let out = scan_sources(
            vec![source(
                "crates/modules/consumer/src/api.rs",
                "pub use other::queries::Row;",
            )],
            &empty_edges(),
        );
        assert!(has_rule(&out, "generated-public-api"));
    }

    #[test]
    fn private_use_is_clean() {
        let out = scan_sources(
            vec![source(
                "crates/modules/consumer/src/api.rs",
                "pub(crate) use crate::queries::Row;",
            )],
            &empty_edges(),
        );
        assert!(out.findings.is_empty(), "{:?}", out.findings);
    }

    #[test]
    fn transport_tag_detected() {
        let out = scan_sources(
            vec![source(
                "crates/modules/consumer/src/domain/order.rs",
                "#[derive(Serialize)]\npub struct Order { id: i64 }",
            )],
            &empty_edges(),
        );
        assert!(has_rule(&out, "transport-tag"));
    }

    #[test]
    fn domain_import_detected() {
        let out = scan_sources(
            vec![source(
                "crates/modules/consumer/src/domain/order.rs",
                "use sqlx::FromRow;\npub struct Order { id: i64 }",
            )],
            &empty_edges(),
        );
        assert!(has_rule(&out, "domain-import"));
    }

    #[test]
    fn domain_pure_is_clean() {
        let out = scan_sources(
            vec![source(
                "crates/modules/consumer/src/domain/order.rs",
                "pub struct Order { id: i64 }",
            )],
            &empty_edges(),
        );
        assert!(out.findings.is_empty(), "{:?}", out.findings);
    }

    #[test]
    fn application_adapter_detected() {
        let out = scan_sources(
            vec![source(
                "crates/modules/consumer/src/application/service.rs",
                "use crate::infrastructure::Repo;",
            )],
            &empty_edges(),
        );
        assert!(has_rule(&out, "application-adapter-import"));
    }

    #[test]
    fn application_ports_clean() {
        let out = scan_sources(
            vec![source(
                "crates/modules/consumer/src/application/service.rs",
                "pub(crate) trait Repo { fn get(&self) {} }",
            )],
            &empty_edges(),
        );
        assert!(out.findings.is_empty(), "{:?}", out.findings);
    }

    #[test]
    fn persistence_and_invalid_tags() {
        let out = scan_sources(
            vec![source(
                "crates/modules/consumer/src/domain/order.rs",
                "#[sqlx(table_name = \"orders\")]\npub struct Order { id: i64 }",
            )],
            &empty_edges(),
        );
        assert!(has_rule(&out, "persistence-tag"));
        let out = scan_sources(
            vec![source(
                "crates/modules/consumer/src/domain/order.rs",
                "#[serde(rename = \"id\"\npub struct Order { id: i64 }",
            )],
            &empty_edges(),
        );
        assert!(has_rule(&out, "invalid-tag"));
        let clean = scan_sources(
            vec![source(
                "crates/modules/consumer/src/domain/order.rs",
                "pub struct Order { id: i64 }",
            )],
            &empty_edges(),
        );
        assert!(clean.findings.is_empty(), "{:?}", clean.findings);
    }

    #[test]
    fn platform_import_detected() {
        let out = scan_sources(
            vec![source(
                "crates/db/src/lib.rs",
                "use modules::consumer::infrastructure::types;",
            )],
            &empty_edges(),
        );
        assert!(has_rule(&out, "platform-import"));
    }

    #[test]
    fn module_cycle_detected() {
        let mut edges: EdgeGraph = HashMap::new();
        edges.insert(
            "crates/modules/alpha".to_string(),
            HashSet::from(["crates/modules/beta".to_string()]),
        );
        edges.insert(
            "crates/modules/beta".to_string(),
            HashSet::from(["crates/modules/alpha".to_string()]),
        );
        let out = scan_sources(
            vec![source("crates/modules/alpha/src/lib.rs", "pub fn x() {}")],
            &edges,
        );
        let cycles: Vec<_> = out
            .findings
            .iter()
            .filter(|finding| finding.rule == "module-cycle")
            .collect();
        assert_eq!(cycles.len(), 1, "{:?}", out.findings);
        assert!(cycles[0].detail.contains("alpha") && cycles[0].detail.contains("beta"));
    }

    #[test]
    fn cross_module_transaction_pub_fn() {
        let out = scan_sources(
            vec![source(
                "crates/modules/consumer/src/infrastructure/repo.rs",
                "pub async fn apply(tx: sqlx::Transaction<'static, sqlx::MySql>, x: i64) { other::apply(tx, x) }",
            )],
            &empty_edges(),
        );
        assert!(has_rule(&out, "cross-module-transaction"));
    }
}
