import React, { useRef } from 'react';
import { useTranslation } from 'react-i18next';
import { Check, Minus } from 'lucide-react';
import { RenniliMark } from '../../../shared/components/RenniliLogo';
import { PhonePagerSvg, PlasticPagerSvg } from './PagerIllustration';
import { useInView } from './useLandingObservers';

export const CompareSection = (): React.JSX.Element => {
  const { t } = useTranslation('pager');
  const compareRef = useRef<HTMLDivElement>(null);
  const compareVisible = useInView(compareRef, { once: true, threshold: 0.12 });

  const rawOld = t('pager.landing.compare.oldItems', { returnObjects: true });
  const oldItems = Array.isArray(rawOld) ? (rawOld as string[]) : [];
  const rawNew = t('pager.landing.compare.newItems', { returnObjects: true });
  const newItems = Array.isArray(rawNew) ? (rawNew as string[]) : [];

  return (
    <section id="why" className="rl-section" aria-labelledby="rl-compare-heading">
      <div className="rl-wrap">
        <div className="rl-section-head">
          <h2 id="rl-compare-heading" className="rl-section-title">
            {t('pager.landing.compare.title')}
          </h2>
          <p className="rl-section-intro">{t('pager.landing.compare.intro')}</p>
        </div>

        <div
          ref={compareRef}
          className={`rl-compare-ledger ${compareVisible ? 'rl-compare-ledger--in' : ''}`}
        >
          <div className="rl-ledger-col rl-ledger-col--old">
            <div className="rl-ledger-head">
              <PlasticPagerSvg className="rl-ledger-svg" />
              <span className="rl-ledger-badge rl-ledger-badge--old">
                {t('pager.landing.compare.oldBadge')}
              </span>
              <h3 className="rl-ledger-title">{t('pager.landing.compare.oldTitle')}</h3>
            </div>
            <ul className="rl-ledger-list">
              {oldItems.map((line) => (
                <li key={line} className="rl-ledger-row rl-ledger-row--old">
                  <span className="rl-ledger-icon rl-ledger-icon--old" aria-hidden="true">
                    <Minus size={14} strokeWidth={2.6} />
                  </span>
                  <span>{line}</span>
                </li>
              ))}
            </ul>
          </div>

          <div className="rl-ledger-col rl-ledger-col--new">
            <div className="rl-ledger-head">
              <PhonePagerSvg className="rl-ledger-svg" />
              <span className="rl-ledger-badge rl-ledger-badge--new">
                <RenniliMark size={15} />
                {t('pager.landing.compare.newBadge')}
              </span>
              <h3 className="rl-ledger-title rl-ledger-title--new">
                {t('pager.landing.compare.newTitle')}
              </h3>
            </div>
            <ul className="rl-ledger-list">
              {newItems.map((line, index) => (
                <li
                  key={line}
                  className="rl-ledger-row rl-ledger-row--new"
                  style={{ transitionDelay: `${index * 55}ms` }}
                >
                  <span className="rl-ledger-icon rl-ledger-icon--new" aria-hidden="true">
                    <Check size={14} strokeWidth={3} />
                  </span>
                  <span>{line}</span>
                </li>
              ))}
            </ul>
          </div>
        </div>
      </div>
    </section>
  );
};
