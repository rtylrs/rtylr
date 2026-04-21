import React, { useCallback, useEffect, useRef, useState } from 'react';
import { motion, AnimatePresence } from 'motion/react';
import { Button } from '@voxire/components';
import { useRouter } from 'next/router';
import { blurFadeIn, blurFadeInSlow, viewportOnce } from '../../../lib/animations';
import styles from './FeaturesSection.module.scss';

/**
 * Apps showcase — tabbed version.
 *
 * The previous incarnation was a sticky-right-column, scroll-linked
 * feed where each of the seven apps occupied ~50vh. Great on a demo
 * recording; terrible on the actual home page — seven viewport
 * heights of scroll just to get past the platform pitch.
 *
 * This version collapses the same content into a single tabbed
 * viewport: one screen tall, one app on display at a time, auto-
 * advancing every 6s (pauses when the visitor hovers or interacts).
 * All seven app images stay preloaded so swaps are instant.
 */

interface Feature {
  id: string;
  category: string;
  title: string;
  description: string;
  ctaText: string;
  image: string;
}

const FEATURES: Feature[] = [
  {
    id: 'pos',
    category: 'POS',
    title: 'A register that keeps up with real service.',
    description:
      'Tickets, tables, splits, tips, and every payment type — card, Apple Pay, cash, gift, comp. Works offline when the WiFi doesn\u2019t. < 200ms per checkout.',
    ctaText: 'Open POS',
    image: '/images/features/pos.jpg',
  },
  {
    id: 'erp',
    category: 'ERP',
    title: 'Inventory that counts itself.',
    description:
      'Recipe-linked depletion as sales ring up. Auto-reorder at threshold. Multi-location transfers with a full audit trail. Your shelves tell the database, not the other way around.',
    ctaText: 'Open ERP',
    image: '/images/features/erp.jpg',
  },
  {
    id: 'hr',
    category: 'HR',
    title: 'The people side, without the paperwork.',
    description:
      'Drag-to-publish schedules. Clock-in from the POS. Payroll that rolls up tips, hours, and deductions automatically. Compliance flags before they become violations.',
    ctaText: 'Open HR',
    image: '/images/features/hr.jpg',
  },
  {
    id: 'crm',
    category: 'CRM',
    title: 'Every customer, remembered.',
    description:
      'Every POS sale enriches the profile. Segment by spend, visit cadence, birthday, tags. Send the campaign from the same app — no Mailchimp detour required.',
    ctaText: 'Open CRM',
    image: '/images/features/crm.jpg',
  },
  {
    id: 'finance',
    category: 'Finance',
    title: 'A real P&L, in real time.',
    description:
      'Recurring invoices that collect themselves. Expenses captured at source. Sales tax by jurisdiction. A P&L that updates with every sale, not at month-end.',
    ctaText: 'Open Finance',
    image: '/images/features/finance.jpg',
  },
  {
    id: 'flow',
    category: 'Flow',
    title: 'Work that moves between apps.',
    description:
      'Kanban boards, automations, and tasks linked to the invoice, customer, or shift they\u2019re actually about. Operations that don\u2019t slip between tools.',
    ctaText: 'Open Flow',
    image: '/images/features/flow.jpg',
  },
  {
    id: 'insights',
    category: 'Insights',
    title: 'The dashboard that sees everything.',
    description:
      'Revenue trends, customer cohorts, product performance, labor margin. Every number pulled live from every other app. No SQL, no nightly ETL.',
    ctaText: 'Open Insights',
    image: '/images/features/insights.jpg',
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
    // Modulo wrap — Insights (last) rolls back to POS (first).
    setActiveIndex((i) => (i + 1) % FEATURES.length);
  }, []);

  // Auto-advance. Pauses while hovered so a visitor reading one
  // app's copy isn't yanked mid-sentence; leaving the section
  // resumes the cycle from whatever tab is currently selected.
  // Every click resets the timer too — the clicked tab always gets
  // its full spotlight window before the carousel rolls on.
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
    // Clicking a tab pins the carousel to that app and resets the
    // countdown. When the visitor is interested in a specific app,
    // they get 6 full seconds on it before the cycle resumes — no
    // getting yanked mid-sentence, no full pause either.
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
            Seven apps, one screen at a time.
            <br />
            <span className={styles.sectionTitleAccent}>See how each fits.</span>
          </h2>
        </motion.div>

        <motion.div className={styles.tabBar} variants={blurFadeIn} role="tablist">
          {FEATURES.map((f, i) => (
            <button
              key={f.id}
              type="button"
              role="tab"
              aria-selected={i === activeIndex}
              aria-controls={`feature-panel-${f.id}`}
              onClick={() => selectFeature(i)}
              className={`${styles.tab} ${i === activeIndex ? styles.tabActive : ''}`}
            >
              {f.category}
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
                initial={{ opacity: 0, y: 12 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -8 }}
                transition={{ duration: 0.32, ease: [0.16, 1, 0.3, 1] }}
              >
                <span className={styles.featureCategory}>{active.category}</span>
                <h3 className={styles.featureTitle}>{active.title}</h3>
                <p className={styles.featureDescription}>{active.description}</p>
                <div className={styles.featureActions}>
                  <Button
                    variant="dark-outline"
                    onClick={() => router.push('/demo')}
                    className={styles.featureButton}
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
                  className={`${styles.showcaseImage} ${i === activeIndex ? styles.active : ''}`}
                  aria-hidden={i !== activeIndex}
                >
                  {/* All images mounted from first paint — swap is a
                      CSS opacity transition rather than a network trip,
                      so the preview is instant once cached. */}
                  <img
                    src={f.image}
                    alt={f.title}
                    className={styles.showcaseImg}
                    loading={i === 0 ? 'eager' : 'lazy'}
                  />
                </div>
              ))}
            </div>

            <div className={styles.progressRow} aria-hidden="true">
              {FEATURES.map((f, i) => (
                <span
                  key={f.id}
                  className={`${styles.progressDot} ${i === activeIndex ? styles.progressDotActive : ''}`}
                />
              ))}
            </div>
          </div>
        </div>
      </div>
    </motion.section>
  );
};

export default FeaturesSection;
