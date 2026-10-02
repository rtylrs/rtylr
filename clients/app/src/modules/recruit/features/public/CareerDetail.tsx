import React, { useEffect } from 'react';
import { useParams } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';

import { usePublicCareer } from '../../shared/hooks/queries/usePublic';
import { applyBrand, resolveJob } from '../../templates/resolver';
import { applySEO, loadAnalytics, emitJobPostingJSONLD } from '../../shared/lib/seo';
import NotFound from './NotFound';

/**
 * CareerDetail - the single-role page at recruit.rtylr.com/:slug/:careerSlug.
 * Template choice comes from the owning business's customization; the
 * same resolver pattern as CareersIndex keeps the two pages visually
 * consistent.
 */
const CareerDetail: React.FC = (): React.JSX.Element => {
  const { slug = '', careerSlug = '' } = useParams({ strict: false }) as {
    slug?: string;
    careerSlug?: string;
  };
  const { t } = useTranslation('recruitCommon');
  const { data, isLoading, error } = usePublicCareer(slug, careerSlug);

  useEffect(() => {
    if (!data) return;
    applyBrand(data.customization);
    applySEO(data.customization, { pageKind: 'career', careerTitle: data.career.title });
    loadAnalytics(data.customization);
    emitJobPostingJSONLD(data.career, data.customization);
  }, [data]);

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

  if (error && !data) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center gap-4 bg-white p-6 text-center">
        <h2 className="text-lg font-semibold text-gray-900">
          {t('recruitPublic.errors.jobDetails')}
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

  const Template = resolveJob(data.customization.template);
  return <Template career={data.career} customization={data.customization} slug={slug} />;
};

export default CareerDetail;
