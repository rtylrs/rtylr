#!/usr/bin/env python3
"""Report insert paths that let business_id fall back to its schema default.

Pure static analysis: it reads .sql files and touches no database, so it is safe
in CI. It exists because the number in the tenancy issue was measured by hand and
would otherwise have to be rediscovered.

Two facts it reports, and the difference between them matters:

  1. Tables declaring `business_id BIGINT NOT NULL DEFAULT 0`. That default turns
     a forgotten tenant into a silently shared one rather than an error.
     `internal/platform/database/schemas/flow/v8_time_tracking.sql:9-12` documents
     the decision to stop adding to this set.

  2. INSERT statements into those tables that omit business_id from their column
     list. These are the statements a `DROP DEFAULT` would turn into a loud error.

What this CANNOT see, and what therefore needs a different fix: an INSERT that
names business_id and is handed a zero. sqlc builds a params struct, an unset
int64 field is 0, and the statement then writes 0 explicitly. Dropping the
default does not catch that, because the column was supplied. Only a CHECK
constraint or validation above the query does.

Usage:
    python3 scripts/check_business_id_defaults.py            # report
    python3 scripts/check_business_id_defaults.py --max N     # fail if omissions exceed N
"""

import argparse
import collections
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SCHEMAS = os.path.join(ROOT, "internal", "platform", "database", "schemas")

CREATE_TABLE = re.compile(
    r"CREATE TABLE(?:\s+IF NOT EXISTS)?\s+`?(\w+)`?\s*\((.*?)\n\)\s*", re.S | re.I
)
DEFAULT_ZERO = re.compile(r"business_id\s+BIGINT\s+NOT NULL\s+DEFAULT\s+0", re.I)
INSERT_INTO = re.compile(r"INSERT\s+INTO\s+`?(\w+)`?\s*\(([^)]*)\)", re.S | re.I)


def sql_files(root, skip_schemas):
    for dirpath, dirnames, filenames in os.walk(root):
        dirnames[:] = [d for d in dirnames if d not in (".git", "voxie", "node_modules")]
        if skip_schemas and dirpath.startswith(SCHEMAS):
            continue
        for name in filenames:
            if name.endswith(".sql"):
                yield os.path.join(dirpath, name)


def defaulted_tables():
    """Tables whose business_id column carries DEFAULT 0, mapped to their schema file."""
    found = {}
    for path in sql_files(SCHEMAS, skip_schemas=False):
        with open(path, encoding="utf-8", errors="replace") as fh:
            text = fh.read()
        # Walk CREATE TABLE bodies so a column is attributed to its own table.
        for match in CREATE_TABLE.finditer(text):
            table, body = match.group(1), match.group(2)
            if DEFAULT_ZERO.search(body):
                found.setdefault(table, os.path.relpath(path, ROOT))
    return found


def omitting_inserts(defaulted):
    """INSERTs into a defaulted table whose column list has no business_id."""
    omits = collections.defaultdict(list)
    covered = 0
    for path in sql_files(ROOT, skip_schemas=True):
        with open(path, encoding="utf-8", errors="replace") as fh:
            text = fh.read()
        for match in INSERT_INTO.finditer(text):
            table, columns = match.group(1), match.group(2)
            if table not in defaulted:
                continue
            names = [c.strip().strip("`") for c in columns.split(",")]
            if "business_id" in names:
                covered += 1
                continue
            line = text[: match.start()].count("\n") + 1
            omits[table].append(f"{os.path.relpath(path, ROOT)}:{line}")
    return omits, covered


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--max",
        type=int,
        default=None,
        help="exit non-zero when the number of omitting inserts exceeds this, "
        "so the count can be ratcheted down but not up",
    )
    args = parser.parse_args()

    defaulted = defaulted_tables()
    omits, covered = omitting_inserts(defaulted)
    total = sum(len(v) for v in omits.values())

    print(f"tables declaring business_id BIGINT NOT NULL DEFAULT 0: {len(defaulted)}")
    print(f"inserts into them that set business_id:                 {covered}")
    print(f"inserts into them that omit business_id:                {total}")

    if omits:
        # Fixtures and production queries fail differently: a fixture seeds
        # tenant-0 rows into a test database, which is what makes a
        # single-tenant test unable to tell a scoped read from an unscoped one.
        fixtures = {t: [l for l in v if "scripts/fixtures/" in l] for t, v in omits.items()}
        queries = {t: [l for l in v if "scripts/fixtures/" not in l] for t, v in omits.items()}

        for label, group in (("production query paths", queries), ("test fixtures", fixtures)):
            rows = {t: v for t, v in group.items() if v}
            count = sum(len(v) for v in rows.values())
            print(f"\n{label} ({count}):")
            for table in sorted(rows):
                print(f"  {table}  [{defaulted[table]}]")
                for loc in rows[table]:
                    print(f"      {loc}")

    if args.max is not None and total > args.max:
        print(f"\nFAIL: {total} omitting inserts exceeds the allowed {args.max}")
        sys.exit(1)

    print("\nok")
    sys.exit(0)


if __name__ == "__main__":
    main()
