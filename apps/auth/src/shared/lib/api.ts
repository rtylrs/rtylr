import { createApiClient, safeFetch } from '@voxire/api';
import type { SafeApi } from '@voxire/api';
import i18next from 'i18next';

// Auth app uses createApiClient but doesn't redirect on 401 (it IS auth).
// The 401 interceptor from createApiClient redirects to auth, which is fine
// because if you're already on auth with an expired token, it just reloads.
const api = createApiClient('auth');

// Inject accept-language header from i18next
api.interceptors.request.use((config) => {
  config.headers['accept-language'] = i18next.language;
  return config;
});

export const safe: SafeApi = safeFetch(api);
export default api;
