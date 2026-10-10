import React from 'react';

import { useConsent } from './useConsent';

/**
 * Strict opt-in cookie banner.
 *
 * Rendered at the app root. Shows only when consent is `undecided`;
 * once the user accepts or declines, the banner hides and we don't
 * prompt again until cookies are cleared.
 *
 * Visual: a slim pill-shaped bar anchored to the bottom-left corner
 * rather than a centered modal. Keeps the landing page hero visible
 * and matches the restrained aesthetic of the rest of the marketing
 * site. Apps can still override the whole thing via `className`.
 *
 * `tone="light"` swaps to a white card for light pages, where the dark bar
 * reads as a foreign object. Buttons carry CONSENT_BUTTON_CLASS so a host
 * with global button styles can exclude them.
 */
interface CookieBannerProps {
  privacyUrl?: string;
  className?: string;
  /** Copy override. Callers that translate the banner pass their own node;
   *  everyone else gets the package's English copy. */
  message?: React.ReactNode;
  declineLabel?: string;
  acceptLabel?: string;
  ariaLabel?: string;
  /** Dark bar (default) or a white card for light pages. */
  tone?: 'dark' | 'light';
  /** Stacking order. Some hosts have their own overlays to sit under. */
  zIndex?: number;
  /** Called after the user accepts - good hook for initialising
   *  analytics that depends on consent. */
  onAccept?: () => void;
  /** Called after the user declines - clear any residual tracking. */
  onDecline?: () => void;
}

/** Class on both banner buttons, so hosts with global button rules can
 *  exclude them (e.g. `:not([class*='consent-btn'])`). */
export const CONSENT_BUTTON_CLASS = 'consent-btn';

/** Style for the privacy link, exported so a translated banner can render
 *  its own link without re-stating the visual treatment. */
export const privacyLinkStyle: React.CSSProperties = {
  display: 'inline-flex',
  alignItems: 'center',
  minHeight: 24,
  color: 'rgba(255,255,255,0.92)',
  textDecoration: 'underline',
  textDecorationColor: 'rgba(255,255,255,0.3)',
  textUnderlineOffset: 2,
};

/** The privacy link for `tone="light"`. */
export const privacyLinkStyleLight: React.CSSProperties = {
  ...privacyLinkStyle,
  color: '#131313',
  textDecorationColor: 'rgba(19,19,19,0.3)',
};

const TONES = {
  dark: {
    card: {
      background: 'rgba(15, 15, 15, 0.92)',
      color: 'rgba(255,255,255,0.92)',
      border: '1px solid rgba(255,255,255,0.08)',
      boxShadow:
        '0 10px 40px rgba(0,0,0,0.35), 0 2px 8px rgba(0,0,0,0.2), inset 0 1px 0 rgba(255,255,255,0.04)',
      fontFamily: 'ui-sans-serif, system-ui, -apple-system, "Segoe UI", Roboto, sans-serif',
    },
    text: 'rgba(255,255,255,0.78)',
    decline: { border: '1px solid rgba(255,255,255,0.12)', color: 'rgba(255,255,255,0.72)' },
    declineHover: { background: 'rgba(255,255,255,0.05)', color: 'rgba(255,255,255,0.92)' },
    accept: { background: 'rgba(255,255,255,0.95)', color: '#0a0a0a' },
    acceptHover: '#ffffff',
    button: { minHeight: 44, minWidth: 44, padding: '10px 16px' },
    acceptPadding: '10px 18px',
  },
  light: {
    card: {
      background: 'rgba(255, 255, 255, 0.96)',
      color: '#131313',
      border: '1px solid rgba(19,19,19,0.08)',
      boxShadow: '0 12px 32px -12px rgba(19,19,19,0.22), 0 1px 2px rgba(19,19,19,0.04)',
      fontFamily: 'inherit',
    },
    text: 'rgba(19,19,19,0.68)',
    decline: { border: '1px solid rgba(19,19,19,0.12)', color: 'rgba(19,19,19,0.7)' },
    declineHover: { background: 'rgba(19,19,19,0.04)', color: '#131313' },
    accept: { background: '#131313', color: '#ffffff' },
    acceptHover: '#000000',
    button: { minHeight: 30, minWidth: 30, padding: '0 12px' },
    acceptPadding: '0 14px',
  },
} as const;

