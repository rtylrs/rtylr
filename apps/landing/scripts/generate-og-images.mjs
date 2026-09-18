#!/usr/bin/env node
/**
 * Per-page Open Graph images.
 *
 * One og-default.png across every page wastes every share: the link to
 * /pricing/ and the link to /legal/terms/ look identical in a Slack
 * unfurl, so the preview carries no information. This renders one image
 * per exported page, headlined with that page's own H1.
 *
 * ## Run it by hand, not in the build
 *
 *     pnpm build                 # produces dist/landing
 *     node apps/landing/scripts/generate-og-images.mjs
 *     pnpm build                 # copies public/og into the export
 *
 * Deliberately NOT wired into `pnpm build`. It needs a Chromium binary,
 * and the deploy path here is a GitOps promotion that builds a tag in CI:
 * adding a browser download to that build is a large change to the
 * riskiest part of the pipeline in exchange for images that change only
 * when page headlines change. The PNGs are committed instead, which is
 * the same offline-generation approach the plan records for voxire.
 *
 * ## Where the content comes from
 *
 * The headline is read out of each built index.html, not from a list
 * kept here. A list would drift from the pages the moment someone
 * reworded an H1, and this file would keep confidently rendering the old
 * one. `--check` exists for exactly that: it reports pages whose
 * headline no longer matches the committed PNG's recorded headline.
 *
 * ## Fonts, and why they are inlined as base64
 *
 * On this project a cover generator once shipped 805 images in the wrong
 * typeface with every check green, because naming a font family that is
 * not installed substitutes silently rather than failing. So no family
 * name is resolved from the system anywhere here: the two vendored TTFs
 * are read off disk and embedded in the template as data URIs, which
 * removes the lookup that failed silently last time.
 *
 * That is a smaller guarantee than it looks, and it was measured rather
 * than assumed. An empty @font-face src and a family name that does not
 * exist render identically in Chromium: both fall back to the default
 * face and lay out text at the same metrics. Omitting a fallback from
 * the font stack does NOT make the text disappear, so "the image is not
 * blank" proves nothing about which typeface produced it.
 *
 * What actually catches substitution is `--verify-font`, which renders
 * the same headline twice, once normally and once with the @font-face
 * src emptied, and asserts the two PNGs differ. That comparison fails if
 * the embedding ever stops taking effect. Nothing else here would.
 *
 * Note the faces are Plus Jakarta Sans and Inter, not the site's DM
 * Sans. DM Sans arrives through next/font as content-hashed woff2 whose
 * filenames change every build and whose family cannot be read back out
 * of the compressed file, so depending on it would reintroduce the drift
 * this script exists to avoid. Both faces used here are already served
 * by the site from public/fonts. Drop a DM Sans TTF in there and change
 * FONTS below if exact brand fidelity is wanted.
 */

