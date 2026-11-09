//! `route-dump <legacy.tsv> [rust.tsv]`: print the normalized table, and
//! with a Rust dump, print the parity diff.

use std::fs;
use std::process::ExitCode;

use route_dump::{diff, parse_dump, render_table};

fn run() -> Result<i32, String> {
    let args: Vec<String> = std::env::args().skip(1).collect();
    if args.is_empty() || args.len() > 2 || args.iter().any(|a| a == "-h" || a == "--help") {
        return Err("usage: route-dump <legacy.tsv> [rust.tsv]".to_string());
    }
    let go = fs::read_to_string(&args[0]).map_err(|e| format!("read {}: {e}", args[0]))?;
    let routes = parse_dump(&go).map_err(|e| format!("{}: {e}", args[0]))?;
    print!("{}", render_table(&routes));
    eprintln!("go routes: {}", routes.len());
    if args.len() == 2 {
        let rust = fs::read_to_string(&args[1]).map_err(|e| format!("read {}: {e}", args[1]))?;
        let rust_routes = parse_dump(&rust).map_err(|e| format!("{}: {e}", args[1]))?;
        let (only_go, only_rust) = diff(&routes, &rust_routes);
        eprintln!("only in go: {}", only_go.len());
        eprintln!("only in rust: {}", only_rust.len());
        for route in &only_go {
            println!(
                "go-only\t{}\t{} {}",
                route.kind, route.method, route.pattern
            );
        }
        for route in &only_rust {
            println!(
                "rust-only\t{}\t{} {}",
                route.kind, route.method, route.pattern
            );
        }
        if only_go.is_empty() && only_rust.is_empty() {
            eprintln!("parity: clean");
            return Ok(0);
        }
        return Ok(1);
    }
    Ok(0)
}

fn main() -> ExitCode {
    match run() {
        Ok(code) => ExitCode::from(code as u8),
        Err(message) => {
            eprintln!("route-dump: {message}");
            ExitCode::FAILURE
        }
    }
}
