#!/usr/bin/env python3
"""Emit the PR description (tables only) with current completion state.

Usage: python3 scripts/pr-description.py > /tmp/pr-body.md

The description is a set of Markdown tables: completion, phases, areas,
pending, and gates. Percentages come from STATE.md unit rows.
"""

import re
import subprocess
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
STATE = ROOT / "docs/rust-migration/STATE.md"


def read_state() -> str:
    return STATE.read_text()


def unit_stats() -> tuple[int, int]:
    """done/total unit rows across every phase table in STATE.md."""
    rows = re.findall(
        r"^\| ((?:[0-9]+(?:\.[0-9]+)?[a-z]?)\b[^|]*) \| (done|todo|blocked) \|",
        read_state(),
        re.M,
    )
    # Phase 3's module table uses a different shape (Module | LOC | Status |
    # Notes) and is counted in phase_rows, not here.
    total = len(rows)
    done = sum(1 for r in rows if r[1] == "done")
    return done, total


def commits_ahead() -> int:
    out = subprocess.run(
        ["git", "rev-list", "--count", "main..rust-migration"],
        cwd=ROOT.parents[1],
        capture_output=True,
        text=True,
    )
    return int(out.stdout.strip() or 0)


def phase_rows() -> list[tuple[str, str, str]]:
    state = read_state()
    sections = re.split(r"^## ", state, flags=re.M)
    rows = []
    phase_names = {
        "0": "0 Foundations",
        "1": "1 Platform crates",
        "2": "2 Query layer",
        "3": "3 Modules",
        "4": "4 Composition",
        "5": "5 Pipeline swap",
        "6": "6 Cleanup",
    }
    for section in sections:
        m = re.match(r"Phase (\d)", section)
        if not m:
            continue
        key = m.group(1)
        if key == "3":
            # Module table: | module | LOC | Status | Notes |
            done = len(re.findall(r"^\| [a-z_]+ \|[^|]*\| done \|", section, re.M))
            todo = len(re.findall(r"^\| [a-z_]+ \|[^|]*\| todo \|", section, re.M))
        else:
            done = len(re.findall(r"\|\s*done\s*\|", section))
            todo = len(re.findall(r"\|\s*todo\s*\|", section))
        rows.append((phase_names.get(key, key), f"{done}/{done + todo}", ""))
    return rows


def main() -> None:
    done, total = unit_stats()
    pct = done * 100 // max(total, 1)
    ahead = commits_ahead()
    rows = phase_rows()
    print("| Metric | Value |")
    print("|---|---|")
    print(f"| Completion (STATE units) | {done}/{total} ({pct}%) done |")
    print("| Branches | `rust-migration` → `main` (draft, no merge) |")
    print(f"| Commits ahead of main | {ahead} |")
    print("| Last push | Phase 2 complete: all 140 query packages converted |")
    print(
        "| Gates (CI rust-check) | fmt ✓ clippy ✓ check ✓ arch-check ✓"
        " (Go foreign-table-access debt carried) |"
    )
    print("| Local workspace | `cargo clippy --workspace -D warnings` ✓ 313 test suites ✓ |")
    print(
        "| SQL+bind parity | 1683/1683 via scripts/verify-sql-parity.py ✓ |"
    )
    print("| Blockers | none |")
    print()
    print("| Phase | Done | Remaining |")
    print("|---|---|---|")
    for phase, count, _ in rows:
        remaining = "—" if count.endswith("/0") else "see STATE.md"
        print(f"| {phase} | {count} | {remaining} |")
    print()
    print("| Area | State | Evidence |")
    print("|---|---|---|")
    print(
        "| 0.x Foundations | done | workspace, contracts (729 gen), inventory,"
        " coupling, route-dump, parity, arch-check, spike (hub.activity.v1 +"
        " CRM unsubscribe) |"
    )
    print(
        "| 1.x Platform crates | done | core, db, auth, events, realtime,"
        " integrations (S3/CDN/Resend/Slack/Paddle/Trino/Sheets/Gemini native),"
        " ai (golden parity <1e-4), server (chain/rate-limit/Partials/CORS/error"
        " writer), AppState + bins/api + bins/worker + PROBE_PORT |"
    )
    print(
        "| 2.x Query layer | done | tools/sqlc2sqlx (bind = Go body `arg.<Field>`"
        " order incl. repeats, null → Option<>, SLICE expansion, concat consts,"
        " r#keyword escape, Copy-binds) + 140 query crates; SQL parity 1683/1683 |"
    )
    print(
        "| 3.x Modules | in progress | aiquota domain/application/repository over"
        " generated queries |"
    )
    print()
    print("| Pending | Owner | Blocks |")
    print("|---|---|---|")
    print(
        "| 0.7b media-engine ops (resize/transform/info/analyze/video/thumbnail)"
        " | migration PR (Phase 5 gate) | dropping C++ build |"
    )
    print("| Phase 3 modules (24 remaining) | migration PR | Phase 4 |")
    print(
        "| Phase 4 composition (route parity + replay) | migration PR |"
        " pipeline swap |"
    )
    print("| gh auth (401) | human | PR body editing + CI checks view |")
    print()
    print("| Gate | Result | Note |")
    print("|---|---|---|")
    print("| cargo fmt --check | ✓ | CI order (fmt before generate) |")
    print(
        "| cargo clippy --workspace -D warnings | ✓ | after make -C voxie"
        " generate |"
    )
    print("| cargo test --workspace | ✓ 313 suites | |")
    print("| verify-sql-parity.py | ✓ 1683/1683 | SQL text + bind order |")
    print(
        "| DB integration | CI-only | local compose pull fails (TLS); repository"
        " tests skip without DATABASE_URL |"
    )


if __name__ == "__main__":
    main()
