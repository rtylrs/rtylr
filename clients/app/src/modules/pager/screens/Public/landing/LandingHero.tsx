import React from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from '@tanstack/react-router';
import { motion } from 'motion/react';
import { ArrowRight, Check } from 'lucide-react';
import { PAGER_HOW_PATH, PAGER_TRY_PATH } from '../../../pagerPaths';
import { HeroInteractiveStage } from './PhoneMockups';

interface LandingHeroProps {
  heroRef: React.RefObject<HTMLElement | null>;
}

export const LandingHero = ({ heroRef }: LandingHeroProps): React.JSX.Element => {
  const { t } = useTranslation('pager');
  const [titleFirst, ...titleRest] = t('pager.landing.hero.title').split('\n');
  const rawStrip = t('pager.landing.hero.strip', { returnObjects: true });
  const stripItems = Array.isArray(rawStrip) ? (rawStrip as string[]) : [];

  return (
    <section ref={heroRef} className="rl-hero">
      <div className="rl-wrap rl-hero-layout">
        <div className="rl-hero-copy">
          <motion.p
            className="rl-eyebrow"
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.42, ease: [0.22, 1, 0.36, 1] }}
          >
            <span className="rl-eyebrow-dot" aria-hidden="true" />
            {t('pager.landing.hero.eyebrow')}
          </motion.p>

          <h1 className="rl-hero-title">
            <span className="rl-line-mask">
              <motion.span
                className="rl-line-inner"
                initial={{ y: '105%', opacity: 0 }}
                animate={{ y: '0%', opacity: 1 }}
                transition={{ duration: 0.56, delay: 0.05, ease: [0.22, 1, 0.36, 1] }}
              >
                {titleFirst}
              </motion.span>
            </span>
            {titleRest.length > 0 && (
              <span className="rl-line-mask">
                <motion.em
                  className="rl-line-inner"
                  initial={{ y: '105%', opacity: 0 }}
                  animate={{ y: '0%', opacity: 1 }}
                  transition={{ duration: 0.56, delay: 0.14, ease: [0.22, 1, 0.36, 1] }}
                >
                  {titleRest.join('\n')}
                </motion.em>
              </span>
            )}
          </h1>

          <motion.p
            className="rl-hero-lead"
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.48, delay: 0.2, ease: [0.22, 1, 0.36, 1] }}
          >
            {t('pager.landing.hero.subtitle')}
          </motion.p>

          <motion.div
            className="rl-hero-ctas"
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.48, delay: 0.26, ease: [0.22, 1, 0.36, 1] }}
          >
            <Link to={PAGER_TRY_PATH} className="rl-btn rl-btn--primary rl-btn--lg">
              {t('pager.landing.hero.primaryCta')}
              <ArrowRight
                size={18}
                strokeWidth={2.4}
                aria-hidden="true"
                className="rtl:rotate-180"
              />
            </Link>
            <Link to={PAGER_HOW_PATH} className="rl-btn rl-btn--secondary rl-btn--lg">
              {t('pager.landing.hero.secondaryCta')}
            </Link>
          </motion.div>

          <motion.p
            className="rl-hero-note"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: 0.45, delay: 0.34 }}
          >
            {t('pager.landing.hero.note')}
          </motion.p>
        </div>

        <HeroInteractiveStage />
      </div>

      {/* Calm studio proof strip anchoring the hero */}
      <div className="rl-hero-specbar">
        <ul className="rl-wrap rl-hero-speclist">
          {stripItems.map((item) => (
            <li key={item} className="rl-hero-specitem">
              <Check size={14} strokeWidth={2.8} aria-hidden="true" />
              <span>{item}</span>
            </li>
          ))}
        </ul>
      </div>
    </section>
  );
};
