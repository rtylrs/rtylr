import axios from 'axios';
import i18next from 'i18next';
import { resolveApiBase } from '@voxire/config';
import { getCookie } from './cookies';

const getBaseUrl = (): string => {
  const envUrl = import.meta.env.VITE_POS_API_URL ?? import.meta.env.VITE_API_URL;
  if (envUrl) return envUrl.replace(/\/$/, '') + '/';
  try {
    const resolved = resolveApiBase();
    if (resolved) return resolved.replace(/\/$/, '') + '/';
  } catch {
    // ignore lookup error
  }
  return 'https://api.voxire.com/';
};

const api = axios.create({
  baseURL: getBaseUrl(),
});

api.interceptors.request.use(
  async (config) => {
    try {
      const token = getCookie('token');
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
    return Promise.reject(error);
  }
);

export default api;
