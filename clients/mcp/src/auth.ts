import { createServer } from "http";
import { createHash, randomBytes } from "crypto";
import { URL } from "url";
import axios from "axios";
import open from "open";

import {
  API_BASE,
  OAUTH_CLIENT_ID,
  OAUTH_ISSUER,
  OAUTH_SCOPES,
  loadConfig,
  saveConfig,
  StoredConfig,
} from "./config.js";

// Returns a valid access token, triggering PKCE OAuth flow if needed.
export async function getAccessToken(): Promise<string> {
  const cfg = loadConfig();

  if (cfg?.access_token && cfg.expires_at > Date.now() + 5 * 60 * 1000) {
    return cfg.access_token;
  }

  if (cfg?.refresh_token) {
    try {
      return await refreshAccessToken(cfg.refresh_token);
    } catch {
      // Fall through to full auth flow.
    }
  }

  return runPKCEFlow();
}

const OAUTH_TIMEOUT_MS = 30_000;

// Single-flight: concurrent callers share one in-flight refresh so we never
// burn a rotated refresh token by racing two refresh requests.
let refreshInFlight: Promise<string> | null = null;

export function refreshAccessToken(refreshToken: string): Promise<string> {
  if (!refreshInFlight) {
    refreshInFlight = doRefreshToken(refreshToken).finally(() => {
      refreshInFlight = null;
    });
  }
  return refreshInFlight;
}

