#!/usr/bin/env node
/**
 * Fails when the test run executed fewer files than it collected.
 *
 * Vitest collects the file list before it runs anything, then reports on what
 * it actually executed. When a worker dies, its queued files leave the run
 * **without the run complaining**: the reporter says "693 passed" and exits
 * as though that were the whole suite.
 *
 * That was observed three times in one session, twice on a gate run that was
 * then quoted as evidence:
 *
 *   collected 695, executed 686  (8 files, ~100 tests, under CPU contention)
 *   collected 695, executed 692  (2 files, at --maxWorkers=2)
 *   collected 695, executed 694  (1 file, machine not busy)
 *
 * Load widens the gap; it does not cause it. Until a run can prove it ran
 * everything it collected, "one flaky test" and "one file that never ran" are
 * indistinguishable from the summary line, which is what made rtylr#988
 * undiagnosable for so long.
 *
 * Reads the JSON report `pnpm test` already writes, so it costs one extra
 * `vitest list` rather than a second run.
 */
import { execFileSync } from 'node:child_process';
import { readFileSync } from 'node:fs';

const REPORT = '.vitest-report.json';

function collectedFileCount() {
  const out = execFileSync('npx', ['vitest', 'list', '--filesOnly'], {
    encoding: 'utf8',
    maxBuffer: 32 * 1024 * 1024,
  });
  return out.split('\n').filter((line) => /\.test\.tsx?$/.test(line.trim())).length;
}

function executedFiles() {
  let report;
  try {
    report = JSON.parse(readFileSync(REPORT, 'utf8'));
  } catch (error) {
    console.error(
      `Could not read ${REPORT}. Run this after \`pnpm test\`, which writes it.\n${String(error)}`
    );
    process.exit(2);
  }
  const results = Array.isArray(report.testResults) ? report.testResults : [];
  return new Set(results.map((r) => String(r.name ?? '')).filter(Boolean));
}

const collected = collectedFileCount();
const executed = executedFiles();

if (collected === 0) {
  console.error('Collected zero test files. That is a configuration problem, not a clean run.');
  process.exit(2);
}

if (executed.size >= collected) {
  console.log(`Every collected test file ran (${executed.size} of ${collected}).`);
  process.exit(0);
}

console.error(
  [
    '',
    `Test files collected: ${collected}`,
    `Test files executed:  ${executed.size}`,
    `MISSING:              ${collected - executed.size}`,
    '',
    'The run reported on fewer files than it collected, so its result does not',
    'cover the whole suite. This is usually a worker dying and taking its queued',
    'files with it; try --maxWorkers=2 to confirm, and see rtylr#988.',
    '',
  ].join('\n')
);
process.exit(1);
