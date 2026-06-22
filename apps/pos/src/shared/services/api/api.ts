import { createApiClient, safeFetch } from '@voxire/api';
import type { SafeApi } from '@voxire/api';
import { createLocalFirstSafeApi, startLocalFirstSync } from './localFirstSafe';
import { resolveApiBaseUrl } from '../server/serverConfig';

const api = createApiClient('pos');
// Override the env-derived base URL with the user-selected backend
// (self-host → VITE_API_URL → hosted default). createApiClient already
// applied the hosted/env default; this lets a self-host deployment win.
api.defaults.baseURL = resolveApiBaseUrl();
const remoteSafe = safeFetch(api);

startLocalFirstSync(remoteSafe);

export const safe: SafeApi = createLocalFirstSafeApi(remoteSafe);
export default api;
