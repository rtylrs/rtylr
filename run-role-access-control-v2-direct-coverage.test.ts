import { spawnSync } from 'node:child_process';
import { readFileSync } from 'node:fs';
import path from 'node:path';
import { describe, expect, it } from 'vitest';

const SCRIPT = path.resolve(__dirname, 'scripts/run-role-access-control-v2-direct-coverage.sh');

// Only the fail-closed path is exercised. Supplying a password would send the
// script on to `test-db.sh` and the onion Go suite, which is not this test's
// job, so the happy path is covered by the static assertion below instead.
const runWithoutPassword = (env: Record<string, string | undefined>) =>
  spawnSync('bash', [SCRIPT], {
    encoding: 'utf8',
    env: { ...process.env, ...env },
  });

describe('run-role-access-control-v2-direct-coverage.sh', () => {
  it('refuses to run when DB_PASSWORD is absent', () => {
    const { DB_PASSWORD: _omitted, ...rest } = process.env;
    const result = spawnSync('bash', [SCRIPT], { encoding: 'utf8', env: rest });

    expect(result.status).toBe(1);
    expect(result.stderr).toContain('DB_PASSWORD is not set');
  });

  it('treats an empty DB_PASSWORD as absent rather than as a password', () => {
    const result = runWithoutPassword({ DB_PASSWORD: '' });

    expect(result.status).toBe(1);
    expect(result.stderr).toContain('DB_PASSWORD is not set');
  });

  it('fails before touching the onion checkout, so the guard cannot be skipped', () => {
    // ONION_DIR is deliberately nonsense. If the credential check ran after the
    // checkout was resolved, this would fail with a cd error instead.
    const result = runWithoutPassword({
      DB_PASSWORD: '',
      ONION_DIR: '/nonexistent-onion-checkout',
    });

    expect(result.status).toBe(1);
    expect(result.stderr).toContain('DB_PASSWORD is not set');
    expect(result.stderr).not.toContain('nonexistent-onion-checkout');
  });

  it('carries no default password in its source', () => {
    const source = readFileSync(SCRIPT, 'utf8');

    // A bash default-substitution on the password is the exact shape that put a
    // credential in this repository. Fail if it ever comes back, whatever the
    // value is, without this test needing to know the value.
    expect(source).not.toMatch(/DB_PASSWORD:-[^}]/);
    expect(source).toContain('export DB_PASSWORD\n');
  });
});
