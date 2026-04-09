import axios from 'axios';

function getToken(): string | null {
  const cookies = document.cookie.split(';').map((c) => c.trim());
  for (const name of ['token', 'rtylr_session', 'auth_token']) {
    const found = cookies.find((c) => c.startsWith(`${name}=`))?.split('=')[1] ?? null;
    if (found) return found;
  }
  return null;
}

const api = axios.create({
  baseURL: 'https://api.voxire.dev/',
  headers: { 'Content-Type': 'application/json' },
});

api.interceptors.request.use((config) => {
  const token = getToken();
  if (token) config.headers['Authorization'] = `Bearer ${token}`;
  return config;
});

api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      window.location.href = `/auth?redirect=crm`;
    }
    return Promise.reject(err);
  },
);

export default api;
