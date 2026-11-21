import React from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from '@tanstack/react-router';
import { ArrowRight } from 'lucide-react';
import { PAGER_FAQ_PATH, PAGER_TRY_PATH } from '../../pagerPaths';
import { ClosingCta } from './landing/ClosingCta';
import { FaqSection } from './landing/FaqSection';
import { FreeSection } from './landing/FreeSection';
import { PublicPageShell } from './landing/PublicPageShell';

/** Dedicated SEO page at /faq answering venue owner questions directly. */
const FaqPage = (): React.JSX.Element => {
  const { t } = useTranslation('pager');

  return (
    <PublicPageShell
      path={PAGER_FAQ_PATH}
      title={t('pager.landing.pages.faq.metaTitle')}
      description={t('pager.landing.pages.faq.metaDescription')}
      keywords={t('pager.landing.meta.keywords')}
    >
      <section className="rl-subhero">
        <div className="rl-wrap rl-subhero-inner">
          <p className="rl-kicker">{t('pager.landing.pages.faq.kicker')}</p>
          <h1 className="rl-subhero-title">{t('pager.landing.pages.faq.heading')}</h1>
          <p className="rl-subhero-lead">{t('pager.landing.pages.faq.intro')}</p>
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
          </div>
        </div>
      </section>

      <FaqSection />
      <FreeSection />
      <ClosingCta />
    </PublicPageShell>
  );
};

export default FaqPage;
