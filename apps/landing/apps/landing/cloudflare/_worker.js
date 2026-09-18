/**
 * rtylr.com edge worker (Cloudflare Pages advanced mode).
 *
 * Deployed automatically as `_worker.js` at the root of the static export
 * (copied there by scripts/generate-agent-artifacts.mjs after `next build`),
 * so this module owns ALL routing for the site:
 *
 *   1. Static assets          -> env.ASSETS, untouched.
 *   2. Documents (/...)       -> RFC 9110 content negotiation between the
 *                                prebuilt text/html page and its text/markdown
 *                                sibling (acceptmarkdown.com compliant):
 *                                q-values, specificity overrides, q=0
 *                                rejections, 406, `Vary: Accept,
 *                                Accept-Encoding`, `Link: rel="alternate"`.
 *   3. Missing pages          -> real HTTP 404. Browsers get the styled
 *                                404.html; agents asking for markdown get a
 *                                short text/markdown body with recovery links.
 *   4. /api/v1/*              -> small read-only JSON API. Errors are always
 *                                structured JSON envelopes (code, message,
 *                                hint, docs) — never HTML.
 *   5. /version               -> served from /version.json (legacy parity).
 *
 * The A/B variant router (cloudflare/variant-router.worker.js) sits in front
 * of this origin and proxies unchanged — negotiation happens here, per
 * variant, because origins differ.
 */

const SITE_URL = 'https://rtylr.com';

/** Media types this origin can produce for document URLs. Order = server preference. */
const DOCUMENT_TYPES = ['text/html', 'text/markdown'];

/** Requests for these extensions bypass negotiation entirely. */
const STATIC_EXT =
  /\.(?:css|m?js|map|png|jpe?g|gif|webp|avif|svg|ico|bmp|woff2?|ttf|otf|eot|xml|txt|json|webmanifest|pdf|mp4|webm|mp3|wav|ogg|zip|wasm|csv)$/i;

// ---------------------------------------------------------------------------
// Pure helpers (exported for unit tests)
// ---------------------------------------------------------------------------

/**
 * Parse an `Accept` header into ranked entries.
 * Returns [] when the header is absent or empty (= "no constraint").
 */
export function parseAccept(header) {
  if (!header) return [];
  const entries = [];
  for (const raw of header.split(',')) {
    const parts = raw.trim().split(';');
    const type = (parts[0] || '').trim().toLowerCase();
    if (!type || !type.includes('/')) continue;
    let q = 1;
    for (const param of parts.slice(1)) {
      const [name, value] = param.split('=');
      if ((name || '').trim().toLowerCase() !== 'q') continue;
      const parsed = Number.parseFloat((value || '').trim());
      if (!Number.isNaN(parsed)) q = Math.min(1, Math.max(0, parsed));
    }
    const specificity = type === '*/*' ? 0 : type.endsWith('/*') ? 1 : 2;
    entries.push({ type, q, specificity });
  }
  return entries;
}

function rangeMatches(range, candidate) {
  if (range === '*/*') return true;
  if (range.endsWith('/*')) return candidate.startsWith(range.slice(0, -1));
  return range === candidate;
}

/**
 * Pick which of `producible` media types best satisfies an Accept header,
 * following RFC 9110 section 12.5.1 and the acceptmarkdown.com parsing rules:
 * highest q wins, more specific ranges override less specific ones (so a
 * rejected text/html is not resurrected by a wildcard allow), ties break
 * toward the client's own ordering, and q=0 means "never send me this".
 *
 * Returns the winning type, or null when nothing acceptable remains (-> 406).
 */
