import { describe, it, before } from 'node:test';
import assert from 'node:assert/strict';
import { existsSync, globSync, readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

// ---------------------------------------------------------------------------
// The blog's editorial and structural gates.
//
// Posts are published by a scheduled task with nobody reading them first,
// so every rule a reviewer would otherwise hold a post to is asserted here.
// A failure stops the publish; the task is not allowed to work around one.
//
// Source checks read content/blog. Build checks read dist/landing, so run
// `pnpm build` first, as for the rest of tests/unit.
// ---------------------------------------------------------------------------

const APP = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../..');
const POSTS = path.join(APP, 'content/blog/posts');
const OUT = path.join(APP, 'dist/landing');

const SEGMENTS = ['small', 'mid', 'large', 'vision'];
const TYPES = ['guide', 'comparison', 'template', 'rules', 'essay', 'shipped', 'market'];
const FORMATS = ['search', 'essay', 'legacy'];
const AUTHORS = ['abed-el-fattah-amouneh', 'mohammad-homsi'];
const BLOCKS = ['h2', 'h3', 'p', 'quote', 'ul', 'ol', 'table', 'cta'];
const MOTIFS = ['ledger', 'board', 'chart', 'schedule', 'stock', 'receipt', 'caret'];
const RESERVED = ['small-business', 'mid-size', 'enterprise', 'vision', 'tag', 'author'];

const TAGS = JSON.parse(readFileSync(path.join(APP, 'content/blog/tags.json'), 'utf8'));

const posts = globSync('*/*.json', { cwd: POSTS })
  .sort()
  .map((rel) => ({ rel, post: JSON.parse(readFileSync(path.join(POSTS, rel), 'utf8')) }));

/** Every string a post renders, for the rules that apply to all text. */
function textsOf(post) {
  const out = [post.title, post.metaTitle, post.description, post.excerpt];
  for (const b of post.blocks) {
    if ('text' in b) out.push(b.text);
    if ('items' in b) out.push(...b.items);
    if (b.type === 'table') out.push(b.caption ?? '', ...b.head, ...b.rows.flat());
  }
  for (const f of post.faq ?? []) out.push(f.question, f.answer);
  for (const r of post.cover.rows) out.push(r.label, r.value ?? '');
  out.push(post.cover.title);
  return out.filter(Boolean);
}

/** Body text only, markup removed: what the word counts measure. */
const bodyText = (post) =>
  post.blocks
    .flatMap((b) => ('text' in b ? [b.text] : 'items' in b ? b.items : b.type === 'table' ? b.rows.flat() : []))
    .concat((post.faq ?? []).flatMap((f) => [f.question, f.answer]))
    .join(' ')
    .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1')
    .replace(/\*\*/g, '');

const wordCount = (s) => s.split(/\s+/).filter(Boolean).length;

const LINK = /\[[^\]]+\]\(([^)\s]+)\)/g;
const linksOf = (post) => textsOf(post).flatMap((s) => [...s.matchAll(LINK)].map((m) => m[1]));
const externalLinks = (post) => [...new Set(linksOf(post).filter((h) => /^https?:\/\//.test(h)))];

const isQuestion = (s) => /[?؟]\s*$/.test(s);

describe('post files', () => {
  it('there are posts to check', () => {
    assert.ok(posts.length >= 3, `expected 3+ post files, found ${posts.length}`);
  });

  for (const { rel, post } of posts) {
    describe(rel, () => {
      it('lives where it says it lives', () => {
        assert.equal(rel, `${post.slug}/${post.locale}.json`);
        assert.match(post.slug, /^[a-z0-9]+(-[a-z0-9]+)*$/);
        assert.ok(!RESERVED.includes(post.slug), 'slug collides with a blog route');
      });

      it('declares a valid segment, type, format, author and tags', () => {
        assert.ok(SEGMENTS.includes(post.segment), `segment ${post.segment}`);
        assert.ok(TYPES.includes(post.type), `type ${post.type}`);
        assert.ok(FORMATS.includes(post.format), `format ${post.format}`);
        assert.ok(AUTHORS.includes(post.author), `author ${post.author}`);
        assert.ok(post.tags.length >= 1 && post.tags.length <= 5, 'one to five tags');
        for (const tag of post.tags) {
          assert.ok(TAGS[tag]?.en && TAGS[tag]?.ar, `tag ${tag} has no label in tags.json`);
        }
      });

      it('has real dates', () => {
        assert.match(post.date, /^\d{4}-\d{2}-\d{2}$/);
        if (post.dateModified) {
          assert.match(post.dateModified, /^\d{4}-\d{2}-\d{2}$/);
          assert.ok(post.dateModified >= post.date, 'modified before it was published');
        }
      });

      it('uses only known block types and a well-formed cover', () => {
        for (const b of post.blocks) assert.ok(BLOCKS.includes(b.type), `block type ${b.type}`);
        for (const b of post.blocks.filter((x) => x.type === 'table')) {
          assert.ok(b.head.length >= 2, 'a table needs two columns');
          for (const row of b.rows) assert.equal(row.length, b.head.length, 'ragged table row');
        }
        assert.ok(MOTIFS.includes(post.cover.motif), `cover motif ${post.cover.motif}`);
        assert.ok(post.cover.rows.length >= 3 && post.cover.rows.length <= 5, 'cover needs 3 to 5 rows');
      });

      it('contains no em dash, en dash or hyphen used as a dash', () => {
        const offenders = textsOf(post).filter((s) => /[—–]|\s-{1,2}\s/.test(s));
        assert.deepEqual(offenders, [], 'dashes are not allowed in post text');
      });

      it('writes the brand in lowercase', () => {
        const offenders = textsOf(post).filter((s) => /\b(RTYLR|Rtylr)\b/.test(s.replace(LINK, '')));
        assert.deepEqual(offenders, []);
      });

      if (post.format === 'search') {
        it('answers in the first 60 words', () => {
          const first = post.blocks[0];
          assert.equal(first?.type, 'p', 'a search post opens with its answer paragraph');
          const n = wordCount(first.text.replace(LINK, ' x '));
          assert.ok(n <= 60, `opening answer is ${n} words`);
        });

        it('is the right length', () => {
          const n = wordCount(bodyText(post));
          const [lo, hi] = post.locale === 'en' ? [1200, 2000] : [900, 2200];
          assert.ok(n >= lo && n <= hi, `${n} words, expected ${lo} to ${hi}`);
        });

        it('has 5 to 8 question headings', () => {
          const q = post.blocks.filter((b) => (b.type === 'h2' || b.type === 'h3') && isQuestion(b.text));
          assert.ok(q.length >= 5 && q.length <= 8, `${q.length} question headings`);
        });

        it('has an original table', () => {
          assert.ok(post.blocks.some((b) => b.type === 'table'), 'no table');
        });

        it('links at least four outside sources', () => {
          const n = externalLinks(post).length;
          assert.ok(n >= 4, `${n} outside sources`);
        });
      }

      if (post.format === 'essay') {
        it('links its sources', () => {
          const n = externalLinks(post).length;
          assert.ok(n >= 3, `${n} outside sources`);
        });
      }

      if (post.format !== 'legacy') {
        it('links every figure where it appears', () => {
          // A market size, a share or a price in a sentence with no link in it
          // is a number the reader has to take on trust.
          const FIGURE = /(\d[\d.,]*\s?(%|percent|billion|million|bn)|[$€£]\s?\d|\d\s?(مليار|مليون|بالمئة|٪))/i;
          const unsourced = post.blocks
            .flatMap((b) => ('text' in b ? [b.text] : 'items' in b ? b.items : []))
            .filter((s) => FIGURE.test(s.replace(LINK, '')) && !/\[[^\]]+\]\(https?:/.test(s));
          assert.deepEqual(unsourced, [], 'figures with no source link in the same block');
        });

        it('states a reading time that matches its length', () => {
          const perMinute = post.locale === 'en' ? 230 : 180;
          const expected = Math.max(1, Math.round(wordCount(bodyText(post)) / perMinute));
          assert.ok(
            Math.abs(post.readingMinutes - expected) <= 1,
            `readingMinutes ${post.readingMinutes}, length says ${expected}`
          );
        });
      }

      if (post.locale === 'ar') {
        it('is written in Arabic, not transliterated English', () => {
          // A short list of the loanwords that give a translated page away.
          // Arabic has its own word for each.
          const LOANS = ['سوفتوير', 'بزنس', 'داشبورد', 'سيستم', 'ماركتنج', 'ديجيتال', 'بلاتفورم', 'إنفنتوري', 'بيرول'];
          const offenders = textsOf(post).filter((s) => LOANS.some((w) => s.includes(w)));
          assert.deepEqual(offenders, []);
          const arabic = (bodyText(post).match(/[؀-ۿ]/g) ?? []).length;
          const latin = (bodyText(post).match(/[A-Za-z]/g) ?? []).length;
          assert.ok(arabic > latin * 4, 'an Arabic post that is mostly Latin script');
        });
      }
    });
  }
});

describe('language pairs', () => {
  const bySlug = new Map();
  for (const { post } of posts) bySlug.set(post.slug, [...(bySlug.get(post.slug) ?? []), post]);

  it('every Arabic post has an English original', () => {
    const orphans = [...bySlug].filter(([, ps]) => ps.some((p) => p.locale === 'ar') && !ps.some((p) => p.locale === 'en'));
    assert.deepEqual(orphans.map(([s]) => s), []);
  });

  it('the two versions agree on what the post is', () => {
    for (const [slug, ps] of bySlug) {
      if (ps.length < 2) continue;
      const [a, b] = ps;
      for (const k of ['segment', 'type', 'format', 'author', 'date']) {
        assert.equal(a[k], b[k], `${slug}: ${k} differs between languages`);
      }
      assert.deepEqual([...a.tags].sort(), [...b.tags].sort(), `${slug}: tags differ`);
    }
  });
});

describe('the built blog', () => {
  before(() => {
    assert.ok(existsSync(OUT), 'run pnpm build first');
  });

  const read = (rel) => readFileSync(path.join(OUT, rel), 'utf8');
  const urlOf = (post) => `${post.locale === 'en' ? '' : `/${post.locale}`}/blog/${post.slug}/`;
  const LD = /<script type="application\/ld\+json"[^>]*>(.*?)<\/script>/gs;

  it('exports every post', () => {
    const missing = posts.filter(({ post }) => !existsSync(path.join(OUT, urlOf(post), 'index.html')));
    assert.deepEqual(missing.map((p) => p.rel), []);
  });

  it('marks every post up as an article with a person as author, and breadcrumbs', () => {
    for (const { post } of posts) {
      const graphs = [...read(`${urlOf(post)}index.html`).matchAll(LD)].map((m) => JSON.parse(m[1]));
      const nodes = graphs.flatMap((g) => g['@graph'] ?? [g]);
      const article = nodes.find((n) => n['@type'] === 'BlogPosting');
      assert.ok(article, `${urlOf(post)} has no BlogPosting`);
      assert.equal(article.author?.['@type'], 'Person', `${urlOf(post)} author is not a Person`);
      assert.equal(article.inLanguage, post.locale);
      assert.ok(nodes.some((n) => n['@type'] === 'BreadcrumbList'), `${urlOf(post)} has no breadcrumbs`);
    }
  });

  it('shows a visible date on every post', () => {
    for (const { post } of posts) {
      const html = read(`${urlOf(post)}index.html`);
      assert.ok(html.includes(`<time dateTime="${post.date}"`), `${urlOf(post)} shows no date`);
    }
  });

  it('every internal link in a post resolves to an exported page', () => {
    const broken = [];
    for (const { post } of posts) {
      for (const href of linksOf(post).filter((h) => h.startsWith('/'))) {
        const route = href.split(/[?#]/)[0];
        const file = path.join(OUT, route.endsWith('/') ? route : `${route}/`, 'index.html');
        if (!existsSync(file)) broken.push([post.slug, post.locale, href]);
      }
    }
    assert.deepEqual(broken, [], 'internal links to pages that do not exist');
  });

  it('pairs every translated post in the sitemap', () => {
    const xml = read('sitemap-0.xml');
    const unpaired = [];
    for (const { post } of posts.filter(({ post: p }) => p.locale === 'ar')) {
      const block = xml.match(new RegExp(`<url><loc>https://rtylr.com${urlOf(post)}</loc>.*?</url>`, 's'))?.[0] ?? '';
      if (!block.includes(`https://rtylr.com/blog/${post.slug}/`)) unpaired.push(post.slug);
    }
    assert.deepEqual(unpaired, [], 'Arabic posts with no hreflang link to their English version');
  });
});
