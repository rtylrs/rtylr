import React from 'react';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { cn } from '../../../../lib/cn';
import { PortalFrame, type PortalView } from './portal/PortalFrame';
import { useDemoStep } from './portal/usePortalDemo';
import styles from './portal/PortalDemo.module.scss';
import local from './ERPDemo.module.scss';

type Status = 'ok' | 'low' | 'out' | 'reorder';

interface Item {
  name: string;
  sku: string;
  stock: number;
  max: number;
  unit: string;
}

const ITEMS: Item[] = [
  { name: 'Shipping boxes, M', sku: 'PK-204', stock: 420, max: 600, unit: 'pcs' },
  { name: 'Label rolls', sku: 'PK-118', stock: 64, max: 120, unit: 'rolls' },
  { name: 'Printer paper', sku: 'OF-031', stock: 38, max: 200, unit: 'reams' },
  { name: 'Packing tape', sku: 'PK-077', stock: 150, max: 240, unit: 'rolls' },
  { name: 'Sample kits', sku: 'SL-012', stock: 0, max: 50, unit: 'kits' },
  { name: 'Branded mugs', sku: 'MR-009', stock: 210, max: 300, unit: 'pcs' },
];

// The item the demo follows: stock drains, it turns Low, a reorder goes out.
const WATCHED = 2;
const WATCHED_STOCK = [74, 22, 22];
const WATCHED_STATUS: Status[] = ['ok', 'low', 'reorder'];
const TONE: Record<Status, string> = { ok: 'green', low: 'amber', out: 'red', reorder: 'accent' };

const statusOf = (item: Item): Status =>
  item.stock === 0 ? 'out' : item.stock / item.max < 0.25 ? 'low' : 'ok';

export const ERPDemo = ({
  isActive,
  onNavigate,
}: {
  isActive: boolean;
  onNavigate?: (view: PortalView) => void;
}): React.JSX.Element => {
  const { t } = useTranslation();
  const step = useDemoStep(isActive, WATCHED_STATUS.length, 2000);
  const cols = t('home.features.demos.erp.cols', { returnObjects: true }) as string[];
  const crumb = t('home.features.demos.erp.crumb', { returnObjects: true }) as string[];

  const rows = ITEMS.map((item, i) =>
    i === WATCHED
      ? { ...item, stock: WATCHED_STOCK[step], status: WATCHED_STATUS[step] }
      : { ...item, status: statusOf(item) },
  );
  const alerts = rows.filter((r) => r.status !== 'ok').length;

  return (
    <PortalFrame
      onNavigate={onNavigate}
      app="erp"
      crumb={crumb}
    >
      <div className={styles.spread}>
        <span className={styles.title}>{crumb[1]}</span>
        <span className={styles.pill} data-tone="amber">
          {t('home.features.demos.erp.alerts', { count: alerts })}
        </span>
      </div>

      <table className={cn(styles.table, local.table)}>
        <thead>
          <tr>
            <th>{cols[0]}</th>
            <th>{cols[1]}</th>
            <th className={styles.num}>{cols[2]}</th>
            <th>{cols[3]}</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((r, i) => {
            const ratio = r.stock / r.max;
            const barColor =
              r.status === 'out' || ratio < 0.15
                ? 'var(--p-red)'
                : ratio < 0.25
                  ? 'var(--p-amber)'
                  : 'var(--p-text)';
            return (
              <tr key={r.sku} className={i === WATCHED ? local.watched : undefined}>
                <td>
                  <div className={local.name}>{r.name}</div>
                  <div className={styles.muted}>{r.sku}</div>
                </td>
                <td className={local.barCell}>
                  <div className={styles.bar}>
                    <span style={{ width: `${Math.max(ratio * 100, 0)}%`, background: barColor }} />
                  </div>
                </td>
                <td className={styles.num}>
                  {r.stock} <span className={styles.muted}>{r.unit}</span>
                </td>
                <td>
                  <motion.span
                    key={r.status}
                    className={styles.pill}
                    data-tone={TONE[r.status]}
                    initial={{ opacity: 0, y: 3 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ duration: 0.3 }}
                  >
                    {t(`home.features.demos.erp.status.${r.status}`)}
                  </motion.span>
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </PortalFrame>
  );
};
