import React from 'react';
import { motion } from 'motion/react';
import { FileSpreadsheet, Download } from 'lucide-react';
import { toast } from 'sonner';
import PageHeader from '../../shared/components/PageHeader';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import { cardStyle, ghostBtn, ACCENT } from '../../shared/lib/tokens';

const PERIODS = [
  { period: 'Q1 2026', collected: 18420, owed: 16200, filed: true, dueDate: '2026-04-15' },
  { period: 'Q4 2025', collected: 22100, owed: 19800, filed: true, dueDate: '2026-01-15' },
  { period: 'Q3 2025', collected: 19800, owed: 17200, filed: true, dueDate: '2025-10-15' },
  { period: 'Q2 2025', collected: 16500, owed: 14900, filed: true, dueDate: '2025-07-15' },
];

const TaxReport: React.FC = () => {
  const totalCollected = PERIODS.reduce((s, p) => s + p.collected, 0);
  const totalOwed = PERIODS.reduce((s, p) => s + p.owed, 0);
  const toFile = PERIODS.filter((p) => !p.filed).length;

  return (
    <motion.div initial={{ opacity: 0, y: 4 }} animate={{ opacity: 1, y: 0 }} style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <PageHeader
        title="Tax Summary"
        breadcrumbs={[{ label: 'Reports' }, { label: 'Tax' }]}
        actions={
          <button style={ghostBtn} onClick={() => toast.success('Report exported')}>
            <Download size={13} /> Export
          </button>
        }
      />

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 14 }}>
        <div style={{ ...cardStyle, padding: 20 }}>
          <div style={{ fontSize: 12, color: 'var(--hub-text-muted)' }}>Tax collected</div>
          <div style={{ fontSize: 22, fontWeight: 700, color: ACCENT, marginTop: 6 }}><MoneyDisplay amount={totalCollected} /></div>
        </div>
        <div style={{ ...cardStyle, padding: 20 }}>
          <div style={{ fontSize: 12, color: 'var(--hub-text-muted)' }}>Tax owed</div>
          <div style={{ fontSize: 22, fontWeight: 700, color: '#f59e0b', marginTop: 6 }}><MoneyDisplay amount={totalOwed} /></div>
        </div>
        <div style={{ ...cardStyle, padding: 20 }}>
          <div style={{ fontSize: 12, color: 'var(--hub-text-muted)' }}>Periods to file</div>
          <div style={{ fontSize: 22, fontWeight: 700, color: 'var(--hub-text)', marginTop: 6 }}>{toFile}</div>
        </div>
      </div>

      <div style={{ ...cardStyle, overflow: 'hidden' }}>
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr style={{ borderBottom: '1px solid var(--hub-border)' }}>
              {['Period', 'Collected', 'Owed', 'Status', 'Due date', 'Actions'].map((h) => (
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
            {PERIODS.map((p) => (
              <tr key={p.period} style={{ borderBottom: '1px solid var(--hub-border)' }}>
                <td style={{ padding: '12px 16px', fontSize: 13, fontWeight: 600 }}>{p.period}</td>
                <td style={{ padding: '12px 16px', fontSize: 13 }}><MoneyDisplay amount={p.collected} /></td>
                <td style={{ padding: '12px 16px', fontSize: 13 }}><MoneyDisplay amount={p.owed} /></td>
                <td style={{ padding: '12px 16px' }}>
                  <span
                    style={{
                      fontSize: 11,
                      padding: '3px 10px',
                      borderRadius: 20,
                      fontWeight: 600,
                      background: p.filed ? 'rgba(16,185,129,0.12)' : 'rgba(245,158,11,0.12)',
                      color: p.filed ? ACCENT : '#f59e0b',
                    }}
                  >
                    {p.filed ? 'Filed' : 'Pending'}
                  </span>
                </td>
                <td style={{ padding: '12px 16px', fontSize: 12, color: 'var(--hub-text-muted)' }}>{p.dueDate}</td>
                <td style={{ padding: '12px 16px' }}>
                  <button style={{ ...ghostBtn, padding: '4px 10px' }} onClick={() => toast.success('Detail view opened')}>
                    <FileSpreadsheet size={12} /> View
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </motion.div>
  );
};

export default TaxReport;
