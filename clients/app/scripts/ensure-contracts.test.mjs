#!/usr/bin/env node
/**
 * Regression tests for ensure-contracts.sh.
 *
 * The defect this guards against is a gate that cannot fail: the original script
 * only asked whether voxie/gen and clients/packages/contracts/src EXISTED, so a
 * present-but-stale tree passed and an orphan route reached every downstream
 * typecheck. "It printed a message" is therefore not evidence here, and neither
 * is "it exited 0 on the real tree": every case below builds a throwaway git
 * repository with a real fixture, runs the script as a subprocess, and asserts
 * the exit code.
 *
 * Fixtures are synthetic contracts, not the real ones, so the orphan injected
 * here is a deliberate drift rather than the one currently in the work tree.
 * That keeps the test deterministic instead of coupled to generated output.
 *
 * Runs on Node's built-in test runner:
 *   node --test scripts/ensure-contracts.test.mjs
 */
import assert from 'node:assert/strict';
import { execFileSync, spawnSync } from 'node:child_process';
import {
  chmodSync,
  mkdirSync,
  mkdtempSync,
  readFileSync,
  rmSync,
  symlinkSync,
  writeFileSync,
} from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import test from 'node:test';

const script = join(dirname(fileURLToPath(import.meta.url)), 'ensure-contracts.sh');
const bash = execFileSync('sh', ['-c', 'command -v bash'], { encoding: 'utf8' }).trim();

const git = (repo, args) =>
  execFileSync('git', args, { cwd: repo, encoding: 'utf8', stdio: ['ignore', 'pipe', 'ignore'] });

/** A .onk service declaring two operations. */
const SERVICE_ONK = `package demo.v1

/// DemoService defines demo endpoints.
service DemoService {
  /// Reads a widget.
  getWidget(GetWidgetRequest) -> GetWidgetResponse @get("/widget")

  /// Writes a widget.
  putWidget(PutWidgetRequest) -> PutWidgetResponse @put("/widget")
}
`;

/**
 * A generated server for the two operations above, in the shape the generator
 * emits. `extra` appends operations that no committed .onk declares, which is
 * the orphan-route defect.
 */
const serverGenGo = (extra = '') =>
  `package v1

func (o *DemoServiceServer) RegisterService(mux *http.ServeMux) {
	mux.Handle("GET /demo/v1/widget", o.wrapHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})), RequestMetadata{Service: "DemoService", Method: "getWidget", HTTPMethod: "GET", Route: "/demo/v1/widget", AuthSchemes: nil}))
	mux.Handle("PUT /demo/v1/widget", o.wrapHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})), RequestMetadata{Service: "DemoService", Method: "putWidget", HTTPMethod: "PUT", Route: "/demo/v1/widget", AuthSchemes: nil}))
${extra}}
`;

/** The orphan: a route and operation that service.onk above never declares. */
const ORPHAN_ROUTE = `	mux.Handle("PATCH /demo/v1/organization/mfa", o.wrapHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})), RequestMetadata{Service: "DemoService", Method: "updateOrganizationMFAEnforcement", HTTPMethod: "PATCH", Route: "/demo/v1/organization/mfa", AuthSchemes: nil}))`;

/**
 * A repository laid out like the real one. `extra` is appended to the generated
 * server, and `omitGenerated` leaves the generated tree absent, so one fixture
 * builder covers a fresh tree, a stale tree and a never-generated checkout.
 */
function makeRepo({ extra = '', omitGenerated = false } = {}) {
  const repo = mkdtempSync(join(tmpdir(), 'ensure-contracts-'));
  git(repo, ['init', '-q', '-b', 'main', '.']);
  git(repo, ['config', 'user.email', 'test@example.com']);
  git(repo, ['config', 'user.name', 'test']);
  git(repo, ['config', 'commit.gpgsign', 'false']);

  const api = join(repo, 'voxie/api/demo/v1');
  mkdirSync(api, { recursive: true });
  writeFileSync(join(repo, 'voxie/onekit.toml'), 'schema_root = "api"\n');
  writeFileSync(join(api, 'service.onk'), SERVICE_ONK);

  if (!omitGenerated) {
    const gen = join(repo, 'voxie/gen/go/demo/v1');
    mkdirSync(gen, { recursive: true });
    writeFileSync(join(gen, 'server.gen.go'), serverGenGo(extra));
    mkdirSync(join(repo, 'clients/packages/contracts/src/demo/v1'), { recursive: true });
    writeFileSync(
      join(repo, 'clients/packages/contracts/src/demo/v1/client.ts'),
      'export const x = 1;\n'
    );
  }

  git(repo, ['add', '-A']);
  git(repo, ['commit', '-qm', 'baseline']);
  return repo;
}

const withRepo = (options, fn) => {
  const repo = makeRepo(options);
  try {
    return fn(repo);
  } finally {
    rmSync(repo, { recursive: true, force: true });
  }
};

