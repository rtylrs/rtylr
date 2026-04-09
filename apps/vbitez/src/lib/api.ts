import axios from 'axios';
import i18next from 'i18next';

function getToken(): string | undefined {
  const cookies = document.cookie.split(';');
  const names = ['token', 'rtylr_session', 'auth_token'];
  for (const cookie of cookies) {
    const [key, value] = cookie.trim().split('=');
    if (names.includes(key)) return decodeURIComponent(value);
  }
  // Fallback to localStorage for backward compat
  return localStorage.getItem('token') ?? undefined;
}

const api = axios.create({
  baseURL: 'https://api.voxire.dev/',
  headers: { 'Content-Type': 'application/json' },
});

api.interceptors.request.use(
  async (config) => {
    try {
      const token = getToken();
      if (token) {
        config.headers['Authorization'] = `Bearer ${token}`;
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
      window.location.href = 'https://auth.rtylr.com?redirect=' + window.location.hostname.split('.')[0];
    }
    return Promise.reject(error);
  }
);

export default api;
