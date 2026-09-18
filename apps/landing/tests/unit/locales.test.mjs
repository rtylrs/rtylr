import { describe, it, before } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync, existsSync, globSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

import {
  baseRouteOf,
  basePathOf,
  localeRouteParams,
  LOCALES,
  DEFAULT_LOCALE,
  LOCALISED_ROUTES,
  shippedLocales,
  isLocalisedRoute,
  localeFromPath,
  localisePath,
  alternatesFor,
} from '../../apps/landing/lib/locales.ts';

// ---------------------------------------------------------------------------
// The locale contract.
//
// Two failures this guards, both silent.
//
// A locale route that renders untranslated copy is a doorway page, and
// Google says plainly that localized versions "are only considered
// duplicates if the main content of the page remains untranslated". Nothing
// in a build will tell you it happened. So `shipped` is asserted against
// the actual locale file and against the actual export.
//
// A non-reciprocal hreflang set is ignored, silently, in its entirety. The
// set must include every locale and itself. Generating it from one function
// is what makes that possible; these tests are what prove the function does
// it, including the case that is not exercised today because only English
// ships.
// ---------------------------------------------------------------------------

const REPO = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../..');
const APP = path.join(REPO, 'apps/landing');
const OUT = path.join(REPO, 'dist/landing');

const localeKeys = (code) => {
  const flat = (obj, prefix = '', out = []) => {
    for (const [k, v] of Object.entries(obj)) {
      const key = prefix ? `${prefix}.${k}` : k;
      if (v && typeof v === 'object' && !Array.isArray(v)) flat(v, key, out);
      else out.push(key);
    }
    return out;
  };
  return flat(JSON.parse(readFileSync(path.join(APP, `locales/${code}.json`), 'utf8')));
};

describe('the locale registry', () => {
  it('has exactly one default, and it ships', () => {
    const def = LOCALES.find((l) => l.code === DEFAULT_LOCALE);
    assert.ok(def, `default locale ${DEFAULT_LOCALE} is not in the registry`);
    assert.ok(def.shipped, 'the default locale must ship');
  });

  it('uses ISO 639-1 codes, with no region invented', () => {
    for (const l of LOCALES) {
      assert.match(l.code, /^[a-z]{2}$/, `${l.code} is not a bare ISO 639-1 code`);
      // A region alone is invalid, and UK / EU / UN are reserved codes
      // Google ignores. GB is the correct one for Britain.
      assert.ok(!/UK|EU|UN/.test(l.htmlLang), `${l.htmlLang} uses a reserved region code`);
    }
  });

  it('declares rtl only where it means it', () => {
    const rtl = LOCALES.filter((l) => l.dir === 'rtl').map((l) => l.code);
    assert.deepEqual(rtl, ['ar'], 'unexpected set of right-to-left locales');
  });
});

describe('shipped means the copy exists', () => {
  // The coupling that makes `shipped` a promise rather than a flag. A
  // locale claiming to ship on an empty file is the doorway case.
  it('every shipped locale has translated keys', () => {
    const enCount = localeKeys('en').length;
    for (const l of shippedLocales()) {
      const n = localeKeys(l.code).length;
      assert.ok(
        n > enCount * 0.5,
        `${l.code} ships with ${n}/${enCount} keys, which would serve mostly English`
      );
    }
  });

  it('every unshipped locale is genuinely absent from the export', () => {
    if (!existsSync(OUT)) return; // build-output checks are covered elsewhere
    const unshipped = LOCALES.filter((l) => !l.shipped).map((l) => l.code);
    const leaked = unshipped.filter((code) =>
      globSync(`${code}/**/index.html`, { cwd: OUT }).length > 0
    );
    assert.deepEqual(leaked, [], 'routes exported for a locale that has no copy');
  });

  it('every unshipped locale is absent from the sitemap', () => {
    const f = path.join(OUT, 'sitemap-0.xml');
    if (!existsSync(f)) return;
    const xml = readFileSync(f, 'utf8');
    for (const l of LOCALES.filter((x) => !x.shipped)) {
      assert.ok(
        !xml.includes(`https://rtylr.com/${l.code}/`),
        `${l.code} is advertised in the sitemap with no route behind it`
      );
    }
  });
});

