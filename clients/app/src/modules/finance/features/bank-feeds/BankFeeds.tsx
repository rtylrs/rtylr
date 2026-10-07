import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { pageTitle } from '../../shared/lib/pageTitle';
import { HelmetContainer } from '@voxire/components';
import { useUserStore } from '../../../../shared/stores/userStore';
import { motion } from 'motion/react';
import { CheckCircle, Link as LinkIcon, Building2, CreditCard, PiggyBank } from 'lucide-react';
import PageHeader from '../../shared/components/PageHeader';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import Modal from '../../shared/components/Modal';
import { SkeletonRow } from '../../shared/components/Skeleton';
import {
  useBankAccounts,
  useBankTransactions,
  useCreateBankAccount,
} from '../../shared/hooks/useBankAccounts';
import {
  cardStyle,
  ghostBtn,
  inputStyle,
  labelStyle,
  primaryBtn,
  ACCENT,
} from '../../shared/lib/tokens';
import { canFinancePermission } from '../../shared/lib/financePermissions';

// The stored value is the account type; only its label is translated.
const ACCOUNT_TYPES = ['checking', 'savings', 'credit'] as const;

const iconFor: Record<
  string,
  React.ComponentType<{ size?: number; style?: React.CSSProperties }>
> = {
  checking: Building2,
  savings: PiggyBank,
  credit: CreditCard,
};

