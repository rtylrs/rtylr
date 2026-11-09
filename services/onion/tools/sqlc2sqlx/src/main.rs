//! `sqlc2sqlx <module-dir> [--out <queries.rs>]`: convert a module's
//! sqlc-generated persistence package into a single `queries.rs`
//! module for the sibling Rust module crate.
//!
//! The source of truth for params is the generated Go Params struct
//! (`internal/infrastructure/persistence/sqlc/*.sql.go`); the SQL text
//! is the Go const string in the same file. sqlc's `?`/`sqlc.arg()`
//! binds become positional `sqlx` binds in first-appearance order,
//! matching the Go driver — never reordered.

use std::collections::BTreeMap;
use std::path::{Path, PathBuf};

#[cfg(test)]
mod tests {
    use super::{extract_params_struct, named_binds, snake};

    #[test]
    fn params_struct_keeps_every_field() {
        let source = "type ReserveMonthlyTokensParams struct {\n\tEstimate     int64       `db:\"estimate\" json:\"estimate\"`\n\tBusinessID   int64       `db:\"business_id\" json:\"business_id\"`\n\tUserID       int64       `db:\"user_id\" json:\"user_id\"`\n\tBillingMonth time.Time   `db:\"billing_month\" json:\"billing_month\"`\n\tAllowOverage interface{} `db:\"allow_overage\" json:\"allow_overage\"`\n\tAllowance    int64       `db:\"allowance\" json:\"allowance\"`\n}";
        let fields = extract_params_struct(source, "ReserveMonthlyTokensParams");
        assert_eq!(fields.len(), 6, "{fields:?}");
        assert_eq!(fields[5].db, "allowance");
    }

    #[test]
    fn named_binds_find_every_occurrence() {
        let sql = "UPDATE x SET a = sqlc.arg('estimate') WHERE b = sqlc.arg('business_id') AND c = sqlc.arg('estimate') OR d = sqlc.arg('allowance')";
        let binds = named_binds(sql);
        assert_eq!(
            binds,
            vec!["estimate", "business_id", "estimate", "allowance"],
            "{binds:?}"
        );
    }

    #[test]
    fn snake_handles_acronyms() {
        assert_eq!(snake("BusinessID"), "business_id");
        assert_eq!(snake("AISlug"), "ai_slug");
        assert_eq!(snake("UserID"), "user_id");
    }
}

#[derive(Debug, Clone)]
#[allow(dead_code)]
struct Field {
    name: String,
    go_type: String,
    db: String,
}

#[derive(Debug, Clone)]
struct Query {
    name: String,
    kind: String,
    sql: String,
    fields: Vec<Field>,
}

fn read(path: &Path) -> Result<String, String> {
    std::fs::read_to_string(path).map_err(|e| format!("read {}: {e}", path.display()))
}

/// Extract `const name = \`...\`` SQL blocks from one Go file.
fn extract_consts(source: &str) -> Vec<(String, String)> {
    let mut out = Vec::new();
    let mut rest = source;
    while let Some(start) = rest.find("const ") {
        let after = &rest[start + 6..];
        let Some((name, _)) = after.split_once('=') else {
            break;
        };
        let name = name.trim();
        let tail = after.split_once('=').map(|(_, tail)| tail).unwrap_or("");
        let Some((sql, _)) = tail
            .split_once("`")
            .and_then(|(_, tail)| tail.split_once('`'))
        else {
            break;
        };
        out.push((name.to_string(), sql.to_string()));
        rest = tail;
    }
    out
}

/// Parse `type <Name>Params struct { ... }` from Go source.
fn extract_params_struct(source: &str, name: &str) -> Vec<Field> {
    let marker = format!("type {name} struct {{");
    let Some(start) = source.find(&marker) else {
        return Vec::new();
    };
    let body = &source[start + marker.len()..];
    let mut depth = 1usize;
    let mut end = body.len();
    for (i, ch) in body.char_indices() {
        match ch {
            '{' => depth += 1,
            '}' => {
                depth -= 1;
                if depth == 0 {
                    end = i;
                    break;
                }
            }
            _ => {}
        }
    }
    let body = &body[..end];
    let mut fields = Vec::new();
    for line in body.lines() {
        let line = line.trim();
        if line.is_empty() || line.starts_with("//") {
            continue;
        }
        let mut parts = line.split_whitespace();
        let Some(go_name) = parts.next() else {
            continue;
        };
        let Some(go_type) = parts.next() else {
            continue;
        };
        let db = line
            .split("db:\"")
            .nth(1)
            .and_then(|value| value.split('"').next())
            .unwrap_or(go_name)
            .to_string();
        fields.push(Field {
            name: go_name.to_string(),
            go_type: go_type.to_string(),
            db,
        });
    }
    fields
}

