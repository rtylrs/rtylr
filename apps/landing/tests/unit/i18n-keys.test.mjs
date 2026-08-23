import { describe, it, before } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync, globSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

// ---------------------------------------------------------------------------
// The i18n contract.
//
// i18next fails soft: ask for a key that does not exist and it renders the
// key name. "footer.link.pos" on the page looks obviously wrong to a human
// looking at that page, and completely invisible in a diff, in a build log,
// or on any of the 20 pages nobody opened. Extraction across 21 pages and
// 30 components is exactly the kind of long mechanical job where one typo
// ships.
//
// So the rules are asserted, not remembered:
//
//   1. Every key a t() call asks for exists in en.json.
//   2. Every key in en.json is asked for by something. This is the check
//      that would have caught the state this file was written in: en.json
//      held 101 strings describing VBitez, the QR-menu product this repo
//      was extracted from, wired to nothing, with the Arabic and French
//      files carrying finished translations of that same wrong product.
//   3. ar.json and fr.json contain no key that en.json does not have, so a
//      locale cannot drift into inventing its own structure.
//   4. Interpolation placeholders match across locales, because a
//      translation that drops {{group}} silently loses the value.
//
// Locale coverage is reported rather than asserted: ar and fr are
// deliberately empty until the copy is written, not machine translated, and
// reviewed. Rule 3 keeps them honest in the meantime.
// ---------------------------------------------------------------------------

const REPO = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../..');
const APP = path.join(REPO, 'apps/landing');
const LOCALES = path.join(APP, 'locales');

