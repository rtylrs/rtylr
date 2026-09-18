import React from 'react';
import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { ArrowLeft } from 'lucide-react';

import Header from '../shared/Header';
import type { JobProps } from '../resolver';
import { cleanMojibake } from '../../shared/lib/sanitize';
import SemanticContent from '../../shared/components/SemanticContent';

const BoldJob: React.FC<JobProps> = ({ career, customization, slug }) => {
  const { t } = useTranslation();
  return (
    <div className="min-h-screen bg-white">
      <Header customization={customization} compact />

      <section style={{ background: 'var(--recruit-accent)' }}>
        <div className="mx-auto max-w-5xl px-6 py-20 text-white">
          <Link
            to="/$slug"
            params={{ slug }}
            className="inline-flex min-h-[44px] items-center gap-2 text-xs font-semibold uppercase tracking-[0.2em] text-white/80 hover:text-white"
          >
            <ArrowLeft className="h-3.5 w-3.5" />
            {t('public.job.allRoles')}
          </Link>
          <h1 className="mt-6 text-5xl font-black leading-[1.05] tracking-tight sm:text-7xl">
            {cleanMojibake(career.title)}
          </h1>
          <div className="mt-6 flex flex-wrap items-center gap-x-4 gap-y-2 text-sm font-medium text-white/85">
            {career.department && <span>{cleanMojibake(career.department)}</span>}
            {career.location && <span>· {cleanMojibake(career.location)}</span>}
            {career.workMode && <span>· {cleanMojibake(career.workMode)}</span>}
            {career.seniority && <span>· {cleanMojibake(career.seniority)}</span>}
            {career.employmentType && <span>· {cleanMojibake(career.employmentType)}</span>}
          </div>
        </div>
      </section>

      <main className="mx-auto max-w-3xl px-6 py-16">
        <article className="text-lg leading-relaxed" style={{ color: 'var(--recruit-fg)' }}>
          <SemanticContent content={career.details} />
        </article>

        {career.requirements && (
          <section className="mt-12">
            <h2 className="text-2xl font-bold">{t('public.job.requirements')}</h2>
            <div className="mt-4 text-base leading-relaxed" style={{ color: 'var(--recruit-fg)' }}>
              <SemanticContent content={career.requirements} />
            </div>
          </section>
        )}

        <Link
          to="/$slug/$careerSlug/apply"
          params={{ slug, careerSlug: career.slug }}
          className="mt-12 inline-flex min-h-[44px] items-center justify-center gap-2 px-8 py-4 text-sm font-bold uppercase tracking-wider text-white transition hover:opacity-90"
          style={{ background: 'var(--recruit-accent)' }}
        >
          Apply now →
        </Link>
      </main>
    </div>
  );
};

export default BoldJob;
