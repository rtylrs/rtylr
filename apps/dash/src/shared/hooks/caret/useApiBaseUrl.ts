/**
 * Resolve the API base URL from Vite env. Centralized so the Caret panel
 * and any future fetch wrappers stay in lockstep.
 */
export function useApiBaseUrl(): string {
  const url = import.meta.env.VITE_API_URL as string | undefined;
  return url ?? 'https://api.voxire.com/';
}
