import React from 'react';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
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
import CTAGraphic from './CTAGraphic';
import Lines from '../../text/Lines';
import styles from './CTASection.module.scss';

const CTASection: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation();
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
        {/* Left, title, CTAs */}
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
            text={t('home.cta.title')}
            trigger="view"
          />

          <motion.p className={styles.ctaSubtitle} variants={blurFadeIn}>
            <Lines text={t('home.cta.subtitle')} />
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
              {t('common.cta.startFree')}
            </motion.button>
            <motion.button
              type="button"
              onClick={handleBookDemo}
              className={styles.outlineBtn}
              whileHover={{ scale: 1.04 }}
              whileTap={{ scale: 0.96 }}
              transition={SPRING_SNAPPY}
            >
              {t('common.cta.bookDemo')}
            </motion.button>
          </motion.div>
        </motion.div>

        {/* Right, interactive graphic - unmasks from the center with a slow settle */}
        <motion.div
          className={styles.right}
          initial={{ opacity: 0, scale: 1.08, clipPath: 'inset(12% 12% 12% 12% round 24px)' }}
          whileInView={{ opacity: 1, scale: 1, clipPath: 'inset(0% 0% 0% 0% round 0px)' }}
          viewport={viewportOnce}
          transition={{ duration: 1.3, ease: EASE_OUT_EXPO }}
        >
          <CTAGraphic />
        </motion.div>
      </div>
    </section>
  );
};

export default CTASection;
