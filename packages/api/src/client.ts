import axios, { type AxiosInstance } from 'axios';

const getBaseUrl = (): string => {
  return (import.meta as ImportMeta & { env: Record<string, string> }).env
    .VITE_API_URL ?? 'https://api.voxire.dev/';
};

const getToken = (): string | null => {
  return (
    document.cookie
      .split(';')
      .map((c) => c.trim())
      .find((c) => c.startsWith('token='))
      ?.split('=')[1] ?? null
  );
};

const clearTokenCookie = (): void => {
  document.cookie = 'token=; Max-Age=0; path=/;';
};

const getAuthUrl = (appId: string): string => {
  const isLocalhost =
    window.location.hostname === 'localhost' ||
    window.location.hostname === '127.0.0.1';
  const base = isLocalhost
    ? 'http://localhost:5176'
    : 'https://auth.rtylr.com';
  return `${base}?redirect=${appId}`;
};

export function createApiClient(appId: string): AxiosInstance {
  const client = axios.create({
    baseURL: getBaseUrl(),
  });

  client.interceptors.request.use(
    (config) => {
      const token = getToken();
      if (token) {
        config.headers['Authorization'] = `Bearer ${token}`;
      }
      return config;
    },
    (error) => Promise.reject(error)
  );

  client.interceptors.response.use(
    (response) => response,
    (error) => {
      if (error.response?.status === 401) {
        clearTokenCookie();
        window.location.href = getAuthUrl(appId);
      }
      return Promise.reject(error);
    }
  );

  return client;
}

export const api: AxiosInstance = createApiClient('hub');