/** Runs the script from the client workspace, the way `pnpm typecheck` does. */
const runCheck = (repo, args = []) =>
  spawnSync('bash', [script, ...args], { cwd: repo, encoding: 'utf8' });

test('a fresh generated tree matching the .onk sources passes', () => {
  withRepo({}, (repo) => {
    const result = runCheck(repo);
    assert.equal(result.status, 0, result.stderr);
    assert.match(result.stdout, /Voxie contracts match voxie\/api/);
  });
});

test('a STALE tree with an orphan route fails: the original existence check passed this', () => {
  withRepo({ extra: `${ORPHAN_ROUTE}\n` }, (repo) => {
    const result = runCheck(repo, ['--no-generate']);
    assert.equal(result.status, 1, `orphan route was not caught: ${result.stdout}${result.stderr}`);
    assert.match(result.stderr, /\(orphan route\): updateOrganizationMFAEnforcement/);
  });
});

test('the stale failure names the remedy', () => {
  withRepo({ extra: `${ORPHAN_ROUTE}\n` }, (repo) => {
    const result = runCheck(repo, ['--no-generate']);
    assert.match(result.stderr, /make -C voxie clean generate/);
  });
});

test('a declared operation missing from the generated server fails', () => {
  withRepo({}, (repo) => {
    // Drop the putWidget registration: the .onk still declares it.
    const gen = join(repo, 'voxie/gen/go/demo/v1/server.gen.go');
    const trimmed = readFileSync(gen, 'utf8')
      .split('\n')
      .filter((line) => !line.includes('putWidget'))
      .join('\n');
    writeFileSync(gen, trimmed);

    const result = runCheck(repo, ['--no-generate']);
    assert.equal(result.status, 1, `missing operation was not caught: ${result.stdout}`);
    assert.match(result.stderr, /absent from generated output: putWidget/);
  });
});

test('a generated server with no committed .onk fails: an orphan in the other direction', () => {
  withRepo({}, (repo) => {
    // An extra generated package the contracts never declared.
    const orphanGen = join(repo, 'voxie/gen/go/ghost/v1');
    mkdirSync(orphanGen, { recursive: true });
    writeFileSync(join(orphanGen, 'server.gen.go'), serverGenGo());

    const result = runCheck(repo, ['--no-generate']);
    assert.equal(result.status, 1, `unbacked generated service was not caught: ${result.stdout}`);
    assert.match(
      result.stderr,
      /ghost\/v1\/server\.gen\.go is generated but no service\.onk declares it/
    );
  });
});

test('a declared service with no generated server fails', () => {
  withRepo({}, (repo) => {
    mkdirSync(join(repo, 'voxie/api/other/v1'), { recursive: true });
    writeFileSync(join(repo, 'voxie/api/other/v1/service.onk'), SERVICE_ONK);

    const result = runCheck(repo, ['--no-generate']);
    assert.equal(result.status, 1, `ungenerated service was not caught: ${result.stdout}`);
    assert.match(result.stderr, /gen\/go\/other\/v1\/server\.gen\.go is missing/);
  });
});

test('a tree with no generated output fails instead of passing as a no-op', () => {
  withRepo({ omitGenerated: true }, (repo) => {
    const result = runCheck(repo, ['--no-generate']);
    assert.equal(result.status, 1, `missing output was not caught: ${result.stdout}`);
    assert.match(result.stderr, /generated contracts are missing/);
  });
});

test('a missing generator tool is a distinguishable exit 3, not a silent pass', () => {
  withRepo({ extra: `${ORPHAN_ROUTE}\n` }, (repo) => {
    // A PATH carrying everything the parity check needs (git, find, sort, comm,
    // mktemp, grep, sed, dirname) but not the generator toolchain (make, go, python3).
    // Regeneration is impossible, so the check must say so rather than quietly
    // reporting success.
    const bin = mkdtempSync(join(tmpdir(), 'ensure-contracts-bin-'));
    try {
      for (const tool of [
        'git',
        'find',
        'sort',
        'comm',
        'mktemp',
        'grep',
        'sed',
        'tr',
        'cat',
        'dirname',
        'rm',
      ]) {
        const resolved = execFileSync('sh', ['-c', `command -v ${tool}`], {
          encoding: 'utf8',
        }).trim();
        symlinkSync(resolved, join(bin, tool));
      }
      const result = spawnSync(bash, [script], {
        cwd: repo,
        encoding: 'utf8',
        env: { ...process.env, PATH: bin },
      });
      assert.equal(
        result.status,
        3,
        `absent tooling must exit 3, got ${result.status}; signal=${result.signal}; error=${result.error}; stderr=${result.stderr}`
      );
      assert.match(result.stderr, /not on PATH/);
      assert.match(result.stderr, /not silently skipped/);
    } finally {
      rmSync(bin, { recursive: true, force: true });
    }
  });
});

