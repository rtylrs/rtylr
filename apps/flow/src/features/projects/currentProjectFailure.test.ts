import { describe, it, expect } from 'vitest';
import fs from 'node:fs';
import path from 'node:path';

/**
 * `useCurrentProject` returns `project: undefined` for two entirely different
 * facts: the project does not exist, and the project list never arrived. A
 * screen that checks only `!project` renders "Project not found." for a 500,
 * telling the owner their project has gone and offering no way back (#1072).
 *
 * The hook exposes `loadFailed` and `refetch` precisely so the two can be told
 * apart. This asserts every consumer actually uses them, which the behavioural
 * tests cannot do for screens whose full render needs a dozen mocks.
 */
const FEATURES = path.resolve(__dirname, '..');

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

describe('every screen that reads useCurrentProject separates missing from failed', () => {
  const consumers = sourceFiles(FEATURES).filter((file) =>
    /useCurrentProject\(/.test(fs.readFileSync(file, 'utf8'))
  );

  it('finds the consumers at all, so this file cannot pass by matching nothing', () => {
    expect(consumers.length).toBeGreaterThanOrEqual(5);
  });

  it.each(consumers.map((f) => [path.relative(FEATURES, f), f]))(
    '%s renders the failed state rather than "not found"',
    (_name, file) => {
      const source = fs.readFileSync(file, 'utf8');
      // A screen that never says "not found" has nothing to disambiguate: it
      // shows the project inline and an absent one is not its concern.
      if (!/project\.notFound|projectNotFound/.test(source)) return;

      expect(source, 'destructure loadFailed from useCurrentProject').toMatch(/loadFailed/);
      expect(source, 'render ListLoadFailed when it is set').toMatch(/ListLoadFailed/);
    }
  );
});
