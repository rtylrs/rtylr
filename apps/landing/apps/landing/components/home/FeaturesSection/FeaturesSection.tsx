import React, { useCallback, useEffect, useRef, useState } from 'react';
import { cn } from '../../../lib/cn';
import { motion, AnimatePresence } from 'motion/react';
import { ArrowUpRight } from 'lucide-react';
import { Button } from '@voxire/components';
import { useRouter } from 'next/router';
import { useTranslation } from 'react-i18next';
import { blurFadeIn, blurFadeInSlow, viewportOnce, SPRING_SMOOTH } from '../../../lib/animations';
import { POSDemo } from './demos/POSDemo';
import { ERPDemo } from './demos/ERPDemo';
import { HRDemo } from './demos/HRDemo';
import { CRMDemo } from './demos/CRMDemo';
import { FinanceDemo } from './demos/FinanceDemo';
import { FlowDemo } from './demos/FlowDemo';
import { InsightsDemo } from './demos/InsightsDemo';
import styles from './FeaturesSection.module.scss';

interface DemoProps {
  isActive: boolean;
}

interface FeatureMeta {
  id: string;
  category: string;
  Demo: React.ComponentType<DemoProps>;
}

interface Feature extends FeatureMeta {
  title: string;
  description: string;
  ctaText: string;
}

// category is the product name, not copy - POS, ERP, HR, CRM, Finance and
// Flow are the same protected, untranslated terms used everywhere else on
// the site. Only title and description are written per language, in
// home.features.items, matched to this array by index.
const FEATURE_META: FeatureMeta[] = [
  { id: 'pos', category: 'POS', Demo: POSDemo },
  { id: 'erp', category: 'ERP', Demo: ERPDemo },
  { id: 'hr', category: 'HR', Demo: HRDemo },
  { id: 'crm', category: 'CRM', Demo: CRMDemo },
  { id: 'finance', category: 'Finance', Demo: FinanceDemo },
  { id: 'flow', category: 'Flow', Demo: FlowDemo },
  { id: 'insights', category: 'Insights', Demo: InsightsDemo },
];

const AUTO_ADVANCE_MS = 6000;

const FeaturesSection: React.FC = (): React.JSX.Element => {
  const router = useRouter();
  const { t } = useTranslation();
  const [activeIndex, setActiveIndex] = useState<number>(0);
  const [isPaused, setIsPaused] = useState<boolean>(false);
  const timerRef = useRef<number | null>(null);

  const items = t('home.features.items', { returnObjects: true }) as unknown as {
    title: string;
    description: string;
  }[];
  const FEATURES: Feature[] = FEATURE_META.map((meta, i) => ({
    ...meta,
    ...items[i],
    ctaText: t('home.features.openCta', { app: meta.category }),
  }));

  const active = FEATURES[activeIndex];

  const advance = useCallback((): void => {
    setActiveIndex((i) => (i + 1) % FEATURES.length);
  }, []);

  useEffect(() => {
    if (isPaused) return;
    timerRef.current = window.setTimeout(advance, AUTO_ADVANCE_MS);
    return () => {
      if (timerRef.current !== null) {
        window.clearTimeout(timerRef.current);
        timerRef.current = null;
      }
    };
  }, [activeIndex, isPaused, advance]);

  const selectFeature = (idx: number): void => {
    setActiveIndex(idx);
  };

  return (
    <motion.section
      className={styles.featuresSection}
      initial="hidden"
      whileInView="visible"
      viewport={viewportOnce}
      variants={blurFadeInSlow}
      onMouseEnter={() => setIsPaused(true)}
      onMouseLeave={() => setIsPaused(false)}
    >
      <div className={styles.container}>
        <motion.div className={styles.header} variants={blurFadeIn}>
          <span className={styles.badge}>{t('home.features.badge')}</span>
          <h2 className={styles.sectionTitle}>
            {t('home.features.title')}
            <br />
            <span className={styles.sectionTitleAccent}>{t('home.features.titleAccent')}</span>
          </h2>
        </motion.div>

        <motion.div className={styles.tabBar} variants={blurFadeIn} role="tablist">
          {FEATURES.map((f, i) => (
            <button
              key={f.id}
              id={`tab-${f.id}`}
              type="button"
              role="tab"
              aria-selected={i === activeIndex}
              aria-controls={`feature-panel-${f.id}`}
              onClick={() => selectFeature(i)}
              className={cn(styles.tab, i === activeIndex && styles.tabActive)}
            >
              {i === activeIndex && (
                <motion.span
                  layoutId="featureTabPill"
                  className={styles.tabPill}
                  transition={SPRING_SMOOTH}
                />
              )}
              <span className={styles.tabLabel}>{f.category}</span>
            </button>
          ))}
        </motion.div>

        <div className={styles.showcaseRow}>
          <div
            className={styles.copyCol}
            id={`feature-panel-${active.id}`}
            role="tabpanel"
            aria-labelledby={`tab-${active.id}`}
          >
            <AnimatePresence mode="wait">
              <motion.div
                key={active.id}
                initial={{ opacity: 0, y: 14 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -10 }}
                transition={{ duration: 0.4, ease: [0.16, 1, 0.3, 1] }}
              >
                <span className={styles.featureCategory}>{active.category}</span>
                <h3 className={styles.featureTitle}>{active.title}</h3>
                <p className={styles.featureDescription}>{active.description}</p>
                <div className={styles.featureActions}>
                  <Button
                    variant="dark-outline"
                    onClick={() => router.push('/demo')}
                    className={styles.featureButton}
                    suffix={<ArrowUpRight size={15} strokeWidth={2.1} aria-hidden="true" />}
                  >
                    {active.ctaText}
                  </Button>
                </div>
              </motion.div>
            </AnimatePresence>
          </div>

          <div className={styles.imgCol}>
            <div className={styles.showcaseImageContainer}>
              {FEATURES.map((f, i) => (
                <div
                  key={f.id}
                  className={cn(styles.showcaseImage, i === activeIndex && styles.active)}
                  aria-hidden={i !== activeIndex}
                >
                  <f.Demo isActive={i === activeIndex} />
                </div>
              ))}
            </div>

            <div
              className={styles.progressRow}
              role="group"
              aria-label={t('home.features.chooseAppLabel')}
            >
              {FEATURES.map((f, i) => (
                <button
                  key={f.id}
                  type="button"
                  aria-label={t('home.features.showLabel', { app: f.category })}
                  aria-pressed={i === activeIndex}
                  onClick={() => selectFeature(i)}
                  className={cn(styles.progressDot, i === activeIndex && styles.progressDotActive)}
                >
                  {i === activeIndex && (
                    <motion.span
                      key={`fill-${activeIndex}`}
                      className={styles.progressFill}
                      initial={{ scaleX: 0 }}
                      animate={{ scaleX: 1 }}
                      transition={{ duration: AUTO_ADVANCE_MS / 1000, ease: 'linear' }}
                    />
                  )}
                </button>
              ))}
            </div>
          </div>
        </div>
      </div>
    </motion.section>
  );
};

export default FeaturesSection;
