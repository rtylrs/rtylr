// Sweep the Go import graph: module-to-module and module-to-platform edges,
// both directions, with counts.
// Output: backend_imports.csv
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { walk, paths, writeCsv, relRepo } from './lib.mjs';

const moduleNames = new Set(
  (await import('./lib.mjs')).moduleDirs().map((m) => m.name),
);

const rows = []; // {from, to, kind, count, files}

function addEdge(from, to, kind, file) {
  let r = rows.find((x) => x.from === from && x.to === to && x.kind === kind);
  if (!r) {
    r = { from, to, kind, count: 0, files: new Set() };
    rows.push(r);
  }
  r.count++;
  if (r.files.size < 6) r.files.add(relRepo(file));
}

const importRe =
  /"github\.com\/voxire\/onion\/(internal\/modules\/([\w]+)|internal\/platform\/([\w]+))(\/[\w/]*)?"/g;

for (const m of moduleNames) {
  for (const f of walk(join(paths.modules, m), ['.go'])) {
    const t = readFileSync(f, 'utf8');
    // Only import blocks.
    const block = /import\s*\(([\s\S]*?)\)/.exec(t);
    const text = block ? block[1] : '';
    if (!text) continue;
    let mm;
    importRe.lastIndex = 0;
    while ((mm = importRe.exec(text))) {
      const mod = mm[2];
      const plat = mm[3];
      if (mod) {
        if (mod === m) continue;
        addEdge(m, mod, 'module-to-module', f);
      } else if (plat) {
        addEdge(m, `platform/${plat}`, 'module-to-platform', f);
      }
    }
  }
}

// Platform-to-module imports are forbidden by policy; measure anyway.
for (const f of walk(paths.platform, ['.go'])) {
  if (f.includes('database/schemas')) continue;
  const rel = f.slice(paths.platform.length + 1);
  const pkg = `platform/${rel.split('/')[0]}`;
  const t = readFileSync(f, 'utf8');
  const block = /import\s*\(([\s\S]*?)\)/.exec(t);
  const text = block ? block[1] : '';
  if (!text) continue;
  let mm;
  importRe.lastIndex = 0;
  while ((mm = importRe.exec(text))) {
    const mod = mm[2];
    if (mod) addEdge(pkg, mod, 'platform-to-module', f);
  }
}

rows.sort((a, b) => b.count - a.count || a.from.localeCompare(b.from));
writeCsv(
  'backend_imports.csv',
  ['from', 'to', 'kind', 'count', 'files'],
  rows.map((r) => [r.from, r.to, r.kind, r.count, [...r.files].join(' ')]),
);
