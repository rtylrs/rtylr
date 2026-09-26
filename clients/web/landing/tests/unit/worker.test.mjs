import { describe, it, beforeEach } from 'node:test';
import assert from 'node:assert/strict';
import worker from '../../cloudflare/_worker.js';
import { htmlCandidates, markdownCandidates, markdownNotFound, jsonError } from '../../cloudflare/_worker.js';

// ---------------------------------------------------------------------------
// In-memory ASSETS binding that mimics Cloudflare Pages asset serving.
// ---------------------------------------------------------------------------

const INDEX_HTML = `<!DOCTYPE html><html><head><title>Home · rtylr</title></head>
<body><nav>ignore me</nav><main><h1>Retail OS</h1><p>One platform.</p></main></body></html>`;
const INDEX_MD = '# Retail OS\n\nOne platform.\n';
const PRICING_HTML = '<html><head></head><body><main><h1>Pricing</h1></main></body></html>';
const PRICING_MD = '# Pricing\n';
const NOT_FOUND_HTML = '<html><head></head><body class="nf">404 styled</body></html>';
const PAGES_JSON = JSON.stringify({
  generatedAt: '2026-08-22T00:00:00.000Z',
  count: 2,
  pages: [{ url: 'https://rtylr.com/', title: 'Retail OS' }],
});
const VERSION_JSON = '{"version":"1.2.3"}';

function contentTypeFor(path) {
  if (path.endsWith('.html')) return 'text/html; charset=utf-8';
  if (path.endsWith('.md')) return 'text/markdown; charset=utf-8';
  if (path.endsWith('.json')) return 'application/json; charset=utf-8';
  return 'text/plain; charset=utf-8';
}

/**
 * Files are keyed by their real on-disk asset path (e.g. '/pricing/index.html').
 * Replicates Cloudflare Pages' actual `html_handling: auto-trailing-slash`
 * asset-server behavior: a literal `.html` request (including `.../index.html`)
 * 308-redirects to its clean URL instead of serving the file — this applies
 * even to requests made internally via env.ASSETS.fetch(), which is exactly
 * what broke production (see apps/landing/cloudflare/_worker.js htmlCandidates).
 * Anything unresolvable 404s like the real bucket.
 */
function makeEnv(files) {
  return {
    ASSETS: {
      async fetch(request) {
        const path = new URL(request.url).pathname;

        if (/\.html?$/i.test(path)) {
          if (!files.has(path)) return new Response('not found', { status: 404 });
          const clean = /\/index\.html$/i.test(path)
            ? path.slice(0, -'index.html'.length)
            : path.replace(/\.html?$/i, '');
          return new Response(null, { status: 308, headers: { Location: clean } });
        }

        if (path.endsWith('/')) {
          const key = `${path}index.html`;
          if (!files.has(key)) return new Response('not found', { status: 404 });
          return new Response(files.get(key), { status: 200, headers: { 'Content-Type': 'text/html; charset=utf-8' } });
        }

        if (!/\.[a-z0-9]+$/i.test(path) && files.has(`${path}/index.html`)) {
          return new Response(null, { status: 308, headers: { Location: `${path}/` } });
        }

        if (!files.has(path)) return new Response('not found', { status: 404 });
        return new Response(files.get(path), { status: 200, headers: { 'Content-Type': contentTypeFor(path) } });
      },
    },
  };
}

const BASE_ENV = new Map([
  ['/index.html', INDEX_HTML],
  ['/index.md', INDEX_MD],
  ['/pricing/index.html', PRICING_HTML],
  ['/pricing/index.md', PRICING_MD],
  ['/404.html', NOT_FOUND_HTML],
  ['/404/index.html', NOT_FOUND_HTML],
  ['/pages.json', PAGES_JSON],
  ['/version.json', VERSION_JSON],
]);

const SITE = 'https://rtylr.com';

