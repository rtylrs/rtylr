import React, { useEffect } from 'react';
import { useParams } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';

import { useBusinessCareers } from '../../shared/hooks/queries/usePublic';
import { applyBrand, resolveIndex } from '../../templates/resolver';
import { applySEO, loadAnalytics } from '../../shared/lib/seo';
import NotFound from './NotFound';

function isNotFoundError(error: Error | null): boolean {
  return typeof error === 'object' && error !== null && 'status' in error && error.status === 404;
}

/**
 * CareersIndex - the list page at recruit.rtylr.com/:slug.
 *
 * Resolves the tenant by slug (the backend does the mapping, we just
 * ask), applies the brand color to the page-wide CSS variable, and
 * then hands off to whichever template the business selected.
 */
const CareersIndex: React.FC = (): React.JSX.Element => {
  const { slug } = useParams({ strict: false });
  const { t } = useTranslation('recruitCommon');
  const { data, isLoading, error } = useBusinessCareers(slug);

  useEffect(() => {
    if (!data?.customization) return;
    applyBrand(data.customization);
    applySEO(data.customization, { pageKind: 'index' });
    loadAnalytics(data.customization);
  }, [data?.customization]);

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div
          className="h-6 w-6 animate-spin rounded-full border-2"
          style={{
            borderColor: 'rgba(0,0,0,0.1)',
            borderTopColor: 'var(--recruit-accent)',
          }}
        />
      </div>
    );
  }

  if (error && !data && isNotFoundError(error)) {
    return <NotFound />;
  }

  if (error && !data) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center gap-4 bg-white p-6 text-center">
        <h2 className="text-lg font-semibold text-gray-900">
          {t('recruitPublic.errors.careersPage')}
        </h2>
        <p className="text-sm text-gray-500">{t('recruitPublic.errors.refreshHint')}</p>
        <button
          type="button"
          onClick={() => window.location.reload()}
          className="rounded-lg bg-gray-900 px-4 py-2 text-sm text-white hover:bg-gray-800"
        >
          {t('recruitPublic.errors.retry')}
        </button>
      </div>
    );
  }

  if (!data) return <NotFound />;

  const Template = resolveIndex(data.customization.template);
  return (
    <Template
      careers={data.careers}
      customization={data.customization}
      sections={data.sections ?? []}
    />
  );
};

export default CareersIndex;
