import axios, { AxiosInstance } from 'axios';
import { API_BASE, loadConfig, saveConfig } from './config.js';
import { getAccessToken } from './auth.js';

let _client: AxiosInstance | null = null;

export function createClient(): AxiosInstance {
  if (_client) return _client;

  _client = axios.create({ baseURL: API_BASE });

  _client.interceptors.request.use(async (config) => {
    const token = await getAccessToken();
    config.headers['Authorization'] = `Bearer ${token}`;
    return config;
  });

  _client.interceptors.response.use(
    (res) => res,
    async (err) => {
      // On 401, clear stored token and retry once.
      if (err.response?.status === 401 && !err.config._retried) {
        const cfg = loadConfig();
        if (cfg?.refresh_token) {
          try {
            const res = await axios.post(`${API_BASE}/oauth/token`, {
              grant_type: 'refresh_token',
              client_id: 'rtylr-mcp',
              refresh_token: cfg.refresh_token,
            });
            saveConfig({
              access_token: res.data.access_token,
              refresh_token: res.data.refresh_token,
              expires_at: Date.now() + res.data.expires_in * 1000,
            });
            err.config._retried = true;
            err.config.headers['Authorization'] = `Bearer ${res.data.access_token}`;
            return _client!.request(err.config);
          } catch {
            // Refresh failed — let the original error propagate.
          }
        }
      }
      return Promise.reject(err);
    }
  );

  return _client;
}

export function apiError(err: unknown): string {
  if (axios.isAxiosError(err)) {
    return err.response?.data?.message ?? err.message;
  }
  return String(err);
}
