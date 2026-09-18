/**
 * Shape of the response from `GET {API_URL}/license/state`.
 *
 * The backend never returns the license key or any secret material here —
 * only enough state for the frontend to decide whether to surface a notice.
 */
export type LicenseMode = 'hosted' | 'self_hosted';

export type LicenseStatus = 'hosted' | 'valid' | 'missing' | 'expired' | 'invalid' | 'unreachable';

export interface LicenseState {
  mode?: LicenseMode | string;
  valid?: boolean;
  status?: LicenseStatus | string;
  reason?: string;
  /** RFC3339 timestamp */
  expiresAt?: string | null;
  /** RFC3339 timestamp */
  graceUntil?: string | null;
  /** RFC3339 timestamp */
  checkedAt?: string;
}

/** Minimal structural type implemented by the generated OneKit client. */
export interface LicenseClient {
  getLicenseState(request: Record<string, never>): Promise<LicenseState>;
}