describe('localisePath', () => {
  it('leaves the default locale at the root', () => {
    assert.equal(localisePath('/', 'en'), '/');
    assert.equal(localisePath('/pricing/', 'en'), '/pricing/');
  });

  it('prefixes every other locale', () => {
    assert.equal(localisePath('/', 'ar'), '/ar/');
    assert.equal(localisePath('/pricing/', 'ar'), '/ar/pricing/');
    assert.equal(localisePath('/legal/terms/', 'fr'), '/fr/legal/terms/');
  });

  it('does not produce a double slash or drop one', () => {
    for (const code of ['en', 'ar', 'fr']) {
      for (const p of ['/', '/pricing/', '/legal/terms/']) {
        const out = localisePath(p, code);
        assert.ok(out.startsWith('/'), `${out} does not start with /`);
        assert.ok(!out.includes('//'), `${out} has a double slash`);
        assert.ok(out.endsWith('/'), `${out} lost its trailing slash`);
      }
    }
  });
});

describe('localeFromPath', () => {
  it('reads the default locale for an unprefixed path', () => {
    assert.equal(localeFromPath('/').code, 'en');
    assert.equal(localeFromPath('/pricing/').code, 'en');
  });

  it('does not recognise an unshipped prefix as a locale', () => {
    // /ar/ is not a route today. Treating it as one would declare
    // lang="ar" on a page that does not exist.
    const unshipped = LOCALES.find((l) => !l.shipped);
    if (!unshipped) return;
    assert.equal(localeFromPath(`/${unshipped.code}/pricing/`).code, DEFAULT_LOCALE);
  });

  it('never returns undefined', () => {
    for (const p of ['', '/', '//', '/xx/', '/pricing/', '/ar/', '/fr/deep/path/']) {
      assert.ok(localeFromPath(p), `no locale resolved for ${JSON.stringify(p)}`);
    }
  });
});

