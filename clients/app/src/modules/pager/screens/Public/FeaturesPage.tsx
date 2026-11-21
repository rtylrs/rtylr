import React from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from '@tanstack/react-router';
import { ArrowRight } from 'lucide-react';
import { PAGER_FEATURES_PATH, PAGER_TRY_PATH, pagerClaimUrl } from '../../pagerPaths';
import { BentoShowcase } from './landing/BentoShowcase';
import { ClosingCta } from './landing/ClosingCta';
import { CompareSection } from './landing/CompareSection';
import { FreeSection } from './landing/FreeSection';
import { PublicPageShell } from './landing/PublicPageShell';
import { VenuesSection } from './landing/VenuesSection';

/** Dedicated SEO page at /features comparing Rennili against legacy plastic buzzers. */
const FeaturesPage = (): React.JSX.Element => {
  const { t } = useTranslation('pager');

  return (
    <PublicPageShell
      path={PAGER_FEATURES_PATH}
      title={t('pager.landing.pages.features.metaTitle')}
      description={t('pager.landing.pages.features.metaDescription')}
      keywords={t('pager.landing.meta.keywords')}
    >
      <section className="rl-subhero">
        <div className="rl-wrap rl-subhero-inner">
          <p className="rl-kicker">{t('pager.landing.pages.features.kicker')}</p>
          <h1 className="rl-subhero-title">{t('pager.landing.pages.features.heading')}</h1>
          <p className="rl-subhero-lead">{t('pager.landing.pages.features.intro')}</p>
          <div className="rl-subhero-actions">
            <Link to={PAGER_TRY_PATH} className="rl-btn rl-btn--primary rl-btn--lg">
              {t('pager.landing.hero.primaryCta')}
              <ArrowRight
                size={18}
                strokeWidth={2.4}
                aria-hidden="true"
                className="rtl:rotate-180"
              />
            </Link>
            <a href={pagerClaimUrl('register')} className="rl-btn rl-btn--secondary rl-btn--lg">
              {t('pager.landing.free.primaryCta')}
            </a>
          </div>
        </div>
      </section>

      <BentoShowcase />
      <CompareSection />
      <VenuesSection />
      <FreeSection />
      <ClosingCta />
    </PublicPageShell>
  );
};

export default FeaturesPage;
