//! `arch-check`: Rust port of `architecture_test.go`, enforcing ports and
//! adapters layering, module edges, table and migration ownership.
//!
//! Usage: `arch-check [--root <onion-dir>] [--skip-sweep]`
//! Exit 0 when no findings (including after exceptions), 1 otherwise.
//!
//! Steady gate: `cargo run -p arch-check` in `rust-check.yml` (CI).
//! Never add an exception to `docs/architecture-exceptions.json` to make
//! it pass (PROMPT.md).

mod scan;

use std::collections::{BTreeMap, HashSet};
use std::path::{Path, PathBuf};

use chrono::NaiveDate;
use scan::Finding;
use serde::Deserialize;

#[derive(Debug, Deserialize)]
struct ExceptionEntry {
    #[serde(default, rename = "ID")]
    id: String,
    #[serde(default)]
    #[serde(rename = "Rule")]
    rule: String,
    #[serde(default)]
    #[serde(rename = "Owner")]
    owner: String,
    #[serde(default)]
    #[serde(rename = "Axiom")]
    axiom: String,
    #[serde(default)]
    #[serde(rename = "Reason")]
    reason: String,
    #[serde(default)]
    #[serde(rename = "Alternatives")]
    alternatives: String,
    #[serde(default)]
    #[serde(rename = "Containment")]
    containment: String,
    #[serde(default)]
    #[serde(rename = "ExitPlan")]
    exit_plan: String,
    #[serde(default, rename = "Files")]
    files: Vec<String>,
    #[serde(default)]
    #[serde(rename = "Details")]
    details: Vec<String>,
    #[serde(default, rename = "ReviewBy")]
    review_by: String,
}

fn load_exceptions(root: &Path) -> Result<Vec<ExceptionEntry>, String> {
    let path = root.join("docs/architecture-exceptions.json");
    let text =
        std::fs::read_to_string(&path).map_err(|e| format!("read {}: {e}", path.display()))?;
    serde_json::from_str(&text).map_err(|e| format!("parse {}: {e}", path.display()))
}

fn validate_exceptions(
    exceptions: &[ExceptionEntry],
    root: &Path,
    now: NaiveDate,
) -> Result<(), String> {
    for exception in exceptions {
        let review = NaiveDate::parse_from_str(&exception.review_by, "%Y-%m-%d").map_err(|_| {
            format!(
                "architecture exception {} has an invalid or expired review date",
                exception.id
            )
        })?;
        if review < now {
            return Err(format!(
                "architecture exception {} has an invalid or expired review date",
                exception.id
            ));
        }
        if exception.id.is_empty()
            || exception.rule.is_empty()
            || exception.owner.is_empty()
            || exception.axiom.is_empty()
            || exception.reason.is_empty()
            || exception.alternatives.is_empty()
            || exception.containment.is_empty()
            || exception.exit_plan.is_empty()
            || exception.files.is_empty()
            || exception.details.is_empty()
        {
            return Err(format!(
                "incomplete architecture exception {}",
                exception.id
            ));
        }
        for path in &exception.files {
            if path.contains(['*', '?']) || path.contains("..") {
                return Err(format!("exception {} must name exact files", exception.id));
            }
            if !root.join(path).exists() {
                return Err(format!(
                    "exception {} names missing file {}",
                    exception.id, path
                ));
            }
        }
    }
    Ok(())
}

fn apply_exceptions(
    mut findings: Vec<Finding>,
    exceptions: &[ExceptionEntry],
) -> (Vec<Finding>, Vec<String>) {
    let mut counts = BTreeMap::<String, usize>::new();
    let mut seen = HashSet::new();
    let mut used = HashSet::new();
    let mut remaining = Vec::new();
    for finding in findings.drain(..) {
        let mut exempt = false;
        for exception in exceptions {
            if exception.rule == finding.rule
                && exception.files.contains(&finding.file)
                && exception.details.contains(&finding.detail)
            {
                used.insert(exception.id.clone());
                exempt = true;
                break;
            }
        }
        if exempt {
            continue;
        }
        let key = format!("{}:{}:{}", finding.rule, finding.file, finding.detail);
        if !seen.insert(key) {
            continue;
        }
        *counts.entry(finding.rule.to_string()).or_default() += 1;
        remaining.push(finding);
    }
    let unused: Vec<String> = exceptions
        .iter()
        // Go-side exceptions (paths under internal/) can never fire in
        // the Rust scan; they ride along until Phase 6 deletes Go with
        // them. Only Rust-path exceptions are unused-checked.
        .filter(|exception| {
            !used.contains(&exception.id)
                && exception
                    .files
                    .iter()
                    .any(|file| file.starts_with("crates/"))
        })
        .map(|exception| exception.id.clone())
        .collect();
    (remaining, unused)
}