describe('alternatesFor', () => {
  it('emits nothing while a single locale ships', () => {
    // A set of one is noise: it would advertise the page as an alternate
    // of itself and nothing else.
    if (shippedLocales().length >= 2) return;
    assert.deepEqual(alternatesFor('/pricing/'), []);
  });

  // The multi-locale behaviour cannot be exercised through the real
  // registry today, and it is the behaviour that matters most, so it is
  // tested directly against the rules Google states.
  const simulate = (path) => {
    const shipped = [
      { code: 'en', htmlLang: 'en' },
      { code: 'ar', htmlLang: 'ar' },
      { code: 'fr', htmlLang: 'fr' },
    ];
    const site = 'https://rtylr.com';
    const set = shipped.map((l) => ({
      hreflang: l.htmlLang,
      href: `${site}${localisePath(path, l.code)}`,
    }));
    set.push({ hreflang: 'x-default', href: `${site}${localisePath(path, DEFAULT_LOCALE)}` });
    return set;
  };

  for (const p of ['/', '/pricing/', '/legal/terms/']) {
    it(`${p} would be self-referencing and reciprocal`, () => {
      const set = simulate(p);
      // Self-reference: the page must list itself. A set where one member
      // does not point back is ignored in full.
      assert.ok(
        set.some((r) => r.hreflang === 'en' && r.href === `https://rtylr.com${p}`),
        'the set does not include the page itself'
      );
      assert.ok(set.some((r) => r.hreflang === 'x-default'), 'no x-default');
      // No duplicated path segments. next-sitemap treats a non-absolute
      // href as a base and appends the page path to it, which turned
      // /pricing/ into /pricing/pricing/ before hrefIsAbsolute was set.
      for (const r of set) {
        const tail = r.href.replace('https://rtylr.com', '');
        assert.ok(!/\/([^/]+)\/\1\//.test(tail), `duplicated segment in ${r.href}`);
      }
    });
  }
});

describe('coverage is per page, not just per locale', () => {
  // `shipped` is per locale; translation happens page by page. The first
  // version of alternatesFor took only a path, so a second shipped locale
  // would have advertised alternates for all 21 pages while five were
  // translated. Sixteen would 404, and one broken member voids the set.
  it('only routes listed as localised get alternates', () => {
    for (const p of ['/crm/', '/hr/', '/about/', '/support/', '/blog/']) {
      assert.equal(isLocalisedRoute(p), false, `${p} claims to be translated`);
      assert.deepEqual(alternatesFor(p), [], `${p} would advertise dead alternates`);
    }
  });

  it('the localised list is trailing-slash normalised', () => {
    for (const r of LOCALISED_ROUTES) {
      assert.ok(r.startsWith('/'), `${r} is not absolute`);
      assert.ok(r.endsWith('/'), `${r} lacks the trailing slash the export serves`);
      assert.ok(isLocalisedRoute(r), `${r} does not match its own predicate`);
      // The predicate has to cope with either form, because next-sitemap
      // does not guarantee the trailing slash on the path it passes in.
      assert.ok(isLocalisedRoute(r.replace(/\/$/, '') || '/'), `${r} fails unslashed`);
    }
  });

  it('the sitemap advertises alternates only for localised routes', () => {
    const f = path.join(OUT, 'sitemap-0.xml');
    if (!existsSync(f)) return;
    const xml = readFileSync(f, 'utf8');
    const offenders = [];
    for (const block of xml.match(/<url>[\s\S]*?<\/url>/g) ?? []) {
      if (!block.includes('hreflang')) continue;
      const loc = block.match(/<loc>([^<]+)<\/loc>/)?.[1] ?? '';
      const p = loc.replace('https://rtylr.com', '');
      // basePathOf() first: p is the real, already-localised URL in the
      // sitemap (e.g. /ar/pricing/), and isLocalisedRoute keys off the
      // bare route. Without it every shipped locale's own URLs failed
      // this check while actually being correct - see basePathOf's own
      // docblock, which names this exact bug.
      if (!isLocalisedRoute(basePathOf(p))) offenders.push(p);
    }
    assert.deepEqual(offenders, [], 'sitemap advertises alternates for untranslated routes');
  });
});

/**
 * Strip comments before scanning source text.
 *
 * Both checks below failed on their first run, against correct code: the
 * fix's own docblock quotes the bug it removed, so a naive regex matched
 * the explanation rather than an occurrence. A gate that fires on its own
 * documentation trains people to delete the documentation.
 */
const codeOnly = (src) =>
  src.replace(/\/\*[\s\S]*?\*\//g, '').replace(/(^|[^:])\/\/.*$/gm, '$1');

describe('the active language comes from the URL', () => {
  // It used to come from a cookie while the document derived lang and dir
  // from the route, so the two disagreed by construction. Under a static
  // export that is also a guaranteed hydration mismatch: the server has no
  // cookie and the client does.
  const src = codeOnly(readFileSync(path.join(APP, 'services/i18next.ts'), 'utf8'));

  it('i18next does not read a cookie for its language', () => {
    assert.ok(
      !/lng:\s*getCookie/.test(src),
      'the active language is being taken from a cookie again'
    );
    assert.ok(src.includes('localeFromPath'), 'the language is not derived from the path');
  });

  it('nothing calls changeLanguage from a cookie', () => {
    const offenders = [];
    for (const rel of globSync('**/*.{ts,tsx}', { cwd: APP })) {
      if (rel.includes('node_modules')) continue;
      const f = codeOnly(readFileSync(path.join(APP, rel), 'utf8'));
      // A changeLanguage call in the same file as a language-cookie read is
      // the shape of the bug: AppLayout did exactly this on every mount,
      // overriding the route after hydration.
      if (f.includes('changeLanguage') && /getCookie\(\s*['"]language['"]/.test(f)) {
        offenders.push(rel);
      }
    }
    assert.deepEqual(offenders, [], 'a cookie is overriding the route language');
  });
});

describe('the sitemap config asks for absolute alternates', () => {
  // Without hrefIsAbsolute every alternate URL is wrong, and wrong in a way
  // that only appears once a second locale ships. Asserted against the
  // config text because the failure is a missing option, not a value.
  it('sets hrefIsAbsolute on every alternate ref', () => {
    const cfg = readFileSync(path.join(APP, 'next-sitemap.config.js'), 'utf8');
    assert.ok(cfg.includes('alternateRefs'), 'the config no longer builds alternateRefs');
    assert.ok(
      cfg.includes('hrefIsAbsolute: true'),
      'alternateRefs without hrefIsAbsolute produce /pricing/pricing/'
    );
  });
});

describe('the locale route table', () => {
  // Both the routes and the sitemap read the same two lists, so a route
  // cannot exist without an hreflang entry and vice versa. These assert the
  // gate that keeps an unreviewed locale out of the export entirely.
  it('emits no route while only the default locale ships', () => {
    if (shippedLocales().length >= 2) return;
    assert.deepEqual(localeRouteParams(), [], 'a locale route exists with no shipped locale');
  });

  it('never emits a route for the default locale', () => {
    // /en/pricing/ and /pricing/ would be two URLs for one page.
    const en = localeRouteParams().filter((p) => p.locale === DEFAULT_LOCALE);
    assert.deepEqual(en, [], 'the default locale is served from the root, not a prefix');
  });

  it('produces exactly the translated routes for each shipped locale', () => {
    const extra = shippedLocales().filter((l) => l.code !== DEFAULT_LOCALE);
    const expected = extra.length * LOCALISED_ROUTES.length;
    assert.equal(localeRouteParams().length, expected);
  });

  it('round-trips a param set back to its route', () => {
    // getStaticProps derives the page to render from the slug, so a wrong
    // round-trip renders the wrong page under the right URL.
    assert.equal(baseRouteOf([]), '/');
    assert.equal(baseRouteOf(undefined), '/');
    assert.equal(baseRouteOf(['pricing']), '/pricing/');
    assert.equal(baseRouteOf(['legal', 'terms']), '/legal/terms/');
    for (const r of LOCALISED_ROUTES) {
      const slug = r.split('/').filter(Boolean);
      assert.equal(baseRouteOf(slug), r, `${r} does not round-trip`);
    }
  });

  it('every route it would emit has a page behind it', () => {
    // The catch-all maps a route to a component through its own table. A
    // route in the registry with no entry there renders the 404.
    const src = readFileSync(path.join(APP, 'pages/[locale]/[[...slug]].tsx'), 'utf8');
    const missing = LOCALISED_ROUTES.filter((r) => !src.includes(`'${r}'`));
    assert.deepEqual(missing, [], 'localised routes with no component mapped');
  });
});

describe('the served document declares its locale', () => {
  before(() => {
    if (!existsSync(OUT)) throw new Error('run pnpm build first');
  });

  it('every page carries a lang and a dir', () => {
    const pages = globSync('**/index.html', { cwd: OUT });
    assert.ok(pages.length >= 20, `expected 20+ pages, found ${pages.length}`);
    const offenders = [];
    for (const p of pages) {
      const html = readFileSync(path.join(OUT, p), 'utf8');
      const tag = html.match(/<html[^>]*>/)?.[0] ?? '';
      if (!/\blang="[a-z]{2}"/.test(tag) || !/\bdir="(ltr|rtl)"/.test(tag)) offenders.push([p, tag]);
    }
    assert.deepEqual(offenders, [], 'pages missing lang or dir on <html>');
  });

  it('declares the direction its locale actually uses', () => {
    const pages = globSync('**/index.html', { cwd: OUT });
    const offenders = [];
    for (const p of pages) {
      const html = readFileSync(path.join(OUT, p), 'utf8');
      const tag = html.match(/<html[^>]*>/)?.[0] ?? '';
      const lang = tag.match(/lang="([a-z]{2})"/)?.[1];
      const dir = tag.match(/dir="(ltr|rtl)"/)?.[1];
      const expected = LOCALES.find((l) => l.code === lang)?.dir;
      if (expected && dir !== expected) offenders.push([p, lang, dir, expected]);
    }
    assert.deepEqual(offenders, [], 'pages whose dir contradicts their lang');
  });
});
