import React from 'react';
import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { ArrowUpRight, Briefcase } from 'lucide-react';
import { EmptyState } from '@voxire/ui';

import Header from '../shared/Header';
import Sections from '../shared/Sections';
import type { IndexProps } from '../resolver';
import { cleanMojibake } from '../../shared/lib/sanitize';

// Bold - saturated accent block, oversized type, aggressive hover.
// For businesses that want the careers page to feel like a campaign
// landing page. Best paired with a strong primary_color.
const BoldIndex: React.FC<IndexProps> = ({ careers, customization, sections }) => {
  const { t } = useTranslation('recruitCommon');
  return (
    <div className="min-h-screen bg-white">
      <Header customization={customization} />
      <main id="main-content">
        <section style={{ background: 'var(--recruit-accent)' }}>
          <div className="mx-auto max-w-6xl px-6 py-28 text-white sm:py-40">
            <div className="mb-6 inline-block border border-white/40 px-3 py-1 text-[10px] uppercase tracking-[0.3em]">
              {cleanMojibake(customization.brandName)} · {t('recruitPublic.index.careersEyebrow')}
            </div>
            <h1 className="whitespace-pre-line text-6xl font-black leading-[0.95] tracking-tight sm:text-8xl">
              {t('recruitPublic.index.heroTitle')}
            </h1>
            {customization.aboutBlurb && (
              <p className="mt-10 max-w-2xl text-xl leading-relaxed text-white/80">
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

        <section className="mx-auto max-w-6xl px-6 py-24">
          <div className="mb-12 flex items-baseline justify-between">
            <h2 className="text-3xl font-bold tracking-tight sm:text-4xl">
              {t('recruitPublic.index.openPositions')}
            </h2>
            <div className="text-sm font-medium" style={{ color: 'var(--recruit-muted)' }}>
              {t('recruitPublic.index.roleCount', { count: careers.length })}
            </div>
          </div>

          <div className="grid gap-0 border-t" style={{ borderColor: 'var(--recruit-border)' }}>
            {careers.length === 0 ? (
              <EmptyState
                icon={Briefcase}
                title={t('recruitCommon.emptyStates.publicRolesNone')}
                action={{
                  label: t('recruitCommon.emptyStates.checkAgain'),
                  onClick: () => window.location.reload(),
                }}
              />
            ) : (
              careers.map((c) => (
                <Link
                  key={c.id}
                  to="/$slug/$careerSlug"
                  params={{ slug: customization.slug, careerSlug: c.slug }}
                  className="group flex items-center justify-between border-b py-8 transition hover:pl-4"
                  style={{ borderColor: 'var(--recruit-border)' }}
                >
                  <div>
                    <div className="text-2xl font-bold leading-tight sm:text-3xl">
                      {cleanMojibake(c.title)}
                    </div>
                    <div
                      className="mt-2 flex items-center gap-3 text-sm font-medium"
                      style={{ color: 'var(--recruit-muted)' }}
                    >
                      {c.department && <span>{cleanMojibake(c.department)}</span>}
                      {c.department && c.location && <span>·</span>}
                      {c.location && <span>{cleanMojibake(c.location)}</span>}
                      {c.workMode && <span>· {cleanMojibake(c.workMode)}</span>}
                    </div>
                  </div>
                  <ArrowUpRight
                    className="h-6 w-6 transition group-hover:rotate-12"
                    style={{ color: 'var(--recruit-accent)' }}
                  />
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

export default BoldIndex;
