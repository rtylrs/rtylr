// Inventory every internal/platform/* package: exported surface, caller
// count (distinct modules importing it), and a live / dead / declared-only
// verdict.
// Output: platform_packages.csv
import { readFileSync, readdirSync, statSync } from 'node:fs';
import { join } from 'node:path';
import { walk, paths, writeCsv } from './lib.mjs';

const pkgs = readdirSync(paths.platform).filter((d) => {
  try {
    return statSync(join(paths.platform, d)).isDirectory();
  } catch {
    return false;
  }
});

// Caller counts from the imports sweep logic (recomputed here to keep this
// script independently runnable).
const callers = new Map(); // pkg -> Set(module)
for (const m of walk(paths.modules, ['.go'])) {
  const mod = m.slice(paths.modules.length + 1).split('/')[0];
  const t = readFileSync(m, 'utf8');
  const block = /import\s*\(([\s\S]*?)\)/.exec(t);
  const text = block ? block[1] : t;
  for (const mm of text.matchAll(
    /"github\.com\/voxire\/onion\/internal\/platform\/([\w]+)(?:\/[\w/]*)?"/g,
  )) {
    if (!callers.has(mm[1])) callers.set(mm[1], new Set());
    callers.get(mm[1]).add(mod);
  }
}
// Platform-internal callers too.
for (const f of walk(paths.platform, ['.go'])) {
  const rel = f.slice(paths.platform.length + 1);
  const from = rel.split('/')[0];
  const t = readFileSync(f, 'utf8');
  for (const mm of t.matchAll(
    /"github\.com\/voxire\/onion\/internal\/platform\/([\w]+)(?:\/[\w/]*)?"/g,
  )) {
    const target = mm[1];
    if (target === from) continue;
    if (!callers.has(target)) callers.set(target, new Set());
    callers.get(target).add(`platform/${from}`);
  }
}
// Composition roots consume platform services too. They are not business
// modules, but counting them keeps runtime-only platform packages from being
// misreported as unused.
for (const f of walk(paths.commands, ['.go'])) {
  const rel = f.slice(paths.commands.length + 1);
  const from = `cmd/${rel.split('/')[0]}`;
  const t = readFileSync(f, 'utf8');
  const block = /import\s*\(([\s\S]*?)\)/.exec(t);
  const text = block ? block[1] : t;
  for (const mm of text.matchAll(
    /"github\.com\/voxire\/onion\/internal\/platform\/([\w]+)(?:\/[\w/]*)?"/g,
  )) {
    if (!callers.has(mm[1])) callers.set(mm[1], new Set());
    callers.get(mm[1]).add(from);
  }
}

const rows = [];
for (const p of pkgs.sort()) {
  const files = walk(join(paths.platform, p), ['.go']);
  let exported = new Set();
  let hasImpl = false; // any concrete func body beyond type/interface decls
  let hasContractOnly = true;
  for (const f of files) {
    const t = readFileSync(f, 'utf8');
    if (f.endsWith('_test.go')) continue;
    for (const m of t.matchAll(/^(?:func\s+(?:\([^)]*\)\s+)?([A-Z]\w*)|(?:type|var|const)\s+([A-Z]\w*))/gm)) {
      exported.add(m[1] || m[2]);
    }
    // Grouped type/var/const blocks declare one exported name per line, so the
    // single-line pattern above cannot see them.
    for (const m of t.matchAll(/^(?:type|var|const)\s*\(([\s\S]*?)^\)/gm)) {
      for (const d of m[1].matchAll(/^\s*([A-Z]\w*)/gm)) exported.add(d[1]);
    }
    // Concrete functions with bodies that are not trivial accessors.
    for (const m of t.matchAll(/^func\s+(?:\([^)]*\)\s+)?([A-Za-z]\w*)\s*\([^)]*\)[^{]*\{/gm)) {
      hasImpl = true;
    }
    if (/interface \{|^type\s+\w+\s+=/m.test(t) === false && /func /.test(t)) hasContractOnly = false;
  }
  const callerSet = callers.get(p) || new Set();
  let verdict;
  const implFiles = files.filter((f) => !f.endsWith('_test.go'));
  if (callerSet.size === 0 && !hasImpl && implFiles.every((f) => /contract\.go$/.test(f)))
    verdict = 'declared-only';
  else if (callerSet.size === 0) verdict = 'dead-or-internal-only';
  else verdict = 'live';
  // Special case: contract-only packages whose port is implemented and wired
  // in another module are live seams even though this directory holds no impl.
  if (
    implFiles.every((f) => /contract\.go$/.test(f)) &&
    callerSet.size > 0
  )
    verdict = 'live-port'; // contract-only but imported: consumers reference its types

  rows.push({
    package: `platform/${p}`,
    go_files: files.filter((f) => !f.endsWith('_test.go')).length,
    exported_surface: [...exported].sort().join(' '),
    caller_count: callerSet.size,
    callers: [...callerSet].sort().slice(0, 14).join(' '),
    verdict,
  });
}
writeCsv(
  'platform_packages.csv',
  ['package', 'go_files', 'exported_surface', 'caller_count', 'callers', 'verdict'],
  rows.map((r) => [r.package, r.go_files, r.exported_surface, r.caller_count, r.callers, r.verdict]),
);
console.log(rows.map((r) => `${r.package}: ${r.verdict} (${r.caller_count})`).join('\n'));
