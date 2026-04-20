import React, { useEffect } from 'react';
import { useParams } from 'react-router-dom';

import { usePublicCareer } from '../../shared/hooks/queries/usePublic';
import { applyBrand, resolveJob } from '../../templates/resolver';
import NotFound from './NotFound';

/**
 * CareerDetail — the single-role page at recruit.rtylr.com/:slug/:careerSlug.
 * Template choice comes from the owning business's customization; the
 * same resolver pattern as CareersIndex keeps the two pages visually
 * consistent.
 */
const CareerDetail: React.FC = (): React.JSX.Element => {
  const { slug = '', careerSlug = '' } = useParams<{ slug: string; careerSlug: string }>();
  const { data, isLoading, error } = usePublicCareer(slug, careerSlug);

  useEffect(() => {
    if (data?.customization) applyBrand(data.customization);
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

  if (error || !data) return <NotFound />;

  const Template = resolveJob(data.customization.template);
  return <Template career={data.career} customization={data.customization} slug={slug} />;
};

export default CareerDetail;
