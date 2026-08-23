import React, { useCallback, useEffect, useRef, useState } from 'react';
import { cn } from '../../../lib/cn';
import { motion, AnimatePresence } from 'motion/react';
import { ArrowUpRight } from 'lucide-react';
import { Button } from '@voxire/components';
import { useRouter } from 'next/router';
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

interface Feature {
  id: string;
  category: string;
  title: string;
  description: string;
  ctaText: string;
  Demo: React.ComponentType<DemoProps>;
}

const FEATURES: Feature[] = [
  {
    id: 'pos',
    category: 'POS',
    title: 'A register that keeps up with real service.',
    description:
      'Tickets, tables, splits, tips, and every payment type - card, Apple Pay, cash, gift, comp. Works offline when the WiFi doesn’t. < 200ms per checkout.',
    ctaText: 'Open POS',
    Demo: POSDemo,
  },
  {
    id: 'erp',
    category: 'ERP',
    title: 'Inventory that counts itself.',
    description:
      'Recipe-linked depletion as sales ring up. Auto-reorder at threshold. Multi-location transfers with a full audit trail. Your shelves tell the database, not the other way around.',
    ctaText: 'Open ERP',
    Demo: ERPDemo,
  },
  {
    id: 'hr',
    category: 'HR',
    title: 'The people side, without the paperwork.',
    description:
      'Drag-to-publish schedules. Clock-in from the POS. Payroll that rolls up tips, hours, and deductions automatically. Compliance flags before they become violations.',
    ctaText: 'Open HR',
    Demo: HRDemo,
  },
  {
    id: 'crm',
    category: 'CRM',
    title: 'Every customer, remembered.',
    description:
      'Every POS sale enriches the profile. Segment by spend, visit cadence, birthday, tags. Send the campaign from the same app - no Mailchimp detour required.',
    ctaText: 'Open CRM',
    Demo: CRMDemo,
  },
  {
    id: 'finance',
    category: 'Finance',
    title: 'A real P&L, in real time.',
    description:
      'Recurring invoices that collect themselves. Expenses captured at source. Sales tax by jurisdiction. A P&L that updates with every sale, not at month-end.',
    ctaText: 'Open Finance',
    Demo: FinanceDemo,
  },
  {
    id: 'flow',
    category: 'Flow',
    title: 'Work that moves between apps.',
    description:
      'Kanban boards, automations, and tasks linked to the invoice, customer, or shift they’re actually about. Operations that don’t slip between tools.',
    ctaText: 'Open Flow',
    Demo: FlowDemo,
  },
  {
    id: 'insights',
    category: 'Insights',
    title: 'The dashboard that sees everything.',
    description:
      'Revenue trends, customer cohorts, product performance, labor margin. Every number pulled live from every other app. No SQL, no nightly ETL.',
    ctaText: 'Open Insights',
    Demo: InsightsDemo,
  },
];

const AUTO_ADVANCE_MS = 6000;

const FeaturesSection: React.FC = (): React.JSX.Element => {
  const router = useRouter();
  const [activeIndex, setActiveIndex] = useState<number>(0);
  const [isPaused, setIsPaused] = useState<boolean>(false);
  const timerRef = useRef<number | null>(null);

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
          <span className={styles.badge}>THE PLATFORM</span>
          <h2 className={styles.sectionTitle}>
            What does each app do?
            <br />
            <span className={styles.sectionTitleAccent}>One screen at a time.</span>
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

            <div className={styles.progressRow} role="group" aria-label="Choose a platform app">
              {FEATURES.map((f, i) => (
                <button
                  key={f.id}
                  type="button"
                  aria-label={`Show ${f.category}`}
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
