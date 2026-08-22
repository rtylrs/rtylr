import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useUserStore } from '@voxire/auth';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { motion } from 'motion/react';
import { CheckCircle, Edit2, Plus, Power, Search, Trash2 } from 'lucide-react';
import PageHeader from '../../shared/components/PageHeader';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import AccountTypeBadge, { AccountType } from '../../shared/components/AccountTypeBadge';
import ConfirmDialog from '../../shared/components/ConfirmDialog';
import Modal from '../../shared/components/Modal';
import { SkeletonRow } from '../../shared/components/Skeleton';
import {
  useChartOfAccounts,
  useCreateAccount,
  useUpdateAccount,
  useDeleteAccount,
} from '../../shared/hooks/useChartOfAccounts';
import {
  cardStyle,
  ghostBtn,
  primaryBtn,
  inputStyle,
  labelStyle,
  ACCENT,
} from '../../shared/lib/tokens';
import { canFinancePermission } from '../../shared/lib/financePermissions';
import { useSettings } from '../../shared/lib/settings';
import type { ChartOfAccountRaw } from '../../shared/schemas/accounts';

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
  const { t } = useTranslation();
  const { data = [], isLoading } = useChartOfAccounts();
  const createMut = useCreateAccount();
  const updateMut = useUpdateAccount();
  const deleteMut = useDeleteAccount();
  // Derive the currency from settings exactly like MoneyDisplay does instead
  // of hardcoding USD.
  const currency = useSettings((s) => s.currency);
  const [filter, setFilter] = useState<AccountType | 'all'>('all');
  const [query, setQuery] = useState('');
  const [showCreate, setShowCreate] = useState(false);
  const [newName, setNewName] = useState('');
  const [newType, setNewType] = useState<(typeof ACCOUNT_TYPES)[number]>('asset');
  const [newCode, setNewCode] = useState('');

  const [editingAccount, setEditingAccount] = useState<ChartOfAccountRaw | null>(null);
  const [editName, setEditName] = useState('');
  const [editCode, setEditCode] = useState('');
  const [confirmToggleId, setConfirmToggleId] = useState<string | null>(null);
  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null);

  const role = useUserStore((state) => state.user?.role ?? state.role);
  // There is no dedicated chart-of-accounts resource in the permission matrix;
  // managing ledger accounts is an accounts-management action, so it is
  // gated on the same 'finance:accounts' resource used for bank accounts.
  const canManageAccounts = canFinancePermission(role, 'finance:accounts');

  const handleCreate = async (): Promise<void> => {
    if (!canManageAccounts) return;
    if (!newName.trim()) {
      toast.error(t('accounts.nameRequired'));
      return;
    }
    try {
      await createMut.mutateAsync({
        name: newName.trim(),
        type: newType,
        currency,
        balance: 0,
        isActive: true,
        code: newCode.trim() || undefined,
      });
      setShowCreate(false);
      setNewName('');
      setNewType('asset');
      setNewCode('');
    } catch {
      // onError in useCreateAccount handles the error toast
    }
  };

  const startEdit = (account: ChartOfAccountRaw): void => {
    setEditingAccount(account);
    setEditName(account.name);
    setEditCode(account.code ?? '');
  };

  const handleUpdate = (): void => {
    if (!editingAccount?.id) return;
    const trimmed = editName.trim();
    if (!trimmed) return;
    if (!editingAccount.type) {
      toast.error(t('accounts.typeRequired'));
      return;
    }
    updateMut.mutate(
      {
        id: editingAccount.id,
        name: trimmed,
        type: editingAccount.type,
        code: editCode.trim(),
        isActive: editingAccount.isActive !== false,
      },
      {
        onSuccess: () => {
          setEditingAccount(null);
        },
      }
    );
  };

  const pendingToggle =
    confirmToggleId == null ? null : (data.find((a) => a.id === confirmToggleId) ?? null);
  const pendingToggleActive = pendingToggle ? pendingToggle.isActive !== false : false;

  const handleToggleActive = (): void => {
    if (!pendingToggle?.id || !pendingToggle.type) return;
    updateMut.mutate({
      id: pendingToggle.id,
      name: pendingToggle.name,
      type: pendingToggle.type,
      isActive: !pendingToggleActive,
    });
  };

  const pendingDelete =
    confirmDeleteId == null ? null : (data.find((a) => a.id === confirmDeleteId) ?? null);

  const handleDelete = (): void => {
    if (!pendingDelete?.id) return;
    deleteMut.mutate(pendingDelete.id);
  };

  const filtered = data.filter((a) => {
    const matchesType = filter === 'all' || a.type === filter;
    const matchesQuery =
      !query ||
      a.name.toLowerCase().includes(query.toLowerCase()) ||
      (a.code ?? '').includes(query);
    return matchesType && matchesQuery;
  });

  // Align on the column key, not on a hardcoded translated label.
  const columns: Array<{ key: string; label: string }> = [
    { key: 'code', label: t('accounts.code') },
    { key: 'name', label: t('accounts.name') },
    { key: 'type', label: t('accounts.type') },
    { key: 'balance', label: t('accounts.balance') },
  ];

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title={`${t('accounts.title')} · rtylr`} />
      <PageHeader
        title={t('accounts.title')}
        subtitle={t('accounts.count', { count: filtered.length })}
        actions={
          canManageAccounts ? (
            <button type="button" style={primaryBtn} onClick={() => setShowCreate(true)}>
              <Plus size={13} /> {t('accounts.new')}
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
          {TYPES.map((type) => (
            <button
              type="button"
              key={type}
              onClick={() => setFilter(type)}
              style={{
                padding: '5px 12px',
                borderRadius: 7,
                border: 'none',
                fontSize: 12,
                fontWeight: 500,
                cursor: 'pointer',
                background: filter === type ? ACCENT : 'transparent',
                color: filter === type ? 'var(--hub-bg)' : 'var(--hub-text-muted)',
                textTransform: 'capitalize',
              }}
            >
              {type}
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
            placeholder={t('accounts.search')}
            aria-label={t('accounts.search')}
            style={{ ...inputStyle, paddingInlineStart: 30 }}
          />
        </div>
      </div>

      {isLoading ? (
        <div style={{ ...cardStyle, padding: 20 }}>
          <SkeletonRow />
        </div>
      ) : (
        <div style={{ ...cardStyle, overflowX: 'auto' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr style={{ borderBottom: '1px solid var(--hub-border)' }}>
                {columns.map((h) => (
                  <th
                    key={h.key}
                    style={{
                      padding: '10px 16px',
                      textAlign: h.key === 'balance' ? 'end' : 'start',
                      fontSize: 11,
                      fontWeight: 600,
                      textTransform: 'uppercase',
                      letterSpacing: '0.07em',
                      color: 'var(--hub-text-dim)',
                    }}
                  >
                    {h.label}
                  </th>
                ))}
                {canManageAccounts && (
                  <th
                    style={{
                      padding: '10px 16px',
                      textAlign: 'end',
                      fontSize: 11,
                      fontWeight: 600,
                      textTransform: 'uppercase',
                      letterSpacing: '0.07em',
                      color: 'var(--hub-text-dim)',
                    }}
                  >
                    {t('common.actions')}
                  </th>
                )}
              </tr>
            </thead>
            <tbody>
              {filtered.map((a) => {
                const active = a.isActive !== false;
                return (
                  <tr
                    key={a.id}
                    style={{
                      borderBottom: '1px solid var(--hub-border)',
                      opacity: active ? 1 : 0.55,
                    }}
                  >
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
                      {!active && (
                        <span
                          style={{
                            marginInlineStart: 8,
                            fontSize: 10,
                            padding: '1px 8px',
                            borderRadius: 10,
                            background: 'var(--hub-hover)',
                            color: 'var(--hub-text-muted)',
                          }}
                        >
                          {t('accounts.inactive')}
                        </span>
                      )}
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
                    {canManageAccounts && a.id && (
                      <td
                        style={{
                          padding: '12px 16px',
                          display: 'flex',
                          gap: 6,
                          justifyContent: 'flex-end',
                        }}
                      >
                        <button
                          type="button"
                          onClick={() => startEdit(a)}
                          className="p-1 rounded hover:bg-white/10 text-white/60 hover:text-white transition"
                          title={t('accounts.editTitle')}
                          aria-label={t('accounts.editAria', { name: a.name })}
                        >
                          <Edit2 size={13} />
                        </button>
                        <button
                          type="button"
                          onClick={() => setConfirmToggleId(a.id ?? null)}
                          className="p-1 rounded hover:bg-white/10 text-white/60 hover:text-white transition"
                          title={
                            active ? t('accounts.deactivateTitle') : t('accounts.reactivateTitle')
                          }
                          aria-label={
                            active
                              ? t('accounts.deactivateAria', { name: a.name })
                              : t('accounts.reactivateAria', { name: a.name })
                          }
                        >
                          <Power size={13} color={active ? undefined : 'var(--finance-accent)'} />
                        </button>
                        <button
                          type="button"
                          onClick={() => setConfirmDeleteId(a.id ?? null)}
                          className="p-1 rounded hover:bg-red-500/10 text-white/60 hover:text-red-400 transition"
                          title={t('accounts.deleteConfirmTitle')}
                          aria-label={t('accounts.deleteAria', { name: a.name })}
                        >
                          <Trash2 size={13} />
                        </button>
                      </td>
                    )}
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
      {showCreate && canManageAccounts && (
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
              {t('accounts.new')}
            </h3>
            <input
              style={{ ...inputStyle, fontSize: 16 }}
              placeholder={t('accounts.name')}
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              aria-label={t('accounts.name')}
            />
            <input
              style={{ ...inputStyle, fontSize: 16 }}
              placeholder={t('accounts.codeOptional')}
              value={newCode}
              onChange={(e) => setNewCode(e.target.value)}
              aria-label={t('accounts.codeOptional')}
            />
            <select
              style={{ ...inputStyle, fontSize: 16 }}
              value={newType}
              onChange={(e) => setNewType(e.target.value as (typeof ACCOUNT_TYPES)[number])}
              aria-label={t('accounts.type')}
            >
              {ACCOUNT_TYPES.map((type) => (
                <option key={type} value={type}>
                  {type.charAt(0).toUpperCase() + type.slice(1)}
                </option>
              ))}
            </select>
            <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
              <button type="button" style={ghostBtn} onClick={() => setShowCreate(false)}>
                {t('common.cancel')}
              </button>
              <button
                type="button"
                style={primaryBtn}
                onClick={() => void handleCreate()}
                disabled={createMut.isPending}
              >
                {createMut.isPending ? t('common.creating') : t('common.create')}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Edit modal - same pattern as the Categories edit modal */}
      <Modal
        open={Boolean(editingAccount)}
        onClose={() => setEditingAccount(null)}
        title={t('accounts.editTitle')}
        footer={
          <>
            <button
              type="button"
              style={ghostBtn}
              onClick={() => setEditingAccount(null)}
              disabled={updateMut.isPending}
            >
              {t('common.cancel')}
            </button>
            <button
              type="button"
              style={{ ...primaryBtn, opacity: updateMut.isPending ? 0.6 : 1 }}
              onClick={handleUpdate}
              disabled={updateMut.isPending || !editName.trim()}
            >
              <CheckCircle size={13} />
              {updateMut.isPending ? t('common.saving') : t('common.save')}
            </button>
          </>
        }
      >
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
            <label style={labelStyle} htmlFor="edit-account-name">
              {t('accounts.name')}
            </label>
            <input
              id="edit-account-name"
              autoFocus
              value={editName}
              onChange={(event) => setEditName(event.target.value)}
              style={{ ...inputStyle, fontSize: 16 }}
            />
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
            <label style={labelStyle} htmlFor="edit-account-code">
              {t('accounts.codeOptional')}
            </label>
            <input
              id="edit-account-code"
              value={editCode}
              onChange={(event) => setEditCode(event.target.value)}
              style={{ ...inputStyle, fontSize: 16 }}
            />
          </div>
        </div>
      </Modal>

      {/* Deactivate / reactivate confirmation */}
      <ConfirmDialog
        open={Boolean(pendingToggle)}
        onClose={() => setConfirmToggleId(null)}
        onConfirm={handleToggleActive}
        title={pendingToggleActive ? t('accounts.deactivateTitle') : t('accounts.reactivateTitle')}
        message={
          pendingToggleActive
            ? t('accounts.deactivateMessage', { name: pendingToggle?.name ?? '' })
            : t('accounts.reactivateMessage', { name: pendingToggle?.name ?? '' })
        }
        confirmLabel={t('common.confirm')}
      />

      {/* Delete confirmation */}
      <ConfirmDialog
        open={Boolean(pendingDelete)}
        onClose={() => setConfirmDeleteId(null)}
        onConfirm={handleDelete}
        title={t('accounts.deleteConfirmTitle')}
        message={t('accounts.deleteConfirmMessage', { name: pendingDelete?.name ?? '' })}
        confirmLabel={t('common.confirm')}
        destructive
      />
    </motion.div>
  );
};

export default ChartOfAccounts;
