import React, { useState } from 'react';
import { motion, AnimatePresence } from 'motion/react';
import { useTranslation } from 'react-i18next';

import { formatMoney } from '../../../../lib/formatMoney';

type Segment = 'all' | 'vip' | 'regular' | 'atrisk' | 'new';

interface Customer {
  initials: string;
  name: string;
  segment: Exclude<Segment, 'all'>;
  visits: number;
  spend: number;
  lastKey:
    | 'yesterday'
    | 'oneDayAgo'
    | 'twoDaysAgo'
    | 'threeDaysAgo'
    | 'fiveDaysAgo'
    | 'oneWeekAgo'
    | 'threeWeeksAgo'
    | 'oneMonthAgo';
}

const CUSTOMERS: Customer[] = [
  {
    initials: 'EA',
    name: 'Eliza Alvarez',
    segment: 'vip',
    visits: 20,
    spend: 4200,
    lastKey: 'twoDaysAgo' as const,
  },
  {
    initials: 'MB',
    name: 'Marco Bianchi',
    segment: 'vip',
    visits: 15,
    spend: 3100,
    lastKey: 'oneDayAgo' as const,
  },
  {
    initials: 'YK',
    name: 'Yasmine Khoury',
    segment: 'regular',
    visits: 8,
    spend: 890,
    lastKey: 'fiveDaysAgo' as const,
  },
  { initials: 'TP', name: 'Tom Park', segment: 'regular', visits: 6, spend: 620, lastKey: 'oneWeekAgo' as const },
  {
    initials: 'SN',
    name: 'Sofia Nour',
    segment: 'atrisk',
    visits: 12,
    spend: 1440,
    lastKey: 'threeWeeksAgo' as const,
  },
  {
    initials: 'RD',
    name: 'Rami Darwish',
    segment: 'atrisk',
    visits: 9,
    spend: 970,
    lastKey: 'oneMonthAgo' as const,
  },
  { initials: 'LW', name: 'Lily Wang', segment: 'new', visits: 1, spend: 48, lastKey: 'yesterday' as const },
  { initials: 'JH', name: 'Jonas Hall', segment: 'new', visits: 2, spend: 95, lastKey: 'threeDaysAgo' as const },
];

// labelKey matches home.features.demos.crm.segments.<labelKey> - one
// translated set shared by the filter tabs and the row tier badge, so the
// two cannot drift the way two copies of '★ VIP' eventually would.
const SEGMENTS: { id: Segment; labelKey: string; color: string }[] = [
  { id: 'all', labelKey: 'all', color: 'var(--ink)' },
  { id: 'vip', labelKey: 'vip', color: '#b45309' },
  { id: 'regular', labelKey: 'regular', color: '#2563eb' },
  { id: 'atrisk', labelKey: 'atRisk', color: '#dc2626' },
  { id: 'new', labelKey: 'new', color: '#16a34a' },
];

const TIER: Record<Exclude<Segment, 'all'>, { labelKey: string; color: string; bg: string }> = {
  vip: { labelKey: 'vip', color: '#b45309', bg: 'rgba(251,191,36,0.15)' },
  regular: { labelKey: 'regular', color: '#2563eb', bg: 'rgba(96,165,250,0.15)' },
  atrisk: { labelKey: 'atRisk', color: '#dc2626', bg: 'rgba(248,113,113,0.15)' },
  new: { labelKey: 'new', color: '#16a34a', bg: 'rgba(52,211,153,0.15)' },
};

export const CRMDemo = ({ isActive }: { isActive: boolean }): React.JSX.Element => {
  const { t, i18n } = useTranslation();
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
          <span style={{ color: 'var(--ink)', fontWeight: 600, fontSize: 14 }}>
            {t('home.features.demos.crm.customers')}
          </span>
          <span style={{ color: 'rgba(19,19,19,0.5)', fontSize: 12 }}>
            {t('home.features.demos.crm.totalCount')}
          </span>
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
              {t(`home.features.demos.crm.segments.${seg.labelKey}`)}
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
        <span>{t('home.features.demos.crm.table.customer')}</span>
        <span style={{ textAlign: 'center' }}>{t('home.features.demos.crm.table.tier')}</span>
        <span style={{ textAlign: 'end' }}>{t('home.features.demos.crm.table.spend')}</span>
        <span style={{ textAlign: 'end' }}>{t('home.features.demos.crm.table.lastVisit')}</span>
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
                    <div style={{ fontSize: 10, color: 'rgba(19,19,19,0.5)' }}>
                      {t('home.features.demos.crm.visitCount', { count: c.visits })}
                    </div>
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
                    {t(`home.features.demos.crm.segments.${tier.labelKey}`)}
                  </span>
                </div>
                <div
                  style={{
                    textAlign: 'end',
                    fontSize: 12,
                    fontWeight: 600,
                    color: 'var(--ink)',
                  }}
                >
                  {formatMoney(c.spend, i18n.language, 0)}
                </div>
                <div
                  style={{
                    textAlign: 'end',
                    fontSize: 11,
                    color: 'rgba(19,19,19,0.5)',
                  }}
                >
                  {t(`home.features.demos.crm.relativeTime.${c.lastKey}`)}
                </div>
              </motion.div>
            );
          })}
        </AnimatePresence>
      </div>
    </div>
  );
};