function req(path, opts = {}) {
  const { accept, method, ...restHeaders } = opts;
  const headers = { ...restHeaders };
  if (accept) headers.accept = accept;
  return new Request(SITE + path, { method, headers });
}

let env;
beforeEach(() => {
  env = makeEnv(new Map(BASE_ENV));
});

// ---------------------------------------------------------------------------
// Content negotiation
// ---------------------------------------------------------------------------

describe('markdown content negotiation', () => {
  it('serves the .md sibling with text/markdown and Vary: Accept, Accept-Encoding', async () => {
    const res = await worker.fetch(req('/', { accept: 'text/markdown' }), env);
    assert.equal(res.status, 200);
    assert.match(res.headers.get('content-type'), /^text\/markdown/);
    const vary = res.headers.get('vary').toLowerCase();
    assert.ok(vary.includes('accept'), 'Vary must include Accept');
    assert.ok(vary.includes('accept-encoding'), 'Vary must include Accept-Encoding');
    assert.equal(await res.text(), INDEX_MD);
  });

  it('negotiates per-path (/pricing/)', async () => {
    const res = await worker.fetch(req('/pricing/', { accept: 'text/markdown' }), env);
    assert.equal(res.status, 200);
    assert.equal(await res.text(), PRICING_MD);
  });

  it('serves HTML by default (no header, */*, browser headers)', async () => {
    for (const accept of [null, '*/*', 'text/html,application/xhtml+xml,*/*;q=0.8']) {
      const res = await worker.fetch(req('/', accept ? { accept } : {}), env);
      assert.equal(res.status, 200);
      assert.match(res.headers.get('content-type'), /text\/html/);
      assert.ok((await res.text()).includes('<main>'), accept ?? '(no accept)');
    }
  });

  it('q=0 on markdown falls back to HTML instead of serving markdown', async () => {
    const res = await worker.fetch(req('/', { accept: 'text/markdown;q=0' }), env);
    assert.match(res.headers.get('content-type'), /text\/html/);
  });

  it('advertises the markdown alternate via Link on HTML responses', async () => {
    const res = await worker.fetch(req('/pricing/'), env);
    const link = res.headers.get('link');
    assert.ok(link?.includes('/pricing/index.md'), link ?? 'missing Link header');
    assert.ok(link.includes('rel="alternate"') && link.includes('type="text/markdown"'));
  });

  it('adds Vary to HTML responses so caches never mix variants', async () => {
    const res = await worker.fetch(req('/'), env);
    const vary = res.headers.get('vary').toLowerCase();
    assert.ok(vary.includes('accept') && vary.includes('accept-encoding'));
  });

  it('returns 406 when the client rejects every producible type', async () => {
    const res = await worker.fetch(req('/', { accept: 'application/pdf' }), env);
    assert.equal(res.status, 406);
    const vary = res.headers.get('vary').toLowerCase();
    assert.ok(vary.includes('accept'));
  });

  it('serves a markdown-flavored 404 when the sibling is missing and the page does not exist', async () => {
    const res = await worker.fetch(
      req('/no-sibling-page/', { accept: 'text/markdown, text/html;q=0.1' }),
      env
    );
    assert.equal(res.status, 404);
    assert.match(res.headers.get('content-type'), /^text\/markdown/);
    assert.ok((await res.text()).startsWith('# 404'));
  });
});

// ---------------------------------------------------------------------------
// Agent-friendly 404s
// ---------------------------------------------------------------------------

