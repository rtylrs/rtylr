import React, { useEffect, useRef, useState } from 'react';
import { motion, useInView, useReducedMotion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { Check } from 'lucide-react';
import { cn } from '../../../lib/cn';
import { blurFadeIn, blurFadeInSlow, staggerBlur, viewportOnce } from '../../../lib/animations';
import { SplitWords } from '../../motion/SplitWords';
import { CaretMark } from '../../CaretMark/CaretMark';
import styles from './HowItWorksSection.module.scss';

const EMAIL = 'sara@voxire.com';

/** Counts 0..max once the card scrolls into view (instantly with reduced motion). */
function useRevealStep(inView: boolean, max: number, intervalMs: number, delayMs = 0): number {
  const reduceMotion = useReducedMotion();
  const [step, setStep] = useState(0);
  useEffect(() => {
    if (!inView) return;
    if (reduceMotion) {
      setStep(max);
      return;
    }
    let timer = 0;
    const start = window.setTimeout(() => {
      timer = window.setInterval(() => {
        setStep((s) => {
          if (s >= max) window.clearInterval(timer);
          return Math.min(s + 1, max);
        });
      }, intervalMs);
    }, delayMs);
    return () => {
      window.clearTimeout(start);
      window.clearInterval(timer);
    };
  }, [inView, reduceMotion, max, intervalMs, delayMs]);
  return step;
}

// ── Step visuals ─────────────────────────────────────────────────────────────

const SignUpVisual: React.FC<{ inView: boolean }> = ({ inView }) => {
  const { t } = useTranslation();
  // One step per typed character, then one more for the confirmed state.
  const step = useRevealStep(inView, EMAIL.length + 1, 70, 300);
  const typed = EMAIL.slice(0, step);
  const done = step > EMAIL.length;

  return (
    <div className={styles.visual}>
      <div className={styles.field}>
        <span className={styles.fieldLabel}>{t('home.howItWorks.visuals.email')}</span>
        <span className={styles.input}>
          {typed}
          {!done && <span className={styles.cursor} />}
        </span>
      </div>
      <span className={cn(styles.button, done && styles.buttonDone)}>
        {done && <Check size={13} strokeWidth={3} />}
        {done ? t('home.howItWorks.visuals.created') : t('home.howItWorks.visuals.create')}
      </span>
    </div>
  );
};

const CaretSetupVisual: React.FC<{ inView: boolean }> = ({ inView }) => {
  const { t } = useTranslation();
  const areas = t('home.howItWorks.visuals.areas', { returnObjects: true }) as string[];
  const step = useRevealStep(inView, areas.length + 1, 450, 600);

  return (
    <div className={styles.visual}>
      <div className={styles.userMsg}>{t('home.howItWorks.visuals.brief')}</div>
      <div className={styles.caretMsg}>
        <span className={styles.caretMark}>
          <CaretMark size={11} />
        </span>
        <div className={styles.setupList}>
          {areas.map((area, i) => {
            const on = step > i;
            return (
              <span key={area} className={cn(styles.setupItem, on && styles.setupItemOn)}>
                <span className={styles.tick}>{on && <Check size={9} strokeWidth={3.5} />}</span>
                {area}
              </span>
            );
          })}
        </div>
      </div>
    </div>
  );
};

const INVITES = [
  { initials: 'MK', color: '#3b82f6', tone: 'neutral' },
  { initials: 'LR', color: '#8b5cf6', tone: 'amber' },
  { initials: 'JD', color: '#f59e0b', tone: 'green' },
];

const TeamVisual: React.FC<{ inView: boolean }> = ({ inView }) => {
  const { t } = useTranslation();
  const people = t('home.howItWorks.visuals.people', { returnObjects: true }) as {
    name: string;
    access: string;
  }[];
  const step = useRevealStep(inView, INVITES.length, 420, 500);

  return (
    <div className={styles.visual}>
      {INVITES.map((inv, i) => (
        <motion.div
          key={inv.initials}
          className={styles.invite}
          initial={false}
          animate={{ opacity: step > i ? 1 : 0, y: step > i ? 0 : 6 }}
          transition={{ duration: 0.35 }}
        >
          <span className={styles.avatar} style={{ background: inv.color }}>
            {inv.initials}
          </span>
          <span className={styles.inviteName}>{people[i].name}</span>
          <span className={styles.access} data-tone={inv.tone}>
            {people[i].access}
          </span>
        </motion.div>
      ))}
    </div>
  );
};

const VISUALS = [SignUpVisual, CaretSetupVisual, TeamVisual];

// ── Section ──────────────────────────────────────────────────────────────────

const HowItWorksSection: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation();
  const steps = t('home.howItWorks.steps', { returnObjects: true }) as unknown as {
    title: string;
    description: string;
  }[];
  const gridRef = useRef<HTMLDivElement>(null);
  const inView = useInView(gridRef, { once: true, margin: '-120px' });

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
            restLinesClassName={styles.sectionTitleMuted}
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
          {steps.map((step, i) => {
            const Visual = VISUALS[i];
            return (
              <motion.div key={step.title} className={styles.card} variants={blurFadeIn}>
                <div className={styles.stage} aria-hidden="true">
                  <Visual inView={inView} />
                </div>
                <div className={styles.copy}>
                  <span className={styles.stepNumber}>{i + 1}</span>
                  <h3 className={styles.stepTitle}>{step.title}</h3>
                  <p className={styles.stepDescription}>{step.description}</p>
                </div>
              </motion.div>
            );
          })}
        </motion.div>
      </div>
    </section>
  );
};

export default HowItWorksSection;
