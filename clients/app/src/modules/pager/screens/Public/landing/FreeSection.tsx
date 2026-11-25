import React from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from '@tanstack/react-router';
import { ArrowRight, Check } from 'lucide-react';
import { PAGER_TRY_PATH, pagerClaimUrl } from '../../../pagerPaths';

export const FreeSection = (): React.JSX.Element => {
  const { t } = useTranslation('pager');
  const rawIncluded = t('pager.landing.free.included', { returnObjects: true });
  const included = Array.isArray(rawIncluded) ? (rawIncluded as string[]) : [];

  return (
    <section id="free" className="rl-section" aria-labelledby="rl-free-heading">
      <div className="rl-wrap">
        <div className="rl-free-card">
          <div className="rl-free-grid">
            <div className="rl-free-display">
              <span className="rl-free-pill">{t('pager.landing.free.badge')}</span>
              <p className="rl-free-figure">{t('pager.landing.free.figure')}</p>
              <p className="rl-free-figure-sub">{t('pager.landing.free.figureSub')}</p>
            </div>

            <div className="rl-free-manifesto">
              <h2 id="rl-free-heading" className="rl-free-heading">
                {t('pager.landing.free.title')}
              </h2>
              <div className="rl-free-why">
                <h3 className="rl-free-why-title">{t('pager.landing.free.whyTitle')}</h3>
                <p className="rl-free-why-body">{t('pager.landing.free.whyBody')}</p>
              </div>

              <div className="rl-free-actions">
                <a href={pagerClaimUrl('register')} className="rl-btn rl-btn--primary rl-btn--lg">
                  {t('pager.landing.free.primaryCta')}
                  <ArrowRight
                    size={18}
                    strokeWidth={2.4}
                    aria-hidden="true"
                    className="rtl:rotate-180"
                  />
                </a>
                <Link to={PAGER_TRY_PATH} className="rl-btn rl-btn--secondary rl-btn--lg">
                  {t('pager.landing.free.secondaryCta')}
                </Link>
              </div>
            </div>
          </div>

          <ul className="rl-free-specs">
            {included.map((item) => (
              <li key={item} className="rl-free-spec">
                <Check size={15} strokeWidth={2.8} aria-hidden="true" />
                <span>{item}</span>
              </li>
            ))}
          </ul>
        </div>
      </div>
    </section>
  );
};
