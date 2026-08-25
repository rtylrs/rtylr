#!/usr/bin/env python3
"""Check that .env.example documents production environment inputs.

This scans production Go files (not tests) for literal environment lookups
and environment-name constants. It is a drift check, not a secret generator:
values remain blank or safe defaults in .env.example.
"""

from pathlib import Path
import re
import sys

ROOT = Path(__file__).resolve().parents[1]
ENV_FILE = ROOT / ".env.example"
IGNORED = {"HOSTNAME", "KUBERNETES_SERVICE_HOST"}
KEY = r"[A-Z][A-Z0-9_]{2,}"


def documented() -> set[str]:
    return {
        match.group(1)
        for line in ENV_FILE.read_text().splitlines()
        if (match := re.match(rf"^({KEY})=", line))
    }


def production_sources() -> list[Path]:
    return [
        path
        for root in (ROOT / "internal", ROOT / "cmd")
        for path in root.rglob("*.go")
        if not path.name.endswith("_test.go")
    ]


def discovered() -> set[str]:
    names: set[str] = set()
    for path in production_sources():
        text = path.read_text(errors="ignore")
        names.update(re.findall(rf'os\.Getenv\("({KEY})"\)', text))
        names.update(re.findall(rf'\b(?:envFirst|envOrDefault|envDurationOrDefault|envFlag)\("({KEY})"', text))
        names.update(re.findall(rf'\b\w*[Ee]nv\w*\s*=\s*"({KEY})"', text))
    return names - IGNORED


def main() -> int:
    missing = sorted(discovered() - documented())
    if missing:
        print(".env.example is missing runtime variables:")
        print("\n".join(f"  {name}=" for name in missing))
        return 1
    print(".env.example covers all discovered production environment variables")
    return 0


if __name__ == "__main__":
    sys.exit(main())