// t('a.b.c') and t("a.b.c"), including the form with interpolation options.
const T_CALL = /\bt\(\s*(['"])([A-Za-z][\w.]*)\1/g;

// t(`a.b.${id}.c`). A key built at runtime cannot be resolved statically,
// so the literal text before the first interpolation is treated as a
// prefix and everything under it counts as used.
//
// This is a deliberate hole in the orphan check and worth naming: adding
// one dynamic call exempts a whole subtree from it. Keep the static prefix
// as long as possible (`contact.channels.` rather than `contact.`) so the
// exemption stays narrow.
const T_TEMPLATE = /\bt\(\s*`([^`$]*)\$\{/g;

const SOURCE_GLOBS = ['pages/**/*.tsx', 'components/**/*.tsx', 'layouts/**/*.tsx'];

let en;
let used;
let usedPrefixes;

/** Flatten {a:{b:'x'}} to {'a.b':'x'}. */
function flatten(obj, prefix = '', out = {}) {
  for (const [k, v] of Object.entries(obj)) {
    const key = prefix ? `${prefix}.${k}` : k;
    if (v && typeof v === 'object' && !Array.isArray(v)) flatten(v, key, out);
    else out[key] = v;
  }
  return out;
}

const loadLocale = (name) => flatten(JSON.parse(readFileSync(path.join(LOCALES, `${name}.json`), 'utf8')));

const placeholders = (s) =>
  typeof s === 'string' ? [...s.matchAll(/\{\{\s*(\w+)\s*\}\}/g)].map((m) => m[1]).sort() : [];

before(() => {
  en = loadLocale('en');

  const files = SOURCE_GLOBS.flatMap((g) => globSync(g, { cwd: APP })).sort();
  // A zero here would make every assertion below vacuously pass.
  assert.ok(files.length > 40, `expected 40+ source files, found ${files.length}`);

  used = new Map();
  usedPrefixes = new Map();
  for (const rel of files) {
    const src = readFileSync(path.join(APP, rel), 'utf8');
    for (const m of src.matchAll(T_CALL)) {
      if (!used.has(m[2])) used.set(m[2], []);
      used.get(m[2]).push(rel);
    }
    for (const m of src.matchAll(T_TEMPLATE)) {
      if (!m[1]) continue;
      if (!usedPrefixes.has(m[1])) usedPrefixes.set(m[1], []);
      usedPrefixes.get(m[1]).push(rel);
    }
  }
});

/** Used directly, or covered by a runtime-built key's static prefix. */
const isUsed = (key) =>
  used.has(key) || [...usedPrefixes.keys()].some((prefix) => key.startsWith(prefix));

describe('translation keys', () => {
  it('every key a component asks for exists in en.json', () => {
    const missing = [...used.keys()].filter((k) => !(k in en)).map((k) => [k, used.get(k)[0]]);
    assert.deepEqual(missing, [], 'keys that would render as their own name');
  });

  it('no key in en.json is unused', () => {
    // Not tidiness. An orphaned namespace is how a previous product's copy
    // survived in this file long enough to be half-translated.
    const orphans = Object.keys(en).filter((k) => !isUsed(k));
    assert.deepEqual(orphans, [], 'keys defined but asked for by nothing');
  });

  it('every runtime-built key prefix matches something in en.json', () => {
    const dead = [...usedPrefixes.entries()]
      .filter(([prefix]) => !Object.keys(en).some((k) => k.startsWith(prefix)))
      .map(([prefix, files]) => [prefix, files[0]]);
    assert.deepEqual(dead, [], 't(`prefix.${...}`) calls that can never resolve');
  });

  it('every leaf in en.json is a legal shape', () => {
    // Three shapes are allowed, and nothing else:
    //   a string
    //   an array of strings        (a feature list, a table's row labels)
    //   an array of flat objects   (question and answer pairs)
    // Lists are arrays rather than numbered keys on purpose: a translator
    // translates a list, and reordering an array cannot silently reassign
    // one tier's feature to another the way features.0 could.
    const legal = (v) =>
      typeof v === 'string' ||
      (Array.isArray(v) &&
        v.length > 0 &&
        (v.every((x) => typeof x === 'string') ||
          v.every(
            (x) =>
              x &&
              typeof x === 'object' &&
              !Array.isArray(x) &&
              Object.values(x).every((y) => typeof y === 'string')
          )));

    const bad = Object.entries(en)
      .filter(([, v]) => !legal(v))
      .map(([k, v]) => [k, Array.isArray(v) ? 'mixed array' : typeof v]);
    assert.deepEqual(bad, [], 'illegal leaf shapes in the locale file');
  });

  it('list leaves have the same length in every locale', () => {
    // A truth table elsewhere in the code is matched to a label list by
    // index (see PLAN_FEATURES). A translation that drops a row shifts
    // every label after it onto the wrong row.
    const mismatched = [];
    for (const loc of ['ar', 'fr']) {
      for (const [k, v] of Object.entries(loadLocale(loc))) {
        if (Array.isArray(en[k]) && Array.isArray(v) && en[k].length !== v.length) {
          mismatched.push([loc, k, en[k].length, v.length]);
        }
      }
    }
    assert.deepEqual(mismatched, [], 'a translated list changed length');
  });

  it('no locale invents keys en.json does not have', () => {
    const strays = [];
    for (const loc of ['ar', 'fr']) {
      for (const k of Object.keys(loadLocale(loc))) {
        if (!(k in en)) strays.push([loc, k]);
      }
    }
    assert.deepEqual(strays, [], 'keys present in a translation but not in the source locale');
  });

  it('translations keep the interpolation placeholders of their source', () => {
    const broken = [];
    for (const loc of ['ar', 'fr']) {
      const other = loadLocale(loc);
      for (const [k, v] of Object.entries(other)) {
        const want = placeholders(en[k]);
        const got = placeholders(v);
        if (want.join() !== got.join()) broken.push([loc, k, want, got]);
      }
    }
    assert.deepEqual(broken, [], 'a translation dropped or renamed an interpolated value');
  });

  it('reports locale coverage', () => {
    const total = Object.keys(en).length;
    for (const loc of ['ar', 'fr']) {
      const n = Object.keys(loadLocale(loc)).length;
      console.log(`      ${loc}: ${n}/${total} keys (${Math.round((n / total) * 100)}%)`);
    }
    assert.ok(total > 0);
  });
});
