import { spawnSync } from 'node:child_process';
import { chmodSync, mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { afterAll, describe, expect, it } from 'vitest';

// The onion checkout is not part of this repository, so the gate has to be
// told where it is. These fixtures stand in for it: one missing path, one
// stub checkout that the gate must actually read.
const WORKSPACE = dirname(fileURLToPath(import.meta.url));
const SCRIPT = join(WORKSPACE, 'scripts/validate-role-access-control-v2-evidence.mjs');

const HARNESS_OPERATIONS = [
  'resource scope denial',
  'wrong tenant or cross-business target',
  'stale authorization version',
  'stale session',
  'revoked assignment',
  'archived role',
  'invalid group source',
  'deactivated user',
  'terminated employee',
  'delegation ceiling',
];

const fixture = mkdtempSync(join(tmpdir(), 'rbac-evidence-'));
afterAll(() => rmSync(fixture, { recursive: true, force: true }));

function run(env: Record<string, string>) {
  return spawnSync(process.execPath, [SCRIPT], {
    encoding: 'utf8',
    env: { ...process.env, ...env },
  });
}

function makeStubOnion() {
  const onion = join(fixture, 'onion');
  mkdirSync(join(onion, 'internal/testutil'), { recursive: true });
  mkdirSync(join(onion, 'internal/modules'), { recursive: true });
  writeFileSync(
    join(onion, 'internal/testutil/direct_authorization.go'),
    HARNESS_OPERATIONS.map((operation) => `// ${operation}`).join('\n')
  );
  return onion;
}

describe('validate-role-access-control-v2-evidence.mjs', () => {
  it('reports a missing onion checkout instead of crashing', () => {
    const missing = join(fixture, 'missing-onion');
    const result = run({ ONION_DIR: missing });

    expect(result.status).toBe(1);
    expect(result.stderr).toContain(`onion checkout not found at ${missing}`);
    expect(result.stderr).toContain('set ONION_DIR');
    expect(result.stderr).not.toContain('ENOENT');
  });

  it('validates against the checkout named by ONION_DIR', () => {
    // rg is only used to list the direct-test files; a stub keeps this test
    // independent of whether ripgrep is installed.
    const bin = join(fixture, 'bin');
    mkdirSync(bin, { recursive: true });
    const rg = join(bin, 'rg');
    writeFileSync(rg, '#!/bin/sh\nexit 0\n');
    chmodSync(rg, 0o755);

    const onion = makeStubOnion();
    const result = run({ ONION_DIR: onion, PATH: `${bin}:${process.env.PATH ?? ''}` });

    expect(result.stderr).toContain('RBAC evidence validation failed');
    expect(result.stderr).not.toContain('direct authorization harness is missing');
    expect(result.stderr).not.toContain('onion checkout not found');
    expect(result.status).toBe(1);
  });
});