export function CookieBanner({
  privacyUrl = '/legal/privacy-policy',
  className,
  message,
  declineLabel = 'Decline',
  acceptLabel = 'Accept',
  ariaLabel = 'Cookie consent',
  tone = 'dark',
  zIndex = 9999,
  onAccept,
  onDecline,
}: CookieBannerProps): React.JSX.Element | null {
  const { state, accept, decline } = useConsent();
  if (state !== 'undecided') return null; // null (pre-mount) and decided states all hide the banner

  const palette = TONES[tone];

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
      aria-label={ariaLabel}
      aria-modal="false"
      aria-describedby="cookie-consent-desc"
      className={className}
      style={{
        position: 'fixed',
        bottom: 16,
        left: 16,
        zIndex,
        maxWidth: 'min(380px, calc(100vw - 32px))',
        padding: '14px 16px',
        borderRadius: 14,
        ...palette.card,
        backdropFilter: 'blur(20px)',
        WebkitBackdropFilter: 'blur(20px)',
        fontSize: 13,
        lineHeight: 1.5,
        letterSpacing: '-0.005em',
        display: 'flex',
        flexDirection: 'column',
        gap: 12,
      }}
    >
      <div id="cookie-consent-desc" style={{ color: palette.text, fontSize: 12.5 }}>
        {message ?? (
          <>
            We use essential cookies to run rtylr, and optional analytics to understand which
            features to build next.{' '}
            <a href={privacyUrl} style={tone === 'light' ? privacyLinkStyleLight : privacyLinkStyle}>
              Privacy
            </a>
            .
          </>
        )}
      </div>
      <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end', alignItems: 'center' }}>
        <button
          type="button"
          onClick={handleDecline}
          className={CONSENT_BUTTON_CLASS}
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            justifyContent: 'center',
            ...palette.button,
            borderRadius: 999,
            ...palette.decline,
            background: 'transparent',
            fontSize: 12.5,
            fontWeight: 500,
            cursor: 'pointer',
            fontFamily: 'inherit',
            letterSpacing: '-0.005em',
            transition: 'background 0.15s ease, color 0.15s ease',
          }}
          onMouseEnter={(e) => {
            (e.currentTarget as HTMLButtonElement).style.background = palette.declineHover.background;
            (e.currentTarget as HTMLButtonElement).style.color = palette.declineHover.color;
          }}
          onMouseLeave={(e) => {
            (e.currentTarget as HTMLButtonElement).style.background = 'transparent';
            (e.currentTarget as HTMLButtonElement).style.color = palette.decline.color;
          }}
        >
          {declineLabel}
        </button>
        <button
          type="button"
          onClick={handleAccept}
          className={CONSENT_BUTTON_CLASS}
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            justifyContent: 'center',
            ...palette.button,
            padding: palette.acceptPadding,
            borderRadius: 999,
            border: 'none',
            ...palette.accept,
            fontSize: 12.5,
            fontWeight: 600,
            cursor: 'pointer',
            fontFamily: 'inherit',
            letterSpacing: '-0.005em',
            transition: 'background 0.15s ease',
          }}
          onMouseEnter={(e) => {
            (e.currentTarget as HTMLButtonElement).style.background = palette.acceptHover;
          }}
          onMouseLeave={(e) => {
            (e.currentTarget as HTMLButtonElement).style.background = palette.accept.background;
          }}
        >
          {acceptLabel}
        </button>
      </div>
    </div>
  );
}
