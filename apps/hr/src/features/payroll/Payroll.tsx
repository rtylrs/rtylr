import React from 'react';
import { motion } from 'motion/react';
import { Play, DollarSign, TrendingUp, Users, Download } from 'lucide-react';
import { PayrollRow } from '../../shared/services/hrApi';
import { usePayroll } from './usePayroll';
import { useIsMobile } from '../../shared/components/ui';

const ACCENT = '#8b5cf6';

function netPay(row: PayrollRow): number {
  return row.basePay + row.overtime - row.deductions;
}

function fmt(n: number): string {
  return `$${n.toLocaleString('en-US', { minimumFractionDigits: 0 })}`;
}

function getInitials(name: string) {
  return name
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);
}

const container = { hidden: {}, show: { transition: { staggerChildren: 0.05 } } };
const rowAnim = {
  hidden: { opacity: 0, x: -6 },
  show: { opacity: 1, x: 0, transition: { duration: 0.18 } },
};

const Payroll: React.FC = () => {
  const isMobile = useIsMobile(900);
  const { data: fetchedPayroll } = usePayroll();
  const payrollData: PayrollRow[] = fetchedPayroll || [];

  const totalBase = payrollData.reduce((s, r) => s + r.basePay, 0);
  const totalOvertime = payrollData.reduce((s, r) => s + r.overtime, 0);
  const totalDeductions = payrollData.reduce((s, r) => s + r.deductions, 0);
  const totalNet = payrollData.reduce((s, r) => s + netPay(r), 0);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
      {/* Header */}
      <div
        style={{
          display: 'flex',
          alignItems: 'flex-start',
          justifyContent: 'space-between',
          gap: 12,
          flexWrap: 'wrap',
        }}
      >
        <div>
          <h2 style={{ margin: 0, fontSize: '18px', fontWeight: 700, color: 'var(--hr-text)' }}>
            Payroll
          </h2>
          <p style={{ margin: '4px 0 0', fontSize: '13px', color: 'var(--hr-text-muted)' }}>
            Pay period: Apr 1 – Apr 30, 2026
          </p>
        </div>
        <div
          style={{
            display: 'flex',
            gap: '8px',
            flexWrap: 'wrap',
            width: isMobile ? '100%' : 'auto',
          }}
        >
          <button
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '8px',
              padding: '9px 16px',
              borderRadius: '8px',
              background: 'var(--hr-surface)',
              border: '1px solid var(--hr-border)',
              color: 'var(--hr-text-muted)',
              fontSize: '13px',
              fontWeight: 600,
              cursor: 'pointer',
            }}
          >
            <Download size={14} />
            Export
          </button>
          <button
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '8px',
              padding: '9px 16px',
              borderRadius: '8px',
              background: ACCENT,
              border: 'none',
              color: '#fff',
              fontSize: '13px',
              fontWeight: 600,
              cursor: 'pointer',
            }}
          >
            <Play size={14} />
            Run Payroll
          </button>
        </div>
      </div>

      {/* Summary cards */}
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))',
          gap: '12px',
        }}
      >
        {[
          {
            label: 'Total Base Pay',
            value: fmt(totalBase),
            icon: DollarSign,
            color: ACCENT,
            bg: 'rgba(139,92,246,0.12)',
          },
          {
            label: 'Total Overtime',
            value: fmt(totalOvertime),
            icon: TrendingUp,
            color: '#f59e0b',
            bg: 'rgba(245,158,11,0.12)',
          },
          {
            label: 'Total Deductions',
            value: fmt(totalDeductions),
            icon: Users,
            color: '#ef4444',
            bg: 'rgba(239,68,68,0.12)',
          },
          {
            label: 'Net Payroll',
            value: fmt(totalNet),
            icon: DollarSign,
            color: '#10b981',
            bg: 'rgba(16,185,129,0.12)',
          },
        ].map((card) => (
          <motion.div
            key={card.label}
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.2 }}
            style={{
              background: card.bg,
              border: `1px solid ${card.color}30`,
              borderRadius: '12px',
              padding: '16px 18px',
              display: 'flex',
              flexDirection: 'column',
              gap: '10px',
            }}
          >
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <span style={{ fontSize: '12px', fontWeight: 500, color: card.color }}>
                {card.label}
              </span>
              <card.icon size={16} style={{ color: card.color }} />
            </div>
            <p style={{ margin: 0, fontSize: '22px', fontWeight: 700, color: card.color }}>
              {card.value}
            </p>
          </motion.div>
        ))}
      </div>

      {/* Table */}
      <motion.div
        variants={container}
        initial="hidden"
        animate="show"
        style={{
          background: 'var(--hr-card)',
          border: '1px solid var(--hr-border)',
          borderRadius: '12px',
          overflow: 'hidden',
        }}
      >
        <div style={{ overflowX: 'auto' }}>
          {/* Table header */}
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: '2fr 1fr 1fr 1fr 1fr 1fr',
              minWidth: 760,
              padding: '12px 20px',
              borderBottom: '1px solid var(--hr-border)',
              background: 'var(--hr-hover)',
            }}
          >
            {['Employee', 'Hours', 'Base Pay', 'Overtime', 'Deductions', 'Net Pay'].map((col) => (
              <span
                key={col}
                style={{
                  fontSize: '11px',
                  fontWeight: 600,
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                  color: 'var(--hr-text-dim)',
                }}
              >
                {col}
              </span>
            ))}
          </div>

          {payrollData.map((row, i) => (
            <motion.div
              key={row.id}
              variants={rowAnim}
              style={{
                display: 'grid',
                gridTemplateColumns: '2fr 1fr 1fr 1fr 1fr 1fr',
                minWidth: 760,
                alignItems: 'center',
                padding: '14px 20px',
                borderBottom: i < payrollData.length - 1 ? '1px solid var(--hr-border)' : 'none',
                transition: 'background 0.12s',
              }}
              whileHover={{ background: 'var(--hr-hover)' } as any}
            >
              {/* Name */}
              <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                <span
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    height: 32,
                    width: 32,
                    borderRadius: '50%',
                    background: 'rgba(139,92,246,0.12)',
                    color: ACCENT,
                    fontSize: '11px',
                    fontWeight: 700,
                    flexShrink: 0,
                  }}
                >
                  {getInitials(row.name)}
                </span>
                <div>
                  <p
                    style={{
                      margin: 0,
                      fontSize: '13px',
                      fontWeight: 600,
                      color: 'var(--hr-text)',
                    }}
                  >
                    {row.name}
                  </p>
                  <p style={{ margin: '2px 0 0', fontSize: '11px', color: 'var(--hr-text-muted)' }}>
                    {row.role}
                  </p>
                </div>
              </div>

              <span style={{ fontSize: '13px', color: 'var(--hr-text)' }}>{row.hoursWorked}h</span>
              <span style={{ fontSize: '13px', color: 'var(--hr-text)' }}>{fmt(row.basePay)}</span>
              <span
                style={{
                  fontSize: '13px',
                  color: row.overtime > 0 ? '#f59e0b' : 'var(--hr-text-muted)',
                }}
              >
                {row.overtime > 0 ? fmt(row.overtime) : '-'}
              </span>
              <span
                style={{
                  fontSize: '13px',
                  color: row.deductions > 0 ? '#ef4444' : 'var(--hr-text-muted)',
                }}
              >
                {row.deductions > 0 ? fmt(row.deductions) : '-'}
              </span>
              <span style={{ fontSize: '13px', fontWeight: 700, color: '#10b981' }}>
                {fmt(netPay(row))}
              </span>
            </motion.div>
          ))}

          {/* Totals row */}
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: '2fr 1fr 1fr 1fr 1fr 1fr',
              minWidth: 760,
              alignItems: 'center',
              padding: '14px 20px',
              background: 'var(--hr-hover)',
              borderTop: `2px solid ${ACCENT}30`,
            }}
          >
            <span
              style={{
                fontSize: '12px',
                fontWeight: 700,
                color: 'var(--hr-text)',
                textTransform: 'uppercase',
                letterSpacing: '0.05em',
              }}
            >
              Total
            </span>
            <span style={{ fontSize: '13px', fontWeight: 600, color: 'var(--hr-text)' }}>
              {payrollData.reduce((s, r) => s + r.hoursWorked, 0)}h
            </span>
            <span style={{ fontSize: '13px', fontWeight: 600, color: 'var(--hr-text)' }}>
              {fmt(totalBase)}
            </span>
            <span style={{ fontSize: '13px', fontWeight: 600, color: '#f59e0b' }}>
              {fmt(totalOvertime)}
            </span>
            <span style={{ fontSize: '13px', fontWeight: 600, color: '#ef4444' }}>
              {fmt(totalDeductions)}
            </span>
            <span style={{ fontSize: '14px', fontWeight: 800, color: '#10b981' }}>
              {fmt(totalNet)}
            </span>
          </div>
        </div>
      </motion.div>
    </div>
  );
};

export default Payroll;
