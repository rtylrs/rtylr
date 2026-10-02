import React from 'react';
import { useTranslation } from 'react-i18next';
import { resolveAppUrl } from '@voxire/config';
import { HelmetContainer } from '@voxire/components';

/**
 * NotFound - rendered when a slug or career slug doesn't resolve.
 * Deliberately neutral (no brand context to draw on) and gently
 * points back at the marketing site rather than stranding the visitor.
 */
const NotFound: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation('recruitCommon');
  return (
    <>
      <HelmetContainer title={t('recruitPublic.notFound.metaTitle')} />
      <div className="flex min-h-screen flex-col items-center justify-center px-6 text-center">
        <div
          className="text-xs uppercase tracking-[0.3em]"
          style={{ color: 'var(--recruit-muted)' }}
        >
          {t('recruitPublic.notFound.eyebrow')}
        </div>
        <h1 className="mt-4 text-4xl font-semibold tracking-tight">
          {t('recruitPublic.notFound.title')}
        </h1>
        <p className="mt-3 max-w-md text-sm" style={{ color: 'var(--recruit-muted)' }}>
          {t('recruitPublic.notFound.body')}
        </p>
        <a
          href={resolveAppUrl('landing')}
          className="mt-8 inline-flex items-center gap-2 rounded-lg border px-5 py-2.5 text-sm font-medium"
          style={{ borderColor: 'var(--recruit-border)' }}
        >
          {t('recruitPublic.notFound.back')}
        </a>
      </div>
    </>
  );
};

export default NotFound;