export function negotiate(acceptHeader, producible) {
  const unconstrained = !acceptHeader || parseAccept(acceptHeader).length === 0;
  if (unconstrained) return producible[0] ?? null;

  const entries = parseAccept(acceptHeader);
  let best = null;
  let bestQ = 0;
  let bestPos = Number.POSITIVE_INFINITY;

  // Most specific matching range for a candidate; returns the entry and its
  // position (position breaks ties toward client ordering).
  const matchFor = (candidate) => {
    let match = null;
    let matchPos = Number.POSITIVE_INFINITY;
    for (let i = 0; i < entries.length; i++) {
      const e = entries[i];
      if (!rangeMatches(e.type, candidate)) continue;
      if (
        match === null ||
        e.specificity > match.specificity ||
        (e.specificity === match.specificity && i < matchPos)
      ) {
        match = e;
        matchPos = i;
      }
    }
    return { match, matchPos };
  };

  for (const candidate of producible) {
    const { match, matchPos } = matchFor(candidate);
    if (match === null || match.q <= 0) continue; // unmatched or explicitly rejected
    if (match.q > bestQ || (match.q === bestQ && matchPos < bestPos)) {
      best = candidate;
      bestQ = match.q;
      bestPos = matchPos;
    }
  }
  if (best) return best;

  // Nothing was positively selected. Per the acceptmarkdown.com guidance,
  // rejecting ONE type (q=0) does not reject the rest: fall back to the most
  // preferred producible type the client did not explicitly reject. Headers
  // that never mention anything we produce (e.g. image/png only) stay
  // unacceptable -> caller sends 406.
  const explicitlyRejected = entries.some((e) => e.q <= 0 && producible.some((c) => rangeMatches(e.type, c)));
  if (!explicitlyRejected) return null;

  for (const candidate of producible) {
    const { match } = matchFor(candidate);
    if (match === null || match.q > 0) return candidate;
  }
  return null;
}

/**
 * Asset path that may hold the HTML representation of a document URL.
 * Matches the `trailingSlash: true` export layout (every page is `<dir>/index.html`).
 *
 * IMPORTANT: this must be the clean, trailing-slash directory URL — never a
 * literal `.../index.html` or `<path>.html` path. Cloudflare Pages' asset
 * server 308-redirects any literal `.html` request to its clean URL
 * (html_handling: auto-trailing-slash), including requests made internally
 * via env.ASSETS.fetch() from this worker. Probing a literal `.html` path
 * therefore never returns 200 — it always looks like a miss.
 */
export function htmlCandidates(pathname) {
  const path = normalizePath(pathname);
  // "index.html" strips to its parent directory (already trailing-slashed);
  // any other ".html" basename strips to a same-named directory instead.
  if (/(^|\/)index\.html?$/i.test(path)) {
    return [path.replace(/index\.html?$/i, '')];
  }
  if (/\.html?$/i.test(path)) {
    return [`${path.replace(/\.html?$/i, '')}/`];
  }
  return [path.endsWith('/') ? path : `${path}/`];
}

/** Asset paths that may hold the Markdown representation of a document URL. */
export function markdownCandidates(pathname) {
  const path = normalizePath(pathname);
  if (path === '/') return ['/index.md'];
  const clean = path.replace(/\/+$/, '');
  const candidates = [];
  if (clean.endsWith('.md')) candidates.push(clean);
  candidates.push(`${clean}/index.md`, `${clean}.md`);
  return [...new Set(candidates)];
}

function normalizePath(pathname) {
  const decoded = safeDecode(pathname);
  if (!decoded.startsWith('/')) return `/${decoded}`;
  return decoded;
}

function safeDecode(value) {
  try {
    return decodeURIComponent(value);
  } catch {
    return value;
  }
}

export function addVary(headers, token) {
  const existing = headers.get('vary');
  if (!existing) {
    headers.set('Vary', token);
    return;
  }
  const tokens = existing.split(',').map((t) => t.trim());
  const lower = tokens.map((t) => t.toLowerCase());
  if (!lower.includes(token.toLowerCase())) tokens.push(token);
  headers.set('Vary', tokens.join(', '));
}

/** Every response whose bytes depend on the Accept header varies on both. */
function varyOnNegotiation(headers) {
  addVary(headers, 'Accept');
  addVary(headers, 'Accept-Encoding');
}

// ---------------------------------------------------------------------------
// Error envelope (structured JSON — agents parse this, not HTML)
// ---------------------------------------------------------------------------

// Advertised, not (yet) individually metered per client — this is a
// read-only, cacheable JSON surface with no per-caller state in the worker.
// The headers still let well-behaved agents self-throttle to the documented
// ceiling instead of guessing; real abuse is stopped at Cloudflare's edge.
const RATE_LIMIT_WINDOW_SECONDS = 60;
const RATE_LIMIT_MAX_REQUESTS = 60;

