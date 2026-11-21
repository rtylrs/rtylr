import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from '@tanstack/react-router';
import { AnimatePresence, motion } from 'motion/react';
import {
  ArrowRight,
  BellRing,
  Check,
  Globe,
  QrCode,
  RotateCcw,
  Shuffle,
  Smartphone,
} from 'lucide-react';
import MatchCode from '../../../shared/components/MatchCode';
import { PAGER_TRY_PATH } from '../../../pagerPaths';

const SHOWCASE_CODES = ['427', '318', '952', '604', '781', '265'] as const;

export const BentoShowcase = (): React.JSX.Element => {
  const { t } = useTranslation('pager');
  const [codeIndex, setCodeIndex] = useState(0);
  const [miniReady, setMiniReady] = useState(false);

  const currentCode = SHOWCASE_CODES[codeIndex];

  const nextCode = (): void => {
    setCodeIndex((prev) => (prev + 1) % SHOWCASE_CODES.length);
  };

  return (
    <section
      id="features"
      className="rl-section rl-section--mist"
      aria-labelledby="rl-bento-heading"
    >
      <div className="rl-wrap">
        <div className="rl-section-head">
          <h2 id="rl-bento-heading" className="rl-section-title">
            {t('pager.landing.bento.title')}
          </h2>
          <p className="rl-section-intro">{t('pager.landing.bento.intro')}</p>
        </div>

        <div className="rl-bento-grid">
          {/* Tile 1 (2-col wide): interactive match code shuffler */}
          <article className="rl-bento-card rl-bento-card--wide rl-bento-shuffler">
            <div className="rl-bento-copy">
              <span className="rl-bento-kicker">{t('pager.landing.bento.shuffler.kicker')}</span>
              <h3 className="rl-bento-title">{t('pager.landing.bento.shuffler.title')}</h3>
              <p className="rl-bento-body">{t('pager.landing.bento.shuffler.body')}</p>
              <motion.button
                type="button"
                whileTap={{ scale: 0.96 }}
                onClick={nextCode}
                className="rl-bento-action"
              >
                <Shuffle size={15} strokeWidth={2.2} aria-hidden="true" />
                <span>{t('pager.landing.bento.shuffler.action')}</span>
              </motion.button>
            </div>

            <div className="rl-bento-shuffler-stage" aria-live="polite">
              <AnimatePresence mode="wait" initial={false}>
                <motion.div
                  key={currentCode}
                  className="rl-code-card"
                  initial={{ opacity: 0, scale: 0.94, y: 6 }}
                  animate={{ opacity: 1, scale: 1, y: 0 }}
                  exit={{ opacity: 0, scale: 0.96, y: -6 }}
                  transition={{ type: 'spring', stiffness: 460, damping: 30 }}
                >
                  <span className="rl-code-caption">{t('pager.ticket.codeLabel')}</span>
                  <MatchCode code={currentCode} className="rl-code-digits" />
                </motion.div>
              </AnimatePresence>
            </div>
          </article>

          {/* Tile 2 (1-col): Interactive One-Tap Staff Mini-Card */}
          <article className="rl-bento-card rl-bento-board">
            <span className="rl-bento-kicker">{t('pager.landing.bento.miniBoard.kicker')}</span>
            <h3 className="rl-bento-title">{t('pager.landing.bento.miniBoard.title')}</h3>
            <p className="rl-bento-body">{t('pager.landing.bento.miniBoard.body')}</p>

            <div className={`rl-mini-ticket ${miniReady ? 'rl-mini-ticket--ready' : ''}`}>
              <div className="rl-mini-ticket-head">
                <strong className="rl-mini-ticket-num" dir="ltr">
                  {t('pager.landing.bento.miniBoard.order')}
                </strong>
                <span className={`rl-mini-pill ${miniReady ? 'rl-mini-pill--ready' : ''}`}>
                  {miniReady
                    ? t('pager.landing.bento.miniBoard.statusReady')
                    : t('pager.landing.bento.miniBoard.statusWaiting')}
                </span>
              </div>
              <motion.button
                type="button"
                whileTap={{ scale: 0.97 }}
                aria-pressed={miniReady}
                onClick={() => setMiniReady((prev) => !prev)}
                className={`rl-mini-btn ${miniReady ? 'rl-mini-btn--ready' : ''}`}
              >
                {miniReady ? (
                  <>
                    <RotateCcw size={14} strokeWidth={2.3} aria-hidden="true" />
                    <span>{t('pager.landing.bento.miniBoard.tapUndo')}</span>
                  </>
                ) : (
                  <>
                    <BellRing size={14} strokeWidth={2.3} aria-hidden="true" />
                    <span>{t('pager.landing.bento.miniBoard.tapReady')}</span>
                  </>
                )}
              </motion.button>
            </div>
          </article>

          {/* Tile 3 (1-col): One Printed QR */}
          <article className="rl-bento-card rl-bento-poster">
            <span className="rl-bento-kicker">{t('pager.landing.bento.poster.kicker')}</span>
            <h3 className="rl-bento-title">{t('pager.landing.bento.poster.title')}</h3>
            <p className="rl-bento-body">{t('pager.landing.bento.poster.body')}</p>
            <div className="rl-mini-poster" aria-hidden="true">
              <QrCode size={48} strokeWidth={1.6} />
            </div>
          </article>

          {/* Tile 4 (1-col): 0 MB Instant Browser */}
          <article className="rl-bento-card rl-bento-browser">
            <span className="rl-bento-kicker">{t('pager.landing.bento.browser.kicker')}</span>
            <h3 className="rl-bento-title">{t('pager.landing.bento.browser.title')}</h3>
            <p className="rl-bento-body">{t('pager.landing.bento.browser.body')}</p>
            <div className="rl-browser-pill" dir="ltr" aria-hidden="true">
              <Globe size={14} strokeWidth={2} />
              <span>rennili.rtylr.com</span>
              <strong>{t('pager.landing.bento.browser.stat')}</strong>
            </div>
          </article>

          {/* Tile 5 (1-col): Honest iPhone & Android Specs */}
          <article className="rl-bento-card rl-bento-platforms">
            <span className="rl-bento-kicker">{t('pager.landing.bento.platforms.kicker')}</span>
            <h3 className="rl-bento-title">{t('pager.landing.bento.platforms.title')}</h3>
            <div className="rl-platform-list">
              <div className="rl-platform-row">
                <span className="rl-platform-icon" aria-hidden="true">
                  <Smartphone size={15} strokeWidth={2.2} />
                </span>
                <div>
                  <strong>{t('pager.landing.bento.platforms.androidTitle')}</strong>
                  <p>{t('pager.landing.bento.platforms.androidDesc')}</p>
                </div>
              </div>
              <div className="rl-platform-row">
                <span className="rl-platform-icon" aria-hidden="true">
                  <Check size={15} strokeWidth={2.5} />
                </span>
                <div>
                  <strong>{t('pager.landing.bento.platforms.iphoneTitle')}</strong>
                  <p>{t('pager.landing.bento.platforms.iphoneDesc')}</p>
                </div>
              </div>
            </div>
          </article>

          {/* Tile 6 (2-col wide, Deep Botanical Forest accent card): Peak Rush Unlimited */}
          <article className="rl-bento-card rl-bento-card--wide rl-bento-card--forest">
            <div className="rl-bento-copy">
              <span className="rl-bento-kicker rl-bento-kicker--light">
                {t('pager.landing.bento.rush.kicker')}
              </span>
              <h3 className="rl-bento-title rl-bento-title--light">
                {t('pager.landing.bento.rush.title')}
              </h3>
              <p className="rl-bento-body rl-bento-body--light">
                {t('pager.landing.bento.rush.body')}
              </p>
              <div className="rl-bento-cta-wrap">
                <Link to={PAGER_TRY_PATH} className="rl-btn rl-btn--on-forest">
                  {t('pager.landing.bento.rush.cta')}
                  <ArrowRight
                    size={16}
                    strokeWidth={2.4}
                    aria-hidden="true"
                    className="rtl:rotate-180"
                  />
                </Link>
              </div>
            </div>
            <div className="rl-bento-infinity" aria-hidden="true">
              <span className="rl-bento-infinity-mark">∞</span>
              <span className="rl-bento-infinity-label">{t('pager.landing.bento.rush.badge')}</span>
            </div>
          </article>
        </div>
      </div>
    </section>
  );
};