const BankFeeds: React.FC = () => {
  const { t } = useTranslation('financeCommon');
  const accountsQuery = useBankAccounts();
  const {
    data: accounts = [],
    isLoading,
    isError: accountsError,
    refetch: refetchAccounts,
  } = accountsQuery;
  const createAccount = useCreateBankAccount();
  const [selected, setSelected] = useState<string | undefined>();
  const [showConnect, setShowConnect] = useState(false);
  const [form, setForm] = useState({
    name: '',
    bankName: '',
    accountType: 'checking',
    lastFour: '',
    balance: '0',
    currency: 'USD',
  });
  const {
    data: transactions = [],
    isLoading: transactionsLoading,
    isError: transactionsError,
    refetch: refetchTransactions,
  } = useBankTransactions(selected);
  const canManageAccounts = canFinancePermission('finance:bank_feeds:manage');

  const total = accounts.reduce((s: number, a) => s + (a.balance ?? 0), 0);

  const handleCreateAccount = (): void => {
    if (!canManageAccounts) return;
    if (!form.name.trim() || !form.bankName.trim() || form.lastFour.length !== 4) return;

    createAccount.mutate(
      {
        name: form.name.trim(),
        bankName: form.bankName.trim(),
        accountType: form.accountType,
        lastFour: form.lastFour,
        balance: Number(form.balance) || 0,
        currency: form.currency.trim().toUpperCase() || 'USD',
      },
      {
        onSuccess: () => {
          setShowConnect(false);
          setForm({
            name: '',
            bankName: '',
            accountType: 'checking',
            lastFour: '',
            balance: '0',
            currency: 'USD',
          });
        },
      }
    );
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title={pageTitle(t('financeCommon.pageTitles.bankFeeds'))} />
      <PageHeader
        title={t('financeBankAccounts.title')}
        subtitle={
          accountsError ? (
            t('financeErrors.generic')
          ) : (
            <>
              {accounts.length} connected accounts <>&middot;</> <MoneyDisplay amount={total} />{' '}
              total balance
            </>
          )
        }
        actions={
          canManageAccounts ? (
            <button type="button" style={primaryBtn} onClick={() => setShowConnect(true)}>
              <LinkIcon size={13} /> Connect Account
            </button>
          ) : undefined
        }
      />

      {isLoading ? (
        <div style={{ ...cardStyle, padding: 20 }}>
          <SkeletonRow />
        </div>
      ) : accountsError ? (
        <div role="alert" style={{ ...cardStyle, padding: 20, color: 'var(--hub-text-muted)' }}>
          <p style={{ margin: '0 0 12px' }}>{t('financeErrors.generic')}</p>
          <button type="button" style={ghostBtn} onClick={() => void refetchAccounts()}>
            {t('financeCommon.refresh')}
          </button>
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
                  type="button"
                  key={a.id}
                  onClick={() => setSelected(selected === a.id ? undefined : a.id)}
                  style={{
                    ...cardStyle,
                    padding: 16,
                    textAlign: 'start',
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
            {transactionsLoading ? (
              <div style={{ padding: 16 }}>
                <SkeletonRow count={3} />
              </div>
            ) : transactionsError ? (
              <div role="alert" style={{ padding: 16 }}>
                <p style={{ margin: '0 0 12px', fontSize: 13, color: 'var(--hub-text)' }}>
                  {t('financeErrors.generic')}
                </p>
                <button type="button" style={ghostBtn} onClick={() => void refetchTransactions()}>
                  {t('financeCommon.refresh')}
                </button>
              </div>
            ) : transactions.length === 0 ? (
              <div
                style={{
                  padding: 20,
                  textAlign: 'center',
                  color: 'var(--hub-text-dim)',
                  fontSize: 13,
                }}
              >
                {t('financeCommon.noData')}
              </div>
            ) : (
              <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                <thead>
                  <tr style={{ borderBottom: '1px solid var(--hub-border)' }}>
                    {['Date', 'Description', 'Category', 'Amount', 'Match'].map((h) => (
                      <th
                        key={h}
                        style={{
                          padding: '10px 16px',
                          textAlign: h === 'Amount' ? 'end' : 'start',
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
                        style={{
                          padding: '10px 16px',
                          fontSize: 12,
                          color: 'var(--hub-text-muted)',
                        }}
                      >
                        {t.date}
                      </td>
                      <td style={{ padding: '10px 16px', fontSize: 13, color: 'var(--hub-text)' }}>
                        {t.description}
                      </td>
                      <td
                        style={{
                          padding: '10px 16px',
                          fontSize: 12,
                          color: 'var(--hub-text-muted)',
                        }}
                      >
                        {t.category}
                      </td>
                      <td
                        style={{
                          padding: '10px 16px',
                          fontSize: 13,
                          fontWeight: 600,
                          textAlign: 'end',
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
            )}
          </div>
        </>
      )}

      <Modal
        open={showConnect && canManageAccounts}
        onClose={() => setShowConnect(false)}
        title={t('financeBankAccounts.addAccount')}
        footer={
          <>
            <button
              type="button"
              style={ghostBtn}
              onClick={() => setShowConnect(false)}
              disabled={createAccount.isPending}
            >
              {t('financeCommon.cancel')}
            </button>
            <button
              type="button"
              style={{ ...primaryBtn, opacity: createAccount.isPending ? 0.6 : 1 }}
              onClick={handleCreateAccount}
              disabled={
                createAccount.isPending ||
                !form.name.trim() ||
                !form.bankName.trim() ||
                form.lastFour.length !== 4
              }
            >
              <CheckCircle size={13} />
              {createAccount.isPending ? 'Adding…' : 'Add account'}
            </button>
          </>
        }
      >
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
            <label style={labelStyle} htmlFor="bank-account-name">
              {t('financeBankAccounts.accountName')}
            </label>
            <input
              id="bank-account-name"
              autoFocus
              value={form.name}
              onChange={(event) => setForm((current) => ({ ...current, name: event.target.value }))}
              placeholder="e.g. Operating account"
              style={inputStyle}
            />
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
            <label style={labelStyle} htmlFor="bank-account-bank">
              {t('financeBankAccounts.bank')}
            </label>
            <input
              id="bank-account-bank"
              value={form.bankName}
              onChange={(event) =>
                setForm((current) => ({ ...current, bankName: event.target.value }))
              }
              placeholder="e.g. Acme Bank"
              style={inputStyle}
            />
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
              <label style={labelStyle} htmlFor="bank-account-type">
                {t('financeBankAccounts.typeLabel')}
              </label>
              <select
                id="bank-account-type"
                value={form.accountType}
                onChange={(event) =>
                  setForm((current) => ({ ...current, accountType: event.target.value }))
                }
                style={inputStyle}
              >
                {ACCOUNT_TYPES.map((type) => (
                  <option key={type} value={type}>
                    {t(`financeBankAccounts.type.${type}`)}
                  </option>
                ))}
              </select>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
              <label style={labelStyle} htmlFor="bank-account-last-four">
                {t('financeBankAccounts.last4')}
              </label>
              <input
                id="bank-account-last-four"
                inputMode="numeric"
                maxLength={4}
                value={form.lastFour}
                onChange={(event) =>
                  setForm((current) => ({
                    ...current,
                    lastFour: event.target.value.replace(/\D/g, '').slice(0, 4),
                  }))
                }
                placeholder="1234"
                style={inputStyle}
              />
            </div>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
              <label style={labelStyle} htmlFor="bank-account-balance">
                {t('financeBankAccounts.openingBalance')}
              </label>
              <input
                id="bank-account-balance"
                type="number"
                min="0"
                step="0.01"
                value={form.balance}
                onChange={(event) =>
                  setForm((current) => ({ ...current, balance: event.target.value }))
                }
                style={inputStyle}
              />
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
              <label style={labelStyle} htmlFor="bank-account-currency">
                {t('financeBankAccounts.currency')}
              </label>
              <input
                id="bank-account-currency"
                maxLength={3}
                value={form.currency}
                onChange={(event) =>
                  setForm((current) => ({ ...current, currency: event.target.value }))
                }
                style={inputStyle}
              />
            </div>
          </div>
          <p style={{ margin: 0, fontSize: 12, color: 'var(--hub-text-dim)' }}>
            This adds the account for tracking. Automatic bank synchronization is not configured
            yet.
          </p>
        </div>
      </Modal>
    </motion.div>
  );
};

export default BankFeeds;
