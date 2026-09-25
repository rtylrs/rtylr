import React from 'react';
import { useTranslation } from '@voxire/i18n';
import { useUserStore } from '@voxire/auth';
import { cn } from '@voxire/ui';
import { AlertTriangle, X } from 'lucide-react';
import { useLicenseState } from './useLicenseState';
import { resolveLicenseNotice, type LicenseNotice } from './notice';
import { isLicenseAdmin } from './isAdmin';
import type { LicenseClient } from './types';

export interface LicenseBannerProps {
  /** The generated OneKit license client for the host application. */
  client: LicenseClient;
  /** Optional extra classes for the outer bar. */
  className?: string;
}

const SESSION_DISMISS_PREFIX = 'rtylr:license-banner-dismissed:';

function dismissKey(notice: LicenseNotice): string {
  return `${SESSION_DISMISS_PREFIX}${notice.kind}`;
}

function isDismissed(notice: LicenseNotice): boolean {
  try {
    return sessionStorage.getItem(dismissKey(notice)) === '1';
  } catch {
    return false;
  }
}

function noticeMessage(
  notice: LicenseNotice,
  t: (key: string, options?: Record<string, unknown>) => string
): string {
  switch (notice.kind) {
    case 'expired':
      return t('license.expired');
    case 'unreachable':
      return t('license.unreachable');
    case 'expiring':
      return t('license.expiring', {
        date: notice.expiresAt ? new Date(notice.expiresAt).toLocaleDateString() : '',
      });
    case 'invalid':
    case 'missing':
    default:
      return t('license.unlicensed');
  }
}

/**
 * Admin-only self-host license notice bar.
 *
 * Renders NOTHING unless the current user is an admin/owner AND the license
 * state warrants a notice. Hosted deployments (and a plainly valid licence)
 * always render nothing, so the hosted product is visually unchanged.
 *
 * Mount it at the top of an app shell (e.g. above page content). It is a
 * complete no-op for non-admins and for hosted/valid installations.
 */
export function LicenseBanner({ client, className }: LicenseBannerProps): React.JSX.Element | null {
  const { t } = useTranslation();
  const user = useUserStore((s) => s.user);
  const role = useUserStore((s) => s.role);
  const admin = isLicenseAdmin(user, role);

  // Only admins ever trigger the network request.
  const { data } = useLicenseState(client, admin);

  const notice = React.useMemo(() => resolveLicenseNotice(data), [data]);

  const [dismissed, setDismissed] = React.useState<boolean>(() =>
    notice ? isDismissed(notice) : false
  );

  React.useEffect((): void => {
    setDismissed(notice ? isDismissed(notice) : false);
  }, [notice]);

  if (!admin || !notice || dismissed) return null;

  const handleDismiss = (): void => {
    try {
      sessionStorage.setItem(dismissKey(notice), '1');
    } catch {
      /* sessionStorage may be unavailable; dismiss for this render only */
    }
    setDismissed(true);
  };

  const isError = notice.severity === 'error';

  return (
    <div
      role="alert"
      aria-live="polite"
      className={cn(
        'flex w-full items-center gap-2 border-b px-4 py-2.5 text-sm',
        isError
          ? 'border-destructive/30 bg-destructive/10 text-destructive'
          : 'border-warning/30 bg-warning/10 text-warning',
        className
      )}
    >
      <AlertTriangle size={16} className="shrink-0" aria-hidden="true" />
      <span className="flex-1 leading-snug">{noticeMessage(notice, t)}</span>
      <button
        type="button"
        onClick={handleDismiss}
        aria-label={t('license.dismiss')}
        className="ml-2 inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-md opacity-70 transition-opacity hover:opacity-100"
      >
        <X size={16} aria-hidden="true" />
      </button>
    </div>
  );
}
