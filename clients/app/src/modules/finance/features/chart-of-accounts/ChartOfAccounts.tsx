import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useUserStore } from '../../../../shared/stores/userStore';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { motion } from 'motion/react';
import { Plus, Search } from 'lucide-react';
import PageHeader from '../../shared/components/PageHeader';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import AccountTypeBadge, { AccountType } from '../../shared/components/AccountTypeBadge';
import { SkeletonRow } from '../../shared/components/Skeleton';
import { useChartOfAccounts, useCreateAccount } from '../../shared/hooks/useChartOfAccounts';
import { cardStyle, ghostBtn, primaryBtn, inputStyle, ACCENT } from '../../shared/lib/tokens';
import { canFinancePermission } from '../../shared/lib/financePermissions';

const TYPES: (AccountType | 'all')[] = [
  'all',
  'asset',
  'liability',
  'equity',
  'revenue',
  'expense',
];

const ACCOUNT_TYPES = ['asset', 'liability', 'equity', 'revenue', 'expense'] as const;

const ChartOfAccounts: React.FC = () => {
  const { t } = useTranslation('financeCommon');
  const { data = [], isLoading } = useChartOfAccounts();
  const createMut = useCreateAccount();
  const [filter, setFilter] = useState<AccountType | 'all'>('all');
  const [query, setQuery] = useState('');
  const [showCreate, setShowCreate] = useState(false);
  const [newName, setNewName] = useState('');
  const [newType, setNewType] = useState<(typeof ACCOUNT_TYPES)[number]>('asset');
  const [newCode, setNewCode] = useState('');
  // There is no dedicated chart-of-accounts resource in the permission matrix;
  // creating a ledger account is an accounts-management action, so it is
  // gated on the same 'finance:accounts:view' resource used for bank accounts.
  const canCreateAccount = canFinancePermission('finance:accounts:manage');

  const handleCreate = async (): Promise<void> => {
    if (!canCreateAccount) return;
    if (!newName.trim()) {
      toast.error(t('financeAccounts.nameRequired'));
      return;
    }
    try {
      await createMut.mutateAsync({
        name: newName.trim(),
        code: newCode.trim() || undefined,
        type: newType,
        currency: 'USD',
        balance: 0,
        isActive: true,
      });
      setShowCreate(false);
      setNewName('');
      setNewCode('');
    } catch {
      // onError in useCreateAccount handles the error toast
    }
  };

  const filtered = data.filter((a) => {
    const matchesType = filter === 'all' || a.type === filter;
    const matchesQuery =
      !query ||
      a.name.toLowerCase().includes(query.toLowerCase()) ||
      (a.code ?? '').includes(query);
    return matchesType && matchesQuery;
  });

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title={`${t('financeAccounts.title')} · rtylr`} />
      <PageHeader
        title={t('financeAccounts.title')}
        subtitle={t('financeAccounts.count', { count: filtered.length })}
        actions={
          canCreateAccount ? (
            <button type="button" style={primaryBtn} onClick={() => setShowCreate(true)}>
              <Plus size={13} /> {t('financeAccounts.new')}
            </button>
          ) : undefined
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
              type="button"
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
                color: filter === t ? 'var(--hub-bg)' : 'var(--hub-text-muted)',
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
            style={{
              position: 'absolute',
              insetInlineStart: 10,
              top: 11,
              color: 'var(--hub-text-dim)',
            }}
          />
          <input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder={t('financeAccounts.search')}
            style={{ ...inputStyle, paddingInlineStart: 30 }}
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
                {[
                  t('financeAccounts.code'),
                  t('financeAccounts.name'),
                  t('financeAccounts.type'),
                  t('financeAccounts.balance'),
                ].map((h) => (
                  <th
                    key={h}
                    style={{
                      padding: '10px 16px',
                      textAlign: h === 'Balance' ? 'end' : 'start',
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
              {filtered.map((a) => (
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
                    <AccountTypeBadge type={(a.type ?? 'asset') as AccountType} />
                  </td>
                  <td
                    style={{
                      padding: '12px 16px',
                      fontSize: 13,
                      fontWeight: 600,
                      textAlign: 'end',
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
      {showCreate && canCreateAccount && (
        <div
          style={{
            position: 'fixed',
            inset: 0,
            background: 'var(--finance-overlay)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            zIndex: 9999,
          }}
          onClick={(e) => {
            if (e.target === e.currentTarget) setShowCreate(false);
          }}
        >
          <div
            style={{
              ...cardStyle,
              padding: 24,
              width: 360,
              display: 'flex',
              flexDirection: 'column',
              gap: 14,
            }}
          >
            <h3 style={{ fontSize: 15, fontWeight: 700, color: 'var(--hub-text)', margin: 0 }}>
              {t('financeAccounts.new')}
            </h3>
            <input
              style={inputStyle}
              placeholder={t('financeAccounts.name')}
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
            />
            <input
              style={inputStyle}
              placeholder={t('financeAccounts.codeOptional')}
              value={newCode}
              onChange={(e) => setNewCode(e.target.value)}
            />
            <select
              style={inputStyle}
              value={newType}
              onChange={(e) => setNewType(e.target.value as (typeof ACCOUNT_TYPES)[number])}
            >
              {ACCOUNT_TYPES.map((t) => (
                <option key={t} value={t}>
                  {t.charAt(0).toUpperCase() + t.slice(1)}
                </option>
              ))}
            </select>
            <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
              <button type="button" style={ghostBtn} onClick={() => setShowCreate(false)}>
                {t('financeCommon.cancel')}
              </button>
              <button
                type="button"
                style={primaryBtn}
                onClick={handleCreate}
                disabled={createMut.isPending}
              >
                {createMut.isPending ? t('financeCommon.creating') : t('financeCommon.create')}
              </button>
            </div>
          </div>
        </div>
      )}
    </motion.div>
  );
};

export default ChartOfAccounts;
