import React from 'react';
import { AnimatePresence, motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { cn } from '../../../../lib/cn';
import { CaretMark } from '../../../CaretMark/CaretMark';
import { PortalFrame, type PortalView } from './portal/PortalFrame';
import { money, useCountUp, useDemoStep } from './portal/usePortalDemo';
import styles from './portal/PortalDemo.module.scss';
import local from './InsightsDemo.module.scss';

const KPIS = [
  { value: 98400, format: (n: number) => money(n), delta: '+12%' },
  { value: 142, format: (n: number) => Math.round(n).toString(), delta: '+8%' },
  { value: 693, format: (n: number) => money(n), delta: '+4%' },
];

// Daily revenue in thousands, Mon..Sun. Tuesday is the dip Caret explains.
const SERIES = [12.1, 9.4, 15.8, 13.2, 16.9, 18.4, 12.6];
const DIP = 1;
// Close to the rendered plot size, so the stretched viewBox barely distorts
// the stroke (vector-effect would fix that but breaks pathLength drawing).
const W = 600;
const H = 160;
const MAX = 20;
const points = SERIES.map((v, i) => [(i / (SERIES.length - 1)) * W, H - (v / MAX) * H] as const);
const line = points.map(([x, y], i) => `${i ? 'L' : 'M'}${x},${y}`).join(' ');
const area = `${line} L${W},${H} L0,${H} Z`;

const Kpi = ({ i, label, isActive }: { i: number; label: string; isActive: boolean }) => {
  const value = useCountUp(KPIS[i].value, isActive);
  return (
    <div className={styles.kpi}>
      <div className={styles.kpiLabel}>{label}</div>
      <div className={styles.kpiValue}>{KPIS[i].format(value)}</div>
      <div className={styles.delta}>{KPIS[i].delta}</div>
    </div>
  );
};

export const InsightsDemo = ({
  isActive,
  onNavigate,
}: {
  isActive: boolean;
  onNavigate?: (view: PortalView) => void;
}): React.JSX.Element => {
  const { t } = useTranslation();
  const step = useDemoStep(isActive, 3, 2000);
  const labels = t('home.features.demos.insights.kpis', { returnObjects: true }) as string[];
  const ranges = t('home.features.demos.insights.ranges', { returnObjects: true }) as string[];
  const days = t('home.features.demos.insights.days', { returnObjects: true }) as string[];
  const answered = step === 2;

  return (
    <PortalFrame
      onNavigate={onNavigate}
      app="insights"
      crumb={t('home.features.demos.insights.crumb', { returnObjects: true }) as string[]}
      actions={
        <span className={styles.segmented}>
          {ranges.map((r, i) => (
            <span key={r} className={cn(styles.segment, i === 0 && styles.segmentOn)}>
              {r}
            </span>
          ))}
        </span>
      }
    >
      <div className={local.layout}>
        <div className={styles.kpis}>
          {labels.map((label, i) => (
            <Kpi key={label} i={i} label={label} isActive={isActive} />
          ))}
        </div>

        <div className={cn(styles.card, local.chartCard)}>
          <div className={local.plot}>
            <svg className={local.chart} viewBox={`0 0 ${W} ${H}`} preserveAspectRatio="none">
              <defs>
                <linearGradient id="insights-fill" x1="0" x2="0" y1="0" y2="1">
                  <stop offset="0%" stopColor="var(--p-accent)" stopOpacity="0.22" />
                  <stop offset="100%" stopColor="var(--p-accent)" stopOpacity="0" />
                </linearGradient>
              </defs>
              <motion.path
                d={area}
                fill="url(#insights-fill)"
                initial={false}
                animate={{ opacity: isActive ? [0, 1] : 1 }}
                transition={{ duration: 0.8, delay: 0.6 }}
              />
              <motion.path
                d={line}
                fill="none"
                stroke="var(--p-accent)"
                strokeWidth="2"
                strokeLinejoin="round"
                initial={false}
                animate={{ pathLength: isActive ? [0, 1] : 1 }}
                transition={{ duration: 1.2, ease: [0.16, 1, 0.3, 1] }}
              />
            </svg>
            {/* Dot over the dip; HTML so it stays round on a stretched chart. */}
            <span
              className={cn(local.dip, answered && local.dipOn)}
              style={{ left: `${(points[DIP][0] / W) * 100}%`, top: `${(points[DIP][1] / H) * 100}%` }}
            />
          </div>
          <div className={local.days}>
            {days.map((d, i) => (
              <span key={d} className={i === DIP && answered ? local.dayOn : undefined}>
                {d}
              </span>
            ))}
          </div>
        </div>

        <div className={local.ask}>
          <AnimatePresence initial={false}>
            {step >= 1 && (
              <motion.div
                key="q"
                className={local.question}
                initial={{ opacity: 0, y: 6 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0 }}
              >
                {t('home.features.demos.insights.question')}
              </motion.div>
            )}
            {answered && (
              <motion.div
                key="a"
                className={local.answer}
                initial={{ opacity: 0, y: 6 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0 }}
              >
                <span className={local.caret}>
                  <CaretMark size={12} />
                </span>
                {t('home.features.demos.insights.answer')}
              </motion.div>
            )}
          </AnimatePresence>
        </div>
      </div>
    </PortalFrame>
  );
};
