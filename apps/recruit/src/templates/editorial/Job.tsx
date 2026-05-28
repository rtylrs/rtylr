import React from 'react';
import { Link } from '@tanstack/react-router';
import { ArrowLeft } from 'lucide-react';

import Header from '../shared/Header';
import type { JobProps } from '../resolver';

const EditorialJob: React.FC<JobProps> = ({ career, customization, slug }) => {
  return (
    <div className="min-h-screen bg-white">
      <Header customization={customization} compact />
      <main className="mx-auto max-w-3xl px-6 py-16">
        <Link
          to={`/${slug}` as any}
          className="inline-flex items-center gap-2 text-xs uppercase tracking-[0.2em]"
          style={{ color: 'var(--recruit-muted)' }}
        >
          <ArrowLeft className="h-3.5 w-3.5" />
          All roles
        </Link>

        <div
          className="mt-10 text-xs uppercase tracking-[0.3em]"
          style={{ color: 'var(--recruit-accent)' }}
        >
          {career.department || 'Open role'} · {career.location || 'Remote'}
        </div>
        <h1 className="mt-4 font-display text-5xl font-semibold leading-[1.05] tracking-tight sm:text-6xl">
          {career.title}
        </h1>

        <div
          className="mt-6 flex flex-wrap items-center gap-x-5 gap-y-2 text-sm"
          style={{ color: 'var(--recruit-muted)' }}
        >
          {career.employmentType && <span>{career.employmentType}</span>}
          {career.workMode && <span>· {career.workMode}</span>}
          {career.seniority && <span>· {career.seniority}</span>}
        </div>

        <article
          className="mt-12 whitespace-pre-wrap text-lg leading-relaxed"
          style={{ color: 'var(--recruit-fg)' }}
        >
          {career.details}
        </article>

        {career.requirements && (
          <section className="mt-12 border-t pt-8" style={{ borderColor: 'var(--recruit-border)' }}>
            <h2 className="font-display text-2xl font-semibold">What we're looking for</h2>
            <div
              className="mt-4 whitespace-pre-wrap text-base leading-relaxed"
              style={{ color: 'var(--recruit-fg)' }}
            >
              {career.requirements}
            </div>
          </section>
        )}

        <div className="mt-16">
          <Link
            to={`/${slug}/${career.slug}/apply` as any}
            className="inline-flex items-center gap-2 border-2 px-8 py-4 text-sm font-semibold uppercase tracking-wider transition hover:bg-[var(--recruit-accent)] hover:text-white"
            style={{ borderColor: 'var(--recruit-accent)', color: 'var(--recruit-accent)' }}
          >
            Apply for this role
          </Link>
        </div>
      </main>
    </div>
  );
};

export default EditorialJob;
