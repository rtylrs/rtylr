import React from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { RENNILI_URL } from '@voxire/config';
import {
  ArrowRight,
  BatteryCharging,
  BellRing,
  QrCode,
  SearchX,
  Smartphone,
  SprayCan,
  Wallet,
} from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import LanguageSwitch from '../../shared/components/LanguageSwitch';
import { ByRtylr, RenniliLogo, RenniliMark } from '../../shared/components/RenniliLogo';
import { PAGER_TRY_PATH, pagerClaimUrl } from '../../pagerPaths';

interface Card {
  title: string;
  body: string;
}

// Icons only; the copy resolves from landing.how.steps and
// landing.compare.items, matched by position.
const STEP_ICONS: LucideIcon[] = [QrCode, Smartphone, BellRing];
const COMPARE_ICONS: LucideIcon[] = [Wallet, BatteryCharging, SprayCan, SearchX];

function cardsFor(value: unknown): Card[] {
  return Array.isArray(value) ? (value as Card[]) : [];
}

const TryLink = (): React.JSX.Element => {
  const { t } = useTranslation('pager');
  return (
    <Link to={PAGER_TRY_PATH} className="pager-button">
      {t('pager.landing.tryCta')}
      <ArrowRight size={18} strokeWidth={2.4} aria-hidden="true" className="rtl:rotate-180" />
    </Link>
  );
};

const Cards = ({
  cards,
  icons,
  numbered = false,
}: {
  cards: Card[];
  icons: LucideIcon[];
  numbered?: boolean;
}): React.JSX.Element => (
  <ol className={`rennili-cards ${cards.length === 3 ? 'rennili-cards--3' : 'rennili-cards--4'}`}>
    {cards.map((card, index) => {
      const Icon = icons[index];
      return (
        <li key={card.title} className="rennili-card">
          <span className="rennili-card-icon" aria-hidden="true">
            {numbered ? index + 1 : Icon ? <Icon size={22} strokeWidth={1.8} /> : null}
          </span>
          <h3>{card.title}</h3>
          <p>{card.body}</p>
        </li>
      );
    })}
  </ol>
);

/** Rennili's landing page, the root of rennili.rtylr.com. */
const Landing = (): React.JSX.Element => {
  const { t } = useTranslation('pager');
  const steps = cardsFor(t('pager.landing.how.steps', { returnObjects: true }));
  const compare = cardsFor(t('pager.landing.compare.items', { returnObjects: true }));
  const [titleFirst, ...titleRest] = t('pager.landing.hero.title').split('\n');

  return (
    <div className="rennili-landing">
      <HelmetContainer
        title={t('pager.landing.meta.title')}
        description={t('pager.landing.meta.description')}
        keywords={t('pager.landing.meta.keywords')}
        image={`${RENNILI_URL}/og-image.png`}
        url={`${RENNILI_URL}/`}
      />
      <header className="rennili-nav">
        <RenniliLogo size={30} />
        <ByRtylr />
        <LanguageSwitch />
      </header>

      <main>
        <section className="rennili-section rennili-hero">
          <div>
            <p className="rennili-eyebrow">{t('pager.landing.hero.eyebrow')}</p>
            <h1 className="rennili-title">
              {titleFirst}
              {titleRest.length > 0 && (
                <>
                  {'\n'}
                  <em>{titleRest.join('\n')}</em>
                </>
              )}
            </h1>
            <p className="rennili-lead">{t('pager.landing.hero.subtitle')}</p>
            <div className="rennili-ctas">
              <TryLink />
              <span className="rennili-note">{t('pager.landing.hero.note')}</span>
            </div>
          </div>
          <div className="rennili-halo" aria-hidden="true">
            <RenniliMark size={200} wiggle />
          </div>
        </section>

        <section className="rennili-section" aria-labelledby="rennili-how">
          <h2 id="rennili-how" className="rennili-section-title">
            {t('pager.landing.how.title')}
          </h2>
          <p className="rennili-section-intro">{t('pager.landing.how.intro')}</p>
          <Cards cards={steps} icons={STEP_ICONS} numbered />
        </section>

        <section className="rennili-section" aria-labelledby="rennili-compare">
          <h2 id="rennili-compare" className="rennili-section-title">
            {t('pager.landing.compare.title')}
          </h2>
          <p className="rennili-section-intro">{t('pager.landing.compare.intro')}</p>
          <Cards cards={compare} icons={COMPARE_ICONS} />
        </section>

        <section className="rennili-section" aria-labelledby="rennili-free">
          <div className="rennili-free">
            <h2 id="rennili-free" className="rennili-section-title">
              {t('pager.landing.free.title')}
            </h2>
            <p className="rennili-section-intro">{t('pager.landing.free.body')}</p>
          </div>
        </section>

        <section className="rennili-section rennili-closing" aria-labelledby="rennili-cta">
          <h2 id="rennili-cta" className="rennili-section-title">
            {t('pager.landing.closing.title')}
          </h2>
          <p className="rennili-section-intro" style={{ marginInline: 'auto' }}>
            {t('pager.landing.closing.subtitle')}
          </p>
          <div className="rennili-ctas">
            <TryLink />
          </div>
          <p className="rennili-note" style={{ marginTop: 20 }}>
            {t('pager.landing.signIn')}{' '}
            <a className="pager-try-link" href={pagerClaimUrl('signIn')}>
              {t('pager.landing.signInLink')}
            </a>
          </p>
        </section>
      </main>

      <footer className="rennili-foot">
        <span>{t('pager.landing.footer')}</span>
        <ByRtylr />
      </footer>
    </div>
  );
};

export default Landing;
