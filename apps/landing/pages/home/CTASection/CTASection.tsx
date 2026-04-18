import React from 'react';
import { motion } from 'motion/react';
import { useRouter } from 'next/router';
import { APP_URLS } from '@voxire/config';
import { blurFadeIn, staggerBlur, viewportOnce } from '../../../lib/animations';
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
        {/* Left — title, CTAs, logo grid */}
        <motion.div
          className={styles.left}
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={staggerBlur}
        >
          <motion.h2 className={styles.ctaTitle} variants={blurFadeIn}>
            It&apos;s your turn
            <br />
            to build.
          </motion.h2>

          <motion.p className={styles.ctaSubtitle} variants={blurFadeIn}>
            Free for up to 5 members. No credit card required.
            <br />
            Launch your entire business in minutes.
          </motion.p>

          <motion.div className={styles.ctaButtons} variants={blurFadeIn}>
            <button type="button" onClick={handleGetStarted} className={styles.primaryBtn}>
              Start for free
            </button>
            <button type="button" onClick={handleBookDemo} className={styles.outlineBtn}>
              Book a demo
            </button>
          </motion.div>

          <motion.div className={styles.brandGrid} variants={blurFadeIn}>
            {BRANDS.map((b) => (
              <div key={b.id} className={styles.brandCell}>
                <span className={styles.brandLabel}>{b.label}</span>
              </div>
            ))}
          </motion.div>
        </motion.div>

        {/* Right — generative image */}
        <motion.div
          className={styles.right}
          initial={{ opacity: 0, scale: 1.02 }}
          whileInView={{ opacity: 1, scale: 1 }}
          viewport={viewportOnce}
          transition={{ duration: 1.1, ease: [0.16, 1, 0.3, 1] }}
        >
          <img
            src="/images/home/hero-banner-top-bar-bg.webp"
            alt=""
            className={styles.rightImg}
            aria-hidden="true"
          />
          <div className={styles.rightFade} />
        </motion.div>
      </div>
    </section>
  );
};

export default CTASection;