describe('agent-friendly 404s', () => {
  it('nonexistent paths return a REAL 404 status', async () => {
    const res = await worker.fetch(req('/some-path-that-does-not-exist/'), env);
    assert.equal(res.status, 404);
  });

  it('agents asking for markdown get a markdown recovery body on 404', async () => {
    const res = await worker.fetch(
      req('/some-path-that-does-not-exist/', { accept: 'text/markdown' }),
      env
    );
    assert.equal(res.status, 404);
    assert.match(res.headers.get('content-type'), /^text\/markdown/);
    const body = await res.text();
    assert.ok(body.startsWith('# 404'));
    assert.ok(body.includes('/sitemap.xml'), 'points at sitemap');
    assert.ok(body.includes('/openapi.json'), 'points at OpenAPI spec');
    assert.ok(body.includes('/llms.txt'), 'points at llms.txt');
    assert.ok(body.includes('[Pricing](/pricing/)'));
    const vary = res.headers.get('vary').toLowerCase();
    assert.ok(vary.includes('accept'));
  });

  it('browsers get the styled HTML 404 page with status 404', async () => {
    const res = await worker.fetch(req('/does-not-exist/'), env);
    assert.equal(res.status, 404);
    assert.match(res.headers.get('content-type'), /text\/html/);
    assert.ok((await res.text()).includes('404 styled'));
  });

  it('markdownNotFound includes the requested path', () => {
    const body = markdownNotFound('/gone-page/');
    assert.ok(body.includes('/gone-page/'));
  });
});

// ---------------------------------------------------------------------------
// JSON API + structured JSON errors
// ---------------------------------------------------------------------------

describe('JSON API', () => {
  it('GET /api/v1/status returns live JSON with endpoint discovery', async () => {
    const res = await worker.fetch(req('/api/v1/status'), env);
    assert.equal(res.status, 200);
    assert.match(res.headers.get('content-type'), /application\/json/);
    const data = await res.json();
    assert.equal(data.service, 'rtylr.com');
    assert.equal(data.status, 'ok');
    assert.equal(data.endpoints.openapi, '/openapi.json');
  });

  it('GET /api/v1/pages wraps the generated page index as JSON', async () => {
    const res = await worker.fetch(req('/api/v1/pages'), env);
    assert.equal(res.status, 200);
    const data = await res.json();
    assert.equal(data.count, 1);
    assert.deepEqual(data.pages[0].url, 'https://rtylr.com/');
  });

  it('unknown API paths return structured JSON 404 with code, message and hint', async () => {
    const res = await worker.fetch(req('/api/nonexistent'), env);
    assert.equal(res.status, 404);
    assert.match(res.headers.get('content-type'), /application\/json/);
    const { error } = await res.json();
    assert.equal(error.status, 404);
    assert.equal(error.code, 'not_found');
    assert.match(error.message, /Unknown API endpoint/);
    assert.match(error.hint, /\/openapi\.json/);
    assert.ok(error.docs.endsWith('/openapi.json'));
  });

  it('unknown API paths return JSON 404 regardless of verb', async () => {
    const res = await worker.fetch(req('/api/nope/', { method: 'POST' }), env);
    assert.equal(res.status, 404);
    const { error } = await res.json();
    assert.equal(error.code, 'not_found');
  });

  it('wrong verbs on known endpoints return JSON 405 plus Allow header', async () => {
    const res = await worker.fetch(req('/api/v1/status', { method: 'POST' }), env);
    assert.equal(res.status, 405);
    assert.equal(res.headers.get('allow'), 'GET, HEAD, OPTIONS');
    const { error } = await res.json();
    assert.equal(error.code, 'method_not_allowed');
    assert.match(error.hint, /GET/);
  });

  it('OPTIONS returns 204 with Allow', async () => {
    const res = await worker.fetch(req('/api/v1/pages', { method: 'OPTIONS' }), env);
    assert.equal(res.status, 204);
    assert.equal(res.headers.get('allow'), 'GET, HEAD, OPTIONS');
  });

  it('advertises RateLimit-* headers on success and error API responses', async () => {
    const ok = await worker.fetch(req('/api/v1/status'), env);
    assert.ok(ok.headers.get('ratelimit-limit'));
    assert.ok(ok.headers.get('ratelimit-remaining'));
    assert.ok(ok.headers.get('ratelimit-reset'));

    const notFound = await worker.fetch(req('/api/nope'), env);
    assert.ok(notFound.headers.get('ratelimit-limit'));
  });

  it('API errors are JSON even when the client asks for markdown', async () => {
    const res = await worker.fetch(req('/api/unknown-endpoint/', { accept: 'text/markdown' }), env);
    assert.equal(res.status, 404);
    assert.match(res.headers.get('content-type'), /application\/json/);
  });

  it('jsonError envelope shape is stable', async () => {
    const res = jsonError(418, 'teapot', 'short and stout', 'brew coffee instead');
    assert.equal(res.status, 418);
    const { error } = await res.json();
    assert.deepEqual(Object.keys(error).sort(), ['code', 'docs', 'hint', 'message', 'status']);
  });
});