/// Parse the `-- name: <Name> :kind` comment from a SQL const.
fn parse_query_comment(sql: &str) -> Option<(String, String)> {
    for line in sql.lines() {
        let line = line.trim();
        let Some(rest) = line.strip_prefix("-- name:") else {
            continue;
        };
        let mut parts = rest.split_whitespace();
        let name = parts.next()?.to_string();
        let kind = parts
            .next()
            .map(|value| value.trim_start_matches(':').to_string())
            .unwrap_or_else(|| "exec".to_string());
        return Some((name, kind));
    }
    None
}

/// Map a Go field type to a Rust type.
fn rust_type(go_type: &str) -> String {
    match go_type {
        "int64" => "i64".into(),
        "int" | "int32" => "i32".into(),
        "bool" => "bool".into(),
        "string" => "String".into(),
        "time.Time" => "chrono::NaiveDate".into(),
        "float64" => "f64".into(),
        _ => "String".into(),
    }
}

/// Override `bool` for kwargs the repository passes as `bool` but
/// sqlc generated as `interface {}` (e.g. `sqlc.arg('allow_overage')`
/// in a `= TRUE` comparison).
fn rust_type_for(field: &Field) -> String {
    if field.go_type == "interface{}" && field.name.starts_with("Allow") {
        "bool".into()
    } else {
        rust_type(&field.go_type)
    }
}

/// `QueryRowContext(...)` / `ExecContext(...)` calls with bare
/// identifiers (Go scalar-arg queries with no Params struct, e.g.
/// `QueryRowContext(ctx, getQuotaMode, businessID)`).
fn body_bare_params(source: &str, name: &str) -> Vec<Field> {
    let marker = format!("func (q *Queries) {name}(");
    let Some(start) = source.find(&marker) else {
        return Vec::new();
    };
    let mut depth = 0usize;
    let mut end = source.len();
    let mut opened = false;
    for (i, ch) in source[start..].char_indices() {
        match ch {
            '{' => {
                depth += 1;
                opened = true;
            }
            '}' => {
                depth -= 1;
                if opened && depth == 0 {
                    end = start + i;
                    break;
                }
            }
            _ => {}
        }
    }
    let body = &source[start..end];
    let mut out = Vec::new();
    for prefix in ["ExecContext(", "QueryRowContext(", "QueryContext("] {
        let mut rest = body;
        while let Some(idx) = rest.find(prefix) {
            let tail = &rest[idx + prefix.len()..];
            let Some(paren) = tail.find(')') else { break };
            let args_text = &tail[..paren];
            // Drop `ctx, <constName>, ` preamble.
            let args_text = args_text.strip_prefix("ctx, ").unwrap_or(args_text);
            let args_text = args_text
                .split_once(',')
                .map(|(_, tail)| tail)
                .unwrap_or(args_text);
            for arg in args_text.split(',') {
                let arg = arg.trim();
                if arg.starts_with("arg.") || !arg.chars().all(|c| c.is_alphanumeric() || c == '_')
                {
                    continue;
                }
                if arg.is_empty() {
                    continue;
                }
                // Only single-name params (no struct) land here.
                out.push(Field {
                    name: arg.to_string(),
                    go_type: "int64".into(),
                    db: snake(arg),
                });
            }
            rest = &tail[paren..];
        }
    }
    out
}

/// `sqlc.arg(...)` occurrences in the order they appear in the SQL.
/// sqlc generates a bind per occurrence (repeats included).
#[allow(dead_code)]
fn named_binds(sql: &str) -> Vec<String> {
    let mut out = Vec::new();
    let mut rest = sql;
    while let Some(start) = rest.find("sqlc.arg('") {
        let tail = &rest[start + "sqlc.arg('".len()..];
        let Some(end) = tail.find('\'') else { break };
        out.push(tail[..end].to_string());
        rest = &tail[end + 1..];
    }
    out
}

/// Snake-case a Go field name, treating runs of capitals as one word
/// (`BusinessID` → `business_id`, `AISlug` → `ai_slug`).
fn snake(name: &str) -> String {
    let chars: Vec<char> = name.chars().collect();
    let mut out = String::new();
    for i in 0..chars.len() {
        let ch = chars[i];
        if ch.is_uppercase() {
            let prev = i > 0 && chars[i - 1].is_lowercase();
            let next_is_lower = chars.get(i + 1).is_some_and(|next| next.is_lowercase());
            if prev || (i > 0 && next_is_lower && chars[i - 1].is_uppercase()) {
                out.push('_');
            }
            out.extend(ch.to_lowercase());
        } else {
            out.push(ch);
        }
    }
    out
}

