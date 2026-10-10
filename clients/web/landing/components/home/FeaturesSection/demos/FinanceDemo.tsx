import React from 'react';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { PortalFrame, type PortalView } from './portal/PortalFrame';
import { money, useCountUp, useDemoStep } from './portal/usePortalDemo';
import styles from './portal/PortalDemo.module.scss';
import local from './FinanceDemo.module.scss';

const KPIS = [
  { value: 48200, delta: '+12.4%' },
  { value: 31650, delta: '+3.1%' },
  { value: 16550, delta: '+31.8%' },
];

// Thousands, money in / money out per month.
const FLOW = [
  [32, 24],
  [36, 27],
  [34, 29],
  [41, 28],
  [44, 30],
  [48, 32],
];
const MAX = 50;

const INVOICES = [
  { id: 'INV-1042', who: 'Voxire', amount: 4200, status: 'paid' },
  { id: 'INV-1043', who: 'Northwind', amount: 1850, status: 'due' },
];
// The overdue invoice the demo resolves: overdue → reminder sent → paid.
const RESOLVING = { id: 'INV-1039', who: 'Lumen Studio', amount: 2600 };
const RESOLVE_STEPS = ['overdue', 'sent', 'paid'] as const;
const TONE: Record<string, string> = { paid: 'green', due: 'neutral', overdue: 'red', sent: 'amber' };

const Kpi = ({ i, label, isActive }: { i: number; label: string; isActive: boolean }) => {
  const { t } = useTranslation();
  const value = useCountUp(KPIS[i].value, isActive);
  return (
    <div className={styles.kpi}>
      <div className={styles.kpiLabel}>{label}</div>
      <div className={styles.kpiValue}>{money(value)}</div>
      <div className={styles.delta}>
        {KPIS[i].delta} <span className={styles.muted}>{t('home.features.demos.finance.vsLast')}</span>
      </div>
    </div>
  );
};

export const FinanceDemo = ({
  isActive,
  onNavigate,
}: {
  isActive: boolean;
  onNavigate?: (view: PortalView) => void;
}): React.JSX.Element => {
  const { t } = useTranslation();
  const step = useDemoStep(isActive, RESOLVE_STEPS.length, 1900);
  const labels = t('home.features.demos.finance.kpis', { returnObjects: true }) as string[];
  const months = t('home.features.demos.finance.months', { returnObjects: true }) as string[];
  const status = RESOLVE_STEPS[step];
  const rows = [{ ...RESOLVING, status }, ...INVOICES];

  return (
    <PortalFrame
      onNavigate={onNavigate}
      app="finance"
      crumb={t('home.features.demos.finance.crumb', { returnObjects: true }) as string[]}
    >
      <div className={local.layout}>
        <div className={styles.kpis}>
          {labels.map((label, i) => (
            <Kpi key={label} i={i} label={label} isActive={isActive} />
          ))}
        </div>

        <div className={local.bottom}>
          <div className={`${styles.card} ${local.chartCard}`}>
            <div className={styles.spread}>
              <span className={local.cardTitle}>{t('home.features.demos.finance.chart')}</span>
              <span className={styles.row}>
                <span className={local.legendIn} /> {t('home.features.demos.finance.in')}
                <span className={local.legendOut} /> {t('home.features.demos.finance.out')}
              </span>
            </div>
            <div className={local.chart}>
              {FLOW.map(([moneyIn, moneyOut], m) => (
                <div key={months[m]} className={local.month}>
                  <div className={local.bars}>
                    {[moneyIn, moneyOut].map((v, k) => (
                      <motion.span
                        key={k}
                        className={k === 0 ? local.barIn : local.barOut}
                        style={{ height: `${(v / MAX) * 100}%` }}
                        initial={false}
                        animate={isActive ? { scaleY: [0, 1] } : { scaleY: 1 }}
                        transition={{ duration: 0.7, delay: 0.15 + m * 0.06, ease: [0.16, 1, 0.3, 1] }}
                      />
                    ))}
                  </div>
                  <span className={styles.muted}>{months[m]}</span>
                </div>
              ))}
            </div>
          </div>

          <div className={`${styles.card} ${local.invoices}`}>
            <span className={local.cardTitle}>{t('home.features.demos.finance.invoices')}</span>
            {rows.map((inv) => (
              <div key={inv.id} className={local.invoice}>
                <div>
                  <div className={local.invoiceWho}>{inv.who}</div>
                  <div className={styles.muted}>{inv.id}</div>
                </div>
                <div className={local.invoiceRight}>
                  <div className={local.invoiceAmount}>{money(inv.amount)}</div>
                  <motion.span
                    key={inv.status}
                    className={styles.pill}
                    data-tone={TONE[inv.status]}
                    initial={{ opacity: 0, y: 3 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ duration: 0.3 }}
                  >
                    {t(`home.features.demos.finance.status.${inv.status}`)}
                  </motion.span>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </PortalFrame>
  );
};
