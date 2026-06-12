import React from 'react';
import { motion } from 'motion/react';
import { useRouter } from 'next/router';
import { APP_URLS } from '@voxire/config';
import {
  blurFadeIn,
  staggerBlur,
  viewportOnce,
  EASE_OUT_EXPO,
  SPRING_SNAPPY,
} from '../../../lib/animations';
import { SplitWords } from '../../motion/SplitWords';
import { navigateExternal } from '../../../lib/navigation';
import styles from './CTASection.module.scss';

const BRANDS = [
  { id: 'stripe', label: 'Stripe' },
  { id: 'square', label: 'Square' },
  { id: 'quickbooks', label: 'QuickBooks' },
  { id: 'xero', label: 'Xero' },
  { id: 'slack', label: 'Slack' },
  { id: 'google', label: 'Google' },
];

const CTASection: React.FC = (): React.JSX.Element => {
  const router = useRouter();

  const handleBookDemo = (): void => {
    void router.push('/demo');
  };

  const handleGetStarted = (): void => {
    navigateExternal(`${APP_URLS.auth}/register`);
  };

  return (
    <section className={styles.ctaSection}>
      <div className={styles.split}>
        {/* Left, title, CTAs, logo grid */}
        <motion.div
          className={styles.left}
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={staggerBlur}
        >
          <SplitWords
            as="h2"
            className={styles.ctaTitle}
            text={"It's your turn\nto build."}
            trigger="view"
          />

          <motion.p className={styles.ctaSubtitle} variants={blurFadeIn}>
            Free for up to 10 members. No credit card.
            <br />
            Launch every department in an afternoon.
          </motion.p>

          <motion.div className={styles.ctaButtons} variants={blurFadeIn}>
            <motion.button
              type="button"
              onClick={handleGetStarted}
              className={styles.primaryBtn}
              whileHover={{ scale: 1.04 }}
              whileTap={{ scale: 0.96 }}
              transition={SPRING_SNAPPY}
            >
              Start for free
            </motion.button>
            <motion.button
              type="button"
              onClick={handleBookDemo}
              className={styles.outlineBtn}
              whileHover={{ scale: 1.04 }}
              whileTap={{ scale: 0.96 }}
              transition={SPRING_SNAPPY}
            >
              Book a demo
            </motion.button>
          </motion.div>

          <motion.div className={styles.brandGrid} variants={blurFadeIn}>
            {BRANDS.map((b) => (
              <div key={b.id} className={styles.brandCell}>
                <span className={styles.brandLabel}>{b.label}</span>
              </div>
            ))}
          </motion.div>
        </motion.div>

        {/* Right, generative image - unmasks from the center with a slow settle */}
        <motion.div
          className={styles.right}
          initial={{ opacity: 0, scale: 1.08, clipPath: 'inset(12% 12% 12% 12% round 24px)' }}
          whileInView={{ opacity: 1, scale: 1, clipPath: 'inset(0% 0% 0% 0% round 0px)' }}
          viewport={viewportOnce}
          transition={{ duration: 1.3, ease: EASE_OUT_EXPO }}
        >
          <img
            src="/images/home/hero-banner-top-bar-bg.webp"
            alt=""
            className={styles.rightImg}
            loading="lazy"
            decoding="async"
          />
        </motion.div>
      </div>
    </section>
  );
};

export default CTASection;
