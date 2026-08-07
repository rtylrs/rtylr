import React, { useState } from 'react';
import { motion, AnimatePresence } from 'motion/react';

type Segment = 'all' | 'vip' | 'regular' | 'atrisk' | 'new';

interface Customer {
  initials: string;
  name: string;
  segment: Exclude<Segment, 'all'>;
  visits: number;
  spend: number;
  last: string;
}

const CUSTOMERS: Customer[] = [
  {
    initials: 'EA',
    name: 'Eliza Alvarez',
    segment: 'vip',
    visits: 20,
    spend: 4200,
    last: '2d ago',
  },
  {
    initials: 'MB',
    name: 'Marco Bianchi',
    segment: 'vip',
    visits: 15,
    spend: 3100,
    last: '1d ago',
  },
  {
    initials: 'YK',
    name: 'Yasmine Khoury',
    segment: 'regular',
    visits: 8,
    spend: 890,
    last: '5d ago',
  },
  { initials: 'TP', name: 'Tom Park', segment: 'regular', visits: 6, spend: 620, last: '1w ago' },
  {
    initials: 'SN',
    name: 'Sofia Nour',
    segment: 'atrisk',
    visits: 12,
    spend: 1440,
    last: '3w ago',
  },
  {
    initials: 'RD',
    name: 'Rami Darwish',
    segment: 'atrisk',
    visits: 9,
    spend: 970,
    last: '1mo ago',
  },
  { initials: 'LW', name: 'Lily Wang', segment: 'new', visits: 1, spend: 48, last: 'Yesterday' },
  { initials: 'JH', name: 'Jonas Hall', segment: 'new', visits: 2, spend: 95, last: '3d ago' },
];

const SEGMENTS: { id: Segment; label: string; color: string }[] = [
  { id: 'all', label: 'All', color: 'var(--ink)' },
  { id: 'vip', label: '★ VIP', color: '#b45309' },
  { id: 'regular', label: 'Regular', color: '#2563eb' },
  { id: 'atrisk', label: 'At risk', color: '#dc2626' },
  { id: 'new', label: 'New', color: '#16a34a' },
];

const TIER: Record<Exclude<Segment, 'all'>, { label: string; color: string; bg: string }> = {
  vip: { label: '★ VIP', color: '#b45309', bg: 'rgba(251,191,36,0.15)' },
  regular: { label: 'Regular', color: '#2563eb', bg: 'rgba(96,165,250,0.15)' },
  atrisk: { label: 'At risk', color: '#dc2626', bg: 'rgba(248,113,113,0.15)' },
  new: { label: 'New', color: '#16a34a', bg: 'rgba(52,211,153,0.15)' },
};

export const CRMDemo = ({ isActive }: { isActive: boolean }): React.JSX.Element => {
  const [active, setActive] = useState<Segment>('all');

  const visible = active === 'all' ? CUSTOMERS : CUSTOMERS.filter((c) => c.segment === active);

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
          flexShrink: 0,
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            marginBottom: 12,
          }}
        >
          <span style={{ color: 'var(--ink)', fontWeight: 600, fontSize: 14 }}>Customers</span>
          <span style={{ color: 'rgba(19,19,19,0.5)', fontSize: 12 }}>1,204 total</span>
        </div>
        {/* Segment tabs */}
        <div style={{ display: 'flex', gap: 6, overflowX: 'auto' }}>
          {SEGMENTS.map((seg) => (
            <button
              key={seg.id}
              onClick={() => setActive(seg.id)}
              style={{
                flexShrink: 0,
                background: active === seg.id ? 'var(--bg-surface-b)' : 'transparent',
                border: `1px solid ${active === seg.id ? 'var(--hairline-strong)' : 'var(--hairline)'}`,
                borderRadius: 999,
                padding: '5px 12px',
                fontSize: 11,
                fontWeight: 500,
                color: active === seg.id ? seg.color : 'rgba(19,19,19,0.5)',
                cursor: 'pointer',
                transition: 'all 0.15s ease',
              }}
            >
              {seg.label}
            </button>
          ))}
        </div>
      </div>

      {/* Column headers */}
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: '1fr 70px 90px 70px',
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
        <span>Customer</span>
        <span style={{ textAlign: 'center' }}>Tier</span>
        <span style={{ textAlign: 'right' }}>Spend</span>
        <span style={{ textAlign: 'right' }}>Last visit</span>
      </div>

      {/* Rows */}
      <div style={{ flex: 1, overflow: 'hidden' }}>
        <AnimatePresence mode="popLayout">
          {visible.map((c) => {
            const tier = TIER[c.segment];
            return (
              <motion.div
                key={c.initials}
                layout
                initial={{ opacity: 0, y: 6 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -4 }}
                transition={{ duration: 0.22 }}
                style={{
                  display: 'grid',
                  gridTemplateColumns: '1fr 70px 90px 70px',
                  padding: '10px 20px',
                  borderBottom: '1px solid var(--hairline)',
                  alignItems: 'center',
                  cursor: 'pointer',
                }}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                  <div
                    style={{
                      width: 30,
                      height: 30,
                      borderRadius: '50%',
                      background: tier.bg,
                      border: `1px solid ${tier.color}44`,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      fontSize: 10,
                      fontWeight: 700,
                      color: tier.color,
                      flexShrink: 0,
                    }}
                  >
                    {c.initials}
                  </div>
                  <div>
                    <div style={{ fontSize: 12, fontWeight: 500, color: 'var(--ink)' }}>
                      {c.name}
                    </div>
                    <div style={{ fontSize: 10, color: 'rgba(19,19,19,0.5)' }}>{c.visits} visits</div>
                  </div>
                </div>
                <div style={{ textAlign: 'center' }}>
                  <span
                    style={{
                      background: tier.bg,
                      color: tier.color,
                      borderRadius: 999,
                      padding: '2px 7px',
                      fontSize: 10,
                      fontWeight: 600,
                    }}
                  >
                    {tier.label}
                  </span>
                </div>
                <div
                  style={{
                    textAlign: 'right',
                    fontSize: 12,
                    fontWeight: 600,
                    color: 'var(--ink)',
                  }}
                >
                  ${c.spend.toLocaleString()}
                </div>
                <div
                  style={{
                    textAlign: 'right',
                    fontSize: 11,
                    color: 'rgba(19,19,19,0.5)',
                  }}
                >
                  {c.last}
                </div>
              </motion.div>
            );
          })}
        </AnimatePresence>
      </div>
    </div>
  );
};
