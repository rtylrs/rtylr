import type { LicenseState } from './types';

/** The kind of notice to surface. `null` means render nothing. */
export type LicenseNoticeKind = 'expired' | 'invalid' | 'missing' | 'unreachable' | 'expiring';

/** Severity drives the colour of the banner. */
export type LicenseNoticeSeverity = 'error' | 'warning';

export interface LicenseNotice {
  kind: LicenseNoticeKind;
  severity: LicenseNoticeSeverity;
  /** Present only for the `expiring` notice. */
  expiresAt?: string;
}

/** Window (days) before expiry within which a "expiring soon" notice shows. */
const EXPIRY_WINDOW_DAYS = 14;
const MS_PER_DAY = 24 * 60 * 60 * 1000;

/**
 * Decides which license notice (if any) is warranted for the given state.
 *
 * Returns `null` when nothing should be shown — notably for hosted
 * deployments (`status: 'hosted'`) and a plainly valid license that is not
 * close to expiry. This keeps the hosted product visually unchanged.
 */
export function resolveLicenseNotice(
  state: LicenseState | undefined,
  now: number = Date.now()
): LicenseNotice | null {
  if (!state) return null;

  switch (state.status) {
    case 'expired':
      return { kind: 'expired', severity: 'error' };
    case 'invalid':
      return { kind: 'invalid', severity: 'error' };
    case 'missing':
      return { kind: 'missing', severity: 'error' };
    case 'unreachable':
      return { kind: 'unreachable', severity: 'warning' };
    case 'valid': {
      if (!state.valid || !state.expiresAt) return null;
      const expiresMs = Date.parse(state.expiresAt);
      if (Number.isNaN(expiresMs)) return null;
      const daysLeft = (expiresMs - now) / MS_PER_DAY;
      if (daysLeft > 0 && daysLeft <= EXPIRY_WINDOW_DAYS) {
        return { kind: 'expiring', severity: 'warning', expiresAt: state.expiresAt };
      }
      return null;
    }
    case 'hosted':
    default:
      return null;
  }
}