// ---------------------------------------------------------------------------
// Routing parity & passthrough
// ---------------------------------------------------------------------------

describe('routing', () => {
  it('/version is served from /version.json (legacy _redirects parity)', async () => {
    const res = await worker.fetch(req('/version'), env);
    assert.equal(res.status, 200);
    assert.equal(await res.text(), VERSION_JSON);
  });

  it('static assets pass through untouched, without negotiation Vary', async () => {
    env = makeEnv(new Map([['/robots.txt', 'User-agent: *\nAllow: /\n']]));
    const res = await worker.fetch(req('/robots.txt'), env);
    assert.equal(res.status, 200);
    assert.ok((await res.text()).startsWith('User-agent'));
  });

  // This used to assert a 200 with the page body. Serving both /pricing and
  // /pricing/ is what #1359 removes: two URLs, one document, and a crawl
  // budget spent proving they are the same. The directory export is still
  // what answers, one redirect later.
  it('extensionless URLs redirect to the directory-style export (/pricing)', async () => {
    const res = await worker.fetch(req('/pricing'), env);
    assert.equal(res.status, 301);
    assert.equal(res.headers.get('location'), `${SITE}/pricing/`);

    const followed = await worker.fetch(req('/pricing/'), env);
    assert.equal(followed.status, 200);
    assert.ok((await followed.text()).includes('<main>'));
  });

  it('the trailing-slash redirect keeps the query string', async () => {
    const res = await worker.fetch(req('/trust?utm_source=newsletter&a=b'), env);
    assert.equal(res.status, 301);
    assert.equal(res.headers.get('location'), `${SITE}/trust/?utm_source=newsletter&a=b`);
  });

  it('the API is never trailing-slash redirected', async () => {
    const res = await worker.fetch(req('/api/v1/status'), env);
    assert.equal(res.status, 200);
    assert.equal(res.headers.get('content-type'), 'application/json; charset=utf-8');
  });

  it('/version is not trailing-slash redirected', async () => {
    const res = await worker.fetch(req('/version'), env);
    assert.equal(res.status, 200);
    assert.equal(await res.text(), VERSION_JSON);
  });

  it('an unknown extension is left alone rather than redirected', async () => {
    const res = await worker.fetch(req('/whatever.bar'), env);
    assert.notEqual(res.status, 301);
  });

  it('explicit /index.html still serves the page', async () => {
    const res = await worker.fetch(req('/index.html'), env);
    assert.equal(res.status, 200);
  });

  it('candidate helpers match the trailingSlash export layout', () => {
    // Must be the clean directory URL, never a literal '.../index.html' or
    // '<path>.html' — Cloudflare Pages 308-redirects those even for
    // env.ASSETS.fetch() calls made from inside this worker, which always
    // looks like a miss to isHtmlResponse() (expects status 200).
    assert.deepEqual(htmlCandidates('/'), ['/']);
    assert.deepEqual(htmlCandidates('/pricing/'), ['/pricing/']);
    assert.deepEqual(htmlCandidates('/pricing'), ['/pricing/']);
    assert.deepEqual(htmlCandidates('/pricing.html'), ['/pricing/']);
    assert.deepEqual(markdownCandidates('/'), ['/index.md']);
    assert.deepEqual(markdownCandidates('/pricing/'), ['/pricing/index.md', '/pricing.md']);
  });

  it('API paths with extensions still get JSON errors, not asset passthrough', async () => {
    const res = await worker.fetch(req('/api/v1/status.json', { accept: 'text/markdown' }), env);
    assert.equal(res.status, 404);
    assert.match(res.headers.get('content-type'), /application\/json/);
    const { error } = await res.json();
    assert.equal(error.code, 'not_found');
  });
});