test('a failing generator exits 4 rather than being treated as success', () => {
  withRepo({ omitGenerated: true }, (repo) => {
    // Stand in for a broken `make -C voxie generate`.
    writeFileSync(
      join(repo, 'voxie/Makefile'),
      'generate:\n\t@echo "generator exploded" >&2\n\t@exit 1\n'
    );
    const result = runCheck(repo);
    assert.equal(result.status, 4, `failing generation must exit 4, got ${result.status}`);
    assert.match(result.stderr, /failed; see its output above/);
  });
});

test('a generator that leaves no output exits 4', () => {
  withRepo({ omitGenerated: true }, (repo) => {
    // Succeeds but produces nothing, so success is not proof of a fresh tree.
    writeFileSync(join(repo, 'voxie/Makefile'), 'generate:\n\t@true\n');
    const result = runCheck(repo);
    assert.equal(result.status, 4, `a no-op generator must exit 4, got ${result.status}`);
    assert.match(result.stderr, /left no generated output/);
  });
});

test('regeneration runs when the tree is stale and the result is re-verified', () => {
  withRepo({ extra: `${ORPHAN_ROUTE}\n` }, (repo) => {
    // A stand-in generator that rewrites the tree the way voxie/Makefile does:
    // rm -rf first, then emit only what the .onk declares.
    const genFile = join(repo, 'voxie/gen/go/demo/v1/server.gen.go');
    // A stand-in generator that rewrites the tree the way voxie/Makefile does:
    // rm -rf first, then emit only what the .onk declares. make runs with
    // $(CURDIR) at voxie/, so the fixture source is addressed from there.
    writeFileSync(
      join(repo, 'voxie/Makefile'),
      'generate:\n\trm -rf gen ../clients/packages/contracts/src\n\tmkdir -p gen/go/demo/v1 ../clients/packages/contracts/src/demo/v1\n\tcp $(CURDIR)/../.fixture-server.gen.go gen/go/demo/v1/server.gen.go\n\ttouch ../clients/packages/contracts/src/demo/v1/client.ts\n'
    );
    writeFileSync(join(repo, '.fixture-server.gen.go'), serverGenGo());

    const result = runCheck(repo);
    assert.equal(result.status, 0, `${result.stdout}${result.stderr}`);
    assert.match(result.stderr, /Generating Voxie contracts/);
    // The orphan is gone: the generator removed it and the check confirmed it.
    assert.doesNotMatch(readFileSync(genFile, 'utf8'), /updateOrganizationMFAEnforcement/);
  });
});

test('regeneration that leaves the drift in place still fails', () => {
  withRepo({ extra: `${ORPHAN_ROUTE}\n` }, (repo) => {
    // Generation "succeeds" but leaves the stale content in place, which is the
    // case a green `make` would otherwise hide. The re-verification must catch
    // it, so this is exit 1 (stale), not 0.
    writeFileSync(join(repo, 'voxie/Makefile'), 'generate:\n\t@true\n');

    const result = runCheck(repo);
    assert.equal(result.status, 1, `a no-op generator must not pass, got ${result.status}`);
    assert.match(result.stderr, /regeneration did not resolve the drift/);
  });
});

test('an unknown argument fails with the layout exit code', () => {
  withRepo({}, (repo) => {
    const result = runCheck(repo, ['--not-a-flag']);
    assert.equal(result.status, 2, `unknown argument must exit 2, got ${result.status}`);
    assert.match(result.stderr, /unknown argument/);
  });
});

test('--help exits 0 and documents the exit codes and the remedy', () => {
  withRepo({}, (repo) => {
    const result = runCheck(repo, ['--help']);
    assert.equal(result.status, 0, result.stderr);
    assert.match(result.stdout, /make -C voxie clean generate/);
    assert.match(result.stdout, /3 {2}a required tool is unavailable/);
  });
});

test('outside a git work tree the check fails loudly', () => {
  const notARepo = mkdtempSync(join(tmpdir(), 'ensure-contracts-nogit-'));
  try {
    const result = spawnSync('bash', [script], {
      cwd: notARepo,
      encoding: 'utf8',
      // `git rev-parse` must not succeed by walking up into an unrelated repo.
      env: { ...process.env, GIT_CEILING_DIRECTORIES: notARepo },
    });
    assert.equal(result.status, 2, `expected the layout exit code, got ${result.status}`);
    assert.match(result.stderr, /not inside a git work tree/);
  } finally {
    rmSync(notARepo, { recursive: true, force: true });
  }
});

test('the script is executable and syntactically valid', () => {
  const syntax = spawnSync('bash', ['-n', script], { encoding: 'utf8' });
  assert.equal(syntax.status, 0, syntax.stderr);
  chmodSync(script, 0o755);
});
