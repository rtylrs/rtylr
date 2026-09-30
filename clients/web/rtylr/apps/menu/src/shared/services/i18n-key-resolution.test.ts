import fs from 'node:fs';
import path from 'node:path';
import { describe, expect, it } from 'vitest';

/**
 * Every literal `t('…')` key in the menu app must resolve in every shipped
 * locale.
 *
 * This is the guard whose absence let `posmenu.*` reach customers as dotted
 * paths on `/b/:id`. The eight keys in `locales/*\/posmenu.json` had no file to
 * live in at all, and nothing in the app, the test suite or CI noticed, because
 * a key that does not exist is indistinguishable from a key that renders
 * correctly unless something walks the call sites and the locale files together.
 *
 * Why the key must be qualified from the tree root: `shared/services/i18next.ts`
 * hands i18next the whole locale tree for every namespace it asks for, so
 * i18next stores the entire tree under `posmenu` and again under every other
 * namespace. The namespace is cosmetic and `t('posmenu.allItems')` resolves
 * while `t('allItems')` under `useTranslation('posmenu')` cannot. That is why
 * the collector records the full dotted key as written, and why a key is
 * resolved by looking it up from the root of the tree.
 *
 * Known limit, kept visible rather than assumed away: only *literal* keys are
 * collected. OrderStatus builds two families dynamically from `STATUS_STEPS`
 * (`` t(`orderstatus.steps.${step}`) ``), so `collectDynamicFamilies` expands
 * those from the array in the same file instead of trusting a hand-written list.
 * A new dynamic family fails the run rather than passing unseen.
 */

const SRC_DIR = path.resolve(__dirname, '..', '..');
const LOCALES_DIR = path.join(SRC_DIR, 'locales');

/**
 * A floor per app, so a regex that silently stops matching cannot make this
 * suite pass by finding nothing to check. Menu is a customer-facing QR menu, not
 * an admin console, so its floor sits far below the other apps by nature rather
 * than by neglect.
 */
const MINIMUM_LITERAL_KEYS = 80;

const SOURCE_FILES = ['.ts', '.tsx'] as const;
const SKIP_DIRS = new Set(['node_modules', 'dist', 'build', 'coverage']);

function sourceFiles(dir: string = SRC_DIR): string[] {
  const out: string[] = [];
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    if (entry.isDirectory()) {
      if (SKIP_DIRS.has(entry.name)) continue;
      out.push(...sourceFiles(path.join(dir, entry.name)));
    } else if (SOURCE_FILES.some((ext) => entry.name.endsWith(ext))) {
      // Test files are not shipped, and a fixture in one would be collected as
      // if the app called it.
      if (/\.(test|spec)\.[jt]sx?$|\.d\.ts$/.test(entry.name)) continue;
      out.push(path.join(dir, entry.name));
    }
  }
  return out;
}

/**
 * Remove comments while leaving string literals alone.
 *
 * Needed because this app documents its own key convention in a comment beside
 * the loader — `t('profile.save')` is an example, not a call. Counting it would
 * fail the suite on a key nobody renders, and deleting the comment to make the
 * suite pass is exactly the wrong trade. Scanning character by character with
 * string state is what keeps a `//` inside a URL from eating the rest of the
 * line, and a `/` inside `'…'` from opening a block comment.
 */
export function stripComments(source: string): string {
  let out = '';
  let quote: string | null = null;
  for (let i = 0; i < source.length; i++) {
    const c = source[i]!;
    const next = source[i + 1];
    if (quote) {
      out += c;
      if (c === '\\') {
        out += next ?? '';
        i++;
      } else if (c === quote) {
        quote = null;
      }
      continue;
    }
    if (c === "'" || c === '"' || c === '`') {
      quote = c;
      out += c;
      continue;
    }
    if (c === '/' && next === '/') {
      while (i < source.length && source[i] !== '\n') i++;
      if (i < source.length) out += '\n';
      continue;
    }
    if (c === '/' && next === '*') {
      i += 2;
      while (i < source.length && !(source[i] === '*' && source[i + 1] === '/')) i++;
      i++;
      continue;
    }
    out += c;
  }
  return out;
}

/**
 * Literal `t('…')` calls, mapped to the first place each is used.
 *
 * `\b` before `t` is what keeps `format('x')` and `'a/b'.split('/')` out of the
 * result; a call is `t` as a whole word followed by `(`.
 */