async function doRefreshToken(refreshToken: string): Promise<string> {
  const res = await axios.post<TokenResponse>(
    `${API_BASE}/common/oauth/v1/token`,
    {
      grant_type: "refresh_token",
      client_id: OAUTH_CLIENT_ID,
      refresh_token: refreshToken,
    },
    { timeout: OAUTH_TIMEOUT_MS },
  );

  // Some authorization servers omit refresh_token on refresh grants — keep the
  // existing one instead of overwriting it with undefined.
  const cfg: StoredConfig = {
    access_token: res.data.access_token,
    refresh_token: res.data.refresh_token ?? refreshToken,
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
  const codeChallenge = base64url(
    createHash("sha256").update(codeVerifier).digest(),
  );
  const state = base64url(randomBytes(16));

  const authorizeURL = new URL(`${API_BASE}/oauth/authorize`);
  authorizeURL.searchParams.set("client_id", OAUTH_CLIENT_ID);
  authorizeURL.searchParams.set("redirect_uri", redirectURI);
  authorizeURL.searchParams.set("scope", OAUTH_SCOPES);
  authorizeURL.searchParams.set("state", state);
  authorizeURL.searchParams.set("code_challenge", codeChallenge);
  authorizeURL.searchParams.set("code_challenge_method", "S256");

  process.stderr.write(
    `\nOpening browser for rtylr authorization...\n${authorizeURL.toString()}\n\n`,
  );

  const code = await waitForCallback(port, state, async () => {
    await open(authorizeURL.toString());
  });

  const res = await axios.post<TokenResponse>(
    `${API_BASE}/common/oauth/v1/token`,
    {
      grant_type: "authorization_code",
      client_id: OAUTH_CLIENT_ID,
      code,
      code_verifier: codeVerifier,
      redirect_uri: redirectURI,
    },
    { timeout: OAUTH_TIMEOUT_MS },
  );

  const cfg: StoredConfig = {
    access_token: res.data.access_token,
    refresh_token: res.data.refresh_token,
    expires_at: Date.now() + res.data.expires_in * 1000,
  };
  saveConfig(cfg);
  process.stderr.write("Authorization successful. Token saved.\n");
  return cfg.access_token;
}

function waitForCallback(
  port: number,
  expectedState: string,
  onReady: () => Promise<void>,
): Promise<string> {
  return new Promise((resolve, reject) => {
    let settled = false;
    let timeoutId: NodeJS.Timeout | undefined;

    const settle = (fn: () => void) => {
      if (settled) return;
      settled = true;
      clearTimeout(timeoutId);
      server.close();
      fn();
    };

    const server = createServer((req, res) => {
      const url = new URL(req.url ?? "/", `http://127.0.0.1:${port}`);

      // Ignore unrelated requests (e.g. browsers fetching /favicon.ico)
      // instead of aborting the flow.
      if (url.pathname !== "/callback") {
        res.writeHead(404, { "Content-Type": "text/plain" });
        res.end("Not found");
        return;
      }

      const code = url.searchParams.get("code");
      const state = url.searchParams.get("state");
      const error = url.searchParams.get("error");
      const iss = url.searchParams.get("iss");

      if (error) {
        res.writeHead(200, { "Content-Type": "text/html" });
        res.end(
          callbackHTML(
            "Authorization denied. You can close this window.",
            false,
          ),
        );
        settle(() => reject(new Error(`OAuth error: ${error}`)));
        return;
      }

      if (!code) {
        res.writeHead(400, { "Content-Type": "text/html" });
        res.end(callbackHTML("Missing authorization code.", false));
        settle(() => reject(new Error("Missing code in callback")));
        return;
      }

      if (state !== expectedState) {
        res.writeHead(400, { "Content-Type": "text/html" });
        res.end(callbackHTML("State mismatch — possible CSRF.", false));
        settle(() => reject(new Error("State mismatch")));
        return;
      }

      // RFC 9207: Validate issuer parameter if present. Onion currently sends
      // its configured identifier, while RFC-compliant issuers use their URL.
      if (iss && !isExpectedIssuer(iss)) {
        res.writeHead(400, { "Content-Type": "text/html" });
        res.end(
          callbackHTML(
            "Issuer mismatch — potential authorization mix-up attack.",
            false,
          ),
        );
        settle(() => reject(new Error(`RFC 9207 issuer mismatch: got ${iss}`)));
        return;
      }

      res.writeHead(200, { "Content-Type": "text/html" });
      res.end(
        callbackHTML(
          "Authorization complete. You can close this window.",
          true,
        ),
      );
      settle(() => resolve(code));
    });

    server.on("error", (err) => {
      settle(() => reject(err));
    });

    server.listen(port, "127.0.0.1", async () => {
      try {
        await onReady();
      } catch (err) {
        settle(() =>
          reject(err instanceof Error ? err : new Error(String(err))),
        );
      }
    });

    // Timeout after 5 minutes.
    timeoutId = setTimeout(
      () => {
        settle(() => reject(new Error("Authorization timed out")));
      },
      5 * 60 * 1000,
    );
  });
}

function findFreePort(): Promise<number> {
  return new Promise((resolve, reject) => {
    const srv = createServer();
    srv.listen(0, "127.0.0.1", () => {
      const addr = srv.address();
      if (!addr || typeof addr === "string") {
        srv.close();
        reject(new Error("Could not find free port"));
        return;
      }
      const port = addr.port;
      srv.close(() => resolve(port));
    });
  });
}

function isExpectedIssuer(issuer: string): boolean {
  if (issuer === OAUTH_ISSUER) return true;
  try {
    return new URL(issuer).origin === new URL(API_BASE).origin;
  } catch {
    return false;
  }
}

function base64url(buf: Buffer): string {
  return buf.toString("base64url");
}

function callbackHTML(message: string, success: boolean): string {
  return `<!DOCTYPE html><html><head><meta charset="UTF-8">
<style>
  body { font-family: -apple-system, sans-serif; display: flex; align-items: center; justify-content: center; min-height: 100vh; margin: 0; background: #0a0a0a; color: #e5e5e5; }
  .card { background: #141414; border: 1px solid #2a2a2a; border-radius: 12px; padding: 40px; text-align: center; }
  .logo { color: #64C5B9; font-size: 20px; font-weight: 700; margin-bottom: 16px; }
  p { color: ${success ? "#64C5B9" : "#ef4444"}; font-size: 14px; }
</style></head><body><div class="card"><div class="logo">rtylr</div><p>${message}</p></div></body></html>`;
}
