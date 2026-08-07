import React, { useState, useEffect } from 'react';
import { motion } from 'motion/react';

interface StockItem {
  name: string;
  unit: string;
  qty: number;
  par: number;
  supplier: string;
}

const ITEMS: StockItem[] = [
  { name: 'Olive Oil 5L', unit: 'btl', qty: 5, par: 24, supplier: 'Rossini' },
  { name: 'Flour 25kg', unit: 'bag', qty: 8, par: 10, supplier: 'GrainCo' },
  { name: 'Espresso Beans', unit: 'kg', qty: 3, par: 15, supplier: 'Nomad' },
  { name: 'Atlantic Salmon', unit: 'kg', qty: 6, par: 8, supplier: 'FreshSea' },
  { name: 'Truffle Oil', unit: 'btl', qty: 0, par: 6, supplier: 'Rossini' },
  { name: 'Heavy Cream', unit: 'L', qty: 14, par: 20, supplier: 'DairyHub' },
  { name: 'Sourdough', unit: 'pcs', qty: 21, par: 30, supplier: 'Artisan' },
];

const status = (qty: number, par: number): { label: string; color: string; bg: string } => {
  const r = qty / par;
  if (r === 0) return { label: 'Out', color: '#ef4444', bg: 'rgba(239,68,68,0.12)' };
  if (r < 0.3) return { label: 'Critical', color: '#f97316', bg: 'rgba(249,115,22,0.12)' };
  if (r < 0.55) return { label: 'Low', color: '#eab308', bg: 'rgba(234,179,8,0.12)' };
  return { label: 'OK', color: '#22c55e', bg: 'rgba(34,197,94,0.1)' };
};

export const ERPDemo = ({ isActive }: { isActive: boolean }): React.JSX.Element => {
  const [items, setItems] = useState<StockItem[]>(ITEMS);
  const [pulse, setPulse] = useState(false);

  useEffect(() => {
    if (!isActive) return;
    const t = setInterval(() => {
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
    return () => clearInterval(t);
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
          <span style={{ color: 'var(--ink)', fontWeight: 600, fontSize: 14 }}>Inventory</span>
          <span style={{ color: 'rgba(19,19,19,0.5)', fontSize: 12, marginLeft: 10 }}>
            {items.length} items
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
            {alerts} alert{alerts !== 1 ? 's' : ''}
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
        <span>Item</span>
        <span>Stock</span>
        <span style={{ textAlign: 'center' }}>On hand</span>
        <span style={{ textAlign: 'center' }}>Status</span>
      </div>

      {/* Rows */}
      <div style={{ flex: 1, overflow: 'hidden' }}>
        {items.map((item, i) => {
          const s = status(item.qty, item.par);
          const pct = Math.min(100, (item.qty / item.par) * 100);
          return (
            <div
              key={item.name}
              style={{
                display: 'grid',
                gridTemplateColumns: '1fr 100px 80px 80px',
                padding: '10px 20px',
                borderBottom: '1px solid var(--hairline)',
                alignItems: 'center',
                background: pulse && s.label !== 'OK' ? 'rgba(19,19,19,0.02)' : 'transparent',
                transition: 'background 0.4s ease',
              }}
            >
              <div>
                <div style={{ color: 'var(--ink)', fontSize: 12, fontWeight: 500 }}>
                  {item.name}
                </div>
                <div style={{ color: 'rgba(19,19,19,0.45)', fontSize: 11 }}>{item.supplier}</div>
              </div>
              <div style={{ paddingRight: 8 }}>
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
                {item.qty} {item.unit}
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
                  {s.label}
                </span>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};