import { readFileSync, writeFileSync, mkdirSync, existsSync, readdirSync, statSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const HERE = path.dirname(fileURLToPath(import.meta.url));
const APP = path.resolve(HERE, '..');
const EXPORT_DIR = path.join(APP, 'dist/landing');
const OUT_DIR = path.join(APP, 'public/og');
const MANIFEST = path.join(OUT_DIR, 'manifest.json');

const WIDTH = 1200;
const HEIGHT = 630;

const FONTS = {
  bold: 'PlusJakartaSans-ExtraBold.ttf',
  body: 'Inter-Regular.ttf',
};

// ---------------------------------------------------------------------------
// Page discovery, from the build output
// ---------------------------------------------------------------------------

/** Every exported index.html, as { route, slug, file }. */
function findPages() {
  const found = [];
  const walk = (dir) => {
    for (const entry of readdirSync(dir)) {
      const full = path.join(dir, entry);
      if (statSync(full).isDirectory()) {
        if (entry === '_next') continue;
        walk(full);
      } else if (entry === 'index.html') {
        const rel = path.relative(EXPORT_DIR, dir).split(path.sep).filter(Boolean);
        found.push({
          route: rel.length ? `/${rel.join('/')}/` : '/',
          slug: rel.length ? rel.join('-') : 'home',
          file: full,
        });
      }
    }
  };
  walk(EXPORT_DIR);
  return found.sort((a, b) => a.route.localeCompare(b.route));
}

const stripTags = (s) =>
  s
    .replace(/<[^>]+>/g, ' ')
    .replace(/&amp;/g, '&')
    .replace(/&#x27;|&apos;/g, "'")
    .replace(/&quot;/g, '"')
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/&nbsp;/g, ' ')
    .replace(/&#(\d+);/g, (_, n) => String.fromCharCode(Number(n)))
    .replace(/\s+/g, ' ')
    // Tags become spaces above, which is right between words and wrong
    // before punctuation. SplitWords puts every token of an h1 in its own
    // span, so "keeps up." arrived here as "keeps up ." and rendered that
    // way on the image.
    .replace(/\s+([.,;:!?%)\]])/g, '$1')
    .replace(/([(\[])\s+/g, '$1')
    .trim();

/**
 * The page's own H1, plus the eyebrow above it where one exists. Falls
 * back to <title> minus the site suffix, and throws if neither is
 * present: a page with no headline is a page whose OG image would be
 * blank, and that should stop the run rather than ship.
 */
function readHeadline({ file, route }) {
  const html = readFileSync(file, 'utf8');

  const h1 = html.match(/<h1[^>]*>([\s\S]*?)<\/h1>/);
  let headline = h1 ? stripTags(h1[1]) : '';

  if (!headline) {
    const title = html.match(/<title[^>]*>([\s\S]*?)<\/title>/);
    headline = title ? stripTags(title[1]).split(/\s[|·]\s/)[0].trim() : '';
  }
  if (!headline) throw new Error(`no h1 and no usable title on ${route}`);

  const desc = html.match(/<meta name="description" content="([^"]*)"/);
  const label = route === '/' ? 'rtylr.com' : route.replace(/^\/|\/$/g, '').replace(/\//g, ' / ');

  return { headline, label, description: desc ? stripTags(desc[1]) : '' };
}

// ---------------------------------------------------------------------------
// Template
// ---------------------------------------------------------------------------

function fontDataUri(fileName) {
  const p = path.join(APP, 'public/fonts', fileName);
  if (!existsSync(p)) throw new Error(`vendored font missing: ${p}`);
  return `data:font/ttf;base64,${readFileSync(p).toString('base64')}`;
}

const escapeHtml = (s) =>
  s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');

/**
 * @param opts.breakFont when true the @font-face src points at nothing.
 *        Used only by --verify-font, to prove the real embedding matters.
 */
function template({ headline, label, breakFont = false }) {
  const bold = breakFont ? 'data:font/ttf;base64,' : fontDataUri(FONTS.bold);
  const body = breakFont ? 'data:font/ttf;base64,' : fontDataUri(FONTS.body);

  // Long headlines get a smaller size rather than an overflow. Measured
  // in characters, which is crude and good enough for two or three lines
  // of a display face at fixed width.
  const n = headline.length;
  const size = n > 78 ? 62 : n > 52 ? 74 : n > 32 ? 88 : 104;

  return `<!doctype html>
<meta charset="utf-8">
<style>
  @font-face { font-family: 'OGBold'; src: url('${bold}') format('truetype'); }
  @font-face { font-family: 'OGBody'; src: url('${body}') format('truetype'); }
  * { margin: 0; padding: 0; box-sizing: border-box; }
  html, body { width: ${WIDTH}px; height: ${HEIGHT}px; }
  body {
    background: #fafafa;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    padding: 72px 80px;
    /* No fallback listed on purpose, so the intended face is the only one
       named. Be clear about what that does and does not buy: Chromium
       still falls back to its default face if 'OGBody' fails to load, and
       the text renders at the same metrics, so this alone cannot be
       detected by looking at the output. --verify-font is what detects
       it. */
    font-family: 'OGBody';
    color: #131313;
  }
  .top { display: flex; align-items: center; justify-content: space-between; }
  .mark {
    font-family: 'OGBold';
    font-size: 34px;
    letter-spacing: -0.02em;
  }
  .label {
    font-family: 'OGBody';
    font-size: 20px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: rgba(19, 19, 19, 0.45);
  }
  h1 {
    font-family: 'OGBold';
    font-size: ${size}px;
    line-height: 1.06;
    letter-spacing: -0.035em;
    max-width: 15.5em;
  }
  .rule { height: 3px; background: #131313; width: 96px; margin-bottom: 28px; }
  .foot {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    font-size: 21px;
    color: rgba(19, 19, 19, 0.55);
  }
  .apps { letter-spacing: 0.01em; }
</style>
<body>
  <div class="top">
    <span class="mark">rtylr</span>
    <span class="label">${escapeHtml(label)}</span>
  </div>
  <div>
    <div class="rule"></div>
    <h1>${escapeHtml(headline)}</h1>
  </div>
  <div class="foot">
    <span>rtylr.com</span>
    <span class="apps">POS · ERP · CRM · HR · Finance · Flow</span>
  </div>
</body>`;
}

// ---------------------------------------------------------------------------
// Render
// ---------------------------------------------------------------------------

async function withPage(fn) {
  let chromium;
  try {
    ({ chromium } = await import('playwright'));
  } catch {
    try {
      ({ chromium } = await import('@playwright/test'));
    } catch {
      throw new Error(
        'Playwright is not resolvable. It is a devDependency of the workspace root:\n' +
          '  pnpm install && pnpm exec playwright install chromium'
      );
    }
  }
  const browser = await chromium.launch();
  try {
    const page = await browser.newPage({
      viewport: { width: WIDTH, height: HEIGHT },
      deviceScaleFactor: 1,
    });
    return await fn(page);
  } finally {
    await browser.close();
  }
}

async function shoot(page, html) {
  await page.setContent(html, { waitUntil: 'load' });
  // Fonts are data URIs so there is no network wait, but the face still
  // has to be parsed before the first paint uses it.
  await page.evaluate(() => document.fonts.ready);
  return page.screenshot({ type: 'png' });
}

// ---------------------------------------------------------------------------
// Modes
// ---------------------------------------------------------------------------

async function generate() {
  const pages = findPages();
  if (!pages.length) {
    throw new Error(`no exported pages under ${EXPORT_DIR}. Run \`pnpm build\` first.`);
  }
  mkdirSync(OUT_DIR, { recursive: true });

  const manifest = {};
  await withPage(async (page) => {
    for (const p of pages) {
      const { headline, label } = readHeadline(p);
      const png = await shoot(page, template({ headline, label }));
      writeFileSync(path.join(OUT_DIR, `${p.slug}.png`), png);
      // Record the headline each PNG was rendered from, so --check can
      // tell a stale image from a current one without re-rendering.
      manifest[p.slug] = { route: p.route, headline };
      console.log(`  ${`${p.slug}.png`.padEnd(34)} ${headline.slice(0, 52)}`);
    }
  });

  writeFileSync(MANIFEST, `${JSON.stringify(manifest, null, 2)}\n`);
  console.log(`\n[og-images] ${pages.length} images written to public/og`);
}

/** Report pages whose headline has changed since the PNG was rendered. */
function check() {
  const pages = findPages();
  if (!existsSync(MANIFEST)) {
    console.error('[og-images] no manifest. Run without --check to generate.');
    process.exit(1);
  }
  const manifest = JSON.parse(readFileSync(MANIFEST, 'utf8'));
  const problems = [];

  for (const p of pages) {
    const png = path.join(OUT_DIR, `${p.slug}.png`);
    if (!existsSync(png)) {
      problems.push(`missing image for ${p.route} (expected public/og/${p.slug}.png)`);
      continue;
    }
    const { headline } = readHeadline(p);
    const recorded = manifest[p.slug]?.headline;
    if (recorded !== headline) {
      problems.push(`stale image for ${p.route}\n      png: ${recorded}\n      now: ${headline}`);
    }
  }
  for (const slug of Object.keys(manifest)) {
    if (!pages.some((p) => p.slug === slug)) problems.push(`orphaned image: ${slug}.png`);
  }

  if (problems.length) {
    console.error(`[og-images] ${problems.length} problem(s):`);
    for (const p of problems) console.error(`    ${p}`);
    process.exit(1);
  }
  console.log(`[og-images] ${pages.length} images current`);
}

/**
 * Prove the embedded font is the one being used, by rendering the same
 * headline with the font-face src emptied and asserting the output
 * differs. Without this, "the font is named in the CSS" is all anyone
 * ever verifies, and that is exactly the check that passed while every
 * cover shipped in the wrong typeface.
 *
 * Expect the broken render to be LARGER, not empty: it falls back to the
 * default serif, which compresses worse than the display face. A failure
 * here means the two are byte-identical, which is the case where the
 * embedded font was never taking effect in the first place.
 */
async function verifyFont() {
  const headline = 'Handgloves 123 quick brown fox';
  const [real, broken] = await withPage(async (page) => [
    await shoot(page, template({ headline, label: 'font check' })),
    await shoot(page, template({ headline, label: 'font check', breakFont: true })),
  ]);

  console.log(`  with embedded font: ${real.length} bytes`);
  console.log(`  with font removed : ${broken.length} bytes`);

  if (real.equals(broken)) {
    console.error(
      '\n[og-images] FAIL: removing the font changed nothing, so the embedded\n' +
        '            face is not being used. Text is rendering in something else.'
    );
    process.exit(1);
  }
  console.log('\n[og-images] font check passed: the embedded face is what renders.');
}

const arg = process.argv[2];
if (arg === '--check') check();
else if (arg === '--verify-font') await verifyFont();
else await generate();
