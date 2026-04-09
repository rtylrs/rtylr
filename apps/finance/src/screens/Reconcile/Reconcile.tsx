import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { Link2, CheckCheck } from 'lucide-react';
import { toast } from 'sonner';
import PageHeader from '../../components/PageHeader';
import MoneyDisplay from '../../components/MoneyDisplay';
import { useBankTransactions } from '../../hooks/useBankAccounts';
import { cardStyle, ghostBtn, primaryBtn, ACCENT } from '../../lib/tokens';

const CANDIDATES = [
  { id: 'INV-001', label: 'Invoice INV-001 — Acme', amount: 14820 },
  { id: 'INV-002', label: 'Invoice INV-002 — Globex', amount: 2400 },
  { id: 'STRIPE-1', label: 'Stripe payout batch 1', amount: 8920 },
  { id: 'EXP-203', label: 'Expense EXP-203 — Figma', amount: -420 },
];

const Reconcile: React.FC = () => {
  const { data: transactions = [] } = useBankTransactions();
  const [matches, setMatches] = useState<Record<string, string>>({});

  const unmatched = transactions.filter((t) => !t.matched && !matches[t.id]);

  const match = (txId: string, candidateId: string) => {
    setMatches((prev) => ({ ...prev, [txId]: candidateId }));
    toast.success(`Matched transaction to ${candidateId}`);
  };

  return (
    <motion.div initial={{ opacity: 0, y: 4 }} animate={{ opacity: 1, y: 0 }} style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <PageHeader
        title="Reconciliation"
        subtitle={`${unmatched.length} transactions awaiting match`}
        actions={
          <button style={primaryBtn} onClick={() => toast.success('Auto-match completed')}>
            <CheckCheck size={13} /> Auto-match all
          </button>
        }
      />

      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        {unmatched.map((t) => (
          <div key={t.id} style={{ ...cardStyle, padding: 16 }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 10 }}>
              <div>
                <div style={{ fontSize: 13, fontWeight: 600, color: 'var(--hub-text)' }}>{t.description}</div>
                <div style={{ fontSize: 11, color: 'var(--hub-text-dim)' }}>{t.date} <>&middot;</> {t.category}</div>
              </div>
              <div style={{ fontSize: 16, fontWeight: 700, color: t.amount >= 0 ? ACCENT : 'var(--hub-text)' }}>
                <MoneyDisplay amount={t.amount} signed />
              </div>
            </div>
            <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
              {CANDIDATES.filter((c) => Math.abs(c.amount) === Math.abs(t.amount) || Math.abs(Math.abs(c.amount) - Math.abs(t.amount)) < 500).map((c) => (
                <button key={c.id} style={{ ...ghostBtn, padding: '6px 12px' }} onClick={() => match(t.id, c.id)}>
                  <Link2 size={12} /> {c.label}
                </button>
              ))}
            </div>
          </div>
        ))}
        {unmatched.length === 0 && (
          <div style={{ ...cardStyle, padding: 30, textAlign: 'center', color: 'var(--hub-text-muted)', fontSize: 13 }}>
            All transactions reconciled.
          </div>
        )}
      </div>
    </motion.div>
  );
};

export default Reconcile;
