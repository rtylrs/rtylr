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
const APP = REPO;
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

// <LegalPageLayout i18nKey="legal.terms" .../>. LegalArticle and
// LegalPageLayout take the whole document's key prefix as a prop and build
// every key inside it at runtime (`t(`${i18nKey}.body.${id}.${blockKey}`)`),
// so there is no static string prefix for T_TEMPLATE to find - the dynamic
// part starts at position zero. The literal value passed at each of the
// three call sites is the actual prefix; this reads it the same way
// T_TEMPLATE reads a template literal's static lead-in.
//
// <Trans i18nKey="trust.packet.body" components={...} /> matches the same
// prop, but is a plain literal leaf key, not a dynamic-prefix root - it
// needs registering like a normal t() call, not exempting a whole subtree.
// The two are told apart below by checking whether the literal value is
// itself a key in en.json.
const I18N_KEY_PROP = /\bi18nKey=["']([\w.-]+)["']/g;

// lib and services are in here because t() is not confined to components:
// lib/faqs.ts exports a useFaq hook that builds keys, and leaving lib out
// made every FAQ key look orphaned.
const SOURCE_GLOBS = [
  'pages/**/*.tsx',
  'components/**/*.tsx',
  'layouts/**/*.tsx',
  'lib/**/*.ts',
  'services/**/*.ts',
];

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

  const files = SOURCE_GLOBS.flatMap((g) => globSync(g, { cwd: APP }))
    .filter((f) => !/\.(test|spec)\.tsx?$/.test(f))
    .sort();
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
    for (const m of src.matchAll(I18N_KEY_PROP)) {
      if (m[1] in en) {
        if (!used.has(m[1])) used.set(m[1], []);
        used.get(m[1]).push(rel);
        continue;
      }
      const prefix = `${m[1]}.`;
      if (!usedPrefixes.has(prefix)) usedPrefixes.set(prefix, []);
      usedPrefixes.get(prefix).push(rel);
    }
  }
});

// i18next's built-in pluralisation (Intl.PluralRules): a call for the bare
// key resolves at runtime to whichever of these suffixed variants matches
// the count's plural category, so the bare key itself need not exist in
// en.json as long as at least one variant does, and a suffixed key in
// en.json is "used" by a t() call for its unsuffixed base.
const PLURAL_SUFFIXES = ['_zero', '_one', '_two', '_few', '_many', '_other'];

const hasPluralVariant = (key) => PLURAL_SUFFIXES.some((s) => `${key}${s}` in en);

const pluralBase = (key) => {
  const suffix = PLURAL_SUFFIXES.find((s) => key.endsWith(s));
  return suffix ? key.slice(0, -suffix.length) : null;
};

/** Used directly, via a plural variant, or covered by a runtime-built key's static prefix. */
const isUsed = (key) =>
  used.has(key) ||
  used.has(pluralBase(key)) ||
  [...usedPrefixes.keys()].some((prefix) => key.startsWith(prefix));

