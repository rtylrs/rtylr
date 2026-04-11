import React, { useState } from 'react';
import { motion } from 'motion/react';
import { Plus, Paperclip, Check } from 'lucide-react';
import { Expense } from '../../shared/services/financeApi';
import { useExpenses } from './useExpenses';

const ACCENT = '#10b981';

type Category = string;

const categoryColors: Record<string, string> = {
  'Food & Bev': '#f59e0b',
  Utilities: '#3b82f6',
  Rent: '#8b5cf6',
  Staff: '#10b981',
  Marketing: '#f43f5e',
  Other: '#71717a',
};

function getCategoryColor(cat: string): string {
  return categoryColors[cat] ?? '#6b7280';
}

const fmt = (n: number) =>
  n.toLocaleString('en-US', { style: 'currency', currency: 'USD', minimumFractionDigits: 0 });

const cardStyle: React.CSSProperties = {
  background: 'var(--hub-card)',
  border: '1px solid var(--hub-border)',
  borderRadius: 12,
  padding: 20,
};

// ── CSS "pie" breakdown (horizontal stacked bar) ──────────────
const CategoryBar: React.FC<{ expenses: Expense[] }> = ({ expenses }) => {
  const categoryTotals = expenses.reduce<Record<string, number>>((acc, exp) => {
    acc[exp.category] = (acc[exp.category] || 0) + exp.amount;
    return acc;
  }, {});
  const grandTotal = Object.values(categoryTotals).reduce((s, v) => s + v, 0);
  const entries = Object.entries(categoryTotals);
  entries.sort((a, b) => b[1] - a[1]);

  if (entries.length === 0)
    return <div style={{ height: 20, background: 'var(--hub-hover)', borderRadius: 8 }} />;

  return (
    <div>
      {/* Stacked bar */}
      <div
        style={{
          display: 'flex',
          borderRadius: 8,
          overflow: 'hidden',
          height: 20,
          marginBottom: 16,
        }}
      >
        {entries.map(([cat, val]) => (
          <div
            key={cat}
            style={{
              width: `${(val / grandTotal) * 100}%`,
              background: getCategoryColor(cat),
              transition: 'width 0.4s ease',
            }}
            title={`${cat}: ${fmt(val)}`}
          />
        ))}
      </div>

      {/* Legend */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '10px 16px' }}>
        {entries.map(([cat, val]) => (
          <div key={cat} style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <span
              style={{
                width: 10,
                height: 10,
                borderRadius: 3,
                background: getCategoryColor(cat),
                flexShrink: 0,
              }}
            />
            <div style={{ minWidth: 0 }}>
              <div
                style={{
                  fontSize: 12,
                  fontWeight: 500,
                  color: 'var(--hub-text)',
                  whiteSpace: 'nowrap',
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                }}
              >
                {cat}
              </div>
              <div style={{ fontSize: 11, color: 'var(--hub-text-muted)' }}>
                {fmt(val)} &middot; {((val / grandTotal) * 100).toFixed(1)}%
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

const Expenses: React.FC = () => {
  const [showAdd, setShowAdd] = useState(false);
  const { data: fetchedExpenses, isLoading: loading } = useExpenses();
  const expenses: Expense[] = fetchedExpenses || [];

  const grandTotal = expenses.reduce((s, e) => s + e.amount, 0);

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
            Expenses
          </h2>
          <p style={{ margin: '2px 0 0', fontSize: 13, color: 'var(--hub-text-muted)' }}>
            Monthly total: <strong style={{ color: 'var(--hub-text)' }}>{fmt(grandTotal)}</strong>
          </p>
        </div>
        <button
          onClick={() => setShowAdd(true)}
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 6,
            padding: '8px 16px',
            borderRadius: 8,
            background: ACCENT,
            color: '#fff',
            border: 'none',
            fontSize: 13,
            fontWeight: 600,
            cursor: 'pointer',
          }}
        >
          <Plus size={15} />
          Add Expense
        </button>
      </div>

      {/* Category breakdown */}
      <div style={cardStyle}>
        <h3 style={{ margin: '0 0 16px', fontSize: 14, fontWeight: 600, color: 'var(--hub-text)' }}>
          Breakdown by Category
        </h3>
        <CategoryBar expenses={expenses} />
      </div>

      {/* Add expense inline form */}
      {showAdd && (
        <motion.div
          initial={{ opacity: 0, height: 0 }}
          animate={{ opacity: 1, height: 'auto' }}
          exit={{ opacity: 0, height: 0 }}
          style={{
            ...cardStyle,
            border: `1px solid ${ACCENT}40`,
            background: 'rgba(16,185,129,0.04)',
          }}
        >
          <h3
            style={{ margin: '0 0 16px', fontSize: 14, fontWeight: 600, color: 'var(--hub-text)' }}
          >
            Add New Expense
          </h3>
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: '1fr 1fr 1fr 1fr',
              gap: 12,
              marginBottom: 16,
            }}
          >
            {[
              { label: 'Date', type: 'date', placeholder: '' },
              { label: 'Description', type: 'text', placeholder: 'e.g. Team lunch' },
              { label: 'Amount ($)', type: 'number', placeholder: '0.00' },
            ].map(({ label, type, placeholder }) => (
              <div key={label} style={{ display: 'flex', flexDirection: 'column', gap: 5 }}>
                <label
                  style={{
                    fontSize: 11,
                    fontWeight: 600,
                    color: 'var(--hub-text-muted)',
                    textTransform: 'uppercase',
                    letterSpacing: '0.06em',
                  }}
                >
                  {label}
                </label>
                <input
                  type={type}
                  placeholder={placeholder}
                  style={{
                    padding: '7px 10px',
                    borderRadius: 7,
                    border: '1px solid var(--hub-border)',
                    background: 'var(--hub-input-bg)',
                    color: 'var(--hub-text)',
                    fontSize: 13,
                    outline: 'none',
                  }}
                />
              </div>
            ))}
            <div style={{ display: 'flex', flexDirection: 'column', gap: 5 }}>
              <label
                style={{
                  fontSize: 11,
                  fontWeight: 600,
                  color: 'var(--hub-text-muted)',
                  textTransform: 'uppercase',
                  letterSpacing: '0.06em',
                }}
              >
                Category
              </label>
              <select
                style={{
                  padding: '7px 10px',
                  borderRadius: 7,
                  border: '1px solid var(--hub-border)',
                  background: 'var(--hub-input-bg)',
                  color: 'var(--hub-text)',
                  fontSize: 13,
                  outline: 'none',
                }}
              >
                {Object.keys(categoryColors).map((c) => (
                  <option key={c} value={c}>
                    {c}
                  </option>
                ))}
              </select>
            </div>
          </div>
          <div style={{ display: 'flex', gap: 8 }}>
            <button
              style={{
                padding: '7px 16px',
                borderRadius: 7,
                background: ACCENT,
                color: '#fff',
                border: 'none',
                fontSize: 13,
                fontWeight: 600,
                cursor: 'pointer',
              }}
              onClick={() => setShowAdd(false)}
            >
              Save Expense
            </button>
            <button
              style={{
                padding: '7px 14px',
                borderRadius: 7,
                background: 'transparent',
                color: 'var(--hub-text-muted)',
                border: '1px solid var(--hub-border)',
                fontSize: 13,
                cursor: 'pointer',
              }}
              onClick={() => setShowAdd(false)}
            >
              Cancel
            </button>
          </div>
        </motion.div>
      )}

      {/* Expenses table */}
      <div
        style={{
          background: 'var(--hub-card)',
          border: '1px solid var(--hub-border)',
          borderRadius: 12,
          overflow: 'hidden',
        }}
      >
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr style={{ borderBottom: '1px solid var(--hub-border)' }}>
              {['Date', 'Description', 'Category', 'Amount', 'Receipt'].map((h) => (
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
            {expenses.map((exp, i) => (
              <tr
                key={exp.id}
                style={{
                  borderBottom: i < expenses.length - 1 ? '1px solid var(--hub-border)' : 'none',
                }}
                onMouseEnter={(e) => (e.currentTarget.style.background = 'var(--hub-hover)')}
                onMouseLeave={(e) => (e.currentTarget.style.background = 'transparent')}
              >
                <td style={{ padding: '11px 16px', fontSize: 12, color: 'var(--hub-text-muted)' }}>
                  {exp.date}
                </td>
                <td style={{ padding: '11px 16px', fontSize: 13, color: 'var(--hub-text)' }}>
                  <div>{exp.description}</div>
                  <div style={{ fontSize: 11, color: 'var(--hub-text-dim)' }}>{exp.id}</div>
                </td>
                <td style={{ padding: '11px 16px' }}>
                  <span
                    style={{
                      display: 'inline-flex',
                      alignItems: 'center',
                      gap: 5,
                      padding: '3px 10px',
                      borderRadius: 20,
                      fontSize: 11,
                      fontWeight: 600,
                      background: `${getCategoryColor(exp.category)}18`,
                      color: getCategoryColor(exp.category),
                    }}
                  >
                    {exp.category}
                  </span>
                </td>
                <td
                  style={{
                    padding: '11px 16px',
                    fontSize: 13,
                    fontWeight: 600,
                    color: '#ef4444',
                    fontVariantNumeric: 'tabular-nums',
                  }}
                >
                  -{fmt(exp.amount)}
                </td>
                <td style={{ padding: '11px 16px' }}>
                  {exp.receipt ? (
                    <span
                      style={{
                        display: 'inline-flex',
                        alignItems: 'center',
                        gap: 4,
                        fontSize: 11,
                        color: '#10b981',
                      }}
                    >
                      <Check size={12} /> Yes
                    </span>
                  ) : (
                    <span
                      style={{
                        display: 'inline-flex',
                        alignItems: 'center',
                        gap: 4,
                        fontSize: 11,
                        color: 'var(--hub-text-dim)',
                      }}
                    >
                      <Paperclip size={12} /> Upload
                    </span>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
          <tfoot>
            <tr
              style={{ borderTop: '1px solid var(--hub-border)', background: 'var(--hub-hover)' }}
            >
              <td
                colSpan={3}
                style={{
                  padding: '10px 16px',
                  fontSize: 12,
                  fontWeight: 600,
                  color: 'var(--hub-text-muted)',
                }}
              >
                Monthly Total
              </td>
              <td
                style={{
                  padding: '10px 16px',
                  fontSize: 14,
                  fontWeight: 700,
                  color: '#ef4444',
                  fontVariantNumeric: 'tabular-nums',
                }}
              >
                -{fmt(grandTotal)}
              </td>
              <td />
            </tr>
          </tfoot>
        </table>
      </div>
    </motion.div>
  );
};

export default Expenses;