/** RFC-draft `RateLimit-*` headers (draft-ietf-httpapi-ratelimit-headers). */
export function rateLimitHeaders() {
  const nowSeconds = Math.floor(Date.now() / 1000);
  const secondsIntoWindow = nowSeconds % RATE_LIMIT_WINDOW_SECONDS;
  return {
    'RateLimit-Limit': String(RATE_LIMIT_MAX_REQUESTS),
    'RateLimit-Remaining': String(RATE_LIMIT_MAX_REQUESTS),
    'RateLimit-Reset': String(RATE_LIMIT_WINDOW_SECONDS - secondsIntoWindow),
    'RateLimit-Policy': `${RATE_LIMIT_MAX_REQUESTS};w=${RATE_LIMIT_WINDOW_SECONDS}`,
  };
}

export function jsonError(status, code, message, hint) {
  const body = {
    error: {
      status,
      code,
      message,
      hint,
      docs: `${SITE_URL}/openapi.json`,
    },
  };
  return new Response(JSON.stringify(body, null, 2) + '\n', {
    status,
    headers: {
      'Content-Type': 'application/json; charset=utf-8',
      'Cache-Control': 'no-store',
      ...rateLimitHeaders(),
    },
  });
}

// ---------------------------------------------------------------------------
// Agent-facing 404 bodies
// ---------------------------------------------------------------------------

const NOT_FOUND_LINKS = [
  ['/', 'Home'],
  ['/pricing/', 'Pricing'],
  ['/pos/', 'POS'],
  ['/erp/', 'ERP'],
  ['/crm/', 'CRM'],
  ['/hr/', 'HR'],
  ['/finance/', 'Finance'],
  ['/flow/', 'Flow'],
  ['/demo/', 'Book a demo'],
  ['/contact/', 'Contact'],
];

export function markdownNotFound(pathname) {
  const links = NOT_FOUND_LINKS.map(([href, label]) => `- [${label}](${href})`).join('\n');
  return [
    `# 404 — Page not found`,
    '',
    `\`${pathname}\` does not exist on ${SITE_URL}.`,
    '',
    '## Where to look next',
    '',
    links,
    '',
    '- [Full page index (sitemap.xml)](/sitemap.xml)',
    `- [Agent index (llms.txt)](/llms.txt)`,
    `- [Site API specification (OpenAPI)](/openapi.json)`,
    '',
  ].join('\n');
}

// ---------------------------------------------------------------------------
// Worker
// ---------------------------------------------------------------------------

async function fetchAsset(env, path, request) {
  const url = new URL(request.url);
  return env.ASSETS.fetch(new Request(url.origin + path, request));
}

function isHtmlResponse(response) {
  const type = response.headers.get('content-type') || '';
  return response.status === 200 && type.includes('text/html');
}

function isMarkdownResponse(response) {
  const type = response.headers.get('content-type') || '';
  return response.status === 200 && !type.includes('text/html');
}

async function findAsset(env, request, candidates, predicate) {
  for (const path of candidates) {
    const response = await fetchAsset(env, path, request);
    if (predicate(response)) return { response, path };
  }
  return null;
}

/** GET/HEAD /api/v1/* — read-only JSON surface described by /openapi.json. */
async function handleApi(request, url, env) {
  const method = request.method.toUpperCase();
  const path = url.pathname.replace(/\/+$/, '') || '/api';

  const sendJson = (data, cacheControl = 'public, max-age=60') =>
    new Response(JSON.stringify(data, null, 2) + '\n', {
      headers: {
        'Content-Type': 'application/json; charset=utf-8',
        'Cache-Control': cacheControl,
        ...rateLimitHeaders(),
      },
    });

  const allowMethods = 'GET, HEAD, OPTIONS';

  const endpoints = {
    '/api/v1/status': () =>
      sendJson({
        service: 'rtylr.com',
        status: 'ok',
        time: new Date().toISOString(),
        endpoints: {
          openapi: '/openapi.json',
          pages: '/api/v1/pages',
          status: '/api/v1/status',
          sitemap: '/sitemap.xml',
        },
      }),
    '/api/v1/pages': async () => {
      const found = await findAsset(env, request, ['/pages.json'], (r) => r.status === 200).catch(
        () => null
      );
      if (!found) {
        return jsonError(
          500,
          'internal_error',
          'The page index is temporarily unavailable.',
          `Use ${SITE_URL}/sitemap.xml instead.`
        );
      }
      const data = await found.response.json();
      const pages = Array.isArray(data.pages) ? data.pages : [];
      return sendJson({ count: pages.length, pages });
    },
  };

  const handler = endpoints[path];
  if (!handler) {
    // Unknown resource -> 404 regardless of the verb used.
    return jsonError(
      404,
      'not_found',
      `Unknown API endpoint: ${path}.`,
      'Fetch /openapi.json for the list of available endpoints.'
    );
  }

  if (method === 'OPTIONS') {
    return new Response(null, { status: 204, headers: { Allow: allowMethods, ...rateLimitHeaders() } });
  }
  if (method !== 'GET' && method !== 'HEAD') {
    const res = jsonError(
      405,
      'method_not_allowed',
      `${method} is not supported on ${path}.`,
      `Use GET (or HEAD). Allowed methods: ${allowMethods}.`
    );
    res.headers.set('Allow', allowMethods);
    return res;
  }

  return handler();
}

