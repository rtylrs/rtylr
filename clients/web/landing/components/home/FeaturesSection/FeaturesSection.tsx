import React, { useCallback, useEffect, useRef, useState } from 'react';
import { cn } from '../../../lib/cn';
import { motion, useInView } from 'motion/react';
import { ArrowRight } from 'lucide-react';
import Link from 'next/link';
import { useTranslation } from 'react-i18next';
import { blurFadeIn, blurFadeInSlow, viewportOnce } from '../../../lib/animations';
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
  href: string;
  Demo: React.ComponentType<DemoProps>;
}

interface Feature extends FeatureMeta {
  label: string;
  title: string;
  description: string;
  ctaText: string;
}

// Same areas, same order as the workspace grid above. Label, title and
// description are written per language in home.features.items, matched to
// this array by index.
const FEATURE_META: FeatureMeta[] = [
  { id: 'flow', href: '/flow', Demo: FlowDemo },
  { id: 'finance', href: '/finance', Demo: FinanceDemo },
  { id: 'crm', href: '/crm', Demo: CRMDemo },
  { id: 'hr', href: '/hr', Demo: HRDemo },
  { id: 'erp', href: '/erp', Demo: ERPDemo },
  { id: 'pos', href: '/pos', Demo: POSDemo },
  { id: 'insights', href: 'https://insights.rtylr.com', Demo: InsightsDemo },
];

const AUTO_ADVANCE_MS = 6000;

/**
 * Attio-style showcase: a vertical list of areas on the left (the active one
 * opens to show its copy and a filling progress rule) and the live demo on
 * the right. Auto-advances; hovering the section pauses it.
 */
const FeaturesSection: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation();
  const [activeIndex, setActiveIndex] = useState<number>(0);
  const [isPaused, setIsPaused] = useState<boolean>(false);
  const timerRef = useRef<number | null>(null);
  const sectionRef = useRef<HTMLElement>(null);
  // Only auto-advance while the section is on screen.
  const inView = useInView(sectionRef, { margin: '-80px' });

  const items = t('home.features.items', { returnObjects: true }) as unknown as {
    label: string;
    title: string;
    description: string;
  }[];
  const FEATURES: Feature[] = FEATURE_META.map((meta, i) => ({
    ...meta,
    ...items[i],
    ctaText: t('home.features.openCta', { capability: items[i].label }),
  }));

  const advance = useCallback((): void => {
    setActiveIndex((i) => (i + 1) % FEATURE_META.length);
  }, []);

  useEffect(() => {
    if (isPaused || !inView) return;
    timerRef.current = window.setTimeout(advance, AUTO_ADVANCE_MS);
    return () => {
      if (timerRef.current !== null) {
        window.clearTimeout(timerRef.current);
        timerRef.current = null;
      }
    };
  }, [activeIndex, isPaused, inView, advance]);

  return (
    <motion.section
      ref={sectionRef}
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
            <span className={styles.sectionTitleAccent}>{t('home.features.titleAccent')}</span>
          </h2>
        </motion.div>

        <div className={styles.layout}>
          <div className={styles.tabList} role="tablist" aria-orientation="vertical">
            {FEATURES.map((f, i) => {
              const isActive = i === activeIndex;
              return (
                <div key={f.id} className={cn(styles.tabItem, isActive && styles.tabItemActive)}>
                  <button
                    id={`tab-${f.id}`}
                    type="button"
                    role="tab"
                    aria-selected={isActive}
                    aria-controls="feature-stage"
                    onClick={() => setActiveIndex(i)}
                    className={styles.tabTrigger}
                  >
                    {f.label}
                  </button>

                  {/* Opacity only, never height: the body swaps at a fixed
                      height in one frame, so the section's height (and
                      everything below it) never moves. Animating height
                      made it dip mid-swap and scroll anchoring shifted the
                      page under the reader. */}
                  {isActive && (
                    <motion.div
                      className={styles.tabBody}
                      initial={{ opacity: 0, y: 4 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ duration: 0.35 }}
                    >
                      <p className={styles.tabText}>
                        <strong>{f.title}</strong> {f.description}
                      </p>
                      <Link href={f.href} className={styles.tabLink}>
                        {f.ctaText}
                        <ArrowRight size={14} aria-hidden="true" />
                      </Link>
                    </motion.div>
                  )}

                  {isActive && (
                    <motion.span
                      key={`fill-${activeIndex}-${isPaused || !inView}`}
                      className={styles.tabProgress}
                      initial={{ scaleX: 0 }}
                      animate={{ scaleX: isPaused || !inView ? 0 : 1 }}
                      transition={{
                        duration: isPaused || !inView ? 0 : AUTO_ADVANCE_MS / 1000,
                        ease: 'linear',
                      }}
                      aria-hidden="true"
                    />
                  )}
                </div>
              );
            })}
          </div>

          <div
            id="feature-stage"
            className={styles.stage}
            role="tabpanel"
            aria-labelledby={`tab-${FEATURES[activeIndex].id}`}
          >
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
          </div>
        </div>
      </div>
    </motion.section>
  );
};

export default FeaturesSection;
