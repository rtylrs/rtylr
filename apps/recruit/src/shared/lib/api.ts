import { createApiClient, safeFetch } from '@voxire/api';
import type { SafeApi } from '@voxire/api';

// Recruit talks to the same backend as every other app, but its public
// pages are anonymous - the axios client still attaches the cookie if
// one happens to be present (so an admin previewing their own page
// doesn't look logged out), and the backend treats /recruit/public/*
// as anonymous regardless.
const api = createApiClient('recruit');
export const safe: SafeApi = safeFetch(api);
export default api;
