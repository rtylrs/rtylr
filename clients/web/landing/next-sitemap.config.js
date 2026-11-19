/**
 * next-sitemap config - generates sitemap.xml + robots.txt at build
 * time so new pages are picked up without anyone remembering to edit
 * the static XML.
 *
 * We keep this separate from next.config.js because next-sitemap has
 * its own CLI and config loader. `pnpm run build` now runs
 * `next build && next-sitemap` - see the app's package.json.
 *
 * The generator reads `out/` (the SSG export dir) after build and
 * writes the sitemap to the same `out/` so the Cloudflare Pages deploy
 * picks it up.
 */

// hreflang comes from lib/locales.ts, the same registry that decides which
// routes exist and what lang the document declares. Generating it centrally
// is the only arrangement that cannot drift: Google treats the HTML-tag,
// HTTP-header and sitemap methods as equivalent and says there is no
// benefit to doing more than one, so the sitemap is the single place.
//
// While only English ships, alternatesFor returns an empty set and no
// alternate refs are emitted. A one-locale hreflang set is noise.
//
// Required in a require() rather than an import: next-sitemap loads this
// file with the CommonJS loader. Node strips the types.
const {
  LOCALES,
  alternatesFor,
  basePathOf,
  inServedSection,
  localisePath,
} = require('./lib/locales.ts');

// lastmod used to be `new Date()` for every URL, so all 23 entries carried the
// build time. Google uses lastmod to prioritise recrawls and stops trusting it
// on sites where it is uniform, which means a daily build would re-stamp every
// page and the one genuinely new page, the day's post, would get no priority.
//
// There are exactly two sources of truth here, and hand-maintaining a dates
// file is not one of them:
//
//   posts   for a post, its own dateModified, or date. For a blog listing,
//           the newest post it can show.
//   git     for everything else, the last commit that touched the page file.
//
// The three legal routes are the exception: pages/legal/*.tsx are thin
// wrappers around LegalPageLayout and the actual document lives in the locale
// files, so a terms rewrite never touches the page file. They take the last
// commit on locales/en.json instead.
const { execSync } = require('node:child_process');
const fs = require('node:fs');
const nodePath = require('node:path');


const BUILD_TIME = new Date().toISOString();

/** Run git in apps/landing, returning null rather than throwing. */
function git(args) {
  try {
    const out = execSync(`git ${args}`, {
      cwd: __dirname,
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'ignore'],
    });
    return out.trim() || null;
  } catch {
    return null;
  }
}

/**
 * A shallow clone answers `git log -1 -- <file>` with the tip commit for every
 * file, which would put the build time back on all 23 URLs while every local
 * check still passed. The build host is not this repo's CI (the landing is
 * built from tryrtylr/deploy), so this cannot be asserted here. Say so loudly
 * instead of shipping a sitemap that silently lost its signal.
 */
const SHALLOW = git('rev-parse --is-shallow-repository') === 'true';
if (SHALLOW) {
  console.warn(
    '[next-sitemap] shallow clone: per-file git dates are unavailable, so every ' +
      'lastmod will collapse to the tip commit. Fetch full history to restore it.'
  );
}

const lastCommitIso = (relative) => git(`log -1 --format=%cI -- ${relative}`);

const LOCALE_SOURCE = 'locales/en.json';
const LEGAL_ROUTES = new Set(['/legal/terms/', '/legal/privacy-policy/', '/legal/security/']);

