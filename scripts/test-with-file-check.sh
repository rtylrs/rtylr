#!/usr/bin/env bash
# Runs the suite, then checks it actually executed every file it collected.
#
# Both run unconditionally, and the worse status wins. Chaining with `&&`
# looked tidier and was wrong twice over: a currently-failing test on main
# (REQ-R23) would mean the file check never executes at all, and chaining with
# `;` would let the check's exit code mask a genuine test failure.
set -uo pipefail

vitest run --reporter=default --reporter=json --outputFile.json=.vitest-report.json
suite=$?

node scripts/assert-every-test-file-ran.mjs
files=$?

if [ "$suite" -ne 0 ]; then
  exit "$suite"
fi
exit "$files"