describe('translation keys', () => {
  it('every key a component asks for exists in en.json', () => {
    const missing = [...used.keys()]
      .filter((k) => !(k in en) && !hasPluralVariant(k))
      .map((k) => [k, used.get(k)[0]]);
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
              // A field inside one of those objects may itself be a list of
              // strings: a partner track carries a benefits list.
              Object.values(x).every(
                (y) =>
                  typeof y === 'string' ||
                  (Array.isArray(y) && y.length > 0 && y.every((z) => typeof z === 'string'))
              )
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
    // A locale may define plural categories English does not need (Arabic
    // has six, English effectively two), so a suffixed key whose base has
    // any variant in en.json is a legal extra plural form, not an invented
    // key.
    const strays = [];
    for (const loc of ['ar', 'fr']) {
      for (const k of Object.keys(loadLocale(loc))) {
        if (k in en) continue;
        const base = pluralBase(k);
        if (base !== null && hasPluralVariant(base)) continue;
        strays.push([loc, k]);
      }
    }
    assert.deepEqual(strays, [], 'keys present in a translation but not in the source locale');
  });

  it('translations keep the interpolation placeholders of their source', () => {
    // For a plural form the source doesn't itself define (an Arabic-only
    // category like _few), compare against whichever sibling variant does
    // exist: every plural form of one key interpolates the same values.
    const sourceValueFor = (k) => {
      if (k in en) return en[k];
      const base = pluralBase(k);
      if (base === null) return undefined;
      const sibling = PLURAL_SUFFIXES.map((s) => `${base}${s}`).find((sk) => sk in en);
      return sibling ? en[sibling] : undefined;
    };
    const broken = [];
    for (const loc of ['ar', 'fr']) {
      const other = loadLocale(loc);
      for (const [k, v] of Object.entries(other)) {
        const want = placeholders(sourceValueFor(k));
        const got = placeholders(v);
        if (want.join() !== got.join()) broken.push([loc, k, want, got]);
      }
    }
    assert.deepEqual(broken, [], 'a translation dropped or renamed an interpolated value');
  });

  it('multi-line values are rendered through something that breaks lines', () => {
    // JSX collapses a newline in text content to a space. Headings on this
    // site are deliberately two lines, and moving them out of the markup
    // turned an explicit <br /> into a \n inside a string, which rendered
    // as one line with no error and no visible diff.
    //
    // Two consumers handle \n: components/text/Lines, and SplitWords, which
    // splits on it itself. Anything else silently loses the break.
    const multiline = Object.entries(en)
      .filter(([, v]) => typeof v === 'string' && v.includes('\n'))
      .map(([k]) => k);
    assert.ok(multiline.length > 0, 'no multi-line values found, so this check proves nothing');

    const sources = SOURCE_GLOBS.flatMap((g) => globSync(g, { cwd: APP }))
      .filter((f) => !/\.(test|spec)\.tsx?$/.test(f))
      .map((f) => readFileSync(path.join(APP, f), 'utf8'))
      .join('\n');

    // A templated key (text={t(`a.${id}.title`)}) covers every key with that
    // prefix and suffix, the same allowance T_TEMPLATE makes for usage.
    const templated = [...sources.matchAll(/text=\{t\(`([^`$]*)\$\{[^}]+\}([^`$]*)`\)\}/g)].map(
      ([, prefix, suffix]) => [prefix, suffix]
    );
    const unbroken = multiline.filter((k) => {
      const call = `t('${k}')`;
      return !(
        sources.includes(`<Lines text={${call}}`) ||
        sources.includes(`text={${call}}`) ||
        templated.some(([prefix, suffix]) => k.startsWith(prefix) && k.endsWith(suffix))
      );
    });
    assert.deepEqual(unbroken, [], 'multi-line copy rendered without a line break');
  });

  it('an extracted file has no hardcoded accessibility text left', () => {
    // aria-label, alt, title and placeholder are read aloud or shown on
    // hover, and none of them appear in a screenshot review. They are the
    // strings that survive an extraction pass: seven were left behind in the
    // two header files, found only by grepping for the attribute rather than
    // by looking at the page.
    //
    // Scoped to files that already call useTranslation, so it gets stricter
    // as extraction proceeds instead of failing on the pages not yet done.
    const ATTR = /\b(aria-label|placeholder|alt|title)="([^"]+)"/g;
    const offenders = [];

    for (const rel of SOURCE_GLOBS.flatMap((g) => globSync(g, { cwd: APP }))) {
      if (/\.(test|spec)\.tsx?$/.test(rel)) continue;
      const src = readFileSync(path.join(APP, rel), 'utf8');
      if (!src.includes('useTranslation')) continue;
      for (const m of src.matchAll(ATTR)) {
        // A URL or an anchor is not copy.
        if (/^(#|https?:|\/|mailto:|tel:)/.test(m[2])) continue;
        offenders.push([rel, m[0]]);
      }
    }
    assert.deepEqual(offenders, [], 'hardcoded strings in accessibility attributes');
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
