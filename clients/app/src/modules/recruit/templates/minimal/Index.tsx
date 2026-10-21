import React from 'react';
import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { ArrowRight, MapPin, Briefcase } from 'lucide-react';
import { EmptyState } from '@voxire/ui';

import Header from '../shared/Header';
import Sections from '../shared/Sections';
import type { IndexProps } from '../resolver';
import { cleanMojibake } from '../../shared/lib/sanitize';

// Minimal - clean single-column list. Calm neutral palette, strong
// hierarchy. Works for every business without customisation; the
// default template choice for new customers.
const MinimalIndex: React.FC<IndexProps> = ({ careers, customization, sections }) => {
  const { t } = useTranslation('recruitCommon');
  return (
    <div className="min-h-screen">
      <Header customization={customization} />
      <main id="main-content" className="mx-auto max-w-3xl px-6 py-16">
        <h1 className="text-3xl font-semibold tracking-tight sm:text-4xl">
          Open roles at {cleanMojibake(customization.brandName)}
        </h1>
        {customization.aboutBlurb && (
          <p className="mt-4 max-w-2xl text-base" style={{ color: 'var(--recruit-muted)' }}>
            {cleanMojibake(customization.aboutBlurb)}
          </p>
        )}
        {customization.heroImageUrl && (
          <div className="mt-6 overflow-hidden rounded-lg">
            <img
              src={customization.heroImageUrl}
              alt={cleanMojibake(customization.brandName) || 'Careers'}
              className="h-64 w-full object-cover"
            />
          </div>
        )}

        <div className="mt-10 space-y-2">
          {careers.length === 0 ? (
            <div className="rounded-lg border" style={{ borderColor: 'var(--recruit-border)' }}>
              <EmptyState
                icon={Briefcase}
                title={t('recruitCommon.emptyStates.publicRolesNone')}
                action={{
                  label: t('recruitCommon.emptyStates.checkAgain'),
                  onClick: () => window.location.reload(),
                }}
              />
            </div>
          ) : (
            careers.map((c) => (
              <Link
                key={c.id}
                to="/$slug/$careerSlug"
                params={{ slug: customization.slug, careerSlug: c.slug }}
                className="group flex items-center justify-between rounded-lg border bg-white px-5 py-4 transition hover:-translate-y-0.5 hover:shadow-sm"
                style={{ borderColor: 'var(--recruit-border)' }}
              >
                <div className="min-w-0">
                  <div className="truncate text-base font-semibold">{cleanMojibake(c.title)}</div>
                  <div
                    className="mt-1 flex items-center gap-4 text-xs"
                    style={{ color: 'var(--recruit-muted)' }}
                  >
                    {c.department && (
                      <span className="inline-flex items-center gap-1">
                        <Briefcase className="h-3 w-3" />
                        {cleanMojibake(c.department)}
                      </span>
                    )}
                    {c.location && (
                      <span className="inline-flex items-center gap-1">
                        <MapPin className="h-3 w-3" />
                        {cleanMojibake(c.location)}
                      </span>
                    )}
                    {c.workMode && <span>{cleanMojibake(c.workMode)}</span>}
                    {c.employmentType && <span>{cleanMojibake(c.employmentType)}</span>}
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
