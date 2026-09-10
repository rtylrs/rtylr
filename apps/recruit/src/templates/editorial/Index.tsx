import React from 'react';
import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { MapPin, Briefcase } from 'lucide-react';
import { EmptyState } from '@voxire/ui';

import Header from '../shared/Header';
import Sections from '../shared/Sections';
import type { IndexProps } from '../resolver';
import { cleanMojibake } from '../../shared/lib/sanitize';

// Editorial - magazine-style. Serif display type, generous whitespace,
// hero panel with the tagline blown up. For businesses that want their
// careers page to read like a feature story rather than a job board.
const EditorialIndex: React.FC<IndexProps> = ({ careers, customization, sections }) => {
  const { t } = useTranslation();
  return (
    <div className="min-h-screen bg-white">
      <Header customization={customization} />
      <main id="main-content">
        <section className="border-b" style={{ borderColor: 'var(--recruit-border)' }}>
          <div className="mx-auto max-w-5xl px-6 py-24 sm:py-32">
            <div
              className="mb-6 inline-block text-xs uppercase tracking-[0.3em]"
              style={{ color: 'var(--recruit-accent)' }}
            >
              Careers · {cleanMojibake(customization.brandName)}
            </div>
            <h1 className="font-display text-5xl font-semibold leading-[1.05] tracking-tight sm:text-7xl">
              {cleanMojibake(customization.tagline) || 'Come build something that matters.'}
            </h1>
            {customization.aboutBlurb && (
              <p
                className="mt-8 max-w-2xl text-lg leading-relaxed"
                style={{ color: 'var(--recruit-muted)' }}
              >
                {cleanMojibake(customization.aboutBlurb)}
              </p>
            )}
            {customization.heroImageUrl && (
              <div className="mt-10 overflow-hidden rounded-lg">
                <img
                  src={customization.heroImageUrl}
                  alt={cleanMojibake(customization.brandName) || 'Careers'}
                  className="h-80 w-full object-cover"
                />
              </div>
            )}
          </div>
        </section>

        <section className="mx-auto max-w-5xl px-6 py-20">
          <h2 className="font-display text-3xl font-semibold tracking-tight sm:text-4xl">
            Open roles
          </h2>
          <div className="mt-10 divide-y" style={{ borderColor: 'var(--recruit-border)' }}>
            {careers.length === 0 ? (
              <EmptyState
                icon={Briefcase}
                title={t('common.emptyStates.publicRolesNone')}
                action={{
                  label: t('common.emptyStates.checkAgain'),
                  onClick: () => window.location.reload(),
                }}
              />
            ) : (
              careers.map((c) => (
                <Link
                  key={c.id}
                  to="/$slug/$careerSlug"
                  params={{ slug: customization.slug, careerSlug: c.slug }}
                  className="group grid grid-cols-1 gap-2 py-8 transition sm:grid-cols-[1fr_auto] sm:items-center sm:gap-8"
                >
                  <div>
                    <div className="font-display text-2xl font-semibold leading-tight transition group-hover:underline">
                      {cleanMojibake(c.title)}
                    </div>
                    {c.department && (
                      <div className="mt-1 text-sm" style={{ color: 'var(--recruit-muted)' }}>
                        {cleanMojibake(c.department)}
                      </div>
                    )}
                  </div>
                  <div
                    className="flex items-center gap-4 text-sm"
                    style={{ color: 'var(--recruit-muted)' }}
                  >
                    {c.location && (
                      <span className="inline-flex items-center gap-1.5">
                        <MapPin className="h-4 w-4" />
                        {cleanMojibake(c.location)}
                      </span>
                    )}
                    {c.workMode && <span>{cleanMojibake(c.workMode)}</span>}
                  </div>
                </Link>
              ))
            )}
          </div>
        </section>
        <Sections sections={sections} />
      </main>
    </div>
  );
};

export default EditorialIndex;
