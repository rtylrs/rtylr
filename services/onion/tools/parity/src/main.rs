//! `parity diff <go.jsonl> <rust.jsonl>` prints normalized diffs;
//! `parity replay --go <url> --rust <url> --corpus <file>` replays a
//! JSONL corpus against both servers (throwaway MySQL from
//! `scripts/test-db.sh`, container `onion-env-db-1`).

use std::fs;
use std::process::ExitCode;

use parity::{diff_runs, parse_corpus, parse_run};

fn run() -> Result<i32, String> {
    let args: Vec<String> = std::env::args().skip(1).collect();
    match args.first().map(String::as_str) {
        Some("diff") if args.len() == 3 => {
            let go = fs::read_to_string(&args[1]).map_err(|e| format!("read {}: {e}", args[1]))?;
            let rust =
                fs::read_to_string(&args[2]).map_err(|e| format!("read {}: {e}", args[2]))?;
            let go = parse_run(&go).map_err(|e| format!("{}: {e}", args[1]))?;
            let rust = parse_run(&rust).map_err(|e| format!("{}: {e}", args[2]))?;
            let diffs = diff_runs(&go, &rust);
            for diff in &diffs {
                println!(
                    "{}\tgo={} rust={} bodies_equal={}",
                    diff.id, diff.go_status, diff.rust_status, diff.bodies_equal
                );
            }
            eprintln!("diffs: {}", diffs.len());
            Ok(if diffs.is_empty() { 0 } else { 1 })
        }
        Some("replay") => {
            // Live replay needs two servers up (scripts/test-db.sh + both
            // binaries). The skeleton only validates the corpus parses.
            let corpus = args
                .iter()
                .position(|a| a == "--corpus")
                .and_then(|i| args.get(i + 1))
                .ok_or("replay needs --corpus <file.jsonl>")?;
            let text = fs::read_to_string(corpus).map_err(|e| format!("read {corpus}: {e}"))?;
            let vectors = parse_corpus(&text).map_err(|e| format!("{corpus}: {e}"))?;
            eprintln!("corpus vectors: {}", vectors.len());
            eprintln!("replay: point --go/--rust at two servers seeded from the same throwaway DB");
            Ok(0)
        }
        _ => Err(
            "usage: parity (diff <go.jsonl> <rust.jsonl> | replay --corpus <file.jsonl>)"
                .to_string(),
        ),
    }
}

fn main() -> ExitCode {
    match run() {
        Ok(code) => ExitCode::from(code as u8),
        Err(message) => {
            eprintln!("parity: {message}");
            ExitCode::FAILURE
        }
    }
}
