import React from 'react';
import { Link } from '@tanstack/react-router';
import { ArrowLeft } from 'lucide-react';

import Header from '../shared/Header';
import type { JobProps } from '../resolver';

const BoldJob: React.FC<JobProps> = ({ career, customization, slug }) => {
  return (
    <div className="min-h-screen bg-white">
      <Header customization={customization} compact />

      <section style={{ background: 'var(--recruit-accent)' }}>
        <div className="mx-auto max-w-5xl px-6 py-20 text-white">
          <Link
            to={`/${slug}` as any}
            className="inline-flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.2em] text-white/80 hover:text-white"
          >
            <ArrowLeft className="h-3.5 w-3.5" />
            All roles
          </Link>
          <h1 className="mt-6 text-5xl font-black leading-[1.05] tracking-tight sm:text-7xl">
            {career.title}
          </h1>
          <div className="mt-6 flex flex-wrap items-center gap-x-4 gap-y-2 text-sm font-medium text-white/85">
            {career.department && <span>{career.department}</span>}
            {career.location && <span>· {career.location}</span>}
            {career.workMode && <span>· {career.workMode}</span>}
            {career.seniority && <span>· {career.seniority}</span>}
            {career.employmentType && <span>· {career.employmentType}</span>}
          </div>
        </div>
      </section>

      <main className="mx-auto max-w-3xl px-6 py-16">
        <article
          className="whitespace-pre-wrap text-lg leading-relaxed"
          style={{ color: 'var(--recruit-fg)' }}
        >
          {career.details}
        </article>

        {career.requirements && (
          <section className="mt-12">
            <h2 className="text-2xl font-bold">Requirements</h2>
            <div
              className="mt-4 whitespace-pre-wrap text-base leading-relaxed"
              style={{ color: 'var(--recruit-fg)' }}
            >
              {career.requirements}
            </div>
          </section>
        )}

        <Link
          to={`/${slug}/${career.slug}/apply` as any}
          className="mt-12 inline-flex items-center gap-2 px-8 py-4 text-sm font-bold uppercase tracking-wider text-white transition hover:opacity-90"
          style={{ background: 'var(--recruit-accent)' }}
        >
          Apply now →
        </Link>
      </main>
    </div>
  );
};

export default BoldJob;
