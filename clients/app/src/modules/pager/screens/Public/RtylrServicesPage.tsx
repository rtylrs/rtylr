import React from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from '@tanstack/react-router';
import { APP_URLS } from '@voxire/config';
import { motion } from 'motion/react';
import { ArrowRight, ArrowUpRight, Check } from 'lucide-react';
import { PAGER_RTYLR_PATH, PAGER_TRY_PATH } from '../../pagerPaths';
import { RtylrBrandText } from '../../shared/components/RenniliLogo';
import { ClosingCta } from './landing/ClosingCta';
import { PublicPageShell } from './landing/PublicPageShell';
import { RTYLR_SERVICES } from './landing/RtylrUpsellBand';

/**
 * Dedicated RTYLR services & upsell page at /rtylr showing the broader RTYLR
 * platform behind Rennili (POS, QR Menu, Inventory, Finance, HR, CRM, Flow, Insights).
 */
const RtylrServicesPage = (): React.JSX.Element => {
  const { t } = useTranslation('pager');
  const rawBenefits = t('pager.landing.rtylr.benefits', { returnObjects: true });
  const benefits = Array.isArray(rawBenefits) ? (rawBenefits as string[]) : [];

  return (
    <PublicPageShell
      path={PAGER_RTYLR_PATH}
      title={t('pager.landing.rtylr.metaTitle')}
      description={t('pager.landing.rtylr.metaDescription')}
      keywords={t('pager.landing.rtylr.metaKeywords')}
    >
      <section className="rl-subhero">
        <div className="rl-wrap rl-subhero-inner">
          <p className="rl-kicker">
            <RtylrBrandText text={t('pager.landing.rtylr.eyebrow')} />
          </p>
          <h1 className="rl-subhero-title">
            <RtylrBrandText text={t('pager.landing.rtylr.heading')} />
          </h1>
          <p className="rl-subhero-lead">
            <RtylrBrandText text={t('pager.landing.rtylr.intro')} />
          </p>
          <div className="rl-subhero-actions">
            <a
              href={`${APP_URLS.landing}/workspace/`}
              target="_blank"
              rel="noreferrer"
              className="rl-btn rl-btn--primary rl-btn--lg"
            >
              <RtylrBrandText text={t('pager.landing.rtylr.visitWorkspace')} />
              <ArrowUpRight size={18} strokeWidth={2.3} aria-hidden="true" />
            </a>
            <a
              href={`${APP_URLS.landing}/demo/`}
              target="_blank"
              rel="noreferrer"
              className="rl-btn rl-btn--secondary rl-btn--lg"
            >
              <RtylrBrandText text={t('pager.landing.rtylr.bookDemo')} />
            </a>
          </div>
        </div>
      </section>

      <section className="rl-section rl-section--mist" aria-labelledby="rl-services-grid-heading">
        <div className="rl-wrap">
          <div className="rl-section-head">
            <p className="rl-kicker">
              <RtylrBrandText text={t('pager.landing.rtylr.suiteKicker')} />
            </p>
            <h2 id="rl-services-grid-heading" className="rl-section-title">
              {t('pager.landing.rtylr.suiteTitle')}
            </h2>
            <p className="rl-section-intro">{t('pager.landing.rtylr.suiteIntro')}</p>
          </div>

          <div className="rl-services-grid">
            {RTYLR_SERVICES.map((service, index) => {
              const Icon = service.icon;
              return (
                <motion.a
                  key={service.id}
                  href={service.href}
                  target="_blank"
                  rel="noreferrer"
                  className="rl-service-card"
                  initial={{ opacity: 0, y: 14 }}
                  whileInView={{ opacity: 1, y: 0 }}
                  viewport={{ once: true, amount: 0.15 }}
                  transition={{
                    duration: 0.4,
                    delay: Math.min(index * 0.04, 0.24),
                    ease: [0.22, 1, 0.36, 1],
                  }}
                >
                  <div className="rl-service-top">
                    <span className="rl-service-icon" aria-hidden="true">
                      <Icon size={20} strokeWidth={1.9} />
                    </span>
                    <ArrowUpRight
                      size={17}
                      strokeWidth={2}
                      className="rl-service-arrow"
                      aria-hidden="true"
                    />
                  </div>
                  <h3 className="rl-service-title">
                    {t(`pager.landing.rtylr.services.${service.id}.title`)}
                  </h3>
                  <p className="rl-service-body">
                    {t(`pager.landing.rtylr.services.${service.id}.body`)}
                  </p>
                </motion.a>
              );
            })}
          </div>
        </div>
      </section>

      <section className="rl-section" aria-labelledby="rl-rtylr-bridge-heading">
        <div className="rl-wrap">
          <div className="rl-free-card">
            <div className="rl-free-grid">
              <div className="rl-free-display">
                <span className="rl-kicker">{t('pager.landing.rtylr.bridgeKicker')}</span>
                <p className="rl-free-figure">{t('pager.brand.name')}</p>
                <p className="rl-free-figure-sub">{t('pager.landing.rtylr.bridgeSub')}</p>
              </div>
              <div className="rl-free-manifesto">
                <h2 id="rl-rtylr-bridge-heading" className="rl-free-heading">
                  {t('pager.landing.rtylr.bridgeTitle')}
                </h2>
                <p className="rl-free-why-body">
                  <RtylrBrandText text={t('pager.landing.rtylr.bridgeBody')} />
                </p>
                <div className="rl-free-actions">
                  <Link to={PAGER_TRY_PATH} className="rl-btn rl-btn--primary rl-btn--lg">
                    {t('pager.landing.hero.primaryCta')}
                    <ArrowRight
                      size={18}
                      strokeWidth={2.4}
                      aria-hidden="true"
                      className="rtl:rotate-180"
                    />
                  </Link>
                  <a
                    href={`${APP_URLS.landing}/hardware/`}
                    target="_blank"
                    rel="noreferrer"
                    className="rl-btn rl-btn--secondary rl-btn--lg"
                  >
                    {t('pager.landing.rtylr.hardwareCta')}
                  </a>
                </div>
              </div>
            </div>

            <ul className="rl-free-specs">
              {benefits.map((item) => (
                <li key={item} className="rl-free-spec">
                  <Check size={15} strokeWidth={2.8} aria-hidden="true" />
                  <span>
                    <RtylrBrandText text={item} />
                  </span>
                </li>
              ))}
            </ul>
          </div>
        </div>
      </section>

      <ClosingCta />
    </PublicPageShell>
  );
};

export default RtylrServicesPage;
