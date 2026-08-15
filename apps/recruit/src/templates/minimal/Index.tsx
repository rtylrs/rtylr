import React from 'react';
import { Link } from '@tanstack/react-router';
import { ArrowRight, MapPin, Briefcase } from 'lucide-react';

import Header from '../shared/Header';
import Sections from '../shared/Sections';
import type { IndexProps } from '../resolver';

// Minimal - clean single-column list. Calm neutral palette, strong
// hierarchy. Works for every business without customisation; the
// default template choice for new customers.
const MinimalIndex: React.FC<IndexProps> = ({ careers, customization, sections }) => {
  return (
    <div className="min-h-screen">
      <Header customization={customization} />
      <main className="mx-auto max-w-3xl px-6 py-16">
        <h1 className="text-3xl font-semibold tracking-tight sm:text-4xl">
          Open roles at {customization.brandName}
        </h1>
        {customization.aboutBlurb && (
          <p className="mt-4 max-w-2xl text-base" style={{ color: 'var(--recruit-muted)' }}>
            {customization.aboutBlurb}
          </p>
        )}
        {customization.heroImageUrl && (
          <div className="mt-6 overflow-hidden rounded-lg">
            <img
              src={customization.heroImageUrl}
              alt={customization.brandName || 'Careers'}
              className="h-64 w-full object-cover"
            />
          </div>
        )}

        <div className="mt-10 space-y-2">
          {careers.length === 0 ? (
            <div
              className="rounded-lg border p-8 text-center text-sm"
              style={{ borderColor: 'var(--recruit-border)', color: 'var(--recruit-muted)' }}
            >
              No open roles right now. Check back soon.
            </div>
          ) : (
            careers.map((c) => (
              <Link
                key={c.id}
                to={`/${customization.slug}/${c.slug}` as any}
                className="group flex items-center justify-between rounded-lg border bg-white px-5 py-4 transition hover:-translate-y-0.5 hover:shadow-sm"
                style={{ borderColor: 'var(--recruit-border)' }}
              >
                <div className="min-w-0">
                  <div className="truncate text-base font-semibold">{c.title}</div>
                  <div
                    className="mt-1 flex items-center gap-4 text-xs"
                    style={{ color: 'var(--recruit-muted)' }}
                  >
                    {c.department && (
                      <span className="inline-flex items-center gap-1">
                        <Briefcase className="h-3 w-3" />
                        {c.department}
                      </span>
                    )}
                    {c.location && (
                      <span className="inline-flex items-center gap-1">
                        <MapPin className="h-3 w-3" />
                        {c.location}
                      </span>
                    )}
                    {c.workMode && <span>{c.workMode}</span>}
                    {c.employmentType && <span>{c.employmentType}</span>}
                  </div>
                </div>
                <ArrowRight
                  className="h-4 w-4 opacity-0 transition group-hover:opacity-100"
                  style={{ color: 'var(--recruit-accent)' }}
                />
              </Link>
            ))
          )}
        </div>
      </main>
      <Sections sections={sections} />
    </div>
  );
};

export default MinimalIndex;
