import React from 'react';

import { useConsent } from './useConsent';

/**
 * Strict opt-in cookie banner.
 *
 * Rendered at the app root. Shows only when consent is `undecided`;
 * once the user accepts or declines, the banner hides and we don't
 * prompt again until cookies are cleared.
 *
 * Styling is intentionally plain — each app wraps this or supplies
 * its own via the `className` prop to match local design. Defaults
 * to a dark bar at the bottom that works on both the landing
 * (light-theme) and recruit admin (dark theme) pages.
 */
interface CookieBannerProps {
  privacyUrl?: string;
  className?: string;
  /** Called after the user accepts — good hook for initialising
   *  analytics that depends on consent. */
  onAccept?: () => void;
  /** Called after the user declines — clear any residual tracking. */
  onDecline?: () => void;
}

export function CookieBanner({
  privacyUrl = '/legal/privacy-policy',
  className,
  onAccept,
  onDecline,
}: CookieBannerProps): React.JSX.Element | null {
  const { state, accept, decline } = useConsent();
  if (state !== 'undecided') return null;

  const handleAccept = (): void => {
    accept();
    onAccept?.();
  };
  const handleDecline = (): void => {
    decline();
    onDecline?.();
  };

  return (
    <div
      role="dialog"
      aria-label="Cookie consent"
      className={className}
      style={{
        position: 'fixed',
        bottom: 16,
        left: 16,
        right: 16,
        zIndex: 9999,
        maxWidth: 560,
        marginInline: 'auto',
        padding: '14px 18px',
        borderRadius: 12,
        background: 'rgba(15, 15, 15, 0.95)',
        color: 'white',
        backdropFilter: 'blur(12px)',
        boxShadow: '0 8px 32px rgba(0,0,0,0.25)',
        fontSize: 13,
        lineHeight: 1.5,
        display: 'flex',
        flexWrap: 'wrap',
        alignItems: 'center',
        gap: 12,
      }}
    >
      <div style={{ flex: 1, minWidth: 240 }}>
        We use essential cookies to run rtylr, and optional analytics to understand which features
        to build next.{' '}
        <a href={privacyUrl} style={{ color: '#64C5B9', textDecoration: 'underline' }}>
          Privacy policy
        </a>
        .
      </div>
      <div style={{ display: 'flex', gap: 8 }}>
        <button
          type="button"
          onClick={handleDecline}
          style={{
            padding: '7px 14px',
            borderRadius: 8,
            border: '1px solid rgba(255,255,255,0.18)',
            background: 'transparent',
            color: 'white',
            fontSize: 13,
            cursor: 'pointer',
          }}
        >
          Decline
        </button>
        <button
          type="button"
          onClick={handleAccept}
          style={{
            padding: '7px 14px',
            borderRadius: 8,
            border: 'none',
            background: '#64C5B9',
            color: '#0a0a0a',
            fontSize: 13,
            fontWeight: 600,
            cursor: 'pointer',
          }}
        >
          Accept
        </button>
      </div>
    </div>
  );
}
