import React from 'react';
import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';

/**
 * Stable deep-link state for routes that are intentionally hidden by a flag.
 * Keeping the current URL makes a bookmark explainable instead of silently
 * sending the user somewhere else.
 */
export function PreviewRouteUnavailable(): React.JSX.Element {
  const { t } = useTranslation();
  return (
    <div
      role="status"
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        minHeight: '320px',
        padding: '32px',
        textAlign: 'center',
        color: 'var(--flow-text)',
      }}
    >
      <h1 style={{ margin: 0, fontSize: 'var(--text-screen)', fontWeight: 600 }}>
        {t('flow.previewRoute.title')}
      </h1>
      <p style={{ maxWidth: '420px', margin: '12px 0 20px', color: 'var(--flow-text-muted)' }}>
        {t('flow.previewRoute.body')}
      </p>
      <Link
        to="/my-issues"
        style={{ color: 'var(--accent)', fontWeight: 600, textDecoration: 'none' }}
      >
        {t('flow.previewRoute.back')}
      </Link>
    </div>
  );
}

export default PreviewRouteUnavailable;
