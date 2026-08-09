import axios from 'axios';
import i18next from 'i18next';
import { getCookie } from './cookies';

const getBaseUrl = (): string => import.meta.env.VITE_API_URL ?? 'https://api.voxire.com/';

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