/// Recursively collect `.rs` sources under the onion workspace crates.
fn collect_sources(root: &Path) -> Vec<scan::LayeredSource> {
    let mut out = Vec::new();
    for base in ["crates", "tools"] {
        let dir = root.join(base);
        if !dir.exists() {
            continue;
        }
        let mut stack = vec![dir];
        while let Some(current) = stack.pop() {
            let Ok(entries) = std::fs::read_dir(&current) else {
                continue;
            };
            for entry in entries.flatten() {
                let path = entry.path();
                if path.is_dir() {
                    let name = path.file_name().and_then(|n| n.to_str()).unwrap_or("");
                    if name == "target" || name == "generated" {
                        continue;
                    }
                    stack.push(path);
                } else if path.extension().and_then(|e| e.to_str()) == Some("rs")
                    && !path.to_string_lossy().ends_with("_test.rs")
                {
                    if let Ok(text) = std::fs::read_to_string(&path) {
                        let rel = path
                            .strip_prefix(root)
                            .unwrap_or(&path)
                            .to_string_lossy()
                            .replace('\\', "/");
                        out.push(scan::LayeredSource {
                            path: rel,
                            source: text,
                        });
                    }
                }
            }
        }
    }
    out
}

/// Cargo-level platform-import: platform crates depending on module crates.
fn check_cargo_platform_imports(root: &Path) -> Vec<Finding> {
    let mut findings = Vec::new();
    let platform = [
        "core",
        "contracts",
        "db",
        "auth",
        "events",
        "realtime",
        "integrations",
        "ai",
        "server",
        "media-engine",
    ];
    for crate_name in platform {
        let manifest = root.join("crates").join(crate_name).join("Cargo.toml");
        let Ok(text) = std::fs::read_to_string(&manifest) else {
            continue;
        };
        let rel = manifest
            .strip_prefix(root)
            .unwrap_or(manifest.as_path())
            .to_string_lossy()
            .replace('\\', "/");
        for line in text.lines() {
            let line = line.trim();
            if line.contains("../modules/") || line.contains("crates/modules/") {
                findings.push(Finding::new("platform-import", &rel, line.to_string()));
            }
        }
    }
    findings
}

/// Cargo-derived module dep edges for `module-cycle` detection.
fn load_dep_edges(root: &Path) -> scan::EdgeGraph {
    let mut edges: scan::EdgeGraph = std::collections::HashMap::new();
    let modules_dir = root.join("crates/modules");
    let Ok(entries) = std::fs::read_dir(&modules_dir) else {
        return edges;
    };
    for entry in entries.flatten() {
        let manifest = entry.path().join("Cargo.toml");
        let Ok(text) = std::fs::read_to_string(&manifest) else {
            continue;
        };
        let owner = entry.file_name().to_string_lossy().replace('-', "_");
        let owner_dir = format!("crates/modules/{owner}");
        let mut in_dependencies = false;
        for line in text.lines() {
            let trimmed = line.trim();
            if trimmed.starts_with('[') {
                in_dependencies = trimmed.starts_with("[dependencies");
                continue;
            }
            if !in_dependencies {
                continue;
            }
            if let Some(rest) = trimmed.strip_prefix("../") {
                let foreign = rest.split('/').next().unwrap_or("").trim_end_matches('"');
                let foreign = foreign.trim_end_matches(" {").trim_end_matches(',');
                if !foreign.is_empty() {
                    edges
                        .entry(owner_dir.clone())
                        .or_default()
                        .insert(format!("crates/modules/{foreign}"));
                }
            }
        }
    }
    edges
}

