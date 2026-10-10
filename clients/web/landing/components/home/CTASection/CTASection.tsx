import React from 'react';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { useRouter } from 'next/router';
import { ArrowRight } from 'lucide-react';
import { APP_URLS } from '@voxire/config';
import { blurFadeIn, staggerBlur, viewportOnce, SPRING_SNAPPY } from '../../../lib/animations';
import { SplitWords } from '../../motion/SplitWords';
import { navigateExternal } from '../../../lib/navigation';
import { track, EV_SIGNUP_CLICK, EV_CTA_CLICK } from '../../../lib/analytics';
import styles from './CTASection.module.scss';

/** Closing call to action: one two-tone line, one reassurance, two buttons. */
const CTASection: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation();
  const router = useRouter();

  const handleBookDemo = (): void => {
    track(EV_CTA_CLICK, { cta: 'demo', page: router.pathname });
    void router.push('/demo');
  };

  const handleGetStarted = (): void => {
    track(EV_SIGNUP_CLICK, { source: 'cta' });
    navigateExternal(`${APP_URLS.auth}/register`);
  };

  return (
    <section className={styles.ctaSection}>
      <motion.div
        className={styles.inner}
        initial="hidden"
        whileInView="visible"
        viewport={viewportOnce}
        variants={staggerBlur}
      >
        <SplitWords
          as="h2"
          className={styles.ctaTitle}
          text={t('home.cta.title')}
          restLinesClassName={styles.ctaTitleMuted}
          trigger="view"
        />

        <motion.p className={styles.ctaSubtitle} variants={blurFadeIn}>
          {t('home.cta.subtitle')}
        </motion.p>

        <motion.div className={styles.ctaButtons} variants={blurFadeIn}>
          <motion.button
            type="button"
            onClick={handleGetStarted}
            className={styles.ctaPrimary}
            whileHover={{ scale: 1.03 }}
            whileTap={{ scale: 0.97 }}
            transition={SPRING_SNAPPY}
          >
            {t('common.cta.getStarted')}
            <ArrowRight size={15} strokeWidth={2.2} aria-hidden="true" />
          </motion.button>
          <motion.button
            type="button"
            onClick={handleBookDemo}
            className={styles.ctaSecondary}
            whileHover={{ scale: 1.03 }}
            whileTap={{ scale: 0.97 }}
            transition={SPRING_SNAPPY}
          >
            {t('common.cta.bookDemo')}
          </motion.button>
        </motion.div>
      </motion.div>
    </section>
  );
};

export default CTASection;
