#!/usr/bin/env node
/**
 * Local verification harness: runs the real cloudflare/_worker.js against the
 * REAL dist/landing output using a filesystem-backed ASSETS binding, and
 * serves it over HTTP so every audit probe can be reproduced with curl.
 *
 * Usage: node scripts/verify-agentic.mjs [port]
 */
import { createServer } from 'node:http';
import { readFileSync, existsSync, statSync } from 'node:fs';
import { join, extname } from 'node:path';
import { fileURLToPath } from 'node:url';
import { dirname } from 'node:path';

const __dirname = dirname(fileURLToPath(import.meta.url));
const DIST = join(__dirname, '..', '..', '..', 'dist', 'landing');
const PORT = Number(process.argv[2] || 8787);

const MIME = {
  '.html': 'text/html; charset=utf-8',
  '.md': 'text/markdown; charset=utf-8',
  '.json': 'application/json; charset=utf-8',
  '.xml': 'application/xml',
  '.txt': 'text/plain; charset=utf-8',
  '.js': 'application/javascript',
  '.css': 'text/css',
  '.png': 'image/png',
  '.svg': 'image/svg+xml',
  '.ico': 'image/x-icon',
  '.woff2': 'font/woff2',
};

const worker = (await import(join(__dirname, '..', 'cloudflare', '_worker.js'))).default;

const env = {
  ASSETS: {
    async fetch(request) {
      const path = decodeURIComponent(new URL(request.url).pathname);
      const candidates = [];
      if (path.endsWith('/')) candidates.push(path + 'index.html');
      else candidates.push(path);
      for (const candidate of candidates) {
        const full = join(DIST, candidate);
        try {
          if (existsSync(full) && statSync(full).isFile()) {
            const body = readFileSync(full);
            return new Response(body, {
              status: 200,
              headers: { 'Content-Type': MIME[extname(full)] ?? 'application/octet-stream' },
            });
          }
        } catch {
          /* fall through */
        }
      }
      return new Response('asset miss', { status: 404 });
    },
  },
};

const server = createServer(async (req, res) => {
  const url = `http://localhost:${PORT}${req.url}`;
  const request = new Request(url, {
    method: req.method,
    headers: req.headers,
  });
  const response = await worker.fetch(request, env);
  res.statusCode = response.status;
  response.headers.forEach((value, key) => res.setHeader(key, value));
  const body = await response.arrayBuffer();
  res.end(Buffer.from(body));
});

server.listen(PORT, () => {
  console.log(`verify-agentic: dist/landing served through _worker.js on http://127.0.0.1:${PORT}`);
});
