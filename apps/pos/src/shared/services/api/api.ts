import { createApiClient, safeFetch } from '@voxire/api';
import type { SafeApi } from '@voxire/api';
import { createLocalFirstSafeApi, startLocalFirstSync } from './localFirstSafe';

const api = createApiClient('pos');
const remoteSafe = safeFetch(api);

startLocalFirstSync(remoteSafe);

export const safe: SafeApi = createLocalFirstSafeApi(remoteSafe);
export default api;
