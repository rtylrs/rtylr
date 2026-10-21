import React from 'react';
import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { ArrowLeft, MapPin, Briefcase, Clock, Award } from 'lucide-react';

import Header from '../shared/Header';
import type { JobProps } from '../resolver';
import { cleanMojibake } from '../../shared/lib/sanitize';
import SemanticContent from '../../shared/components/SemanticContent';

const MinimalJob: React.FC<JobProps> = ({ career, customization, slug }) => {
  const { t } = useTranslation('recruitCommon');
  return (
    <div className="min-h-screen">
      <Header customization={customization} compact />
      <main className="mx-auto max-w-3xl px-6 py-12">
        <Link
          to="/$slug"
          params={{ slug }}
          className="inline-flex min-h-[44px] items-center gap-2 text-sm font-medium hover:underline"
          style={{ color: 'var(--recruit-muted)' }}
        >
          <ArrowLeft className="h-4 w-4" />
          {t('recruitPublic.job.allRoles')}
        </Link>

        <h1 className="mt-6 text-3xl font-semibold tracking-tight sm:text-4xl">
          {cleanMojibake(career.title)}
        </h1>

        <div
          className="mt-4 flex flex-wrap items-center gap-x-5 gap-y-2 text-sm"
          style={{ color: 'var(--recruit-muted)' }}
        >
          {career.department && (
            <span className="inline-flex items-center gap-1.5">
              <Briefcase className="h-4 w-4" />
              {cleanMojibake(career.department)}
            </span>
          )}
          {career.location && (
            <span className="inline-flex items-center gap-1.5">
              <MapPin className="h-4 w-4" />
              {cleanMojibake(career.location)}
            </span>
          )}
          {career.workMode && (
            <span className="inline-flex items-center gap-1.5">
              <Clock className="h-4 w-4" />
              {cleanMojibake(career.workMode)}
            </span>
          )}
          {career.seniority && (
            <span className="inline-flex items-center gap-1.5">
              <Award className="h-4 w-4" />
              {cleanMojibake(career.seniority)}
            </span>
          )}
          {career.employmentType && <span>{cleanMojibake(career.employmentType)}</span>}
        </div>

        <article
          className="mt-10 max-w-none text-sm leading-relaxed"
          style={{ color: 'var(--recruit-fg)' }}
        >
          <SemanticContent content={career.details} />
        </article>

        {career.requirements && (
          <section className="mt-8">
            <h2 className="text-sm font-semibold uppercase tracking-wider">
              {t('recruitPublic.job.requirements')}
            </h2>
            <div className="mt-3 text-sm leading-relaxed" style={{ color: 'var(--recruit-fg)' }}>
              <SemanticContent content={career.requirements} />
            </div>
          </section>
        )}

        {(career.salaryMin > 0 || career.salaryMax > 0) && (
          <section className="mt-8">
            <h2 className="text-sm font-semibold uppercase tracking-wider">
              {t('recruitPublic.job.compensation')}
            </h2>
            <div className="mt-2 text-sm">
              {career.salaryCurrency || '$'}
              {career.salaryMin.toLocaleString()}
              {career.salaryMax > 0 && ` – ${career.salaryMax.toLocaleString()}`}
            </div>
          </section>
        )}

        <Link
          to="/$slug/$careerSlug/apply"
          params={{ slug, careerSlug: career.slug }}
          className="mt-12 inline-flex min-h-[44px] items-center justify-center gap-2 rounded-lg px-6 py-3 text-sm font-semibold text-white transition hover:opacity-90"
          style={{ background: 'var(--recruit-accent)' }}
        >
          {t('recruitPublic.job.apply')}
        </Link>
      </main>
    </div>
  );
};

export default MinimalJob;
