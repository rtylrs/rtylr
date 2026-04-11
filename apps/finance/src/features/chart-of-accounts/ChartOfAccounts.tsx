import React, { useState } from 'react';
import { motion } from 'motion/react';
import { Plus, Search } from 'lucide-react';
import { toast } from 'sonner';
import PageHeader from '../../shared/components/PageHeader';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import AccountTypeBadge, { AccountType } from '../../shared/components/AccountTypeBadge';
import { SkeletonRow } from '../../shared/components/Skeleton';
import { useChartOfAccounts } from '../../shared/hooks/useChartOfAccounts';
import { cardStyle, ghostBtn, primaryBtn, inputStyle, ACCENT } from '../../shared/lib/tokens';

const TYPES: (AccountType | 'all')[] = [
  'all',
  'asset',
  'liability',
  'equity',
  'revenue',
  'expense',
];

const ChartOfAccounts: React.FC = () => {
  const { data = [], isLoading } = useChartOfAccounts();
  const [filter, setFilter] = useState<AccountType | 'all'>('all');
  const [query, setQuery] = useState('');

  const filtered = data.filter((a: any) => {
    const matchesType = filter === 'all' || a.type === filter;
    const matchesQuery =
      !query || a.name.toLowerCase().includes(query.toLowerCase()) || a.code.includes(query);
    return matchesType && matchesQuery;
  });

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <PageHeader
        title="Chart of Accounts"
        subtitle={`${filtered.length} accounts`}
        actions={
          <button style={primaryBtn} onClick={() => toast.success('Account created')}>
            <Plus size={13} /> New Account
          </button>
        }
      />

      <div style={{ display: 'flex', gap: 12 }}>
        <div
          style={{
            display: 'flex',
            gap: 4,
            padding: 4,
            background: 'var(--hub-surface)',
            borderRadius: 10,
            border: '1px solid var(--hub-border)',
          }}
        >
          {TYPES.map((t) => (
            <button
              key={t}
              onClick={() => setFilter(t)}
              style={{
                padding: '5px 12px',
                borderRadius: 7,
                border: 'none',
                fontSize: 12,
                fontWeight: 500,
                cursor: 'pointer',
                background: filter === t ? ACCENT : 'transparent',
                color: filter === t ? '#fff' : 'var(--hub-text-muted)',
                textTransform: 'capitalize',
              }}
            >
              {t}
            </button>
          ))}
        </div>
        <div style={{ position: 'relative', flex: 1, maxWidth: 280 }}>
          <Search
            size={13}
            style={{ position: 'absolute', left: 10, top: 11, color: 'var(--hub-text-dim)' }}
          />
          <input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search accounts"
            style={{ ...inputStyle, paddingLeft: 30 }}
          />
        </div>
      </div>

      {isLoading ? (
        <div style={{ ...cardStyle, padding: 20 }}>
          <SkeletonRow />
        </div>
      ) : (
        <div style={{ ...cardStyle, overflow: 'hidden' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr style={{ borderBottom: '1px solid var(--hub-border)' }}>
                {['Code', 'Name', 'Type', 'Balance'].map((h) => (
                  <th
                    key={h}
                    style={{
                      padding: '10px 16px',
                      textAlign: h === 'Balance' ? 'right' : 'left',
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
              {filtered.map((a: any) => (
                <tr key={a.id} style={{ borderBottom: '1px solid var(--hub-border)' }}>
                  <td
                    style={{
                      padding: '12px 16px',
                      fontSize: 13,
                      fontWeight: 600,
                      color: 'var(--hub-text)',
                      fontVariantNumeric: 'tabular-nums',
                    }}
                  >
                    {a.code}
                  </td>
                  <td style={{ padding: '12px 16px', fontSize: 13, color: 'var(--hub-text)' }}>
                    {a.name}
                  </td>
                  <td style={{ padding: '12px 16px' }}>
                    <AccountTypeBadge type={a.type} />
                  </td>
                  <td
                    style={{
                      padding: '12px 16px',
                      fontSize: 13,
                      fontWeight: 600,
                      textAlign: 'right',
                    }}
                  >
                    <MoneyDisplay amount={a.balance} />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </motion.div>
  );
};

export default ChartOfAccounts;
