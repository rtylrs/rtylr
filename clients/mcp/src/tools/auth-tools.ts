import type { Tool } from './types.js';
import { ok, err } from './types.js';
import { loadConfig, clearConfig, OAUTH_CLIENT_ID } from '../config.js';

export const authStatus: Tool = {
  name: 'auth_status',
  description: 'Check authentication status — returns current user info if logged in, or indicates not authenticated.',
  inputSchema: { type: 'object', properties: {} },
  handler: async (_input, api) => {
    const cfg = loadConfig();
    if (!cfg?.access_token) {
      return err('Not authenticated. The next tool call will open your browser to authorize access.');
    }
    try {
      const res = await api.get('/common/oauth/v1/me');
      return ok(res.data);
    } catch {
      return err('Token expired or invalid. Re-authenticate.');
    }
  },
};

export const authLogout: Tool = {
  name: 'auth_logout',
  description: 'Log out — revokes the current access token and clears local credentials.',
  inputSchema: { type: 'object', properties: {} },
  handler: async (_input, api) => {
    const cfg = loadConfig();
    if (!cfg?.access_token) {
      return ok({ message: 'Already logged out.' });
    }
    try {
      // Revocation keys off the refresh token; the access token expires on its own.
      if (cfg.refresh_token) {
        await api.post('/common/oauth/v1/revoke', { token: cfg.refresh_token, client_id: OAUTH_CLIENT_ID });
      }
    } catch {
      // best-effort revoke — clear locally regardless
    }
    clearConfig();
    return ok({ message: 'Logged out successfully.' });
  },
};

export const authTools = [authStatus, authLogout];