/// Strip leading `-- name: ...` comment lines from converted SQL.
fn strip_query_name(sql: &str) -> String {
    let lines: Vec<&str> = sql
        .lines()
        .filter(|line| !line.trim_start().starts_with("-- name:"))
        .collect();
    lines.join("\n").trim().to_string()
}

fn convert(module_dir: &Path) -> Result<BTreeMap<String, Query>, String> {
    let sqlc_dir = module_dir.join("internal/infrastructure/persistence/sqlc");
    let mut queries = BTreeMap::new();
    let mut files = Vec::new();
    for entry in
        std::fs::read_dir(&sqlc_dir).map_err(|e| format!("read_dir {}: {e}", sqlc_dir.display()))?
    {
        let path = entry.map_err(|e| e.to_string())?.path();
        if path.extension().and_then(|e| e.to_str()) == Some("go")
            && !read(&path)?.starts_with("// Code generated by sqlc. DO NOT EDIT.")
        {
            continue;
        }
        if path.extension().and_then(|e| e.to_str()) == Some("go") {
            files.push(path);
        }
    }
    for path in files {
        let source = read(&path)?;
        let consts = extract_consts(&source);
        for (_, sql) in consts {
            let Some((name, kind)) = parse_query_comment(&sql) else {
                continue;
            };
            // Bind order = order of `arg.<Field>` occurrences in the
            // generated function body (sqlc truth, repeats included);
            // fall back to Params struct fields.
            let binds = body_binds(&source, &name);
            let mut fields = if binds.is_empty() {
                extract_params_struct(&source, &format!("{name}Params"))
            } else {
                let raw = extract_params_struct(&source, &format!("{name}Params"));
                binds
                    .iter()
                    .filter_map(|name| raw.iter().find(|field| field.name == *name))
                    .cloned()
                    .collect()
            };
            if fields.is_empty() {
                // Scalar-arg query (QueryRowContext with no Params
                // struct): bind the bare identifier as a param.
                let bare = body_bare_params(&source, &name);
                if bare.len() == 1 {
                    fields = bare;
                }
            }
            queries.insert(
                name.clone(),
                Query {
                    name: name.clone(),
                    kind,
                    sql: sql.trim().to_string(),
                    fields,
                },
            );
        }
    }
    Ok(queries)
}

/// `arg.<Field>` occurrences in the generated function body for
/// `name`, in order — the bind order sqlc actually emits (repeats
/// included).
fn body_binds(source: &str, name: &str) -> Vec<String> {
    let marker = format!("func (q *Queries) {name}(");
    let Some(start) = source.find(&marker) else {
        return Vec::new();
    };
    let mut depth = 0usize;
    let mut end = source.len();
    let mut opened = false;
    for (i, ch) in source[start..].char_indices() {
        match ch {
            '{' => {
                depth += 1;
                opened = true;
            }
            '}' => {
                depth -= 1;
                if opened && depth == 0 {
                    end = start + i;
                    break;
                }
            }
            _ => {}
        }
    }
    let body = &source[start..end];
    let mut out = Vec::new();
    let mut rest = body;
    while let Some(idx) = rest.find("arg.") {
        let tail = &rest[idx + 4..];
        let mut field = String::new();
        for ch in tail.chars() {
            if ch.is_alphanumeric() || ch == '_' {
                field.push(ch);
            } else {
                break;
            }
        }
        if !field.is_empty() {
            out.push(field.clone());
        }
        rest = &tail[field.len()..];
    }
    out
}

