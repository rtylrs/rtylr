import { describe, it, before } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync, existsSync } from 'node:fs';
import { globSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

import { basePathOf } from '../../lib/locales.ts';

// ---------------------------------------------------------------------------
// Standing gates, asserted against the build output rather than the source.
//
// The distinction matters and has been got wrong twice on this project:
// build-time artifacts do not exist in the source tree, and a source-tree
// grep concluded "missing" both times. Everything here reads dist/landing.
//
// These are the gates the SEO plan lists as non-negotiable, plus the two
// that were actually violated when they were first checked:
//
//   - exactly one <h1> per page, no skipped heading levels. Seven pages
//     failed this. Two of them failed because of a shared footer whose
//     column labels were <h3>, so /demo/ and /404/ went h1 -> h3 with
//     nothing wrong on the pages themselves.
//   - no price asserted in JSON-LD anywhere. rtylr publishes no prices.
//   - no FAQPage. Google removed the rich result in May 2026, and the
//     markup here had become the only copy of the answers in the served
//     HTML, which made deleting it a content regression until the visible
//     sections landed.
//   - FAQ answers present in the visible HTML, not only in a script tag.
// ---------------------------------------------------------------------------

const REPO = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../..');
const OUT = path.join(REPO, 'dist/landing');

const LD = /<script type="application\/ld\+json"[^>]*>(.*?)<\/script>/gs;
const HEADING = /<h([1-6])[^>]*>/g;

let pages = [];

before(() => {
  assert.ok(
    existsSync(OUT),
    `no build output at ${OUT}. Run \`pnpm build\` before \`pnpm test:unit\`.`
  );
  pages = globSync('**/index.html', { cwd: OUT }).sort();
  // A silent zero here would make every test below vacuously pass, which
  // is the exact "check that cannot fail" this project keeps hitting.
  assert.ok(pages.length >= 20, `expected 20+ exported pages, found ${pages.length}`);
});

const read = (p) => readFileSync(path.join(OUT, p), 'utf8');

describe('release artifact', () => {
  it('publishes a version document', () => {
    const data = JSON.parse(read('version.json'));
    assert.match(data.version, /^\d+\.\d+\.\d+/);
  });
});
const jsonLd = (html) => [...html.matchAll(LD)].map((m) => m[1]);
const stripScripts = (html) => html.replace(/<script.*?<\/script>/gs, '');

describe('heading structure', () => {
  it('every page has exactly one h1', () => {
    const offenders = pages
      .map((p) => [p, [...read(p).matchAll(HEADING)].filter((m) => m[1] === '1').length])
      .filter(([, n]) => n !== 1);
    assert.deepEqual(offenders, [], 'pages whose h1 count is not 1');
  });

  it('no page skips a heading level', () => {
    const offenders = [];
    for (const p of pages) {
      const levels = [...read(p).matchAll(HEADING)].map((m) => Number(m[1]));
      const skips = levels
        .slice(1)
        .map((lvl, i) => [levels[i], lvl])
        .filter(([prev, lvl]) => lvl > prev + 1);
      if (skips.length) offenders.push([p, skips[0]]);
    }
    assert.deepEqual(offenders, [], 'pages that jump more than one heading level');
  });
});

// Snippet bands. A title outside 25 to 65 characters or a description
// outside 110 to 170 gets rewritten or truncated by Google, so the two lines
// a searcher actually reads stop being the two lines we wrote.
//
// Two traps are baked in deliberately. Next stamps `data-next-head` on every
// head tag it owns, so a `<title>` pattern without an attribute allowance
// matches nothing and the whole gate passes on an empty set. And the served
// HTML is entity-encoded, so `&#x27;` has to be decoded before anything is
// counted or every description holding an apostrophe is measured wrong.
// Both are asserted below rather than assumed.
const TITLE = /<title[^>]*>(.*?)<\/title>/s;
const DESCRIPTION = /<meta name="description" content="(.*?)"/s;

const ENTITIES = { amp: '&', lt: '<', gt: '>', quot: '"', apos: "'", nbsp: ' ' };
const decode = (s) =>
  s
    .replace(/&#x([0-9a-f]+);/gi, (_, h) => String.fromCodePoint(parseInt(h, 16)))
    .replace(/&#(\d+);/g, (_, d) => String.fromCodePoint(Number(d)))
    .replace(/&(amp|lt|gt|quot|apos|nbsp);/g, (_, n) => ENTITIES[n]);

// /404/ is not indexable and has no snippet to win.
const SNIPPET_PAGES = () => pages.filter((p) => p !== '404/index.html');

describe('snippet lengths', () => {
  it('every indexable page has a non-empty title and description', () => {
    const missing = SNIPPET_PAGES().filter((p) => {
      const html = read(p);
      return !TITLE.exec(html)?.[1] || !DESCRIPTION.exec(html)?.[1];
    });
    assert.deepEqual(missing, [], 'a page with no title or description measures as zero, not as a failure');
  });

  it('every title is 25 to 65 characters', () => {
    const offenders = [];
    for (const p of SNIPPET_PAGES()) {
      const title = decode(TITLE.exec(read(p))?.[1] ?? '');
      const n = [...title].length;
      if (n < 25 || n > 65) offenders.push([p, n, title]);
    }
    assert.deepEqual(offenders, [], 'titles outside the 25 to 65 character band');
  });

  it('every meta description is 110 to 170 characters', () => {
    const offenders = [];
    for (const p of SNIPPET_PAGES()) {
      const desc = decode(DESCRIPTION.exec(read(p))?.[1] ?? '');
      const n = [...desc].length;
      if (n < 110 || n > 170) offenders.push([p, n, desc]);
    }
    assert.deepEqual(offenders, [], 'descriptions outside the 110 to 170 character band');
  });
});

describe('structured data', () => {
  it('every page emits at least one JSON-LD block', () => {
    const missing = pages.filter((p) => jsonLd(read(p)).length === 0);
    assert.deepEqual(missing, [], 'pages with no structured data');
  });

  it('every JSON-LD block parses', () => {
    for (const p of pages) {
      for (const block of jsonLd(read(p))) {
        assert.doesNotThrow(() => JSON.parse(block), `invalid JSON-LD on ${p}`);
      }
    }
  });

  it('no page asserts a price', () => {
    const offenders = [];
    for (const p of pages) {
      for (const block of jsonLd(read(p))) {
        for (const key of ['"price"', '"lowPrice"', '"highPrice"', '"priceCurrency"', '"Offer"']) {
          if (block.includes(key)) offenders.push([p, key]);
        }
      }
    }
    assert.deepEqual(offenders, [], 'rtylr publishes no prices, so none may be marked up');
  });

  it('no page emits FAQPage', () => {
    const offenders = pages.filter((p) => jsonLd(read(p)).some((b) => b.includes('"FAQPage"')));
    assert.deepEqual(offenders, [], 'the FAQ rich result was removed by Google in May 2026');
  });

  // The WebSite schema used to advertise a SearchAction pointing at
  // /search?q={search_term_string}. No such route is exported, so Google
  // followed the markup and logged a 404. Sitelinks search box, the only
  // feature the block fed, was retired, so there is nothing to restore.
  it('no page advertises a search endpoint', () => {
    const offenders = [];
    for (const p of pages) {
      for (const block of jsonLd(read(p))) {
        for (const needle of ['SearchAction', '/search?']) {
          if (block.includes(needle)) offenders.push([p, needle]);
        }
      }
    }
    assert.deepEqual(offenders, [], 'structured data must not point at a route the export lacks');
  });

  it('a legal page is not a software application', () => {
    const types = jsonLd(read('legal/privacy-policy/index.html'))
      .flatMap((b) => {
        const d = JSON.parse(b);
        return d['@graph'] ?? [d];
      })
      .map((n) => n['@type']);
    assert.ok(types.includes('WebPage'));
    assert.ok(!types.includes('SoftwareApplication'), 'privacy policy claimed to be an app');
    assert.ok(!types.includes('Product'));
  });

  it('each module page declares its own applicationSubCategory', () => {
    const seen = new Map();
    for (const mod of ['pos', 'erp', 'crm', 'hr', 'finance', 'flow']) {
      const sub = jsonLd(read(`${mod}/index.html`))
        .flatMap((b) => {
          const d = JSON.parse(b);
          return d['@graph'] ?? [d];
        })
        .find((n) => n.applicationSubCategory)?.applicationSubCategory;
      assert.ok(sub, `${mod} has no applicationSubCategory`);
      seen.set(mod, sub);
    }
    // Distinct, or the sub-category conveys nothing: every one of these
    // pages already shared an identical applicationCategory.
    assert.equal(new Set(seen.values()).size, seen.size, `sub-categories collide: ${[...seen]}`);
  });
});

describe('FAQ answers reach a crawler', () => {
  // The point of removing FAQPage was that the questions were marked up
  // and rendered nowhere. If an answer is only inside a script tag, the
  // regression is back in the other direction.
  const CASES = [
    ['pricing/index.html', 'Upgrade and the new features unlock instantly'],
    ['pos/index.html', 'keeps a full local copy of the menu'],
    ['erp/index.html', 'stock transfers between locations'],
    ['crm/index.html', 'loyalty points and campaign touchpoints'],
    ['hr/index.html', 'tip pooling, minor work-hour limits'],
    ['finance/index.html', 'balance sheet and cash flow reports'],
    ['flow/index.html', 'boards, lists, timeline views, cycles'],
    ['index.html', 'stops stitching a register, a spreadsheet'],
    ['demo/index.html', 'the questions a signup page cannot answer'],
  ];

  for (const [page, needle] of CASES) {
    it(`${page} carries its answer text outside any script tag`, () => {
      assert.ok(
        stripScripts(read(page)).includes(needle),
        `answer text missing from the visible HTML of ${page}`
      );
    });
  }
});

describe('Open Graph images', () => {
  // HelmetContainer derives /og/<slug>.png from the canonical, and
  // generate-og-images.mjs derives the same slug from the export path.
  // Two copies of one rule, which is the shape that drifts. This asserts
  // the declared image is a file that exists, so a mismatch fails here
  // instead of rendering a broken unfurl on a shared link.
  const ogOf = (html) => html.match(/<meta property="og:image" content="([^"]+)"/)?.[1];

  it('every page declares an og:image', () => {
    const missing = pages.filter((p) => !ogOf(read(p)));
    assert.deepEqual(missing, [], 'pages with no og:image');
  });

  it('every declared og:image exists in the build output', () => {
    const broken = [];
    for (const p of pages) {
      const url = ogOf(read(p));
      const rel = url.replace('https://rtylr.com/', '');
      if (!existsSync(path.join(OUT, rel))) broken.push([p, url]);
    }
    assert.deepEqual(broken, [], 'og:image URLs pointing at files that do not exist');
  });

  // A page's export path, as basePathOf() expects it: 'ar/pricing/index.html'
  // -> '/ar/pricing/', 'index.html' -> '/'.
  const routeOf = (p) => `/${p.replace(/index\.html$/, '')}`;

  it('no two indexable pages share the same og:image', () => {
    // The defect this replaces: one og-default.png on all 24 pages, so
    // every shared link looked the same. 404 is allowed the default.
    //
    // A locale sibling of the same route is allowed to share: /pricing/,
    // /ar/pricing/ and /fr/pricing/ all cover one route and one image is
    // generated for it, not one per route per locale. basePathOf() is what
    // tells apart "same route, different locale" (fine) from "different
    // route, same image" (the actual bug this test exists to catch).
    const byImage = new Map();
    for (const p of pages.filter((x) => x !== '404/index.html')) {
      const url = ogOf(read(p));
      byImage.set(url, [...(byImage.get(url) ?? []), p]);
    }
    const shared = [...byImage]
      .filter(([, ps]) => ps.length > 1)
      .filter(([, ps]) => new Set(ps.map((p) => basePathOf(routeOf(p)))).size > 1);
    assert.deepEqual(shared, [], 'pages sharing an OG image across different routes');
  });

  it('the og image dimensions match the declared width and height', () => {
    // 1200x630 is asserted in the meta tags, so the PNGs must be that.
    // Read the IHDR chunk rather than trusting the generator.
    for (const p of pages) {
      const rel = ogOf(read(p)).replace('https://rtylr.com/', '');
      const buf = readFileSync(path.join(OUT, rel));
      assert.equal(buf.readUInt32BE(16), 1200, `${rel} width`);
      assert.equal(buf.readUInt32BE(20), 630, `${rel} height`);
    }
  });
});

describe('translated copy reaches the page', () => {
  // Two different failures, one per gate, and it is worth being exact
  // about which is which because they look identical on the page:
  //
  //   a key a component asks for that en.json does not define
  //     -> tests/unit/i18n-keys, at the source. NOT this gate: the key
  //        list here is derived from en.json, so a key missing from
  //        en.json is absent from both sides and cannot be compared.
  //
  //   a key that exists in en.json and still renders as its own name
  //     -> this gate. That is a resolution failure: the provider not
  //        mounted, a namespace not loaded, a build that dropped the
  //        locale file. The source looks perfectly correct.
  //
  // Verified by injecting "integrations.api.docs" into a built page.
  //
  // It checks for the EXACT keys en.json defines rather than for anything
  // shaped like a dotted word. The shape approach failed both ways: it
  // hardcoded five namespaces while the file grew to fourteen, so real
  // leaks went unseen, and it flagged a false positive, because
  // /integrations/ renders an API sample payload containing
  // "integrations.delivery.uber", which is page content and not a key.
  const KEYS = (() => {
    const flatten = (obj, prefix = '', out = []) => {
      for (const [k, v] of Object.entries(obj)) {
        const key = prefix ? `${prefix}.${k}` : k;
        if (v && typeof v === 'object' && !Array.isArray(v)) flatten(v, key, out);
        else out.push(key);
      }
      return out;
    };
    const en = JSON.parse(readFileSync(path.join(REPO, 'locales/en.json'), 'utf8'));
    return flatten(en);
  })();

  it('has keys to check against', () => {
    // Guards the whole check: an empty list would pass everything.
    assert.ok(KEYS.length > 100, `expected 100+ keys, found ${KEYS.length}`);
  });

  it('no page renders a raw translation key', () => {
    const offenders = [];
    for (const p of pages) {
      const visible = stripScripts(read(p)).replace(/<[^>]+>/g, ' ');
      for (const key of KEYS) {
        if (visible.includes(key)) offenders.push([p, key]);
      }
    }
    assert.deepEqual(offenders, [], 'untranslated keys rendered as their own name');
  });

  it('the footer still renders its English labels', () => {
    // The extraction moved hundreds of strings out of the markup. If the
    // provider ever stops resolving, the page does not crash, it just loses
    // its text, which is the kind of thing a build log never shows.
    const visible = stripScripts(read('pos/index.html'));
    for (const label of ['PRODUCT', 'RESOURCES', 'DEVELOPERS', 'COMPANY', 'TERMS OF USE']) {
      assert.ok(visible.includes(label), `footer lost "${label}"`);
    }
  });
});

describe('optional third-party embeds', () => {
  it('does not load the Google publisher library before a visitor chooses', () => {
    for (const p of pages) {
      assert.ok(!read(p).includes('https://news.google.com/swg/js/v1/publisher.js'), p);
      assert.ok(!read(p).includes('google-add-preferred-source-btn'), p);
    }
  });
});

// lastmod was `new Date()` for every URL, so all 23 entries carried the build
// time and Google had no way to tell the day's new post from the legal page
// nobody touched. Two assertions, because either alone can pass on a broken
// sitemap: that the values are not uniform, and that the blog values are the
// dates the posts actually declare.
// Thin pages. /demo/ and /contact/ are where every CTA on the site points,
// and they were a heading and a form: 139 and 164 words of visible text.
// Google has nothing to match on a page that thin and an extractor has
// nothing to lift.
//
// The threshold is the one the audit script uses. This gate exists because
// every other check on that fix lives either in production or in a script
// that is not merged, so nothing here would notice the copy being deleted.
describe('page word counts', () => {
  const MINIMUM = 250;
  const COUNTED = ['demo/index.html', 'contact/index.html', 'blog/index.html'];

  const visibleWords = (html) =>
    stripScripts(html)
      .replace(/<style[\s\S]*?<\/style>/g, '')
      .replace(/<[^>]+>/g, ' ')
      .replace(/&[a-z]+;|&#x?[0-9a-f]+;/gi, ' ')
      .split(/\s+/)
      .filter(Boolean).length;

  for (const page of COUNTED) {
    it(`${page} carries more than ${MINIMUM} words of visible text`, () => {
      const n = visibleWords(read(page));
      assert.ok(n > MINIMUM, `${page} has ${n} visible words, under the ${MINIMUM} threshold`);
    });
  }
});

describe('sitemap lastmod', () => {
  const SITEMAP = path.join(OUT, 'sitemap-0.xml');
  let entries = [];

  before(() => {
    assert.ok(existsSync(SITEMAP), `no sitemap at ${SITEMAP}. Run \`pnpm build\` first.`);
    const xml = readFileSync(SITEMAP, 'utf8');
    entries = [...xml.matchAll(/<url>(.*?)<\/url>/gs)].map((m) => ({
      loc: /<loc>(.*?)<\/loc>/s.exec(m[1])?.[1] ?? '',
      lastmod: /<lastmod>(.*?)<\/lastmod>/s.exec(m[1])?.[1] ?? '',
    }));
    // Same guard as the page fixture: a zero-length parse would make every
    // assertion below vacuously true.
    assert.ok(entries.length >= 20, `expected 20+ sitemap URLs, found ${entries.length}`);
  });

  it('every URL carries a lastmod', () => {
    const missing = entries.filter((e) => !e.lastmod).map((e) => e.loc);
    assert.deepEqual(missing, [], 'sitemap URLs with no lastmod');
  });

  it('lastmod is not one uniform build timestamp', () => {
    const distinct = new Set(entries.map((e) => e.lastmod));
    assert.ok(
      distinct.size >= 2,
      `every URL shares lastmod ${[...distinct][0]}, which is the build time bug`
    );
  });

  it('each blog post carries the date it declares', () => {
    const posts = JSON.parse(readFileSync(path.join(REPO, 'lib/blog/posts.json'), 'utf8'));
    assert.ok(posts.length > 0, 'no posts to check');
    const offenders = [];
    for (const post of posts) {
      const entry = entries.find((e) => e.loc.endsWith(`/blog/${post.slug}/`));
      assert.ok(entry, `${post.slug} is missing from the sitemap`);
      const expected = new Date(post.dateModified || post.date).toISOString().slice(0, 10);
      if (entry.lastmod.slice(0, 10) !== expected) {
        offenders.push([post.slug, entry.lastmod, expected]);
      }
    }
    assert.deepEqual(offenders, [], 'blog lastmod does not match posts.json');
  });
});

describe('canonical URLs', () => {
  // /404/ is the one page that should not claim a canonical, because it
  // is not a page anyone should land on deliberately. It is excluded here
  // and held to the stricter rule below instead.
  const INDEXABLE = () => pages.filter((p) => p !== '404/index.html');

  it('every indexable page declares a canonical on rtylr.com', () => {
    const offenders = [];
    for (const p of INDEXABLE()) {
      const m = read(p).match(/<link rel="canonical" href="([^"]+)"/);
      if (!m || !m[1].startsWith('https://rtylr.com')) offenders.push([p, m?.[1] ?? null]);
    }
    assert.deepEqual(offenders, [], 'pages with a missing or off-host canonical');
  });

  it('every canonical ends in a trailing slash, matching the served URL', () => {
    // trailingSlash: true in next.config.js. A canonical without the
    // slash points at a URL that redirects, which is a self-inflicted
    // redirect on every page.
    const offenders = [];
    for (const p of INDEXABLE()) {
      const href = read(p).match(/<link rel="canonical" href="([^"]+)"/)?.[1];
      if (href && !href.endsWith('/')) offenders.push([p, href]);
    }
    assert.deepEqual(offenders, [], 'canonicals that do not match the served trailing-slash URL');
  });

  it('404 is noindex and claims no canonical', () => {
    const html = read('404/index.html');
    assert.match(html, /<meta name="robots" content="noindex/);
    assert.ok(!/<link rel="canonical"/.test(html), '404 should not claim to be canonical');
  });

  it('no indexable page carries a noindex', () => {
    const offenders = INDEXABLE().filter((p) => /content="noindex/.test(read(p)));
    assert.deepEqual(offenders, [], 'pages accidentally excluded from the index');
  });
});
