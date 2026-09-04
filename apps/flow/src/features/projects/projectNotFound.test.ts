import fs from 'node:fs';
import path from 'node:path';
import { describe, expect, it } from 'vitest';

/**
 * A 500 is not a 404, and the sentence has to differ.
 *
 * `useCurrentProject` returned only `{ project, isLoading }`, so a failed
 * projects request left `project` undefined once `isLoading` cleared, and both
 * screens rendered the hardcoded English "Project not found." That told a user
 * their project had been deleted when the server had merely failed.
 *
 * These are source assertions because the branch is reached through a query
 * hook two levels up; the behaviour of the failure notice itself is covered by
 * the four screens in ListLoadFailed's own suite.
 */
const SCREENS = ['ProjectDashboard.tsx', 'ProjectTimeline.tsx'] as const;
const DIR = __dirname;

describe('a failed project load is not a missing project', () => {
  it('useCurrentProject reports the failure separately from the data', () => {
    const source = fs.readFileSync(path.join(DIR, 'useCurrentProject.ts'), 'utf8');
    expect(source).toContain('loadFailed: hasQueryFailed(projectsQuery)');
    expect(source).toContain('refetch: projectsQuery.refetch');
  });

  it.each(SCREENS)('%s renders the failure notice rather than a not-found line', (screen) => {
    const source = fs.readFileSync(path.join(DIR, screen), 'utf8');
    expect(source).toContain('if (projectLoadFailed)');
    expect(source).toContain('<ListLoadFailed onRetry=');
  });

  it.each(SCREENS)('%s no longer hardcodes the English sentence', (screen) => {
    const source = fs.readFileSync(path.join(DIR, screen), 'utf8');
    // Comments stripped first: the comment explaining the fix quotes the old
    // sentence, and a naive search for it matches that rather than any code.
    const code = source
      .replace(/\/\*[\s\S]*?\*\//g, '')
      .split('\n')
      .filter((line) => !line.trim().startsWith('//'))
      .join('\n');
    expect(code).not.toContain('Project not found.');
    expect(code).toContain("t('flow.projectNotFound')");
  });

  it.each(SCREENS)('%s checks the failure before the missing-project branch', (screen) => {
    // Order matters: with `!project` first, a failure would still take the
    // not-found path, because a failed request also leaves project undefined.
    const source = fs.readFileSync(path.join(DIR, screen), 'utf8');
    expect(source.indexOf('if (projectLoadFailed)')).toBeLessThan(source.indexOf('if (!project)'));
  });
});
