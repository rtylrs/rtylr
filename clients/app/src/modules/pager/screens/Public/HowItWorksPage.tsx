import React from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from '@tanstack/react-router';
import { ArrowRight, CheckCircle2, Printer, QrCode, Smartphone } from 'lucide-react';
import { PAGER_HOW_PATH, PAGER_TRY_PATH, pagerClaimUrl } from '../../pagerPaths';
import { ClosingCta } from './landing/ClosingCta';
import { HowItWorks } from './landing/HowItWorks';
import { PublicPageShell } from './landing/PublicPageShell';

const SETUP_ICONS = [Printer, QrCode, Smartphone, CheckCircle2] as const;

interface SetupStep {
  title: string;
  body: string;
}

/** Dedicated SEO page at /how-it-works explaining the full scan-to-ring workflow. */
const HowItWorksPage = (): React.JSX.Element => {
  const { t } = useTranslation('pager');
  const rawSetup = t('pager.landing.pages.how.setupSteps', { returnObjects: true });
  const setupSteps = Array.isArray(rawSetup) ? (rawSetup as SetupStep[]) : [];

  return (
    <PublicPageShell
      path={PAGER_HOW_PATH}
      title={t('pager.landing.pages.how.metaTitle')}
      description={t('pager.landing.pages.how.metaDescription')}
      keywords={t('pager.landing.meta.keywords')}
    >
      <section className="rl-subhero">
        <div className="rl-wrap rl-subhero-inner">
          <p className="rl-kicker">{t('pager.landing.pages.how.kicker')}</p>
          <h1 className="rl-subhero-title">{t('pager.landing.pages.how.heading')}</h1>
          <p className="rl-subhero-lead">{t('pager.landing.pages.how.intro')}</p>
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

      <HowItWorks />

      <section className="rl-section" aria-labelledby="rl-setup-heading">
        <div className="rl-wrap">
          <div className="rl-section-head">
            <p className="rl-kicker">{t('pager.landing.pages.how.setupKicker')}</p>
            <h2 id="rl-setup-heading" className="rl-section-title">
              {t('pager.landing.pages.how.setupTitle')}
            </h2>
            <p className="rl-section-intro">{t('pager.landing.pages.how.setupIntro')}</p>
          </div>

          <ol className="rl-setup-grid">
            {setupSteps.map((step, index) => {
              const Icon = SETUP_ICONS[index % SETUP_ICONS.length];
              return (
                <li key={step.title} className="rl-setup-card">
                  <div className="rl-setup-head">
                    <span className="rl-setup-icon" aria-hidden="true">
                      <Icon size={19} strokeWidth={2} />
                    </span>
                    <span className="rl-setup-num" dir="ltr">
                      0{index + 1}
                    </span>
                  </div>
                  <h3 className="rl-setup-title">{step.title}</h3>
                  <p className="rl-setup-body">{step.body}</p>
                </li>
              );
            })}
          </ol>
        </div>
      </section>

      <ClosingCta />
    </PublicPageShell>
  );
};

export default HowItWorksPage;
