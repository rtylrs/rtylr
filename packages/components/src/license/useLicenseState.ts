import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import { unwrap, type SafeApi } from '@voxire/api';
import type { LicenseState } from './types';

/** How long a fetched license state is considered fresh (5 minutes). */
const LICENSE_STALE_TIME = 5 * 60 * 1000;

/**
 * Fetches the self-host license state from `GET /license/state` on the app's
 * existing API base, via the provided `safeFetch` client.
 *
 * Returns no secrets — only the public license state used to decide whether an
 * admin-facing notice should be shown. In hosted deployments the backend
 * returns `{ status: 'hosted', valid: true }`, so callers render nothing.
 *
 * @param safe    - the app's `SafeApi` instance (from `safeFetch(api)`)
 * @param enabled - when false, no request is made (e.g. for non-admins)
 */
export function useLicenseState(
  safe: SafeApi,
  enabled: boolean = true
): UseQueryResult<LicenseState, unknown> {
  return useQuery<LicenseState, unknown>({
    queryKey: ['license', 'state'],
    queryFn: (): Promise<LicenseState> => unwrap(safe.get<LicenseState>('/license/state')),
    enabled,
    staleTime: LICENSE_STALE_TIME,
    retry: 1,
    refetchOnWindowFocus: false,
  });
}
