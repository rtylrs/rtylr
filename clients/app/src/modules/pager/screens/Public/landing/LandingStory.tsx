import React from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from '@tanstack/react-router';
import { ArrowRight, Check, InfinityIcon, Languages, QrCode, Smartphone, X } from 'lucide-react';
import {
  PAGER_HOW_PATH,
  PAGER_RTYLR_PATH,
  PAGER_TRY_PATH,
  pagerClaimUrl,
} from '../../../pagerPaths';
import { PhoneFrame, ScreenCode, ScreenReady, ScreenScan, ScreenWaiting } from './PhoneMockups';

// The landing page tells each part of the story once: the highlights, the steps,
// the comparison, then one band that says it is free and who makes it.

function list(value: unknown): string[] {
  return Array.isArray(value)
    ? value.filter((item): item is string => typeof item === 'string')
    : [];
}

const STEP_SCREENS = [ScreenScan, ScreenCode, ScreenWaiting, ScreenReady] as const;

export const StepsSection = (): React.JSX.Element => {
  const { t } = useTranslation('pager');
  const chapters = list(t('pager.landing.how.chapters', { returnObjects: true }));

  return (
    <section className="rl-section rl-steps" aria-labelledby="rl-steps-heading">
      <div className="rl-wrap">
        <div className="rl-section-head rl-section-head--center">
          <p className="rl-kicker">{t('pager.landing.nav.how')}</p>
          <h2 id="rl-steps-heading" className="rl-section-title">
            {t('pager.landing.how.title')}
          </h2>
          <p className="rl-section-intro">{t('pager.landing.how.intro')}</p>
        </div>

        <ol className="rl-steps-list">
          {STEP_SCREENS.map((Screen, index) => (
            <li key={chapters[index] ?? index} className="rl-step">
              <div className="rl-step-phone">
                <PhoneFrame variant="compact" ready={index === 3}>
                  <Screen />
                </PhoneFrame>
              </div>
              <p className="rl-step-num">
                <span>{String(index + 1).padStart(2, '0')}</span>
                {chapters[index]}
              </p>
              <h3 className="rl-step-title">{t(`pager.landing.how.steps.${index}.title`)}</h3>
              <p className="rl-step-body">{t(`pager.landing.how.steps.${index}.body`)}</p>
            </li>
          ))}
        </ol>

        <p className="rl-steps-more">
          <Link to={PAGER_HOW_PATH} className="rl-text-link">
            {t('pager.landing.hero.secondaryCta')}
            <ArrowRight size={16} strokeWidth={2.4} aria-hidden="true" className="rtl:rotate-180" />
          </Link>
        </p>
      </div>
    </section>
  );
};

// In the same order as pager.landing.highlights.
const HIGHLIGHT_ICONS = [Smartphone, QrCode, InfinityIcon, Languages] as const;

export const HighlightsBand = (): React.JSX.Element => {
  const { t } = useTranslation('pager');
  const highlights = list(t('pager.landing.highlights', { returnObjects: true }));

  return (
    <section className="rl-highlights" aria-label={t('pager.landing.highlightsLabel')}>
      <ul className="rl-wrap rl-highlights-list">
        {highlights.map((label, index) => {
          const Icon = HIGHLIGHT_ICONS[index] ?? Smartphone;
          return (
            <li key={label} className="rl-highlight">
              <span className="rl-highlight-icon" aria-hidden="true">
                <Icon size={20} strokeWidth={2.2} />
              </span>
              <span className="rl-highlight-label">{label}</span>
            </li>
          );
        })}
      </ul>
    </section>
  );
};

export const CompareTable = (): React.JSX.Element => {
  const { t } = useTranslation('pager');
  const oldItems = list(t('pager.landing.compare.oldItems', { returnObjects: true }));
  const newItems = list(t('pager.landing.compare.newItems', { returnObjects: true }));
  // The two lists are written as pairs: each old pain beside what replaces it.
  const rows = newItems.map((next, index) => ({ old: oldItems[index] ?? '', next }));

  return (
    <section className="rl-section rl-section--mist" aria-labelledby="rl-vs-heading">
      <div className="rl-wrap">
        <div className="rl-section-head rl-section-head--center">
          <h2 id="rl-vs-heading" className="rl-section-title">
            {t('pager.landing.compare.title')}
          </h2>
          <p className="rl-section-intro">{t('pager.landing.compare.intro')}</p>
        </div>

        <table className="rl-vs">
          <thead>
            <tr>
              <th scope="col" className="rl-vs-head rl-vs-head--old">
                {t('pager.landing.compare.oldTitle')}
              </th>
              <th scope="col" className="rl-vs-head rl-vs-head--new">
                {t('pager.landing.compare.newTitle')}
              </th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.next}>
                <td className="rl-vs-cell rl-vs-cell--old">
                  <X size={16} strokeWidth={2.4} aria-hidden="true" />
                  <span>{row.old}</span>
                </td>
                <td className="rl-vs-cell rl-vs-cell--new">
                  <Check size={16} strokeWidth={2.6} aria-hidden="true" />
                  <span>{row.next}</span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
};

export const FreeBand = (): React.JSX.Element => {
  const { t } = useTranslation('pager');
  const included = list(t('pager.landing.free.included', { returnObjects: true }));

  return (
    <section className="rl-section" aria-labelledby="rl-free-heading">
      <div className="rl-wrap">
        <div className="rl-free-card">
          <div className="rl-free-main">
            <p className="rl-free-badge">{t('pager.landing.free.badge')}</p>
            <h2 id="rl-free-heading" className="rl-free-title">
              {t('pager.landing.free.title')}
            </h2>
            <p className="rl-free-body">{t('pager.landing.free.whyBody')}</p>
            <ul className="rl-free-included">
              {included.map((item) => (
                <li key={item}>
                  <Check size={15} strokeWidth={2.8} aria-hidden="true" />
                  {item}
                </li>
              ))}
            </ul>
          </div>
          <div className="rl-free-actions">
            <a className="rl-btn rl-btn--on-forest rl-btn--lg" href={pagerClaimUrl('register')}>
              {t('pager.landing.free.primaryCta')}
              <ArrowRight
                size={18}
                strokeWidth={2.4}
                aria-hidden="true"
                className="rtl:rotate-180"
              />
            </a>
            <Link to={PAGER_TRY_PATH} className="rl-btn rl-btn--on-accent rl-btn--lg">
              {t('pager.landing.free.secondaryCta')}
            </Link>
            <p className="rl-free-signin">
              {t('pager.landing.closing.signIn')}{' '}
              <a href={pagerClaimUrl('signIn')}>{t('pager.landing.closing.signInLink')}</a>
            </p>
          </div>
        </div>

        {/* RTYLR speaks in its own colours, as in the strip above the header. */}
        <div className="rl-rtylr-bar">
          <div>
            <p className="rl-rtylr-kicker">{t('pager.landing.rtylr.eyebrow')}</p>
            <p className="rl-rtylr-title">{t('pager.landing.rtylr.teaserTitle')}</p>
          </div>
          <Link to={PAGER_RTYLR_PATH} className="rl-rtylr-link">
            {t('pager.landing.rtylr.teaserCta')}
            <ArrowRight size={16} strokeWidth={2.4} aria-hidden="true" className="rtl:rotate-180" />
          </Link>
        </div>
      </div>
    </section>
  );
};
