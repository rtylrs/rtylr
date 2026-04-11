import axios from 'axios';
import { FLOW_API_BASE_URL } from './flowApiConfig';

function getToken(): string | undefined {
  const cookies = document.cookie.split(';');
  const names = ['token', 'rtylr_session', 'auth_token'];
  for (const cookie of cookies) {
    const [key, value] = cookie.trim().split('=');
    if (names.includes(key)) return decodeURIComponent(value);
  }
  return undefined;
}

const api = axios.create({
  baseURL: FLOW_API_BASE_URL,
  headers: { 'Content-Type': 'application/json' },
});

api.interceptors.request.use((config) => {
  const token = getToken();
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      window.location.href = 'https://auth.rtylr.com?redirect=' + window.location.hostname.split('.')[0];
    }
    return Promise.reject(error);
  }
);

export default api;
