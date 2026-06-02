import axios from 'axios';
import i18next from 'i18next';

const getBaseUrl = () => {
  return 'https://api.voxire.com/';
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
