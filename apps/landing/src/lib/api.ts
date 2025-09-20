import axios from 'axios';
import i18next from 'i18next';

const getBaseUrl = () => {
  // Only access window on client side
  if (typeof window === 'undefined') {
    // Server-side: use production URL
    return 'https://onion.voxire.dev/';
  }

  const isLocalhost =
    window.location.hostname === 'localhost' ||
    window.location.hostname === '127.0.0.1';
  return isLocalhost ? 'http://localhost:9999/' : 'https://onion.voxire.dev/';
};

const api = axios.create({
  baseURL: getBaseUrl(),
});

api.interceptors.request.use(
  async (config) => {
    try {
      // Only access localStorage on client side
      if (typeof window !== 'undefined') {
        const user = localStorage.getItem('token');
        if (user) {
          config.headers['authorization'] = `Bearer ${user}`;
        }
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