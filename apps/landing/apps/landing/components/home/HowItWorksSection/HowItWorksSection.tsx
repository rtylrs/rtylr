import React, { useRef } from 'react';
import { motion, useScroll, useSpring } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { blurFadeIn, blurFadeInSlow, staggerBlur, viewportOnce } from '../../../lib/animations';
import { SplitWords } from '../../motion/SplitWords';
import styles from './HowItWorksSection.module.scss';

const HowItWorksSection: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation();
  const steps = t('home.howItWorks.steps', { returnObjects: true }) as unknown as {
    title: string;
    description: string;
  }[];

  const gridRef = useRef<HTMLDivElement>(null);

  // The accent line draws across the top of the grid as it scrolls
  // through the viewport, pacing the reader through the four steps.
  const { scrollYProgress } = useScroll({
    target: gridRef,
    offset: ['start 0.85', 'end 0.45'],
  });
  const lineProgress = useSpring(scrollYProgress, { stiffness: 140, damping: 30, mass: 0.5 });

  return (
    <section className={styles.howItWorksSection}>
      <div className={styles.container}>
        <motion.div
          className={styles.header}
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={blurFadeInSlow}
        >
          <span className={styles.eyebrow}>{t('home.howItWorks.eyebrow')}</span>
          <SplitWords
            as="h2"
            className={styles.sectionTitle}
            text={t('home.howItWorks.title')}
            trigger="view"
          />
        </motion.div>

        <motion.div
          ref={gridRef}
          className={styles.grid}
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={staggerBlur}
        >
          <motion.span
            className={styles.gridProgress}
            style={{ scaleX: lineProgress }}
            aria-hidden="true"
          />
          {steps.map((step, i) => (
            <motion.div key={step.title} className={styles.step} variants={blurFadeIn}>
              {/* The number is derived, not copy: it is the position. */}
              <div className={styles.stepNumber}>{String(i + 1).padStart(2, '0')}</div>
              <h3 className={styles.stepTitle}>{step.title}</h3>
              <p className={styles.stepDescription}>{step.description}</p>
            </motion.div>
          ))}
        </motion.div>
      </div>
    </section>
  );
};

export default HowItWorksSection;
