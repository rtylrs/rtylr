import React from 'react';
import { Link } from '@tanstack/react-router';
import { ArrowLeft, MapPin, Briefcase, Clock, Award } from 'lucide-react';

import Header from '../shared/Header';
import type { JobProps } from '../resolver';

const MinimalJob: React.FC<JobProps> = ({ career, customization, slug }) => {
  return (
    <div className="min-h-screen">
      <Header customization={customization} compact />
      <main className="mx-auto max-w-3xl px-6 py-12">
        <Link
          to={`/${slug}` as any}
          className="inline-flex items-center gap-2 text-sm font-medium hover:underline"
          style={{ color: 'var(--recruit-muted)' }}
        >
          <ArrowLeft className="h-4 w-4" />
          All roles
        </Link>

        <h1 className="mt-6 text-3xl font-semibold tracking-tight sm:text-4xl">{career.title}</h1>

        <div
          className="mt-4 flex flex-wrap items-center gap-x-5 gap-y-2 text-sm"
          style={{ color: 'var(--recruit-muted)' }}
        >
          {career.department && (
            <span className="inline-flex items-center gap-1.5">
              <Briefcase className="h-4 w-4" />
              {career.department}
            </span>
          )}
          {career.location && (
            <span className="inline-flex items-center gap-1.5">
              <MapPin className="h-4 w-4" />
              {career.location}
            </span>
          )}
          {career.workMode && (
            <span className="inline-flex items-center gap-1.5">
              <Clock className="h-4 w-4" />
              {career.workMode}
            </span>
          )}
          {career.seniority && (
            <span className="inline-flex items-center gap-1.5">
              <Award className="h-4 w-4" />
              {career.seniority}
            </span>
          )}
          {career.employmentType && <span>{career.employmentType}</span>}
        </div>

        <article
          className="prose prose-sm mt-10 max-w-none whitespace-pre-wrap leading-relaxed"
          style={{ color: 'var(--recruit-fg)' }}
        >
          {career.details}
        </article>

        {career.requirements && (
          <section className="mt-8">
            <h2 className="text-sm font-semibold uppercase tracking-wider">Requirements</h2>
            <div
              className="mt-3 whitespace-pre-wrap text-sm leading-relaxed"
              style={{ color: 'var(--recruit-fg)' }}
            >
              {career.requirements}
            </div>
          </section>
        )}

        {(career.salaryMin > 0 || career.salaryMax > 0) && (
          <section className="mt-8">
            <h2 className="text-sm font-semibold uppercase tracking-wider">Compensation</h2>
            <div className="mt-2 text-sm">
              {career.salaryCurrency || '$'}
              {career.salaryMin.toLocaleString()}
              {career.salaryMax > 0 && ` – ${career.salaryMax.toLocaleString()}`}
            </div>
          </section>
        )}

        <Link
          to={`/${slug}/${career.slug}/apply` as any}
          className="mt-12 inline-flex items-center gap-2 rounded-lg px-6 py-3 text-sm font-semibold text-white transition hover:opacity-90"
          style={{ background: 'var(--recruit-accent)' }}
        >
          Apply for this role
        </Link>
      </main>
    </div>
  );
};

export default MinimalJob;
