import axios, { AxiosInstance } from 'axios';
import { API_BASE, loadConfig, saveConfig } from './config.js';
import { getAccessToken, refreshAccessToken } from './auth.js';

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
            const accessToken = await refreshAccessToken(cfg.refresh_token);
            err.config._retried = true;
            err.config.headers['Authorization'] = `Bearer ${accessToken}`;
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
