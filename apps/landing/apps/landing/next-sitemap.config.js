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

/** @type {import('next-sitemap').IConfig} */
module.exports = {
  siteUrl: 'https://rtylr.com',
  generateRobotsTxt: false, // we maintain robots.txt by hand - it has AI-bot allowances
  outDir: '../../dist/landing', // matches next.config.js output directory
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

    return {
      loc: path,
      changefreq,
      priority,
      lastmod: new Date().toISOString(),
    };
  },
};
