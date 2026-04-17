import { createApiClient, safeFetch } from '@voxire/api';
import type { SafeApi } from '@voxire/api';
import i18next from 'i18next';

const api = createApiClient('dash');

// Dash-specific: inject accept-language header from i18next
api.interceptors.request.use((config) => {
  config.headers['accept-language'] = i18next.language;
  return config;
});

export const safe: SafeApi = safeFetch(api);
export default api;
