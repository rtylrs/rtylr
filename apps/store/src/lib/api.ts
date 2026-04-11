import { createApiClient, safeFetch } from '@voxire/api';
import type { SafeApi } from '@voxire/api';

const api = createApiClient('store');
export const safe: SafeApi = safeFetch(api);
export default api;