// ---------------------------------------------------------------------------
// Security headers
// ---------------------------------------------------------------------------

describe('security headers', () => {
  const REQUIRED = {
    'strict-transport-security': 'max-age=31536000; includeSubDomains',
    'x-content-type-options': 'nosniff',
    'referrer-policy': 'strict-origin-when-cross-origin',
    'permissions-policy': 'camera=(), microphone=(), geolocation=()',
  };

  const assertSecured = (res, what) => {
    for (const [name, value] of Object.entries(REQUIRED)) {
      assert.equal(res.headers.get(name), value, `${what} is missing ${name}`);
    }
    assert.ok(res.headers.get('content-security-policy'), `${what} is missing the CSP`);
    assert.equal(
      res.headers.get('content-security-policy-report-only'),
      null,
      `${what} still carries the report-only header`
    );
  };

  it('the HTML response carries them', async () => {
    const res = await worker.fetch(req('/pricing/'), env);
    assert.equal(res.status, 200);
    assertSecured(res, 'the HTML response');
  });

  it('the markdown response carries them', async () => {
    const res = await worker.fetch(req('/pricing/', { accept: 'text/markdown' }), env);
    assert.equal(res.status, 200);
    assert.match(res.headers.get('content-type'), /text\/markdown/);
    assertSecured(res, 'the markdown response');
  });

  it('the styled 404 carries them', async () => {
    const res = await worker.fetch(req('/nope/'), env);
    assert.equal(res.status, 404);
    assertSecured(res, 'the styled 404');
  });

  it('the markdown 404 carries them', async () => {
    const res = await worker.fetch(req('/nope/', { accept: 'text/markdown' }), env);
    assert.equal(res.status, 404);
    assertSecured(res, 'the markdown 404');
  });

  it('the 406 carries them', async () => {
    const res = await worker.fetch(req('/pricing/', { accept: 'text/html;q=0, text/markdown;q=0' }), env);
    assert.equal(res.status, 406);
    assertSecured(res, 'the 406');
  });

  // Assets are served straight off the bucket by env.ASSETS. Asserting the
  // headers are absent keeps the boundary honest: if that ever changes, this
  // test says so rather than the coverage quietly shifting.
  it('static assets are left untouched', async () => {
    env = makeEnv(new Map([['/robots.txt', 'User-agent: *\n']]));
    const res = await worker.fetch(req('/robots.txt'), env);
    assert.equal(res.status, 200);
    assert.equal(res.headers.get('strict-transport-security'), null);
  });

  it('HSTS does not claim preload', async () => {
    const res = await worker.fetch(req('/pricing/'), env);
    // Preload commits every *.rtylr.com subdomain to HTTPS-only, and the app
    // subdomains are not in this repo.
    assert.ok(!/preload/i.test(res.headers.get('strict-transport-security')));
  });

  it('the CSP is enforcing, not report-only', async () => {
    const res = await worker.fetch(req('/pricing/'), env);
    assert.equal(res.headers.get('content-security-policy-report-only'), null);
    const policy = res.headers.get('content-security-policy');
    assert.match(policy, /default-src 'self'/);
    assert.match(policy, /https:\/\/www\.googletagmanager\.com/);
    // The variant lookup in lib/featureVariant.ts calls api.voxire.com, so the
    // policy has to allow it or every visitor is pinned to the default.
    assert.match(policy, /connect-src[^;]*https:\/\/api\.voxire\.com/);
  });
});
