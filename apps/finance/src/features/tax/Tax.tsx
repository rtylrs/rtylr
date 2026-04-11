import React, { useState } from 'react';
import { motion } from 'motion/react';
import { Download, Settings, CheckCircle, Clock, AlertTriangle } from 'lucide-react';
import { TaxPeriod, TaxConfig as ApiTaxConfig } from '../../shared/services/financeApi';
import { useTaxReport } from '../../shared/hooks/useTax';

const ACCENT = '#10b981';

const fmt = (n: number) =>
  n.toLocaleString('en-US', { style: 'currency', currency: 'USD', minimumFractionDigits: 2 });

const statusStyle: Record<TaxPeriod['status'], React.CSSProperties> = {
  Filed: { background: 'rgba(16,185,129,0.12)', color: '#10b981' },
  Pending: { background: 'rgba(245,158,11,0.12)', color: '#f59e0b' },
  Due: { background: 'rgba(239,68,68,0.12)', color: '#ef4444' },
};

const statusIcon: Record<TaxPeriod['status'], React.ReactNode> = {
  Filed: <CheckCircle size={11} />,
  Pending: <Clock size={11} />,
  Due: <AlertTriangle size={11} />,
};

const cardStyle: React.CSSProperties = {
  background: 'var(--hub-card)',
  border: '1px solid var(--hub-border)',
  borderRadius: 12,
  padding: 20,
};

// ── Tax rate config ───────────────────────────────────────────
interface TaxConfig {
  name: string;
  rate: number;
}
const defaultRates: TaxConfig[] = [
  { name: 'Standard VAT', rate: 7.0 },
  { name: 'Reduced VAT', rate: 3.5 },
  { name: 'Services Tax', rate: 5.0 },
];

