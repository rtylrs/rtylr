import axios from 'axios';
import i18next from 'i18next';

const getBaseUrl = () => {
  return 'https://api.voxire.dev/';
};

const getAuthUrl = () => {
  const isLocalhost =
    window.location.hostname === 'localhost' ||
    window.location.hostname === '127.0.0.1';
  return isLocalhost
    ? 'http://localhost:5176?redirect=hub'
    : 'https://auth.rtylr.com?redirect=hub';
};

const getTokenFromCookie = (): string | null =>
  document.cookie
    .split(';')
    .map((c) => c.trim())
    .find((c) => c.startsWith('token='))
    ?.split('=')[1] ?? null;

const api = axios.create({
  baseURL: getBaseUrl(),
});

api.interceptors.request.use(
  async (config) => {
    try {
      const token = getTokenFromCookie();
      if (token) {
        config.headers['authorization'] = `Bearer ${token}`;
      }
      config.headers['accept-language'] = i18next.language;
      return config;
    } catch (error) {
      return Promise.reject(error);
    }
  },
  (error) => {
    return Promise.reject(error);
  }
);

api.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    if (error.response?.status === 401) {
      document.cookie =
        'token=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
      window.location.href = getAuthUrl();
    }
    return Promise.reject(error);
  }
);

export default api;
