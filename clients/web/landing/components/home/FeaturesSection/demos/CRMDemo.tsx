import React from 'react';
import { AnimatePresence, motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { cn } from '../../../../lib/cn';
import { PortalFrame, type PortalView } from './portal/PortalFrame';
import { money, useDemoStep } from './portal/usePortalDemo';
import styles from './portal/PortalDemo.module.scss';
import local from './CRMDemo.module.scss';

type Status = 'active' | 'atRisk' | 'new';

interface Customer {
  name: string;
  initials: string;
  color: string;
  status: Status;
  ltv: number;
  activity: number; // index into home.features.demos.crm.activity
  when: string;
}

const CUSTOMERS: Customer[] = [
  { name: 'Voxire', initials: 'VX', color: '#09090b', status: 'active', ltv: 48200, activity: 0, when: '2h' },
  { name: 'Northwind', initials: 'NW', color: '#3b82f6', status: 'active', ltv: 31900, activity: 1, when: '1d' },
  { name: 'Lumen Studio', initials: 'LS', color: '#f59e0b', status: 'atRisk', ltv: 12400, activity: 2, when: '3w' },
  { name: 'Harbor & Co', initials: 'HC', color: '#10b981', status: 'active', ltv: 9800, activity: 3, when: '2d' },
  { name: 'Kite Labs', initials: 'KL', color: '#8b5cf6', status: 'new', ltv: 1200, activity: 5, when: '5h' },
  { name: 'Atlas Freight', initials: 'AF', color: '#ef4444', status: 'atRisk', ltv: 7600, activity: 4, when: '1mo' },
  { name: 'Orbit Media', initials: 'OM', color: '#64c5b9', status: 'active', ltv: 6400, activity: 1, when: '4d' },
];

const TONE: Record<Status, string> = { active: 'green', atRisk: 'red', new: 'accent' };
// Segment tabs, in order: All, Active, At risk, New.
const FILTERS: (Status | null)[] = [null, 'active', 'atRisk', 'new'];

export const CRMDemo = ({
  isActive,
  onNavigate,
}: {
  isActive: boolean;
  onNavigate?: (view: PortalView) => void;
}): React.JSX.Element => {
  const { t } = useTranslation();
  const step = useDemoStep(isActive, 3, 2200, 0);
  const segments = t('home.features.demos.crm.segments', { returnObjects: true }) as string[];
  const cols = t('home.features.demos.crm.cols', { returnObjects: true }) as string[];
  const activity = t('home.features.demos.crm.activity', { returnObjects: true }) as string[];
  const filter = FILTERS[step];
  const rows = CUSTOMERS.filter((c) => !filter || c.status === filter);

  return (
    <PortalFrame
      onNavigate={onNavigate}
      app="crm"
      crumb={t('home.features.demos.crm.crumb', { returnObjects: true }) as string[]}
    >
      <div className={styles.spread}>
        <span className={styles.segmented}>
          {segments.map((s, i) => (
            <span key={s} className={cn(styles.segment, i === step && styles.segmentOn)}>
              {s}
            </span>
          ))}
        </span>
        <span className={styles.muted}>
          {t('home.features.demos.crm.total', { count: rows.length })}
        </span>
      </div>

      <table className={cn(styles.table, local.table)}>
        <thead>
          <tr>
            <th>{cols[0]}</th>
            <th>{cols[1]}</th>
            <th className={styles.num}>{cols[2]}</th>
            <th className={styles.num}>{cols[3]}</th>
          </tr>
        </thead>
        <tbody>
          <AnimatePresence initial={false} mode="popLayout">
            {rows.map((c) => (
              <motion.tr
                key={c.name}
                layout
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                transition={{ duration: 0.3 }}
              >
                <td>
                  <span className={styles.row}>
                    <span className={styles.avatar} style={{ background: c.color }}>
                      {c.initials}
                    </span>
                    <span className={local.name}>{c.name}</span>
                  </span>
                </td>
                <td>
                  <span className={styles.pill} data-tone={TONE[c.status]}>
                    {t(`home.features.demos.crm.status.${c.status}`)}
                  </span>
                </td>
                <td className={styles.num}>{money(c.ltv)}</td>
                <td className={cn(styles.num, styles.muted)}>
                  {activity[c.activity]} · {c.when}
                </td>
              </motion.tr>
            ))}
          </AnimatePresence>
        </tbody>
      </table>
    </PortalFrame>
  );
};
