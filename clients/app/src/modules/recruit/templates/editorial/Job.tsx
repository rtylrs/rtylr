import React from 'react';
import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { ArrowLeft } from 'lucide-react';

import Header from '../shared/Header';
import type { JobProps } from '../resolver';
import { cleanMojibake } from '../../shared/lib/sanitize';
import SemanticContent from '../../shared/components/SemanticContent';

const EditorialJob: React.FC<JobProps> = ({ career, customization, slug }) => {
  const { t } = useTranslation('recruitCommon');
  return (
    <div className="min-h-screen bg-white">
      <Header customization={customization} compact />
      <main className="mx-auto max-w-3xl px-6 py-16">
        <Link
          to="/$slug"
          params={{ slug }}
          className="inline-flex min-h-[44px] items-center gap-2 text-xs uppercase tracking-[0.2em]"
          style={{ color: 'var(--recruit-muted)' }}
        >
          <ArrowLeft className="h-3.5 w-3.5" />
          {t('recruitPublic.job.allRoles')}
        </Link>

        <div
          className="mt-6 text-xs uppercase tracking-[0.3em]"
          style={{ color: 'var(--recruit-accent)' }}
        >
          {cleanMojibake(career.department) || 'Open role'} ·{' '}
          {cleanMojibake(career.location) || 'Remote'}
        </div>
        <h1 className="mt-4 font-display text-5xl font-semibold leading-[1.05] tracking-tight sm:text-6xl">
          {cleanMojibake(career.title)}
        </h1>

        <div
          className="mt-6 flex flex-wrap items-center gap-x-5 gap-y-2 text-sm"
          style={{ color: 'var(--recruit-muted)' }}
        >
          {career.employmentType && <span>{cleanMojibake(career.employmentType)}</span>}
          {career.workMode && <span>· {cleanMojibake(career.workMode)}</span>}
          {career.seniority && <span>· {cleanMojibake(career.seniority)}</span>}
        </div>

        <article className="mt-12 text-lg leading-relaxed" style={{ color: 'var(--recruit-fg)' }}>
          <SemanticContent content={career.details} />
        </article>

        {career.requirements && (
          <section className="mt-12 border-t pt-8" style={{ borderColor: 'var(--recruit-border)' }}>
            <h2 className="font-display text-2xl font-semibold">
              {t('recruitPublic.job.lookingFor')}
            </h2>
            <div className="mt-4 text-base leading-relaxed" style={{ color: 'var(--recruit-fg)' }}>
              <SemanticContent content={career.requirements} />
            </div>
          </section>
        )}

        <div className="mt-16">
          <Link
            to="/$slug/$careerSlug/apply"
            params={{ slug, careerSlug: career.slug }}
            className="inline-flex min-h-[44px] items-center justify-center gap-2 border-2 px-8 py-4 text-sm font-semibold uppercase tracking-wider transition hover:bg-[var(--recruit-accent)] hover:text-white"
            style={{ borderColor: 'var(--recruit-accent)', color: 'var(--recruit-accent)' }}
          >
            {t('recruitPublic.job.apply')}
          </Link>
        </div>
      </main>
    </div>
  );
};

export default EditorialJob;