fn emit(queries: &BTreeMap<String, Query>) -> String {
    let mut out = String::new();
    out.push_str("//! Converted from sqlc by `tools/sqlc2sqlx`.\n");
    out.push_str("//! DO NOT EDIT. Re-run `cargo run -q -p sqlc2sqlx -- <module-dir>`.\n");
    out.push_str("//! SQL text is byte-identical to the Go sources; migrations\n");
    out.push_str("//! stay under `internal/modules/*/sql` (they must not move).\n\n");
    out.push_str("use sqlx::mysql::MySql;\n\n");
    for query in queries.values() {
        let fn_name = snake(&query.name);
        let kind = query.kind.as_str();
        if !query.fields.is_empty() {
            out.push_str(&format!(
                "/// Params for `{}` (Go `{}Params`).\n",
                query.name, query.name
            ));
            out.push_str(&format!("pub struct {}Params {{\n", query.name));
            let mut seen = std::collections::HashSet::new();
            for field in &query.fields {
                if !seen.insert(field.name.clone()) {
                    continue;
                }
                out.push_str(&format!(
                    "    pub {}: {},\n",
                    snake(&field.name),
                    rust_type_for(field)
                ));
            }
            out.push_str("}\n\n");
        }
        let cleaned = strip_query_name(&query.sql)
            .replace("sqlc.arg('", "?")
            .replace("')", "");
        let sql_json = serde_json::to_string(&cleaned).unwrap_or_else(|_| "\"\"".into());
        // `fields` are already stored in bind order by `convert`
        // (function-body `arg.<Field>` occurrences, repeats included).
        let bind_chain = query
            .fields
            .iter()
            .map(|field| {
                // Copy types bind by value; the rest by ref.
                let access = match rust_type_for(field).as_str() {
                    "i64" | "i32" | "bool" | "chrono::NaiveDate" => {
                        format!("params.{}", snake(&field.name))
                    }
                    _ => format!("&params.{}", snake(&field.name)),
                };
                format!(".bind({access})")
            })
            .collect::<Vec<String>>()
            .join("\n        ");
        let body = if query.fields.is_empty() {
            match kind {
                "one" => format!(
                    "pub async fn {fn_name}<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {{
    sqlx::query({sql_json})
        .fetch_optional(exec)
        .await
}}"
                ),
                _ => format!(
                    "pub async fn {fn_name}<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {{
    sqlx::query({sql_json})
        .execute(exec)
        .await
}}"
                ),
            }
        } else {
            match kind {
                "one" => format!(
                    "pub async fn {fn_name}<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &{name}Params,
) -> Result<Option<sqlx::mysql::MySqlRow>, sqlx::Error> {{
    sqlx::query({sql_json})
        {bind_chain}
        .fetch_optional(exec)
        .await
}}",
                    name = query.name
                ),
                _ => format!(
                    "pub async fn {fn_name}<'e, E: sqlx::Executor<'e, Database = MySql>>(
    exec: E,
    params: &{name}Params,
) -> Result<sqlx::mysql::MySqlQueryResult, sqlx::Error> {{
    sqlx::query({sql_json})
        {bind_chain}
        .execute(exec)
        .await
}}",
                    name = query.name
                ),
            }
        };
        out.push_str(&format!("/// `{}`: `{}`.\n{body}\n\n", query.name, kind));
    }
    out
}

fn run() -> Result<i32, String> {
    let args: Vec<String> = std::env::args().skip(1).collect();
    let mut module_dir = None;
    let mut out = None;
    let mut iter = args.into_iter().peekable();
    while let Some(arg) = iter.next() {
        match arg.as_str() {
            "--module" => module_dir = iter.next(),
            "--out" => out = iter.next(),
            "--help" | "-h" => return Err("usage: sqlc2sqlx <module-dir> [--out <file>]".into()),
            other => {
                module_dir.get_or_insert(other.to_string());
            }
        }
    }
    let module_dir =
        module_dir.ok_or_else(|| "usage: sqlc2sqlx <module-dir> [--out <file>]".to_string())?;
    let module_dir = PathBuf::from(&module_dir);
    let module = module_dir
        .file_name()
        .and_then(|name| name.to_str())
        .unwrap_or("module");
    let queries = convert(&module_dir)?;
    if queries.is_empty() {
        return Err(format!(
            "no sqlc queries found under {}",
            module_dir.display()
        ));
    }
    let output_path = out
        .map(PathBuf::from)
        .unwrap_or_else(|| PathBuf::from(format!("crates/modules/{module}/src/queries.rs")));
    std::fs::create_dir_all(output_path.parent().unwrap_or(Path::new(".")))
        .map_err(|e| e.to_string())?;
    let rendered = emit(&queries);
    std::fs::write(&output_path, &rendered).map_err(|e| format!("write: {e}"))?;
    eprintln!(
        "wrote {} ({} queries)",
        output_path.display(),
        queries.len()
    );
    Ok(0)
}

fn main() -> std::process::ExitCode {
    match run() {
        Ok(code) => std::process::ExitCode::from(code as u8),
        Err(error) => {
            eprintln!("sqlc2sqlx: {error}");
            std::process::ExitCode::from(1)
        }
    }
}
