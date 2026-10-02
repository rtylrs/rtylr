#!/usr/bin/env node
/**
 * Regression tests for enforce-onekit-boundary.mjs.
 *
 * The guard is a gate, so "it printed a message" is not evidence: every case
 * here builds a throwaway git repository, plants a real violation, runs the
 * script as a subprocess and asserts the exit code. The three original defects
 * all produced exit 0 on a violation and are covered explicitly.
 *
 * Runs on Node's built-in test runner: `node --test scripts/enforce-onekit-boundary.test.mjs`
 */
import assert from 'node:assert/strict';
import { execFileSync, spawnSync } from 'node:child_process';
import { mkdtempSync, mkdirSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import test from 'node:test';

const script = join(dirname(fileURLToPath(import.meta.url)), 'enforce-onekit-boundary.mjs');
const packageRoot = join(dirname(fileURLToPath(import.meta.url)), '..');

/** A clean app file the guard must not flag. */
const CLEAN = 'export const value = 1;\n';

/**
 * The real violation this guard exists for: an app importing Axios instead of
 * going through the generated OneKit client. Reproduces the line that
 * `src/modules/recruit/shared/lib/api.ts` actually committed in 2a01cb638a.
 */
const VIOLATION =
  "import type { AxiosRequestConfig } from 'axios';\nexport const cfg: AxiosRequestConfig = {};\n";

const git = (repo, args) =>
  execFileSync('git', args, { cwd: repo, encoding: 'utf8', stdio: ['ignore', 'pipe', 'ignore'] });

/**
 * Creates a repository laid out like the real one (the client workspace sits at
 * clients/app, not at the repository root) with a passing baseline commit
 * and an `origin` remote, so base resolution has something real to resolve.
 */
function makeRepo() {
  const repo = mkdtempSync(join(tmpdir(), 'onekit-boundary-'));
  git(repo, ['init', '-q', '-b', 'main', '.']);
  git(repo, ['config', 'user.email', 'test@example.com']);
  git(repo, ['config', 'user.name', 'test']);
  git(repo, ['config', 'commit.gpgsign', 'false']);

  const appDir = join(repo, 'clients/app/entrypoints/auth');
  const app = join(appDir, 'src');
  mkdirSync(app, { recursive: true });
  mkdirSync(join(repo, 'clients/app/scripts'), { recursive: true });
  writeFileSync(join(app, 'clean.ts'), CLEAN);
  writeFileSync(
    join(appDir, 'package.json'),
    '{"name":"@rtylr/auth","dependencies":{"@voxire/voxie":"workspace:*"}}\n'
  );
  writeFileSync(
    join(repo, 'clients/app/scripts/enforce-onekit-boundary.mjs'),
    execFileSync(
      'node',
      [
        '-e',
        `process.stdout.write(require('node:fs').readFileSync(${JSON.stringify(script)}, 'utf8'))`,
      ],
      {
        encoding: 'utf8',
      }
    )
  );
  git(repo, ['add', '-A']);
  git(repo, ['commit', '-qm', 'baseline']);

  // A real base to diff against. Without it the guard has no origin to compare
  // to, and the correct behaviour is to fail loudly, not to pass.
  const origin = mkdtempSync(join(tmpdir(), 'onekit-origin-'));
  git(origin, ['init', '-q', '--bare', '.']);
  git(repo, ['remote', 'add', 'origin', origin]);
  git(repo, ['push', '-q', 'origin', 'main']);
  git(repo, ['branch', '--set-upstream-to=origin/main', 'main']);

  return { repo, origin, app, appDir };
}

/** Runs the guard from the client workspace, the way `pnpm contracts:enforce` does. */
function runGuard(repo, args = ['--changed-only']) {
  return spawnSync('node', ['scripts/enforce-onekit-boundary.mjs', ...args], {
    cwd: join(repo, 'clients/app'),
    encoding: 'utf8',
    // The ambient GITHUB_* variables would otherwise decide the base.
    env: { ...process.env, GITHUB_BASE_REF: '', GITHUB_EVENT_BEFORE: '', GITHUB_SHA: '' },
  });
}

const withRepo = (fn) => {
  const fixture = makeRepo();
  try {
    return fn(fixture);
  } finally {
    rmSync(fixture.repo, { recursive: true, force: true });
    rmSync(fixture.origin, { recursive: true, force: true });
  }
};

test('a clean tree passes', () => {
  withRepo(({ repo }) => {
    const result = runGuard(repo);
    assert.equal(result.status, 0, result.stderr);
    assert.match(result.stdout, /no new app transport violations found/);
  });
});

test('a STAGED violation fails: the pre-commit hook runs after git add', () => {
  withRepo(({ repo, app }) => {
    writeFileSync(join(app, 'ClockIn.ts'), VIOLATION);
    git(repo, ['add', 'clients/app/entrypoints/auth/src/ClockIn.ts']);

    const result = runGuard(repo);
    assert.equal(result.status, 1, `staged violation was not caught: ${result.stdout}`);
    assert.match(result.stderr, /entrypoints\/auth\/src\/ClockIn\.ts/);
    assert.match(result.stderr, /AXIOS_IMPORT/);
  });
});

test('an UNSTAGED violation in a tracked file fails', () => {
  withRepo(({ repo, app }) => {
    writeFileSync(join(app, 'clean.ts'), `${CLEAN}\nexport const probe = () => fetch('/probe');\n`);

    const result = runGuard(repo);
    assert.equal(result.status, 1, `unstaged violation was not caught: ${result.stdout}`);
    assert.match(result.stderr, /RAW_FETCH/);
  });
});

test('an UNTRACKED app source file is checked as new', () => {
  withRepo(({ repo, app }) => {
    writeFileSync(join(app, 'New.ts'), VIOLATION);

    const result = runGuard(repo);
    assert.equal(result.status, 1, `untracked violation was not caught: ${result.stdout}`);
  });
});

test('a violation COMMITTED on the branch is caught by diffing against the base', () => {
  withRepo(({ repo, app }) => {
    writeFileSync(join(app, 'ClockIn.ts'), VIOLATION);
    git(repo, ['add', '-A']);
    git(repo, ['commit', '-qm', 'introduce axios import']);

    // The base moved with the commit, so only an unresolvable-base case differs
    // below; here the branch is ahead of origin/main by one commit.
    const result = runGuard(repo);
    assert.equal(result.status, 1, `committed violation was not caught: ${result.stdout}`);
    assert.match(result.stderr, /AXIOS_IMPORT/);
  });
});

test('a pre-existing violation on the base is not re-reported', () => {
  withRepo(({ repo, app }) => {
    // The violation is already on the base, so it is not something this branch
    // introduced. The guard reports added lines, not pre-existing ones.
    writeFileSync(join(app, 'Legacy.ts'), VIOLATION);
    git(repo, ['add', '-A']);
    git(repo, ['commit', '-qm', 'add axios import']);
    git(repo, ['push', '-q', 'origin', 'main']);

    writeFileSync(join(app, 'clean.ts'), `${CLEAN}export const other = 2;\n`);

    const result = runGuard(repo);
    assert.equal(result.status, 0, `pre-existing violation was re-reported: ${result.stderr}`);
  });
});

test('removing a violation that this branch added passes', () => {
  withRepo(({ repo, app }) => {
    writeFileSync(join(app, 'Legacy.ts'), VIOLATION);
    git(repo, ['add', '-A']);
    git(repo, ['commit', '-qm', 'add axios import']);
    // Deleting the file drops its lines from the net diff against the base, so
    // the branch no longer introduces a violation.
    rmSync(join(app, 'Legacy.ts'));
    git(repo, ['add', '-A']);
    git(repo, ['commit', '-qm', 'remove axios import']);

    const result = runGuard(repo);
    assert.equal(result.status, 0, `removing a violation should pass: ${result.stderr}`);
  });
});

test('an unresolvable base fails loudly instead of passing', () => {
  withRepo(({ repo, app }) => {
    writeFileSync(join(app, 'ClockIn.ts'), VIOLATION);
    git(repo, ['add', '-A']);
    git(repo, ['commit', '-qm', 'introduce axios import']);
    // Detach and drop every remote, so no base can be resolved.
    git(repo, ['checkout', '-q', '--detach', 'HEAD']);
    git(repo, ['remote', 'remove', 'origin']);

    const result = runGuard(repo);
    assert.equal(result.status, 2, `unresolvable base must exit 2, got ${result.status}`);
    assert.match(result.stderr, /no diff base could be resolved/);
  });
});

test('a shallow clone fails loudly instead of silently passing', () => {
  withRepo(({ repo, origin }) => {
    // Clone depth 1 from the local origin: the merge base is unreachable, which
    // is the state CI checkouts are in. The guard must not report "no
    // violations" just because it could not see the branch.
    const shallow = mkdtempSync(join(tmpdir(), 'onekit-shallow-'));
    try {
      git(shallow, ['init', '-q', '-b', 'main', '.']);
      git(shallow, ['config', 'user.email', 'test@example.com']);
      git(shallow, ['config', 'user.name', 'test']);
      git(shallow, ['remote', 'add', 'origin', origin]);
      git(shallow, ['fetch', '-q', '--depth=1', 'origin', 'main']);
      git(shallow, ['checkout', '-q', 'main']);

      const app = join(shallow, 'clients/app/entrypoints/auth/src');
      writeFileSync(join(app, 'ClockIn.ts'), VIOLATION);
      git(shallow, ['add', '-A']);
      git(shallow, ['commit', '-qm', 'introduce axios import']);

      const result = runGuard(shallow);
      assert.notEqual(result.status, 0, 'a shallow clone must not silently pass');
      assert.match(result.stderr, /shallow|base/i);
    } finally {
      rmSync(shallow, { recursive: true, force: true });
    }
  });
});

test('the workspace-relative apps/ prefix is resolved against the repository root', () => {
  withRepo(({ repo, app }) => {
    // The real repository has no top-level apps/ directory; the guard must
    // still match clients/app/apps.
    writeFileSync(join(app, 'ClockIn.ts'), VIOLATION);
    git(repo, ['add', '-A']);
    const result = runGuard(repo);
    assert.equal(result.status, 1, `apps/ prefix did not match: ${result.stdout}`);
  });
});

test('an app without the OneKit dependency is reported', () => {
  withRepo(({ repo, appDir }) => {
    writeFileSync(join(appDir, 'package.json'), '{"name":"@rtylr/auth","dependencies":{}}\n');
    git(repo, ['add', '-A']);
    const result = runGuard(repo);
    assert.equal(result.status, 1);
    assert.match(result.stderr, /MISSING_ONEKIT_DEPENDENCY/);
  });
});

test('--all scans the working tree regardless of the base', () => {
  withRepo(({ repo, app }) => {
    writeFileSync(join(app, 'Legacy.ts'), "import axios from 'axios';\n");
    git(repo, ['remote', 'remove', 'origin']);

    const result = runGuard(repo, ['--all']);
    assert.equal(result.status, 1, `--all did not find the violation: ${result.stdout}`);
  });
});

test('the allowlisted raw fetch call is not reported', () => {
  withRepo(({ repo }) => {
    const file = join(repo, 'clients/app/src/modules/flow/shared/lib/realtime.ts');
    mkdirSync(join(file, '..'), { recursive: true });
    writeFileSync(
      join(repo, 'clients/app/package.json'),
      '{"name":"@rtylr/portal","dependencies":{"@voxire/voxie":"workspace:*"}}\n'
    );
    writeFileSync(file, "export const send = () => fetch('/ws');\n");
    git(repo, ['add', '-A']);

    const result = runGuard(repo);
    assert.equal(result.status, 0, `allowlisted file was reported: ${result.stderr}`);
  });
});

test('the guard the package script invokes exists at the expected path', () => {
  const manifest = JSON.parse(
    execFileSync(
      'node',
      [
        '-e',
        `process.stdout.write(require('node:fs').readFileSync('${join(packageRoot, 'package.json')}', 'utf8'))`,
      ],
      {
        encoding: 'utf8',
      }
    )
  );
  assert.equal(
    manifest.scripts['contracts:enforce'],
    'node scripts/enforce-onekit-boundary.mjs --changed-only'
  );
});

test('large migrations still inspect staged and untracked transport violations', () => {
  withRepo(({ repo, app }) => {
    writeFileSync(join(app, 'large.ts'), '// migration padding\n'.repeat(140_000) + VIOLATION);
    for (const staged of [false, true]) {
      if (staged) git(repo, ['add', 'clients/app/entrypoints/auth/src/large.ts']);
      const result = runGuard(repo);
      assert.equal(result.status, 1, result.stderr);
      assert.match(result.stderr, /AXIOS_IMPORT/);
      assert.doesNotMatch(result.stderr, /RangeError/);
    }
  });
});
