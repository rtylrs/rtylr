import { createServer } from 'http';
import { createHash, randomBytes } from 'crypto';
import { URL } from 'url';
import axios from 'axios';
import open from 'open';

import {
  API_BASE,
  OAUTH_CLIENT_ID,
  OAUTH_SCOPES,
  loadConfig,
  saveConfig,
  StoredConfig,
} from './config.js';

// Returns a valid access token, triggering PKCE OAuth flow if needed.
export async function getAccessToken(): Promise<string> {
  const cfg = loadConfig();

  if (cfg?.access_token) {
    // If the token expires in more than 5 minutes, use it directly.
    if (cfg.expires_at > Date.now() + 5 * 60 * 1000) {
      return cfg.access_token;
    }
    // Try to refresh.
    if (cfg.refresh_token) {
      try {
        return await refreshToken(cfg.refresh_token);
      } catch {
        // Fall through to full auth flow.
      }
    }
  }

  return runPKCEFlow();
}

async function refreshToken(refreshToken: string): Promise<string> {
  const res = await axios.post<TokenResponse>(`${API_BASE}/oauth/token`, {
    grant_type: 'refresh_token',
    client_id: OAUTH_CLIENT_ID,
    refresh_token: refreshToken,
  });

  const cfg: StoredConfig = {
    access_token: res.data.access_token,
    refresh_token: res.data.refresh_token,
    expires_at: Date.now() + res.data.expires_in * 1000,
  };
  saveConfig(cfg);
  return cfg.access_token;
}

interface TokenResponse {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_in: number;
}

async function runPKCEFlow(): Promise<string> {
  const port = await findFreePort();
  const redirectURI = `http://127.0.0.1:${port}/callback`;

  const codeVerifier = base64url(randomBytes(64));
  const codeChallenge = base64url(createHash('sha256').update(codeVerifier).digest());

  const state = base64url(randomBytes(16));

  const authorizeURL = new URL(`${API_BASE}/oauth/authorize`);
  authorizeURL.searchParams.set('client_id', OAUTH_CLIENT_ID);
  authorizeURL.searchParams.set('redirect_uri', redirectURI);
  authorizeURL.searchParams.set('scope', OAUTH_SCOPES);
  authorizeURL.searchParams.set('state', state);
  authorizeURL.searchParams.set('code_challenge', codeChallenge);
  authorizeURL.searchParams.set('code_challenge_method', 'S256');

  process.stderr.write(`\nOpening browser for rtylr authorization...\n${authorizeURL.toString()}\n\n`);

  const code = await waitForCallback(port, state, async () => {
    await open(authorizeURL.toString());
  });

  const res = await axios.post<TokenResponse>(`${API_BASE}/oauth/token`, {
    grant_type: 'authorization_code',
    client_id: OAUTH_CLIENT_ID,
    code,
    code_verifier: codeVerifier,
    redirect_uri: redirectURI,
  });

  const cfg: StoredConfig = {
    access_token: res.data.access_token,
    refresh_token: res.data.refresh_token,
    expires_at: Date.now() + res.data.expires_in * 1000,
  };
  saveConfig(cfg);
  process.stderr.write('Authorization successful. Token saved.\n');
  return cfg.access_token;
}

function waitForCallback(port: number, expectedState: string, onReady: () => Promise<void>): Promise<string> {
  return new Promise((resolve, reject) => {
    const server = createServer((req, res) => {
      const url = new URL(req.url ?? '/', `http://127.0.0.1:${port}`);
      const code = url.searchParams.get('code');
      const state = url.searchParams.get('state');
      const error = url.searchParams.get('error');

      if (error) {
        res.writeHead(200, { 'Content-Type': 'text/html' });
        res.end(callbackHTML('Authorization denied. You can close this window.', false));
        server.close();
        reject(new Error(`OAuth error: ${error}`));
        return;
      }

      if (!code) {
        res.writeHead(400, { 'Content-Type': 'text/html' });
        res.end(callbackHTML('Missing authorization code.', false));
        server.close();
        reject(new Error('Missing code in callback'));
        return;
      }

      if (state !== expectedState) {
        res.writeHead(400, { 'Content-Type': 'text/html' });
        res.end(callbackHTML('State mismatch — possible CSRF.', false));
        server.close();
        reject(new Error('State mismatch'));
        return;
      }

      res.writeHead(200, { 'Content-Type': 'text/html' });
      res.end(callbackHTML('Authorization complete. You can close this window.', true));
      server.close();
      resolve(code);
    });

    server.listen(port, '127.0.0.1', async () => {
      try {
        await onReady();
      } catch (err) {
        server.close();
        reject(err);
      }
    });

    // Timeout after 5 minutes.
    setTimeout(() => {
      server.close();
      reject(new Error('Authorization timed out'));
    }, 5 * 60 * 1000);
  });
}

function findFreePort(): Promise<number> {
  return new Promise((resolve, reject) => {
    const srv = createServer();
    srv.listen(0, '127.0.0.1', () => {
      const addr = srv.address();
      if (!addr || typeof addr === 'string') {
        srv.close();
        reject(new Error('Could not find free port'));
        return;
      }
      const port = addr.port;
      srv.close(() => resolve(port));
    });
  });
}

function base64url(buf: Buffer): string {
  return buf.toString('base64').replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
}

function callbackHTML(message: string, success: boolean): string {
  return `<!DOCTYPE html><html><head><meta charset="UTF-8">
<style>
  body { font-family: -apple-system, sans-serif; display: flex; align-items: center; justify-content: center; min-height: 100vh; margin: 0; background: #0a0a0a; color: #e5e5e5; }
  .card { background: #141414; border: 1px solid #2a2a2a; border-radius: 12px; padding: 40px; text-align: center; }
  .logo { color: #64C5B9; font-size: 20px; font-weight: 700; margin-bottom: 16px; }
  p { color: ${success ? '#64C5B9' : '#ef4444'}; font-size: 14px; }
</style></head><body><div class="card"><div class="logo">rtylr</div><p>${message}</p></div></body></html>`;
}
