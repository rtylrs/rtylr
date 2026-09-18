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
const { alternatesFor, basePathOf, shippedLocales } = require('./lib/locales.ts');

/** @type {import('next-sitemap').IConfig} */
module.exports = {
  siteUrl: 'https://rtylr.com',
  generateRobotsTxt: false, // we maintain robots.txt by hand - it has AI-bot allowances
  outDir: 'dist/landing', // matches next.config.js output directory
  // Unshipped locale prefixes can never appear here, because no route is
  // exported for them. The exclusion is belt and braces: if one ever leaks
  // into the export, it must not also be advertised.
  exclude: [
    '/404',
    '/homepage-legacy*',
    ...['ar', 'fr']
      .filter((code) => !shippedLocales().some((l) => l.code === code))
      .flatMap((code) => [`/${code}`, `/${code}/*`]),
  ],
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

    const high = ['/pos/', '/erp/', '/hr/', '/crm/', '/finance/', '/flow/', '/pricing/'];
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
    const alternateRefs = alternatesFor(basePathOf(norm)).map(({ hreflang, href }) => ({
      href,
      hreflang,
      hrefIsAbsolute: true,
    }));

    return {
      loc: path,
      changefreq,
      priority,
      lastmod: new Date().toISOString(),
      ...(alternateRefs.length ? { alternateRefs } : {}),
    };
  },
};