export function collectLiteralKeys(dir: string = SRC_DIR): Map<string, string> {
  const used = new Map<string, string>();
  const call = /\bt\(\s*(['"`])([A-Za-z0-9_]+(?:\.[A-Za-z0-9_]+)*)\1/g;
  for (const file of sourceFiles(dir)) {
    const code = stripComments(fs.readFileSync(file, 'utf8'));
    for (const match of code.matchAll(call)) {
      const key = match[2]!;
      if (!used.has(key)) used.set(key, path.relative(dir, file));
    }
  }
  return used;
}

/** `const STATUS_STEPS = ['received', …] as const;` — name and its members. */
function stringArrays(raw: string): Map<string, string[]> {
  const out = new Map<string, string[]>();
  for (const arr of raw.matchAll(
    /const\s+([A-Za-z_][A-Za-z0-9_]*)\s*(?::[^=]+)?=\s*\[([^\]]*)\]/g
  )) {
    const members = [...arr[2]!.matchAll(/'([^']+)'|"([^"]+)"/g)].map((v) => (v[1] ?? v[2])!);
    if (members.length > 0) out.set(arr[1]!, members);
  }
  return out;
}

/** `STATUS_STEPS.map((step) => …)` — which array each callback parameter belongs to. */
function mappedParams(raw: string): { array: string; param: string }[] {
  return [
    ...raw.matchAll(/([A-Za-z_][A-Za-z0-9_]*)\s*\.\s*map\(\s*\(\s*([A-Za-z_][A-Za-z0-9_]*)/g),
  ].map((m) => ({ array: m[1]!, param: m[2]! }));
}

/**
 * Keys built by a template literal, expanded from the array that drives them.
 *
 * `t(`orderstatus.steps.${step}`)` alone is invisible to a literal-only
 * collector, and that blind spot is how `nav.createItem` and `nav.createBanner`
 * reached three locale files as nothing at all while both screens rendered
 * English. Deriving the values from the code that produces them closes it here
 * without a hand-maintained list to forget.
 *
 * A prefix whose driving array cannot be found is returned as an expansion
 * failure, and `it` below asserts there are none. Silently skipping it would
 * reintroduce the blind spot being guarded against.
 */
export function collectDynamicFamilies(dir: string = SRC_DIR): {
  keys: Map<string, string>;
  unexpanded: string[];
} {
  const keys = new Map<string, string>();
  const unexpanded: string[] = [];
  const call = /\bt\(\s*`([^`]*\$\{[^`]*`)/g;
  for (const file of sourceFiles(dir)) {
    const raw = fs.readFileSync(file, 'utf8');
    const code = stripComments(raw);
    const arrays = stringArrays(raw);
    const mapped = mappedParams(raw);
    for (const match of code.matchAll(call)) {
      const template = match[1]!;
      const where = path.relative(dir, file);
      // Pair each literal segment with the expression that fills it, so the
      // trailing empty piece of `` `a.b.${x}` `` is not read as a prefix.
      for (const seg of template.matchAll(/([^$]*)\$\{([^}]*)\}/g)) {
        const prefix = seg[1]!.endsWith('.') ? seg[1]! : `${seg[1]!}.`;
        const identifiers = new Set(seg[2]!.match(/[A-Za-z_][A-Za-z0-9_]*/g) ?? []);
        const values = new Set<string>();
        for (const [name, members] of arrays) {
          // The array is named directly, as in `${STATUS_STEPS[i]}`, or is
          // mapped and the expression is the callback's parameter, as in
          // `STATUS_STEPS.map((step) => …)`.
          const drives =
            identifiers.has(name) ||
            mapped.some((m) => m.array === name && identifiers.has(m.param));
          if (!drives) continue;
          for (const member of members) values.add(member);
        }
        if (values.size === 0) {
          unexpanded.push(`${template}  (${where})`);
          continue;
        }
        for (const value of values) keys.set(`${prefix}${value}`, where);
      }
    }
  }
  return { keys, unexpanded };
}

export function localeLanguages(localesDir: string = LOCALES_DIR): string[] {
  return fs
    .readdirSync(localesDir, { withFileTypes: true })
    .filter((e) => e.isDirectory())
    .map((e) => e.name)
    .sort();
}

export function localeNamespaces(localesDir: string, lang: string): string[] {
  return fs
    .readdirSync(path.join(localesDir, lang))
    .filter((f) => f.endsWith('.json'))
    .map((f) => f.replace(/\.json$/, ''))
    .sort();
}

/** Flatten a namespace file into root-qualified leaf keys, plural forms included. */
export function availableKeys(localesDir: string, lang: string): Set<string> {
  const keys = new Set<string>();
  const walk = (node: unknown, prefix: string): void => {
    if (node === null || typeof node !== 'object' || Array.isArray(node)) {
      if (prefix) keys.add(prefix);
      return;
    }
    for (const [name, value] of Object.entries(node)) {
      walk(value, prefix ? `${prefix}.${name}` : name);
    }
  };
  for (const ns of localeNamespaces(localesDir, lang)) {
    const file = path.join(localesDir, lang, `${ns}.json`);
    walk(JSON.parse(fs.readFileSync(file, 'utf8')), ns);
  }
  return keys;
}

/**
 * Keys that resolve nowhere, with where each is used.
 *
 * A plural key is accepted by its suffixed forms, because the bare form
 * deliberately does not exist in a resource file.
 */
export function unresolvedKeys(used: Map<string, string>, available: Set<string>): string[] {
  const out: string[] = [];
  for (const [key, where] of used) {
    const leaf = key.split('.').pop()!;
    const plural = [...available].some(
      (a) => a.startsWith(`${key}_`) || a.startsWith(`${key}_${leaf}_`)
    );
    if (!available.has(key) && !plural) out.push(`${key}  (${where})`);
  }
  return out.sort();
}

const languages = localeLanguages();
const literal = collectLiteralKeys();
const dynamic = collectDynamicFamilies();
const used = new Map([...literal, ...dynamic.keys]);

describe('menu i18n keys resolve', () => {
  it('finds translation calls to check, so a passing run means something', () => {
    expect(literal.size).toBeGreaterThanOrEqual(MINIMUM_LITERAL_KEYS);
  });

  it('ships more than one language, so cross-locale checks are meaningful', () => {
    expect(languages.length).toBeGreaterThan(1);
    expect(languages).toContain('en');
  });

  it('expands every dynamically built key, rather than passing on what it cannot see', () => {
    expect(dynamic.unexpanded).toEqual([]);
    expect(dynamic.keys.size).toBeGreaterThan(0);
  });

  it.each(languages)('every key used in the app resolves in %s', (lang) => {
    expect(unresolvedKeys(used, availableKeys(LOCALES_DIR, lang))).toEqual([]);
  });

  it.each(languages)('%s declares the same namespace files as en', (lang) => {
    expect(localeNamespaces(LOCALES_DIR, lang)).toEqual(localeNamespaces(LOCALES_DIR, 'en'));
  });

  it('ships no namespace file that no call site reads', () => {
    // A file nothing reads is translation effort that reaches no customer, and
    // a locale that ships one the others do not is a locale that renders keys.
    const read = new Set([...used.keys()].map((k) => k.split('.')[0]!));
    for (const lang of languages) {
      expect(localeNamespaces(LOCALES_DIR, lang).filter((ns) => !read.has(ns))).toEqual([]);
    }
  });

  it('resolves the posmenu keys on the public /b/:id page in every locale', () => {
    // The regression this suite was restored for. Named explicitly so the
    // failure reads as the customer-facing bug it is, not as a list of keys.
    const posmenuKeys = [...literal.keys()].filter((k) => k.startsWith('posmenu.'));
    expect(posmenuKeys.length).toBeGreaterThan(0);
    for (const lang of languages) {
      expect(
        unresolvedKeys(
          new Map(posmenuKeys.map((k) => [k, 'PosMenu.tsx'])),
          availableKeys(LOCALES_DIR, lang)
        )
      ).toEqual([]);
    }
  });
});

/**
 * Unit coverage for the machinery above.
 *
 * The suite above runs over the real app, which cannot reach the cases that
 * matter most here: a plural key, where the bare form deliberately does not
 * exist, and a key quoted in a comment, which would be a false positive. Both
 * are silent failures in opposite directions, so both get a fixture rather than
 * waiting for the app to contain one.
 */
function fixture(files: Record<string, string>): string {
  const dir = fs.mkdtempSync(path.join(process.env.TMPDIR ?? '/tmp', 'menu-i18n-'));
  for (const [rel, body] of Object.entries(files)) {
    const full = path.join(dir, rel);
    fs.mkdirSync(path.dirname(full), { recursive: true });
    fs.writeFileSync(full, body);
  }
  return dir;
}

describe('stripComments', () => {
  it('drops a line comment but keeps a call before it on the same line', () => {
    expect(stripComments("const a = t('one.alpha'); // t('one.beta')")).toBe(
      "const a = t('one.alpha'); "
    );
  });

  it('keeps a // inside a string, so a URL does not swallow the rest of the line', () => {
    const line = "const u = 'https://x.dev'; const k = t('one.alpha');";
    expect(stripComments(line)).toBe(line);
  });
});

describe('collectLiteralKeys', () => {
  it('finds a call whatever precedes it, including inside an object literal', () => {
    const dir = fixture({
      'a.tsx': [
        "const x = t('one.alpha');",
        "const o = { label: t('one.beta') };",
        "const y = cond ? t('one.gamma') : t('one.delta');",
        "const z = f(t('one.epsilon'));",
      ].join('\n'),
    });

    expect([...collectLiteralKeys(dir).keys()].sort()).toEqual([
      'one.alpha',
      'one.beta',
      'one.delta',
      'one.epsilon',
      'one.gamma',
    ]);
  });

  it('does not treat an identifier ending in t as a translation call', () => {
    const dir = fixture({
      'a.tsx': ["const parts = 'a/b'.split('/');", "const s = format('x');"].join('\n'),
    });

    expect([...collectLiteralKeys(dir).keys()]).toEqual([]);
  });

  it('ignores a key quoted in a comment, which documents rather than calls', () => {
    const dir = fixture({
      'a.tsx': [
        "// so t('common.save') maps to locales/{lang}/common.json",
        '/* measured: t(\'emptyStates.rolesNone\') -> "emptyStates.rolesNone" */',
        "const real = t('one.alpha');",
      ].join('\n'),
    });

    expect([...collectLiteralKeys(dir).keys()]).toEqual(['one.alpha']);
  });

  it('records the first place a key is used', () => {
    const dir = fixture({ 'a.tsx': ["const a = t('one.alpha');"].join('\n') });

    expect(collectLiteralKeys(dir).get('one.alpha')).toBe('a.tsx');
  });
});

describe('collectDynamicFamilies', () => {
  it('expands a template key from the array that drives it', () => {
    const dir = fixture({
      'a.tsx': [
        "const STEPS = ['received', 'ready'] as const;",
        'const a = t(`one.steps.${STEPS[i]}`);',
      ].join('\n'),
    });

    const { keys, unexpanded } = collectDynamicFamilies(dir);
    expect(unexpanded).toEqual([]);
    expect([...keys.keys()].sort()).toEqual(['one.steps.ready', 'one.steps.received']);
  });

  it('expands a template key from the parameter of a map over that array', () => {
    const dir = fixture({
      'a.tsx': [
        "const STEPS = ['received', 'ready'] as const;",
        'STEPS.map((step) => t(`one.steps.${step}`));',
      ].join('\n'),
    });

    expect([...collectDynamicFamilies(dir).keys.keys()].sort()).toEqual([
      'one.steps.ready',
      'one.steps.received',
    ]);
  });

  it('reports a family it cannot expand instead of skipping it silently', () => {
    const dir = fixture({ 'a.tsx': ['const a = t(`one.${who}.label`);'].join('\n') });

    const { keys, unexpanded } = collectDynamicFamilies(dir);
    expect(keys.size).toBe(0);
    expect(unexpanded).toHaveLength(1);
  });
});

describe('unresolvedKeys', () => {
  const available = (): Set<string> => {
    const dir = fixture({
      'en/one.json': JSON.stringify({ alpha: 'A', count_one: 'one', count_other: 'many' }),
    });
    return availableKeys(dir, 'en');
  };

  it('accepts a plural key by its suffixed forms, since the bare key never exists', () => {
    expect(unresolvedKeys(new Map([['one.count', 'a.tsx']]), available())).toEqual([]);
  });

  it('reports a key that resolves in no form, with where it is used', () => {
    expect(unresolvedKeys(new Map([['one.missing', 'a.tsx:7']]), available())).toEqual([
      'one.missing  (a.tsx:7)',
    ]);
  });

  it('reports an unqualified key, which is the defect the guard exists for', () => {
    expect(unresolvedKeys(new Map([['alpha', 'a.tsx:3']]), available())).toEqual([
      'alpha  (a.tsx:3)',
    ]);
  });
});
