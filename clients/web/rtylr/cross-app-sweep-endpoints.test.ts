import { spawnSync } from 'node:child_process';
import { cpSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { afterAll, describe, expect, it } from 'vitest';

// The sweep derives ROOT from its own location (apps/rtylr) but the shared
// packages stayed at the monorepo root, so it has to look outside the
// workspace for them. The fixtures reproduce that layout: a workspace under
// apps/rtylr with its own apps/ and docs/, and packages either at the
// monorepo root (current layout), under the workspace (pre-move layout), or
// missing entirely.
const WORKSPACE = dirname(fileURLToPath(import.meta.url));
const SCRIPT = join(WORKSPACE, 'scripts/cross-app/sweep-endpoints.mjs');

const fixtures: string[] = [];

afterAll(() => {
  for (const dir of fixtures) rmSync(dir, { recursive: true, force: true });
});

function makeFixture(packages: 'root' | 'local' | 'none') {
  const root = mkdtempSync(join(tmpdir(), 'sweep-endpoints-'));
  fixtures.push(root);
  const workspace = join(root, 'apps/rtylr');
  const script = join(workspace, 'scripts/cross-app/sweep-endpoints.mjs');
  const csv = join(workspace, 'docs/cross-app/inventory/endpoints.csv');

  mkdirSync(join(workspace, 'scripts/cross-app'), { recursive: true });
  mkdirSync(join(workspace, 'docs/cross-app/inventory'), { recursive: true });
  cpSync(SCRIPT, script);
  writeFileSync(join(workspace, 'docs/cross-app/inventory/contracts.csv'), 'domain,capability\n');
  mkdirSync(join(workspace, 'apps/dash/src'), { recursive: true });
  writeFileSync(join(workspace, 'apps/dash/src/client.ts'), 'safe.get("/license/state");\n');

  const writePackage = (base: string, file: string) => {
    mkdirSync(join(base, 'api/src'), { recursive: true });
    writeFileSync(join(base, 'api/src', file), 'fetch("/upload");\n');
  };
  if (packages === 'root') writePackage(join(root, 'packages'), 'shared.ts');
  if (packages === 'local') {
    writePackage(join(workspace, 'packages'), 'local.ts');
    writePackage(join(root, 'packages'), 'shared.ts');
  }
  return { script, csv };
}

function run(script: string) {
  return spawnSync(process.execPath, [script], { encoding: 'utf8' });
}

describe('sweep-endpoints.mjs', () => {
  it('resolves shared packages at the monorepo root', () => {
    const { script, csv } = makeFixture('root');
    const result = run(script);

    expect(result.stderr).toBe('');
    expect(result.status).toBe(0);
    const rows = readFileSync(csv, 'utf8');
    expect(rows).toContain('apps/dash/src/client.ts');
    expect(rows).toContain('packages/api/src/shared.ts');
  });

  it('keeps preferring a packages/ directory inside the workspace', () => {
    const { script, csv } = makeFixture('local');
    const result = run(script);

    expect(result.status).toBe(0);
    const rows = readFileSync(csv, 'utf8');
    expect(rows).toContain('packages/api/src/local.ts');
    expect(rows).not.toContain('packages/api/src/shared.ts');
  });

  it('fails with a clear message when no packages directory exists', () => {
    const { script } = makeFixture('none');
    const result = run(script);

    expect(result.status).toBe(1);
    expect(result.stderr).toContain('shared packages not found');
    expect(result.stderr).not.toContain('ENOENT');
  });
});
