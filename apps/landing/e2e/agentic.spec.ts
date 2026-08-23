import { test, expect } from './fixtures';
import { readFileSync, readdirSync } from 'node:fs';
import { join } from 'node:path';

/**
 * Agentic-readiness artifacts. The site is a static export served by a
 * Cloudflare Pages advanced-mode worker (cloudflare/_worker.js). `serve`
 * cannot execute the worker, so these specs assert the BUILD ARTIFACTS the
 * worker depends on exist and are well-formed; runtime negotiation behavior
 * is covered by tests/unit/worker.test.mjs against an in-memory ASSETS mock.
 */

const LANDING_URL = process.env.LANDING_URL ?? 'http://127.0.0.1:4200';

// Playwright runs from the repo root; resolve dist from there (no __dirname:
// spec files are transpiled to ESM where it doesn't exist).
const DIST_DIR = join(process.cwd(), 'dist', 'landing');

function walkHtml(dir: string, out: string[] = []): string[] {
  for (const entry of readdirSync(dir, { withFileTypes: true })) {
    const full = join(dir, entry.name);
    if (entry.isDirectory()) {
      if (entry.name === '_next') continue;
      walkHtml(full, out);
    } else if (entry.isFile() && entry.name.endsWith('.html')) {
      out.push(full);
    }
  }
  return out;
}

test.describe('Agentic artifacts', () => {
  test('openapi.json is served and describes the site API', async ({ request }) => {
    const res = await request.get(`${LANDING_URL}/openapi.json`);
    expect(res.status()).toBe(200);
    const spec = await res.json();
    expect(spec.openapi).toMatch(/^3\.\d+\.\d+$/);
    expect(Object.keys(spec.paths)).toContain('/api/v1/status');
    expect(Object.keys(spec.paths)).toContain('/api/v1/pages');
  });

  test('llms.txt is served with page index + machine surfaces', async ({ request }) => {
    const res = await request.get(`${LANDING_URL}/llms.txt`);
    expect(res.status()).toBe(200);
    const body = await res.text();
    expect(body).toContain('# rtylr');
    expect(body).toContain('## Pages');
    expect(body).toContain('/openapi.json');
    expect(body).toContain('Accept: text/markdown');
  });

  test('llms-full.txt serves every page body in one request', async ({ request }) => {
    // llms.txt is an index of links: an agent still has to fetch 23 URLs to
    // read the site. This is the whole thing in one response.
    const res = await request.get(`${LANDING_URL}/llms-full.txt`);
    expect(res.status()).toBe(200);
    const body = await res.text();

    // One "Source:" line per page, and it must agree with pages.json
    // rather than with a number written here.
    const indexed = (await (await request.get(`${LANDING_URL}/pages.json`)).json()) as {
      count: number;
    };
    const sources = body.match(/^Source: https:\/\/rtylr\.com/gm) ?? [];
    expect(sources.length).toBe(indexed.count);

    // And it must carry real page text, not just headings. If the body
    // conversion ever returns empty, the file would still look plausible.
    expect(body).toContain('rtylr POS is the point-of-sale app inside rtylr');
    expect(body).toContain('keeps a full local copy of the menu');
    expect(body.length).toBeGreaterThan(50_000);
  });

  test('every HTML page has a markdown sibling next to it', () => {
    const htmlFiles = walkHtml(DIST_DIR).filter(
      // The generator deliberately skips the 404 page(s): the edge worker
      // renders its own markdown 404 body.
      (f) => !/(^|[/\\])404([/\\]|\.html$)/.test(f.slice(DIST_DIR.length))
    );
    expect(htmlFiles.length).toBeGreaterThan(10);
    for (const file of htmlFiles) {
      const sibling = join(file.replace(/[^/\\]+$/, ''), 'index.md');
      const md = readFileSync(sibling, 'utf8');
      expect(md.length, sibling).toBeGreaterThan(50);
      expect(md, sibling).not.toMatch(/<script|<html/i);
    }
  });

  test('markdown siblings contain real content (homepage sample)', async ({ request }) => {
    const res = await request.get(`${LANDING_URL}/index.md`);
    expect(res.status()).toBe(200);
    const md = await res.text();
    expect(md).toMatch(/^#\s/m);
    // No raw HTML may leak into the markdown representation.
    expect(md).not.toMatch(/<(?:div|span|section|main|nav|footer)[\s>]/i);
  });

  test('pages.json index matches the sitemap URL set', async ({ request }) => {
    const pages = JSON.parse(await (await request.get(`${LANDING_URL}/pages.json`)).text());
    expect(pages.count).toBe(pages.pages.length);
    expect(pages.count).toBeGreaterThan(10);
    const urls: string[] = pages.pages.map((page: { url: string }) => page.url);
    expect(urls).toContain('https://rtylr.com/');
    for (const url of urls) {
      expect(url.startsWith('https://rtylr.com/')).toBe(true);
    }
  });

  test('_worker.js ships in the deploy output', () => {
    const worker = readFileSync(join(DIST_DIR, '_worker.js'), 'utf8');
    expect(worker).toContain('text/markdown');
    expect(worker).toContain('Vary');
    expect(worker).toContain('/api/v1/status');
  });
});
