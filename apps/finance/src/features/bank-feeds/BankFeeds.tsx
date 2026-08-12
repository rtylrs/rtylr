import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import { Link as LinkIcon, Building2, CreditCard, PiggyBank } from 'lucide-react';
import { toast } from 'sonner';
import PageHeader from '../../shared/components/PageHeader';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import { SkeletonRow } from '../../shared/components/Skeleton';
import { useBankAccounts, useBankTransactions } from '../../shared/hooks/useBankAccounts';
import { cardStyle, ghostBtn, primaryBtn, ACCENT } from '../../shared/lib/tokens';

const iconFor: Record<
  string,
  React.ComponentType<{ size?: number; style?: React.CSSProperties }>
> = {
  checking: Building2,
  savings: PiggyBank,
  credit: CreditCard,
};

const BankFeeds: React.FC = () => {
  const { data: accounts = [], isLoading } = useBankAccounts();
  const [selected, setSelected] = useState<string | undefined>();
  const { data: transactions = [] } = useBankTransactions(selected);

  const total = accounts.reduce((s: number, a) => s + (a.balance ?? 0), 0);

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title="Bank feeds · rtylr" />
      <PageHeader
        title="Bank Feeds"
        subtitle={
          <>
            {accounts.length} connected accounts <>&middot;</> <MoneyDisplay amount={total} /> total
            balance
          </>
        }
        actions={
          <button style={primaryBtn} onClick={() => toast.info('Bank connection coming soon')}>
            <LinkIcon size={13} /> Connect Account
          </button>
        }
      />

      {isLoading ? (
        <div style={{ ...cardStyle, padding: 20 }}>
          <SkeletonRow />
        </div>
      ) : (
        <>
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fill, minmax(260px, 1fr))',
              gap: 14,
            }}
          >
            {accounts.map((a) => {
              const Icon = iconFor[a.type ?? ''] ?? Building2;
              return (
                <button
                  key={a.id}
                  onClick={() => setSelected(selected === a.id ? undefined : a.id)}
                  style={{
                    ...cardStyle,
                    padding: 16,
                    textAlign: 'left',
                    cursor: 'pointer',
                    borderColor: selected === a.id ? ACCENT : 'var(--hub-border)',
                  }}
                >
                  <div
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'space-between',
                      marginBottom: 10,
                    }}
                  >
                    <span
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        width: 36,
                        height: 36,
                        borderRadius: 10,
                        background: 'var(--finance-accent-soft)',
                        color: ACCENT,
                      }}
                    >
                      <Icon size={16} />
                    </span>
                    <span
                      style={{
                        fontSize: 11,
                        color: 'var(--hub-text-dim)',
                        textTransform: 'uppercase',
                      }}
                    >
                      {a.type}
                    </span>
                  </div>
                  <div style={{ fontSize: 13, fontWeight: 600, color: 'var(--hub-text)' }}>
                    {a.institution} <>&middot;</>
                    <>&middot;</>
                    <>&middot;</>
                    <>&middot;</> {a.mask}
                  </div>
                  <div
                    style={{
                      fontSize: 20,
                      fontWeight: 700,
                      color: (a.balance ?? 0) < 0 ? 'var(--finance-danger)' : 'var(--hub-text)',
                      marginTop: 6,
                    }}
                  >
                    <MoneyDisplay amount={a.balance} />
                  </div>
                </button>
              );
            })}
          </div>

          <div style={{ ...cardStyle, overflow: 'hidden' }}>
            <div
              style={{
                padding: '14px 16px',
                borderBottom: '1px solid var(--hub-border)',
                fontSize: 12,
                color: 'var(--hub-text-muted)',
              }}
            >
              {selected ? 'Account transactions' : 'All recent transactions'}
            </div>
            <table style={{ width: '100%', borderCollapse: 'collapse' }}>
              <thead>
                <tr style={{ borderBottom: '1px solid var(--hub-border)' }}>
                  {['Date', 'Description', 'Category', 'Amount', 'Match'].map((h) => (
                    <th
                      key={h}
                      style={{
                        padding: '10px 16px',
                        textAlign: h === 'Amount' ? 'right' : 'left',
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
                {transactions.map((t) => (
                  <tr key={t.id} style={{ borderBottom: '1px solid var(--hub-border)' }}>
                    <td
                      style={{ padding: '10px 16px', fontSize: 12, color: 'var(--hub-text-muted)' }}
                    >
                      {t.date}
                    </td>
                    <td style={{ padding: '10px 16px', fontSize: 13, color: 'var(--hub-text)' }}>
                      {t.description}
                    </td>
                    <td
                      style={{ padding: '10px 16px', fontSize: 12, color: 'var(--hub-text-muted)' }}
                    >
                      {t.category}
                    </td>
                    <td
                      style={{
                        padding: '10px 16px',
                        fontSize: 13,
                        fontWeight: 600,
                        textAlign: 'right',
                        color: (t.amount ?? 0) >= 0 ? ACCENT : 'var(--hub-text)',
                      }}
                    >
                      <MoneyDisplay amount={t.amount ?? 0} signed />
                    </td>
                    <td style={{ padding: '10px 16px', fontSize: 11 }}>
                      <span
                        style={{
                          padding: '2px 8px',
                          borderRadius: 20,
                          background: t.matched
                            ? 'var(--finance-accent-soft)'
                            : 'var(--finance-warning-soft)',
                          color: t.matched ? ACCENT : 'var(--finance-warning)',
                          fontWeight: 600,
                        }}
                      >
                        {t.matched ? 'Matched' : 'Pending'}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </>
      )}
    </motion.div>
  );
};

export default BankFeeds;
