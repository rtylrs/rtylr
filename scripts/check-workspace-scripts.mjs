#!/usr/bin/env node
/**
 * Guards the failure that broke a production deploy.
 *
 * `pnpm --filter <app> run build` puts the package's own node_modules/.bin on
 * PATH, not the workspace root's and not another workspace member's. auth and
 * menu declared no devDependencies at all and relied on the rtylr workspace
 * root's linked binaries, so their recursive build worked while the deploy
 * pipeline's per-app build died with "vite: not found".
 *
 * This checks every workspace member's scripts for:
 *   - a binary that the package does not declare (directly or via the
 *     repository root), and
 *   - a referenced `scripts/...` file that does not exist.
 */
import { existsSync, readFileSync, readdirSync, statSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const ROOT = resolve(dirname(fileURLToPath(import.meta.url)), '..');

/** Binary name -> the package that provides it. */
const BINARY_PACKAGE = {
  tsc: 'typescript',
  vite: 'vite',
  vitest: 'vitest',
  next: 'next',
  playwright: '@playwright/test',
  tauri: '@tauri-apps/cli',
  tsx: 'tsx',
  wrangler: 'wrangler',
  serve: 'serve',
  eslint: 'eslint',
};

/** Commands that come from the system or the package manager, not a package. */
const SYSTEM_COMMANDS = new Set([
  'cd', 'echo', 'printf', 'node', 'pnpm', 'npx', 'npm', 'bash', 'sh', 'rm',
  'cp', 'mv', 'mkdir', 'true', 'false', 'exit', 'cargo', 'git', 'python3',
  'docker', 'aws', 'gh', 'test', 'for', 'do', 'done', 'if', 'then', 'fi',
  'set', 'source', 'wget', 'sleep', 'chmod',
]);

function workspaceMembers() {
  const dirs = ['apps/rtylr/apps', 'apps', 'packages'];
  const seen = new Set();
  for (const dir of dirs) {
    const abs = join(ROOT, dir);
    if (!existsSync(abs)) continue;
    for (const entry of readdirSync(abs)) {
      const pkg = join(abs, entry, 'package.json');
      if (existsSync(pkg) && statSync(join(abs, entry)).isDirectory()) {
        seen.add(join(dir, entry));
      }
    }
  }
  return [...seen].sort();
}

const declaredIn = (pkg, names) => names.some((n) => n in (pkg.dependencies ?? {}) || n in (pkg.devDependencies ?? {}));

const rootPkg = JSON.parse(readFileSync(join(ROOT, 'package.json'), 'utf8'));
const problems = [];

for (const member of workspaceMembers()) {
  const dir = join(ROOT, member);
  const pkg = JSON.parse(readFileSync(join(dir, 'package.json'), 'utf8'));
  const scripts = pkg.scripts ?? {};

  for (const [name, command] of Object.entries(scripts)) {
    if (typeof command !== 'string') continue;

    // Referenced local files must exist.
    for (const token of command.match(/(?:\.\/|scripts\/)[\w./@-]+\.(?:sh|mjs|cjs|js)/g) ?? []) {
      if (!existsSync(join(dir, token.replace(/^\.\//, '')))) {
        problems.push(`${member}\t${name}: ${token} does not exist`);
      }
    }

    // Every command position must be a declared binary or a system command.
    const tokens = command.split(/[\s&|;]+/).filter(Boolean);
    tokens.forEach((token, index) => {
      const previous = tokens[index - 1];
      const atCommandPosition = index === 0 || previous === '&&' || previous === ';' || previous === '|';
      if (!atCommandPosition || token.startsWith('-') || token.startsWith('.') || token.startsWith('$') || token.endsWith('=')) return;
      if (SYSTEM_COMMANDS.has(token)) return;

      const provider = BINARY_PACKAGE[token] ?? token;
      if (declaredIn(pkg, [provider]) || declaredIn(rootPkg, [provider])) return;
      problems.push(`${member}\t${name}: "${token}" is not declared (${provider})`);
    });
  }
}

if (problems.length === 0) {
  console.log(`Workspace scripts check passed (${workspaceMembers().length} members).`);
  process.exit(0);
}

console.error(problems.join('\n'));
console.error(`\n${problems.length} problem(s). Declare the tool in the package that runs it.`);
process.exit(1);