export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    const pathname = url.pathname;

    // 1. Read-only JSON API — errors are structured JSON, never HTML.
    //    Checked before the static passthrough so guesses like
    //    /api/v1/status.json get a JSON 404 instead of an HTML one.
    if (pathname === '/api' || pathname.startsWith('/api/')) {
      return handleApi(request, url, env);
    }

    // 2. Machine-readable static assets pass through untouched.
    if (STATIC_EXT.test(pathname)) {
      return env.ASSETS.fetch(request);
    }

    // 3. Legacy parity with the previous _redirects rule: /version -> /version.json.
    if (pathname === '/version' || pathname === '/version/') {
      return fetchAsset(env, '/version.json', request);
    }

    const accept = request.headers.get('accept');
    const chosen = negotiate(accept, DOCUMENT_TYPES);

    // Client rejected every representation we can produce.
    if (chosen === null) {
      const res = new Response(
        `Not Acceptable\n\nAvailable representations: ${DOCUMENT_TYPES.join(', ')}\n`,
        { status: 406, headers: { 'Content-Type': 'text/plain; charset=utf-8' } }
      );
      varyOnNegotiation(res.headers);
      return res;
    }

    // 4. Markdown representation.
    if (chosen === 'text/markdown') {
      const md = await findAsset(env, request, markdownCandidates(pathname), isMarkdownResponse);
      if (md) {
        const res = new Response(md.response.body, md.response);
        res.headers.set('Content-Type', 'text/markdown; charset=utf-8');
        varyOnNegotiation(res.headers);
        return res;
      }
      // No markdown sibling: continue to HTML below; if that also misses,
      // the 404 flow picks the representation the client prefers.
    }

    // 5. HTML representation (or markdown-preferred fallthrough).
    const html = await findAsset(env, request, htmlCandidates(pathname), isHtmlResponse);

    if (html) {
      const res = new Response(html.response.body, html.response);
      varyOnNegotiation(res.headers);
      // Advertise the markdown alternate for discovery-style agents.
      const md = await findAsset(env, request, markdownCandidates(pathname), isMarkdownResponse);
      if (md) {
        const linkValue = `<${md.path}>; rel="alternate"; type="text/markdown"`;
        const existingLink = res.headers.get('link');
        res.headers.set('Link', existingLink ? `${existingLink}, ${linkValue}` : linkValue);
      }
      return res;
    }

    // 6. Real HTTP 404 — markdown body for agents, styled HTML for browsers.
    const notFoundChoice = negotiate(accept, DOCUMENT_TYPES);
    if (notFoundChoice === 'text/markdown') {
      const res = new Response(markdownNotFound(pathname), {
        status: 404,
        headers: {
          'Content-Type': 'text/markdown; charset=utf-8',
          'Cache-Control': 'public, max-age=60',
          'X-Robots-Tag': 'noindex',
        },
      });
      varyOnNegotiation(res.headers);
      return res;
    }

    // '/404/' (the clean directory URL), not the literal '/404.html' file —
    // same 308-on-literal-.html redirect issue as htmlCandidates() above.
    const styled = await fetchAsset(env, '/404/', request);
    if (styled && isHtmlResponse(styled)) {
      const res = new Response(styled.body, { status: 404, headers: styled.headers });
      varyOnNegotiation(res.headers);
      return res;
    }

    const res = new Response('Not Found\n', {
      status: 404,
      headers: { 'Content-Type': 'text/plain; charset=utf-8' },
    });
    varyOnNegotiation(res.headers);
    return res;
  },
};
