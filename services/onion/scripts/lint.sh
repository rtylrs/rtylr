#!/bin/bash
set -euo pipefail

# In CI we want a clean failure with full context (no --fix, all issues, non-zero exit).
# Locally we --fix automatically and still propagate the exit code.
if [ "${CI:-false}" = "true" ]; then
    if [ -n "${LINT_NEW_FROM_REV:-}" ] && [ "${LINT_NEW_FROM_REV}" != "0000000000000000000000000000000000000000" ]; then
        exec golangci-lint run --new-from-rev="${LINT_NEW_FROM_REV}" --issues-exit-code=1 --max-issues-per-linter=0 --max-same-issues=0
    fi

    exec golangci-lint run --issues-exit-code=1 --max-issues-per-linter=0 --max-same-issues=0
fi

exec golangci-lint run --fix --issues-exit-code=1