/** pages/pricing/index.tsx -> /pricing/, pages/legal/terms.tsx -> /legal/terms/. */
function routeOf(relative) {
  const withoutExt = relative.replace(/^pages\//, '').replace(/\.tsx$/, '');
  // Dynamic segments are not real routes: the blog is covered by its post files and
  // [locale] is not exported while only English ships.
  if (withoutExt.includes('[') || withoutExt.startsWith('_') || withoutExt === '404') return null;
  const route = withoutExt === 'index' ? '/' : `/${withoutExt.replace(/\/index$/, '')}/`;
  return route;
}

function pageFiles(dir = 'pages') {
  return fs.readdirSync(nodePath.join(__dirname, dir), { withFileTypes: true }).flatMap((e) => {
    const relative = `${dir}/${e.name}`;
    if (e.isDirectory()) return pageFiles(relative);
    return e.name.endsWith('.tsx') ? [relative] : [];
  });
}

/** route -> ISO date of the last commit that touched the page behind it. */
const ROUTE_LASTMOD = new Map();
for (const file of pageFiles()) {
  const route = routeOf(file);
  if (!route) continue;
  const iso = lastCommitIso(LEGAL_ROUTES.has(route) ? LOCALE_SOURCE : file);
  if (iso) ROUTE_LASTMOD.set(route, iso);
}

// One folder per post under content/blog/posts, one JSON file per locale.
const POSTS_DIR = nodePath.join(__dirname, 'content/blog/posts');
const posts = fs
  .readdirSync(POSTS_DIR, { withFileTypes: true })
  .filter((e) => e.isDirectory())
  .flatMap((e) =>
    ['en', 'ar']
      .map((code) => nodePath.join(POSTS_DIR, e.name, `${code}.json`))
      .filter((f) => fs.existsSync(f))
      .map((f) => JSON.parse(fs.readFileSync(f, 'utf8')))
  );

const blogPrefix = (locale) => (locale === 'en' ? '/blog/' : `/${locale}/blog/`);
const isoOf = (post) => new Date(post.dateModified || post.date).toISOString();

/** Post URL -> ISO date, from the post's own file. */
const POST_LASTMOD = new Map(posts.map((p) => [`${blogPrefix(p.locale)}${p.slug}/`, isoOf(p)]));

/** Newest post date per locale, for the listing pages that show them. */
const NEWEST_IN = new Map();
for (const p of posts) {
  const iso = isoOf(p);
  if (!NEWEST_IN.has(p.locale) || NEWEST_IN.get(p.locale) < iso) NEWEST_IN.set(p.locale, iso);
}

/** The build time is the last resort, not the default. */
function lastmodFor(norm) {
  if (POST_LASTMOD.has(norm)) return POST_LASTMOD.get(norm);
  const listing = norm.match(/^\/(?:(ar)\/)?blog\//);
  if (listing) return NEWEST_IN.get(listing[1] ?? 'en') ?? BUILD_TIME;
  return ROUTE_LASTMOD.get(norm) ?? BUILD_TIME;
}

const OUT_DIR = nodePath.join(__dirname, 'dist/landing');
const exported = (route) => fs.existsSync(nodePath.join(OUT_DIR, route, 'index.html'));

/**
 * hreflang for a section a locale serves on its own (the blog). A post may
 * exist in one language only, so the set is read off the export: two
 * locales are alternates for a route exactly when both exported it.
 */
function sectionAlternates(base) {
  const members = LOCALES.filter(
    (l) => (l.shipped || inServedSection(base, l)) && exported(localisePath(base, l.code))
  );
  if (members.length < 2) return [];
  const set = members.map((l) => ({
    hreflang: l.htmlLang,
    href: `https://rtylr.com${localisePath(base, l.code)}`,
  }));
  set.push({ hreflang: 'x-default', href: `https://rtylr.com${base}` });
  return set;
}

/** Whether a URL under an unshipped locale prefix is a page that exists on purpose. */
function servedPath(norm) {
  const code = norm.split('/').filter(Boolean)[0];
  const locale = LOCALES.find((l) => l.code === code);
  if (!locale || locale.shipped) return true;
  return inServedSection(basePathOf(norm), locale);
}

/** @type {import('next-sitemap').IConfig} */
module.exports = {
  siteUrl: 'https://rtylr.com',
  generateRobotsTxt: false, // we maintain robots.txt by hand - it has AI-bot allowances
  outDir: 'dist/landing', // matches next.config.js output directory
  // Unshipped locale prefixes are dropped in transform() below rather than
  // here, because a locale may serve one section (the Arabic blog) while the
  // rest of its routes stay unpublished.
  exclude: ['/404', '/homepage-legacy*'],
  // Trailing slashes match next.config's trailingSlash: true so the
  // URLs in the sitemap match the served URLs 1:1.
  trailingSlash: true,
  changefreq: 'weekly',
  priority: 0.7,
  // Per-path priority overrides - homepage, product pages, and
  // pricing earn higher priority than legal or newsroom.
  transform: async (config, path) => {
    // Normalise both sides so "/pos" and "/pos/" hit the same bucket.
    // next-sitemap's `path` argument isn't guaranteed to include the
    // trailing slash regardless of the global trailingSlash option.
    const norm = path === '/' ? '/' : path.replace(/\/$/, '') + '/';
    // An unshipped locale's route that leaked into the export must not
    // also be advertised.
    if (!servedPath(norm)) return null;

    const high = ['/workspace/', '/pricing/'];
    const medium = ['/demo/', '/contact/', '/hardware/', '/integrations/', '/support/'];
    const low = ['/legal/terms/', '/legal/privacy-policy/', '/legal/security/'];

    let priority = 0.7;
    let changefreq = 'weekly';
    if (norm === '/') {
      priority = 1.0;
    } else if (high.includes(norm)) {
      priority = 0.9;
    } else if (medium.includes(norm)) {
      priority = 0.7;
      changefreq = 'monthly';
    } else if (low.includes(norm)) {
      priority = 0.3;
      changefreq = 'yearly';
    }

    // basePathOf() first: this transform runs for every URL in the export,
    // including /ar/pricing/ and /fr/pricing/ themselves, and alternatesFor
    // matches against the bare route. Without it, only the English URLs
    // matched LOCALISED_ROUTES and got an alternate set - the /ar/* and
    // /fr/* pages each got an empty one, which is non-reciprocal hreflang
    // and Google discards the whole set silently.
    //
    // hrefIsAbsolute is required, not optional. Without it next-sitemap
    // treats href as a BASE and appends the page path to it, so a full URL
    // like https://rtylr.com/pricing/ comes out as
    // https://rtylr.com/pricing/pricing/. Every alternate would 404 and
    // Google would drop the whole set silently. The default is designed for
    // multi-domain setups (fr.example.com) where href really is a base.
    const base = basePathOf(norm);
    const alternates = base.startsWith('/blog/') ? sectionAlternates(base) : alternatesFor(base);
    const alternateRefs = alternates.map(({ hreflang, href }) => ({
      href,
      hreflang,
      hrefIsAbsolute: true,
    }));

    return {
      loc: path,
      changefreq,
      priority,
      lastmod: lastmodFor(norm),
      ...(alternateRefs.length ? { alternateRefs } : {}),
    };
  },
};
