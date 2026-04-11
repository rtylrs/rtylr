import axios from 'axios';

function getToken(): string | null {
  const cookies = document.cookie.split(';').map((c) => c.trim());
  const cookieNames = ['rtylr_session', 'rtylr_token', 'auth_token', 'token'];
  for (const name of cookieNames) {
    const found = cookies.find((c) => c.startsWith(`${name}=`));
    if (found) return found.split('=').slice(1).join('=') || null;
  }
  return null;
}

const api = axios.create({
  baseURL: 'https://api.voxire.dev/',
  headers: {
    'Content-Type': 'application/json',
  },
});

api.interceptors.request.use((config) => {
  const token = getToken();
  if (token) {
    config.headers = config.headers ?? {};
    config.headers['Authorization'] = `Bearer ${token}`;
  }
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