const Tax: React.FC = () => {
  const [showConfig, setShowConfig] = useState(false);
  const [rates, setRates] = useState<TaxConfig[]>(defaultRates);
  const { data: taxReport, isLoading: loading } = useTaxReport();

  const vatSummary = taxReport?.vatSummary ?? { collected: 0, owed: 0, paid: 0 };
  const taxPeriods: TaxPeriod[] = taxReport?.periods ?? [];
  const unremitted = vatSummary.owed - vatSummary.paid;

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.25 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      {/* Header */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div>
          <h2 style={{ margin: 0, fontSize: 18, fontWeight: 700, color: 'var(--hub-text)' }}>
            Tax
          </h2>
          <p style={{ margin: '2px 0 0', fontSize: 13, color: 'var(--hub-text-muted)' }}>
            VAT / Sales tax summary and filing periods
          </p>
        </div>
        <div style={{ display: 'flex', gap: 8 }}>
          <button
            onClick={() => setShowConfig(!showConfig)}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 6,
              padding: '7px 14px',
              borderRadius: 8,
              border: '1px solid var(--hub-border)',
              background: showConfig ? 'var(--hub-active)' : 'var(--hub-hover)',
              color: 'var(--hub-text-muted)',
              fontSize: 13,
              cursor: 'pointer',
            }}
          >
            <Settings size={14} />
            Tax Rates
          </button>
          <button
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 6,
              padding: '7px 14px',
              borderRadius: 8,
              background: ACCENT,
              color: '#fff',
              border: 'none',
              fontSize: 13,
              fontWeight: 600,
              cursor: 'pointer',
            }}
          >
            <Download size={14} />
            Export for Accountant
          </button>
        </div>
      </div>

      {/* VAT Summary cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 16 }}>
        {[
          {
            label: 'Tax Collected (YTD)',
            value: vatSummary.collected,
            color: '#10b981',
            note: 'From invoices',
          },
          { label: 'Tax Owed', value: vatSummary.owed, color: '#f59e0b', note: 'Net payable' },
          {
            label: 'Tax Paid',
            value: vatSummary.paid,
            color: '#3b82f6',
            note: 'Remitted to authority',
          },
        ].map(({ label, value, color, note }) => (
          <div
            key={label}
            style={{
              background: 'var(--hub-card)',
              border: '1px solid var(--hub-border)',
              borderRadius: 12,
              padding: 20,
            }}
          >
            <div
              style={{
                fontSize: 11,
                color: 'var(--hub-text-muted)',
                fontWeight: 600,
                textTransform: 'uppercase',
                letterSpacing: '0.07em',
                marginBottom: 10,
              }}
            >
              {label}
            </div>
            <div
              style={{
                fontSize: 24,
                fontWeight: 700,
                color,
                fontVariantNumeric: 'tabular-nums',
                marginBottom: 4,
              }}
            >
              {fmt(value)}
            </div>
            <div style={{ fontSize: 11, color: 'var(--hub-text-dim)' }}>{note}</div>
          </div>
        ))}
      </div>

      {/* Unremitted alert */}
      {unremitted > 0 && (
        <div
          style={{
            display: 'flex',
            alignItems: 'flex-start',
            gap: 12,
            padding: '14px 16px',
            borderRadius: 10,
            background: 'rgba(245,158,11,0.08)',
            border: '1px solid rgba(245,158,11,0.25)',
          }}
        >
          <AlertTriangle size={16} style={{ color: '#f59e0b', flexShrink: 0, marginTop: 1 }} />
          <div>
            <p
              style={{
                margin: 0,
                fontSize: 13,
                fontWeight: 600,
                color: '#f59e0b',
                marginBottom: 2,
              }}
            >
              Unremitted Tax: {fmt(unremitted)}
            </p>
            <p style={{ margin: 0, fontSize: 12, color: 'var(--hub-text-muted)' }}>
              You have collected more tax than you have remitted. Ensure you file before the next
              due date.
            </p>
          </div>
        </div>
      )}

      {/* Tax rate config */}
      {showConfig && (
        <motion.div
          initial={{ opacity: 0, y: -8 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.2 }}
          style={{ ...cardStyle, border: `1px solid ${ACCENT}30` }}
        >
          <h3
            style={{ margin: '0 0 14px', fontSize: 14, fontWeight: 600, color: 'var(--hub-text)' }}
          >
            Tax Rate Configuration
          </h3>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
            {rates.map((r, i) => (
              <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                <input
                  value={r.name}
                  onChange={(e) => {
                    const next = [...rates];
                    next[i] = { ...next[i], name: e.target.value };
                    setRates(next);
                  }}
                  style={{
                    flex: 1,
                    padding: '7px 12px',
                    borderRadius: 7,
                    border: '1px solid var(--hub-border)',
                    background: 'var(--hub-input-bg)',
                    color: 'var(--hub-text)',
                    fontSize: 13,
                    outline: 'none',
                  }}
                />
                <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                  <input
                    type="number"
                    value={r.rate}
                    step={0.5}
                    onChange={(e) => {
                      const next = [...rates];
                      next[i] = { ...next[i], rate: parseFloat(e.target.value) };
                      setRates(next);
                    }}
                    style={{
                      width: 72,
                      padding: '7px 12px',
                      borderRadius: 7,
                      border: '1px solid var(--hub-border)',
                      background: 'var(--hub-input-bg)',
                      color: 'var(--hub-text)',
                      fontSize: 13,
                      outline: 'none',
                      textAlign: 'right',
                    }}
                  />
                  <span style={{ fontSize: 13, color: 'var(--hub-text-muted)' }}>%</span>
                </div>
              </div>
            ))}
          </div>
          <button
            style={{
              marginTop: 14,
              padding: '7px 16px',
              borderRadius: 7,
              background: ACCENT,
              color: '#fff',
              border: 'none',
              fontSize: 13,
              fontWeight: 600,
              cursor: 'pointer',
            }}
            onClick={() => setShowConfig(false)}
          >
            Save Rates
          </button>
        </motion.div>
      )}

      {/* Tax periods table */}
      <div
        style={{
          background: 'var(--hub-card)',
          border: '1px solid var(--hub-border)',
          borderRadius: 12,
          overflow: 'hidden',
        }}
      >
        <div
          style={{
            padding: '14px 20px',
            borderBottom: '1px solid var(--hub-border)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
          }}
        >
          <span style={{ fontSize: 14, fontWeight: 600, color: 'var(--hub-text)' }}>
            Filing Periods
          </span>
        </div>
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr style={{ borderBottom: '1px solid var(--hub-border)' }}>
              {[
                'Period',
                'Revenue',
                'Taxable Amount',
                'Rate',
                'Tax Collected',
                'Status',
                'Due Date',
              ].map((h) => (
                <th
                  key={h}
                  style={{
                    padding: '10px 16px',
                    textAlign: 'left',
                    fontSize: 11,
                    fontWeight: 600,
                    textTransform: 'uppercase',
                    letterSpacing: '0.07em',
                    color: 'var(--hub-text-dim)',
                  }}
                >
                  {h}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {taxPeriods.map((tp, i) => (
              <tr
                key={tp.id}
                style={{
                  borderBottom: i < taxPeriods.length - 1 ? '1px solid var(--hub-border)' : 'none',
                }}
                onMouseEnter={(e) => (e.currentTarget.style.background = 'var(--hub-hover)')}
                onMouseLeave={(e) => (e.currentTarget.style.background = 'transparent')}
              >
                <td style={{ padding: '12px 16px' }}>
                  <div style={{ fontSize: 13, fontWeight: 500, color: 'var(--hub-text)' }}>
                    {tp.period}
                  </div>
                  <div style={{ fontSize: 11, color: 'var(--hub-text-dim)' }}>{tp.id}</div>
                </td>
                <td
                  style={{
                    padding: '12px 16px',
                    fontSize: 13,
                    color: 'var(--hub-text)',
                    fontVariantNumeric: 'tabular-nums',
                  }}
                >
                  {fmt(tp.revenue)}
                </td>
                <td
                  style={{
                    padding: '12px 16px',
                    fontSize: 13,
                    color: 'var(--hub-text)',
                    fontVariantNumeric: 'tabular-nums',
                  }}
                >
                  {fmt(tp.taxableAmount)}
                </td>
                <td style={{ padding: '12px 16px', fontSize: 13, fontWeight: 600, color: ACCENT }}>
                  {(tp.taxRate * 100).toFixed(1)}%
                </td>
                <td
                  style={{
                    padding: '12px 16px',
                    fontSize: 13,
                    fontWeight: 600,
                    color: 'var(--hub-text)',
                    fontVariantNumeric: 'tabular-nums',
                  }}
                >
                  {fmt(tp.collected)}
                </td>
                <td style={{ padding: '12px 16px' }}>
                  <span
                    style={{
                      display: 'inline-flex',
                      alignItems: 'center',
                      gap: 5,
                      padding: '3px 10px',
                      borderRadius: 20,
                      fontSize: 11,
                      fontWeight: 600,
                      ...statusStyle[tp.status],
                    }}
                  >
                    {statusIcon[tp.status]}
                    {tp.status}
                  </span>
                </td>
                <td
                  style={{
                    padding: '12px 16px',
                    fontSize: 12,
                    color: tp.status === 'Due' ? '#ef4444' : 'var(--hub-text-muted)',
                  }}
                >
                  {tp.dueDate}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </motion.div>
  );
};

export default Tax;
