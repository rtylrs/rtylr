import type { Tool } from './types.js';
import { ok, err } from './types.js';
import { loadConfig, clearConfig } from '../config.js';

export const authStatus: Tool = {
  name: 'auth_status',
  description: 'Check authentication status — returns current user info if logged in, or indicates not authenticated.',
  inputSchema: { type: 'object', properties: {} },
  handler: async (_input, api) => {
    const cfg = loadConfig();
    if (!cfg?.access_token) {
      return err('Not authenticated. Run `rtylr-mcp login` to authenticate.');
    }
    try {
      const res = await api.get('/oauth/me');
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
      await api.post('/oauth/revoke', { token: cfg.access_token, token_type_hint: 'access_token' });
    } catch {
      // best-effort revoke — clear locally regardless
    }
    clearConfig();
    return ok({ message: 'Logged out successfully.' });
  },
};

export const authTools = [authStatus, authLogout];
