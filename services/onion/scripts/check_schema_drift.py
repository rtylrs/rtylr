#!/usr/bin/env python3
"""Detect schema drift between the SQL schema files and a live database.

This is a MANUAL operator tool — it connects to a database and must never run in
CI (CI does not touch databases). Run it before/after a manual migration to
confirm the deployed code's expected schema matches the database.

It reports:
  - tables defined in the schema files but MISSING from the DB
    (cause of "Table doesn't exist" errors)
  - columns defined in the schema files but MISSING from the DB
    (cause of "Unknown column" 500s — e.g. the users.two_factor_enabled outage)
  - tables present in the DB but defined in NO schema file (informational —
    a reproducibility gap, not a runtime error)

Exit code is non-zero when any missing table/column is found, so it can gate a
deploy step.

Usage:
    set -a; . ./.env; set +a            # load DB_HOST/PORT/USER/PASSWORD/NAME
    python3 scripts/check_schema_drift.py
"""

import os
import re
import sys
import glob
import subprocess

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

# Every place a CREATE TABLE / ALTER TABLE ADD COLUMN can live.
SCHEMA_GLOBS = [
    os.path.join(ROOT, "internal/modules/**/sql/**/*.sql"),
    os.path.join(ROOT, "internal/platform/**/sql/**/*.sql"),
]

KEYWORDS = {
    "primary", "unique", "key", "index", "constraint", "foreign",
    "check", "fulltext", "spatial",
}

# Tables that legitimately live only in the DB and are not app schema:
# migration-tool bookkeeping, etc. Not treated as drift.
IGNORE_DB_ONLY = {"schema_migrations"}


def strip_comments(s):
    s = re.sub(r"/\*.*?\*/", "", s, flags=re.S)
    out = []
    for line in s.splitlines():
        i = line.find("--")
        if i != -1:
            line = line[:i]
        out.append(line)
    return "\n".join(out)


def table_body(text, open_idx):
    depth = 0
    for j in range(open_idx, len(text)):
        c = text[j]
        if c == "(":
            depth += 1
        elif c == ")":
            depth -= 1
            if depth == 0:
                return text[open_idx + 1:j]
    return None


def split_top_commas(body):
    items, depth, cur, instr = [], 0, [], None
    for c in body:
        if instr:
            cur.append(c)
            if c == instr:
                instr = None
            continue
        if c in ("'", '"', "`"):
            instr = c
            cur.append(c)
        elif c == "(":
            depth += 1
            cur.append(c)
        elif c == ")":
            depth -= 1
            cur.append(c)
        elif c == "," and depth == 0:
            items.append("".join(cur))
            cur = []
        else:
            cur.append(c)
    if cur:
        items.append("".join(cur))
    return items


def column_name(item):
    m = re.match(r"^\s*`?([A-Za-z_][A-Za-z0-9_]*)`?", item)
    if not m:
        return None
    name = m.group(1).lower()
    return None if name in KEYWORDS else name


def parse_expected():
    expected = {}  # table -> set(columns)
    files = []
    for g in SCHEMA_GLOBS:
        files += glob.glob(g, recursive=True)
    for path in sorted(set(files)):
        text = strip_comments(open(path, errors="ignore").read())
        for m in re.finditer(
            r"CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?`?([A-Za-z_]\w*)`?\s*\(",
            text, re.I,
        ):
            t = m.group(1).lower()
            body = table_body(text, m.end() - 1)
            if body is None:
                continue
            expected.setdefault(t, set())
            for item in split_top_commas(body):
                c = column_name(item)
                if c:
                    expected[t].add(c)
        for m in re.finditer(
            r"ALTER\s+TABLE\s+`?([A-Za-z_]\w*)`?\s+(.*?);", text, re.I | re.S
        ):
            t = m.group(1).lower()
            for am in re.finditer(
                r"ADD\s+(?:COLUMN\s+)?(?:IF\s+NOT\s+EXISTS\s+)?`?([A-Za-z_]\w*)`?",
                m.group(2), re.I,
            ):
                c = am.group(1).lower()
                if c not in KEYWORDS:
                    expected.setdefault(t, set()).add(c)
    return expected


def fetch_db_columns():
    env = os.environ
    for var in ("DB_HOST", "DB_USER", "DB_NAME"):
        if not env.get(var):
            sys.exit(f"error: {var} is not set — source ./.env first")
    cmd = [
        "mysql",
        "-h", env["DB_HOST"],
        "-P", env.get("DB_PORT", "3306"),
        "-u", env["DB_USER"],
        "-D", env["DB_NAME"],
        "-N", "-B",
        "-e",
        # Base tables only — views (e.g. product_stock_total) are derived, not
        # schema drift.
        "SELECT c.TABLE_NAME, c.COLUMN_NAME FROM information_schema.COLUMNS c "
        "JOIN information_schema.TABLES t "
        "  ON t.TABLE_SCHEMA = c.TABLE_SCHEMA AND t.TABLE_NAME = c.TABLE_NAME "
        f"WHERE c.TABLE_SCHEMA = '{env['DB_NAME']}' "
        "  AND t.TABLE_TYPE = 'BASE TABLE';",
    ]
    run_env = dict(env)
    run_env["MYSQL_PWD"] = env.get("DB_PASSWORD", "")
    out = subprocess.run(cmd, capture_output=True, text=True, env=run_env, timeout=30)
    if out.returncode != 0:
        sys.exit(f"error: mysql query failed: {out.stderr.strip()}")
    db = {}
    for line in out.stdout.splitlines():
        parts = line.split("\t")
        if len(parts) < 2:
            continue
        db.setdefault(parts[0].lower(), set()).add(parts[1].lower())
    return db


def main():
    expected = parse_expected()
    db = fetch_db_columns()

    missing_tables = sorted(set(expected) - set(db))
    db_only = sorted(set(db) - set(expected) - IGNORE_DB_ONLY)
    missing_columns = {}
    for t in sorted(set(expected) & set(db)):
        miss = expected[t] - db[t]
        if miss:
            missing_columns[t] = sorted(miss)

    print(f"schema files: {len(expected)} tables   db: {len(db)} tables\n")

    print(f"### tables in schema but MISSING from db ({len(missing_tables)}):")
    for t in missing_tables:
        print(f"  - {t}")

    print(f"\n### columns in schema but MISSING from db ({len(missing_columns)} tables):")
    for t, cols in missing_columns.items():
        print(f"  - {t}: {', '.join(cols)}")

    print(f"\n### tables only in db, not defined in any schema file ({len(db_only)}):")
    print("  " + (", ".join(db_only) if db_only else "(none)"))

    drift = bool(missing_tables or missing_columns)
    print("\n" + ("DRIFT DETECTED" if drift else "no drift — schema and db agree"))
    sys.exit(1 if drift else 0)


if __name__ == "__main__":
    main()
