import React, { useState, useEffect } from 'react';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';

interface StockItem {
  itemIndex: number;
  unitKey: 'btl' | 'bag' | 'kg' | 'L' | 'pcs';
  qty: number;
  par: number;
  supplier: string;
}

// itemIndex points at home.features.demos.erp.items[i] - the display name
// is translated, everything else here (unit, quantities, supplier) is
// presentational or a proper noun and stays local.
const ITEMS: StockItem[] = [
  { itemIndex: 0, unitKey: 'btl', qty: 5, par: 24, supplier: 'Rossini' },
  { itemIndex: 1, unitKey: 'bag', qty: 8, par: 10, supplier: 'GrainCo' },
  { itemIndex: 2, unitKey: 'kg', qty: 3, par: 15, supplier: 'Nomad' },
  { itemIndex: 3, unitKey: 'kg', qty: 6, par: 8, supplier: 'FreshSea' },
  { itemIndex: 4, unitKey: 'btl', qty: 0, par: 6, supplier: 'Rossini' },
  { itemIndex: 5, unitKey: 'L', qty: 14, par: 20, supplier: 'DairyHub' },
  { itemIndex: 6, unitKey: 'pcs', qty: 21, par: 30, supplier: 'Artisan' },
];

const statusOf = (qty: number, par: number): { key: string; color: string; bg: string } => {
  const r = qty / par;
  if (r === 0) return { key: 'out', color: '#ef4444', bg: 'rgba(239,68,68,0.12)' };
  if (r < 0.3) return { key: 'critical', color: '#f97316', bg: 'rgba(249,115,22,0.12)' };
  if (r < 0.55) return { key: 'low', color: '#eab308', bg: 'rgba(234,179,8,0.12)' };
  return { key: 'ok', color: '#22c55e', bg: 'rgba(34,197,94,0.1)' };
};

export const ERPDemo = ({ isActive }: { isActive: boolean }): React.JSX.Element => {
  const { t } = useTranslation();
  const [items, setItems] = useState<StockItem[]>(ITEMS);
  const itemNames = t('home.features.demos.erp.items', { returnObjects: true }) as string[];
  const [pulse, setPulse] = useState(false);

  useEffect(() => {
    if (!isActive) return;
    const timer = setInterval(() => {
      setItems((prev) =>
        prev.map((item) => {
          if (item.qty > 0 && Math.random() < 0.25) {
            return { ...item, qty: Math.max(0, item.qty - 1) };
          }
          return item;
        })
      );
      setPulse(true);
      setTimeout(() => setPulse(false), 600);
    }, 2500);
    return () => clearInterval(timer);
  }, [isActive]);

  const alerts = items.filter((i) => i.qty / i.par < 0.3).length;

  return (
    <div
      style={{
        position: 'absolute',
        inset: 0,
        background: 'var(--bg-surface-a)',
        fontFamily: 'inherit',
        fontSize: 13,
        display: 'flex',
        flexDirection: 'column',
        overflow: 'hidden',
      }}
    >
      {/* Header */}
      <div
        style={{
          padding: '16px 20px 14px',
          borderBottom: '1px solid var(--hairline)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          flexShrink: 0,
        }}
      >
        <div>
          <span style={{ color: 'var(--ink)', fontWeight: 600, fontSize: 14 }}>
            {t('home.features.demos.erp.inventory')}
          </span>
          <span style={{ color: 'rgba(19,19,19,0.5)', fontSize: 12, marginInlineStart: 10 }}>
            {t('home.features.demos.erp.itemCount', { count: items.length })}
          </span>
        </div>
        {alerts > 0 && (
          <div
            style={{
              background: 'rgba(249,115,22,0.12)',
              border: '1px solid rgba(249,115,22,0.3)',
              borderRadius: 999,
              padding: '3px 10px',
              fontSize: 11,
              fontWeight: 600,
              color: '#f97316',
            }}
          >
            {t('home.features.demos.erp.alertCount', { count: alerts })}
          </div>
        )}
      </div>

      {/* Table header */}
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: '1fr 100px 80px 80px',
          padding: '8px 20px',
          fontSize: 10,
          fontWeight: 600,
          letterSpacing: '0.1em',
          color: 'rgba(19,19,19,0.45)',
          textTransform: 'uppercase',
          borderBottom: '1px solid var(--hairline)',
          flexShrink: 0,
        }}
      >
        <span>{t('home.features.demos.erp.table.item')}</span>
        <span>{t('home.features.demos.erp.table.stock')}</span>
        <span style={{ textAlign: 'center' }}>{t('home.features.demos.erp.table.onHand')}</span>
        <span style={{ textAlign: 'center' }}>{t('home.features.demos.erp.table.status')}</span>
      </div>

      {/* Rows */}
      <div style={{ flex: 1, overflow: 'hidden' }}>
        {items.map((item, i) => {
          const s = statusOf(item.qty, item.par);
          const pct = Math.min(100, (item.qty / item.par) * 100);
          return (
            <div
              key={item.itemIndex}
              style={{
                display: 'grid',
                gridTemplateColumns: '1fr 100px 80px 80px',
                padding: '10px 20px',
                borderBottom: '1px solid var(--hairline)',
                alignItems: 'center',
                background: pulse && s.key !== 'ok' ? 'rgba(19,19,19,0.02)' : 'transparent',
                transition: 'background 0.4s ease',
              }}
            >
              <div>
                <div style={{ color: 'var(--ink)', fontSize: 12, fontWeight: 500 }}>
                  {itemNames[item.itemIndex]}
                </div>
                <div style={{ color: 'rgba(19,19,19,0.45)', fontSize: 11 }}>{item.supplier}</div>
              </div>
              <div style={{ paddingInlineEnd: 8 }}>
                <div
                  style={{
                    height: 4,
                    background: 'var(--hairline)',
                    borderRadius: 999,
                    overflow: 'hidden',
                  }}
                >
                  <motion.div
                    animate={{ width: `${pct}%` }}
                    transition={{ duration: 0.6, ease: 'easeOut' }}
                    style={{
                      height: '100%',
                      background: s.color,
                      borderRadius: 999,
                    }}
                  />
                </div>
              </div>
              <div style={{ textAlign: 'center', fontSize: 12, color: 'rgba(19,19,19,0.65)' }}>
                {/* Non-breaking space, not JSX whitespace: a quantity should
                    never wrap away from its unit, in any language. */}
                {item.qty}
                {' '}
                {t(`home.features.demos.erp.units.${item.unitKey}`)}
              </div>
              <div style={{ textAlign: 'center' }}>
                <span
                  style={{
                    background: s.bg,
                    color: s.color,
                    borderRadius: 999,
                    padding: '2px 8px',
                    fontSize: 10,
                    fontWeight: 600,
                  }}
                >
                  {t(`home.features.demos.erp.status.${s.key}`)}
                </span>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};