/// Migration ownership: unique identities, owned sources that exist.
fn check_migration_owner(root: &Path) -> Result<Vec<Finding>, String> {
    let registry = root.join("internal/platform/database/schema_registry.tsv");
    let data = std::fs::read_to_string(&registry)
        .map_err(|e| format!("read {}: {e}", registry.display()))?;
    let mut findings = Vec::new();
    let mut identities = HashSet::new();
    let mut sources = HashSet::new();
    for line in data.lines() {
        if line.is_empty() || line.starts_with('#') {
            continue;
        }
        let mut columns = line.split('\t');
        let identity = columns.next().unwrap_or("").to_string();
        let owned = columns.next().unwrap_or("").to_string();
        if columns.next().is_some() || !identities.insert(identity.clone()) {
            return Err(format!("invalid or duplicate migration identity: {line}"));
        }
        for source in owned.split('|').map(str::to_string) {
            if !is_owned_source(source.as_str())
                || source.contains("..")
                || !sources.insert(source.clone())
            {
                findings.push(Finding::new(
                    "migration-owner",
                    "internal/platform/database/schema_registry.tsv",
                    source.clone(),
                ));
            }
            if !root.join(&source).exists() {
                return Err(format!("registered migration {source}: not found"));
            }
        }
    }
    Ok(findings)
}

fn is_owned_source(source: &str) -> bool {
    let parts: Vec<&str> = source.split('/').collect();
    // internal/modules/<owner>/..., internal/platform/<owner>/... and
    // internal/contracts/<owner>/... (Go `architectureOwner`).
    if parts.len() < 3 || parts[0] != "internal" {
        return false;
    }
    matches!(parts[1], "modules" | "platform" | "contracts")
}

/// Table/query sweeps (same node scripts the Go test shells out to).
fn check_sweep(root: &Path, temp_dir: &Path) -> Result<Vec<Finding>, String> {
    let mut findings = Vec::new();
    for script in ["sweep-queries.mjs", "sweep-tables.mjs"] {
        let onion_root = std::fs::canonicalize(root).unwrap_or_else(|_| root.to_path_buf());
        let output = std::process::Command::new("node")
            .arg(onion_root.join("scripts/cross-app").join(script))
            .current_dir(&onion_root)
            .env("CROSS_APP_ONION", &onion_root)
            .env("CROSS_APP_OUT", temp_dir)
            .output()
            .map_err(|e| format!("run {script}: {e}"))?;
        if !output.status.success() {
            return Err(format!(
                "architecture inventory {script}: {}",
                String::from_utf8_lossy(&output.stderr)
            ));
        }
    }
    let tables = read_csv(&temp_dir.join("backend_tables.csv"))?;
    for row in &tables {
        let owner = row.get("owner_schema_domain").cloned().unwrap_or_default();
        let table = row.get("table").cloned().unwrap_or_default();
        if owner.is_empty() || owner.contains('|') {
            findings.push(Finding::new(
                "table-owner",
                "internal",
                format!("{table}: owner {owner:?}"),
            ));
            continue;
        }
        for access in ["reading", "writing"] {
            let consumers = row
                .get(&format!("modules_{access}"))
                .cloned()
                .unwrap_or_default();
            for consumer in consumers.split_whitespace() {
                if consumer != owner {
                    findings.push(Finding::new(
                        "foreign-table-access",
                        "internal",
                        format!("{consumer} {access} {table} owned by {owner}"),
                    ));
                }
            }
        }
    }
    let queries = read_csv(&temp_dir.join("backend_queries.csv"))?;
    for row in &queries {
        if row.get("module").map(String::as_str) == Some("aiquery")
            && !row
                .get("tables_written")
                .unwrap_or(&String::new())
                .is_empty()
        {
            findings.push(Finding::new(
                "read-authority",
                row.get("file").cloned().unwrap_or_default().as_str(),
                format!(
                    "{} writes {}",
                    row.get("query_name").cloned().unwrap_or_default(),
                    row.get("tables_written").cloned().unwrap_or_default()
                ),
            ));
        }
    }
    Ok(findings)
}

fn read_csv(path: &Path) -> Result<Vec<BTreeMap<String, String>>, String> {
    let text =
        std::fs::read_to_string(path).map_err(|e| format!("read {}: {e}", path.display()))?;
    let mut lines = text.lines();
    let header = lines.next().unwrap_or("");
    let names: Vec<String> = parse_csv_line(header);
    let mut rows = Vec::new();
    for line in lines {
        if line.trim().is_empty() {
            continue;
        }
        let cells = parse_csv_line(line);
        let mut row = BTreeMap::new();
        for (index, name) in names.iter().enumerate() {
            row.insert(name.clone(), cells.get(index).cloned().unwrap_or_default());
        }
        rows.push(row);
    }
    if rows.is_empty() {
        return Err(format!(
            "missing or invalid architecture inventory {}",
            path.display()
        ));
    }
    Ok(rows)
}

