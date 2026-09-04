#!/usr/bin/env python3
"""Reject handwritten camelCase JSON tags in module code."""

from __future__ import annotations

import re
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
TAG_PATTERN = re.compile(r'`json:"(?P<tag>[^"]*)"`')


def main() -> int:
	findings: list[str] = []
	modules_root = ROOT / "internal/modules"
	for path in modules_root.rglob("*.go"):
		if path.name.endswith("_test.go"):
			continue
		text = path.read_text(encoding="utf-8")
		for match in TAG_PATTERN.finditer(text):
			tag = match.group("tag")
			if any(character.isupper() for character in tag):
				line = text.count("\n", 0, match.start()) + 1
				findings.append(f"{path.relative_to(ROOT)}:{line}: json:{tag}")

	if findings:
		for finding in findings:
			print(finding, file=sys.stderr)
		return 1

	print("JSON tag check passed")
	return 0


if __name__ == "__main__":
	raise SystemExit(main())
