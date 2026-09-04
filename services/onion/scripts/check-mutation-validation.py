#!/usr/bin/env python3
"""Ensure raw chi mutation routes are limited to transport-specific handlers."""

from __future__ import annotations

import re
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
MUTATION = re.compile(r"\b(?:Post|Put|Patch|Delete)\s*\(")

# These handlers cannot use the JSON framework wrapper because they own a
# protocol-specific request or response. Ordinary JSON mutations must use a
# Register*WithValidation helper instead.
ALLOWED_RAW_MUTATIONS = {
    "internal/modules/careers/internal/routes/routes.go": ("handlers.ApplyHandler",),
    "internal/modules/echo/module.go": ("streamHandler.ServeHTTP",),
    "internal/modules/hr_reports/internal/routes/routes.go": ("handler.ServeHTTP",),
    "internal/modules/pos_orders/internal/routes/routes.go": (
        "PaddlePOSWebhookHTTPHandler",
        "IssuePOSWebSocketTicket",
    ),
    "internal/modules/subscriptions/module.go": ("PaddleWebhookHTTPHandler",),
    "internal/modules/voxhire/internal/routes/routes.go": (
        "handlers.ApplicationsHandler",
        "handlers.CareerApplyHandler",
    ),
    "internal/modules/workspaces/module.go": ("realtime.IssueRealtimeTicket",),
}


def main() -> int:
    findings: list[str] = []
    for path in (ROOT / "internal/modules").rglob("*.go"):
        if path.name.endswith("_test.go") or path.name not in {"routes.go", "module.go"}:
            continue
        relative = str(path.relative_to(ROOT))
        allowed = ALLOWED_RAW_MUTATIONS.get(relative, ())
        for line_number, line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
            if not MUTATION.search(line):
                continue
            if "func " in line:
                continue
            if ".service." in line:
                continue
            if not any(marker in line for marker in allowed):
                findings.append(f"{relative}:{line_number}: {line.strip()}")

    if findings:
        for finding in findings:
            print(finding, file=sys.stderr)
        return 1

    print("mutation validation check passed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
