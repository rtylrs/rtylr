import React from 'react';
import { useTranslation } from 'react-i18next';
import { APP_URLS } from '@voxire/config';

/**
 * Rennili's mark: a phone mid-shake with motion lines either side. The same
 * drawing as `entrypoints/rennili/public/brand/rennili-mark.svg`, inline so it
 * takes the text colour and can wiggle.
 */
export function RenniliMark({
  size = 28,
  wiggle = false,
  className = '',
}: {
  size?: number;
  wiggle?: boolean;
  className?: string;
}): React.JSX.Element {
  const classes = ['rennili-mark', wiggle ? 'rennili-mark--wiggle' : '', className]
    .filter(Boolean)
    .join(' ');
  return (
    <svg
      className={classes}
      width={size}
      height={size}
      viewBox="0 0 64 64"
      fill="none"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      focusable="false"
    >
      <g transform="rotate(-12 32 32)">
        <rect x="21.5" y="9.5" width="21" height="45" rx="6" strokeWidth="4.5" />
        <path d="M29 47.5h6" strokeWidth="4" />
      </g>
      <path d="M11.5 25c-2.6 4.6-2.6 9.4 0 14" strokeWidth="4" />
      <path d="M5 20.5c-4.4 7.6-4.4 15.4 0 23" strokeWidth="4" />
      <path d="M52.5 25c2.6 4.6 2.6 9.4 0 14" strokeWidth="4" />
      <path d="M59 20.5c4.4 7.6 4.4 15.4 0 23" strokeWidth="4" />
    </svg>
  );
}

/** Mark plus wordmark; the name comes from `pager.brand.name` in each language. */
export function RenniliLogo({
  size = 28,
  wiggle = false,
}: {
  size?: number;
  wiggle?: boolean;
}): React.JSX.Element {
  const { t } = useTranslation('pager');
  return (
    <span className="rennili-logo" style={{ fontSize: Math.round(size * 0.78) }}>
      <RenniliMark size={size} wiggle={wiggle} />
      <span className="rennili-wordmark">{t('pager.brand.name')}</span>
    </span>
  );
}

/** The small "by RTYLR" badge on the customer pages and the landing. */
export function ByRtylr(): React.JSX.Element {
  const { t } = useTranslation('pager');
  return (
    <a className="rennili-by" href={APP_URLS.landing} target="_blank" rel="noreferrer">
      {t('pager.brand.by')}
    </a>
  );
}
