import React from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from '@tanstack/react-router';
import { ArrowRight } from 'lucide-react';
import { PAGER_TRY_PATH, pagerClaimUrl } from '../../../pagerPaths';

export const ClosingCta = (): React.JSX.Element => {
  const { t } = useTranslation('pager');

  return (
    <section className="rl-closing" aria-labelledby="rl-closing-heading">
      <div className="rl-wrap">
        <div className="rl-closing-card">
          <div className="rl-closing-rings" aria-hidden="true">
            <span className="rl-closing-ring rl-closing-ring--1" />
            <span className="rl-closing-ring rl-closing-ring--2" />
            <span className="rl-closing-ring rl-closing-ring--3" />
          </div>

          <div className="rl-closing-inner">
            <h2 id="rl-closing-heading" className="rl-closing-title">
              {t('pager.landing.closing.title')}
            </h2>
            <p className="rl-closing-sub">{t('pager.landing.closing.subtitle')}</p>
            <div className="rl-closing-actions">
              <Link to={PAGER_TRY_PATH} className="rl-btn rl-btn--on-forest rl-btn--lg">
                {t('pager.landing.closing.primaryCta')}
                <ArrowRight
                  size={18}
                  strokeWidth={2.4}
                  aria-hidden="true"
                  className="rtl:rotate-180"
                />
              </Link>
              <a href={pagerClaimUrl('register')} className="rl-btn rl-btn--on-accent rl-btn--lg">
                {t('pager.landing.closing.secondaryCta')}
              </a>
            </div>
            <p className="rl-closing-signin">
              {t('pager.landing.closing.signIn')}{' '}
              <a href={pagerClaimUrl('signIn')} className="rl-closing-signin-link">
                {t('pager.landing.closing.signInLink')}
              </a>
            </p>
          </div>
        </div>
      </div>
    </section>
  );
};
