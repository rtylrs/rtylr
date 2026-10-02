import React from 'react';
import { useTranslation } from 'react-i18next';
import { resolveAppUrl, resolveAuthUrl } from '@voxire/config';
import { ArrowRight } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';

/**
 * RootLanding - what a candidate sees if they hit recruit.rtylr.com
 * with no business slug. Almost nobody reaches this in practice (the
 * entry is always a direct link to a business's page), but we render
 * a tight marketing blurb rather than a 404 so crawlers get something
 * indexable.
 */
const RootLanding: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation('recruitCommon');
  return (
    <>
      <HelmetContainer title={t('recruitPublic.landing.metaTitle')} />
      <main
        id="main-content"
        className="flex min-h-screen flex-col items-center justify-center px-6 text-center"
      >
        <div
          className="text-xs uppercase tracking-[0.3em]"
          style={{ color: 'var(--recruit-muted)' }}
        >
          rtylr · recruit
        </div>
        <h1 className="mt-4 max-w-3xl whitespace-pre-line text-5xl font-semibold leading-tight tracking-tight sm:text-6xl">
          {t('recruitPublic.landing.title')}
        </h1>
        <p className="mt-6 max-w-xl text-base" style={{ color: 'var(--recruit-muted)' }}>
          {t('recruitPublic.landing.body')}
        </p>
        <div className="mt-10 flex items-center gap-3">
          <a
            href={`${resolveAppUrl('landing')}/pricing`}
            className="inline-flex min-h-[44px] items-center gap-2 rounded-lg px-6 py-3 text-sm font-semibold text-white transition hover:opacity-90"
            style={{ background: 'var(--recruit-accent)' }}
          >
            {t('recruitPublic.landing.pricing')}
            <ArrowRight className="h-4 w-4" />
          </a>
          <a
            href={`${resolveAuthUrl()}?redirect=recruit`}
            className="inline-flex min-h-[44px] items-center gap-2 rounded-lg border px-6 py-3 text-sm font-semibold"
            style={{ borderColor: 'var(--recruit-border)' }}
          >
            {t('recruitPublic.landing.startFree')}
          </a>
        </div>
      </main>
    </>
  );
};

export default RootLanding;
