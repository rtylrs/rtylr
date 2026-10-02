import { describe, it, expect } from 'vitest';
import fs from 'node:fs';
import path from 'node:path';

/**
 * An empty list, a failed request and a loading list are three different facts
 * and must read differently (#1072). With every request forced to 500, Flow
 * told an owner "No tasks assigned - Nothing is on your plate right now. Nice
 * work!" and "All caught up!", which is worse than showing nothing: it reports
 * success while the server is down.
 *
 * This asserts the third state exists wherever the first two do. Most of these
 * screens need a dozen mocks to render, so a source assertion buys coverage
 * that behavioural tests would not repay, and it covers screens added later.
 */
const FEATURES = path.resolve(__dirname, '../../features');

function sourceFiles(dir: string, into: string[] = []): string[] {
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      sourceFiles(full, into);
      continue;
    }
    if (!/\.tsx$/.test(entry.name)) continue;
    if (/\.test\.tsx$/.test(entry.name)) continue;
    into.push(full);
  }
  return into;
}

/**
 * A screen qualifies when it both renders an empty state and reads a query
 * that can fail. A component handed its data as props has nothing to
 * distinguish and is not this file's business.
 */
function rendersAnEmptyStateForFetchedData(source: string): boolean {
  const hasEmptyState = /<EmptyState\b/.test(source);
  const readsAQuery = /\b(?:const|let)\s+\w*[Qq]uery\s*=|isLoading|isPending/.test(source);
  return hasEmptyState && readsAQuery;
}

describe('every screen with an empty state can also say the request failed', () => {
  const screens = sourceFiles(FEATURES).filter((file) =>
    rendersAnEmptyStateForFetchedData(fs.readFileSync(file, 'utf8'))
  );

  it('finds screens at all, so this file cannot pass by matching nothing', () => {
    expect(screens.length).toBeGreaterThanOrEqual(10);
  });

  it.each(screens.map((f) => [path.relative(FEATURES, f), f]))(
    '%s distinguishes a failed request from an empty list',
    (_name, file) => {
      const source = fs.readFileSync(file, 'utf8');
      // Either the shared third state, or the screen's own wording where the
      // failure has a specific cause worth naming. Directory and Presence
      // mention permissions, and Leave names which of its two lists failed.
      const hasFailureBranch =
        /ListLoadFailed/.test(source) || /hasQueryFailed|isError/.test(source);
      expect(hasFailureBranch, 'render a failed state, not the empty one').toBe(true);
    }
  );
});
