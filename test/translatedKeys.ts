import fs from 'node:fs';
import path from 'node:path';

/**
 * Finds lists whose React `key` is a string that `t()` produced.
 *
 * `key` is React's identity for a list item. When the key is a translation,
 * switching language changes every key, so React unmounts and remounts the whole
 * list: entrance animations replay and any state inside a row is discarded
 * (rtylr#1324).
 *
 * Only an array built in the source counts. A label that arrives from the API is
 * server data and is stable across languages, so `key={row.label}` over
 * `data.rows` is fine and must not be reported — getting that wrong is how the
 * first count of this was overstated.
 */
export interface TranslatedKeyHit {
  file: string;
  line: number;
  detail: string;
}

const KEY_ON_LABEL = /key=\{\s*(?:([A-Za-z_][A-Za-z0-9_]*)\.)?label\s*\}/;

/** The `label:` entries of the array literal that ends at `endLine`, if any. */
function inlineArrayLabels(lines: string[], endLine: number): string[] | null {
  let depth = 0;
  for (let i = endLine; i >= 0 && i > endLine - 80; i -= 1) {
    depth += (lines[i].match(/]/g) ?? []).length - (lines[i].match(/\[/g) ?? []).length;
    if (depth <= 0) {
      const block = lines.slice(i, endLine + 1).join('\n');
      return [...block.matchAll(/label:\s*([^,\n]+)/g)].map((m) => m[1]);
    }
  }
  return null;
}

function labelsFor(lines: string[], keyLine: number): string[] | null {
  let mapLine: number | null = null;
  for (let i = keyLine; i >= 0 && i > keyLine - 25; i -= 1) {
    if (lines[i].includes('.map(')) {
      mapLine = i;
      break;
    }
  }
  if (mapLine === null) return null;

  // `].map(` — the array is written right there.
  if (/\]\s*\.map\(/.test(lines[mapLine])) return inlineArrayLabels(lines, mapLine);

  // `something.map(` — resolve the variable's declaration in this file.
  const named = /([A-Za-z_][A-Za-z0-9_]*)\s*\.map\(/.exec(lines[mapLine]);
  if (!named) return null;
  const declIndex = lines.findIndex((l) => new RegExp(`\\b(const|let)\\s+${named[1]}\\b`).test(l));
  if (declIndex < 0) return null;
  const decl = lines.slice(declIndex, declIndex + 40).join('\n');
  // A declaration that reads a query result or a prop is server data.
  if (/\?\.|useQuery|\bdata\b/.test(lines[declIndex])) return [];
  return [...decl.matchAll(/label:\s*([^,\n]+)/g)].map((m) => m[1]);
}

export function collectTranslatedKeys(srcDir: string): TranslatedKeyHit[] {
  const hits: TranslatedKeyHit[] = [];
  const walk = (dir: string): void => {
    for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
      const full = path.join(dir, entry.name);
      if (entry.isDirectory()) {
        if (entry.name === 'node_modules' || entry.name === 'locales') continue;
        walk(full);
        continue;
      }
      if (!/\.tsx$/.test(entry.name) || /\.test\.tsx$/.test(entry.name)) continue;
      const lines = fs.readFileSync(full, 'utf8').split('\n');
      lines.forEach((line, i) => {
        if (!KEY_ON_LABEL.test(line)) return;
        const labels = labelsFor(lines, i);
        if (labels && labels.some((l) => /\bt\(/.test(l))) {
          hits.push({ file: full, line: i + 1, detail: line.trim() });
        }
      });
    }
  };
  walk(srcDir);
  return hits;
}

export function formatTranslatedKeys(hits: TranslatedKeyHit[], root: string): string[] {
  return hits.map((h) => `${path.relative(root, h.file)}:${h.line}`).sort();
}