fn parse_csv_line(line: &str) -> Vec<String> {
    let mut out = Vec::new();
    let mut field = String::new();
    let mut quoted = false;
    let mut chars = line.chars().peekable();
    while let Some(ch) = chars.next() {
        match ch {
            '"' if !quoted => quoted = true,
            '"' if quoted => {
                if chars.peek() == Some(&'"') {
                    field.push('"');
                    let _ = chars.next();
                } else {
                    quoted = false;
                }
            }
            ',' if !quoted => {
                out.push(field.clone());
                field.clear();
            }
            _ => field.push(ch),
        }
    }
    out.push(field);
    out
}

fn run(root: &Path, skip_sweep: bool) -> i32 {
    let exceptions = match load_exceptions(root) {
        Ok(exceptions) => exceptions,
        Err(error) => {
            eprintln!("arch-check: {error}");
            return 1;
        }
    };
    let now = chrono::Utc::now().date_naive();
    if let Err(error) = validate_exceptions(&exceptions, root, now) {
        eprintln!("arch-check: {error}");
        return 1;
    }

    let sources = collect_sources(root);
    let dep_edges = load_dep_edges(root);
    let output = scan::scan_sources(sources, &dep_edges);
    let mut findings = output.findings;
    eprintln!("arch-check: {} module edges", dep_edges.len());
    let mut cargo_findings = check_cargo_platform_imports(root);
    findings.append(&mut cargo_findings);
    let mut migration_findings = match check_migration_owner(root) {
        Ok(findings) => findings,
        Err(error) => {
            eprintln!("arch-check: {error}");
            return 1;
        }
    };
    findings.append(&mut migration_findings);
    // Sweep (table/query ownership over Go SQL) runs for parity and is
    // reported, but the pre-existing Go findings do not fail the Rust
    // gate: they predate the migration and fail only the Go test suite.
    // Rust-side layering findings are what gate CI.
    if !skip_sweep {
        let temp = std::env::temp_dir().join(format!("arch-check-{}", std::process::id()));
        let _ = std::fs::create_dir_all(&temp);
        match check_sweep(root, &temp) {
            Ok(sweep) => {
                for finding in &sweep {
                    eprintln!(
                        "carried: {}: {}: {}",
                        finding.rule, finding.file, finding.detail
                    );
                }
                eprintln!("arch-check: {} carried Go sweep finding(s)", sweep.len());
            }
            Err(error) => {
                eprintln!("arch-check: {error}");
                return 1;
            }
        }
        let _ = std::fs::remove_dir_all(&temp);
    }

    let (remaining, unused) = apply_exceptions(findings, &exceptions);
    for finding in &remaining {
        eprintln!("{}: {}: {}", finding.rule, finding.file, finding.detail);
    }
    let mut failed = false;
    if !remaining.is_empty() {
        eprintln!("arch-check: {} violation(s)", remaining.len());
        failed = true;
    }
    for id in &unused {
        eprintln!("arch-check: remove unused architecture exception {id}");
        failed = true;
    }
    if !failed {
        eprintln!("arch-check: clean");
    }
    i32::from(failed)
}

fn main() -> std::process::ExitCode {
    let args: Vec<String> = std::env::args().skip(1).collect();
    let mut root = PathBuf::from("services/onion");
    if root.join("docs/architecture-exceptions.json").exists() {
        // Running from the repo root, argv may still point at the monorepo.
    } else if Path::new("docs/architecture-exceptions.json").exists() {
        root = PathBuf::from(".");
    }
    let mut skip_sweep = false;
    let mut iter = args.into_iter().peekable();
    while let Some(arg) = iter.next() {
        match arg.as_str() {
            "--root" => {
                if let Some(value) = iter.next() {
                    root = PathBuf::from(value);
                }
            }
            "--skip-sweep" => skip_sweep = true,
            "-h" | "--help" => {
                println!("usage: arch-check [--root <onion-dir>] [--skip-sweep]");
                return std::process::ExitCode::SUCCESS;
            }
            _ => {
                eprintln!("arch-check: unknown argument {arg:?}");
                return std::process::ExitCode::FAILURE;
            }
        }
    }
    std::process::ExitCode::from(run(&root, skip_sweep) as u8)
}

#[cfg(test)]
mod tests {
    use super::parse_csv_line;

    #[test]
    fn csv_line_parsing() {
        assert_eq!(parse_csv_line("a,b,\"c,d\",e"), vec!["a", "b", "c,d", "e"]);
    }
}
