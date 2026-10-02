import { spawnSync } from 'node:child_process';
import { cpSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { afterAll, describe, expect, it } from 'vitest';

// Mirror the single Portal package and its sibling shared packages.
const WORKSPACE = dirname(fileURLToPath(import.meta.url));
const SCRIPT = join(WORKSPACE, 'scripts/cross-app/sweep-endpoints.mjs');

const fixtures: string[] = [];

afterAll(() => {
  for (const dir of fixtures) rmSync(dir, { recursive: true, force: true });
});

function makeFixture(packages: 'root' | 'local' | 'none') {
  const root = mkdtempSync(join(tmpdir(), 'sweep-endpoints-'));
  fixtures.push(root);
  const workspace = join(root, 'clients/app');
  const script = join(workspace, 'scripts/cross-app/sweep-endpoints.mjs');
  const csv = join(workspace, 'docs/cross-app/inventory/endpoints.csv');

  mkdirSync(join(workspace, 'scripts/cross-app'), { recursive: true });
  mkdirSync(join(workspace, 'docs/cross-app/inventory'), { recursive: true });
  cpSync(SCRIPT, script);
  mkdirSync(join(workspace, 'entrypoints'), { recursive: true });
  writeFileSync(join(workspace, 'docs/cross-app/inventory/contracts.csv'), 'domain,capability\n');
  mkdirSync(join(workspace, 'src'), { recursive: true });
  writeFileSync(join(workspace, 'src/client.ts'), 'safe.get("/health");\n');

  const writePackage = (base: string, file: string) => {
    mkdirSync(join(base, 'api/src'), { recursive: true });
    writeFileSync(join(base, 'api/src', file), 'fetch("/upload");\n');
  };
  if (packages === 'root') writePackage(join(root, 'clients/packages'), 'shared.ts');
  if (packages === 'local') {
    writePackage(join(workspace, 'packages'), 'local.ts');
    writePackage(join(root, 'clients/packages'), 'shared.ts');
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
    expect(rows).toContain('src/client.ts');
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
