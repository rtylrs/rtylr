import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useUserStore } from '@voxire/auth';
import { motion } from 'motion/react';
import {
  Download,
  Settings,
  CheckCircle,
  Clock,
  AlertTriangle,
  Landmark,
  Plus,
} from 'lucide-react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { EmptyState } from '@voxire/ui';
import { TaxPeriod, TaxConfig } from '../../shared/services/financeApi';
import { useTaxReport, useTaxConfig, useCreateTaxConfig } from '../../shared/hooks/useTax';
import { exportTaxSummaryCsv } from '../../shared/lib/exports';
import { canFinancePermission } from '../../shared/lib/financePermissions';

const ACCENT = 'var(--finance-accent)';

const fmt = (n: number) =>
  n.toLocaleString('en-US', { style: 'currency', currency: 'USD', minimumFractionDigits: 2 });

const statusStyle: Record<TaxPeriod['status'], React.CSSProperties> = {
  Filed: {
    background: 'color-mix(in srgb, var(--finance-accent) 12%, transparent)',
    color: 'var(--finance-accent)',
  },
  Pending: {
    background: 'color-mix(in srgb, var(--finance-warning) 12%, transparent)',
    color: 'var(--finance-warning)',
  },
  Due: {
    background: 'color-mix(in srgb, var(--finance-danger) 12%, transparent)',
    color: 'var(--finance-danger)',
  },
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
interface EditableRate {
  name: string;
  rate: number;
}

const Tax: React.FC = () => {
  const { t } = useTranslation();
  const [showConfig, setShowConfig] = useState(false);
  const { data: taxConfigData } = useTaxConfig();
  const saveRates = useCreateTaxConfig();
  const role = useUserStore((state) => state.user?.role ?? state.role);
  const canManageTax = canFinancePermission(role, 'finance:tax');

  // Server rows are the source of truth and render read-only: the generated
  // report client exposes only listTaxConfig/createTaxConfig - there is no
  // update or delete RPC, so an existing rate cannot be edited from here
  // without silently creating a duplicate config.
  const serverRates: TaxConfig[] = (taxConfigData ?? []).map((c) => ({
    id: c.id,
    name: c.name,
    rate: c.rate,
  }));
  // Local state holds ONLY unsaved new rows, never copies of server data, so
  // it cannot go stale and shadow fresh results after ['tax-config'] refetches.
  const [pendingRates, setPendingRates] = useState<EditableRate[]>([]);

  const addPendingRate = (): void => {
    setPendingRates((prev) => [...prev, { name: '', rate: 0 }]);
  };

  const updatePendingRate = (index: number, patch: Partial<EditableRate>): void => {
    setPendingRates((prev) => prev.map((r, i) => (i === index ? { ...r, ...patch } : r)));
  };

  const removePendingRate = (index: number): void => {
    setPendingRates((prev) => prev.filter((_, i) => i !== index));
  };

  const { data: taxReport } = useTaxReport();

  const handleSaveRates = async (): Promise<void> => {
    if (pendingRates.length === 0) return;
    const invalid = pendingRates.find((r) => !r.name.trim() || isNaN(r.rate) || r.rate < 0);
    if (invalid) {
      toast.error(t('tax.ratesInvalid'));
      return;
    }
    let created = 0;
    try {
      for (const r of pendingRates) {
        await saveRates.mutateAsync({ name: r.name.trim(), rate: r.rate });
        created += 1;
      }
    } catch {
      // Earlier creates already live on the server; report the partial outcome
      // honestly and keep only the rows that were not saved.
      const failed = pendingRates.length - created;
      toast.warning(t('tax.ratesPartial', { created, failed }));
      setPendingRates((prev) => prev.slice(created));
      return;
    }
    // Drop the saved rows so the panel reflects fresh ['tax-config'] data.
    setPendingRates([]);
    toast.success(t('tax.ratesSaved', { count: created }));
  };

  const taxCollected = taxReport?.taxCollected ?? 0;
  const taxOwed = taxReport?.taxOwed ?? 0;
  const taxPeriods: TaxPeriod[] = [];
  const unremitted = taxOwed;

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.25 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title="Tax · rtylr" />
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
          {canManageTax && (
            <button
              type="button"
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
          )}
          <button
            type="button"
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 6,
              padding: '7px 14px',
              borderRadius: 8,
              background: ACCENT,
              color: 'var(--finance-white)',
              border: 'none',
              fontSize: 13,
              fontWeight: 600,
              cursor: 'pointer',
            }}
            onClick={() =>
              exportTaxSummaryCsv([
                {
                  period: taxReport?.period ?? 'Current Period',
                  collected: taxCollected,
                  owed: taxOwed,
                  paid: 0,
                },
                ...taxPeriods.map((p) => ({
                  period: p.period,
                  collected: p.collected,
                  owed: p.taxableAmount * p.taxRate,
                  paid: p.status === 'Filed' ? p.collected : 0,
                })),
              ])
            }
            aria-label="Export tax summary for accountant"
          >
            <Download size={14} />
            Export for Accountant
          </button>
        </div>
      </div>

      {/* VAT Summary cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 16 }}>
        {[
          {
            label: 'Tax Collected',
            value: taxCollected,
            color: 'var(--finance-accent)',
            note: 'From invoices',
          },
          {
            label: 'Tax Owed',
            value: taxOwed,
            color: 'var(--finance-warning)',
            note: 'Net payable',
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
            background: 'color-mix(in srgb, var(--finance-warning) 8%, transparent)',
            border: '1px solid color-mix(in srgb, var(--finance-warning) 25%, transparent)',
          }}
        >
          <AlertTriangle
            size={16}
            style={{ color: 'var(--finance-warning)', flexShrink: 0, marginTop: 1 }}
          />
          <div>
            <p
              style={{
                margin: 0,
                fontSize: 13,
                fontWeight: 600,
                color: 'var(--finance-warning)',
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
      {showConfig && canManageTax && (
        <motion.div
          initial={{ opacity: 0, y: -8 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.2 }}
          style={{
            ...cardStyle,
            border: '1px solid color-mix(in srgb, var(--finance-accent) 30%, transparent)',
          }}
        >
          <h3
            style={{ margin: '0 0 14px', fontSize: 14, fontWeight: 600, color: 'var(--hub-text)' }}
          >
            Tax Rate Configuration
          </h3>
          {serverRates.length === 0 && (
            <EmptyState
              size="compact"
              icon={Landmark}
              title={t('tax.ratesEmptyTitle')}
              description={t('tax.ratesEmptyDescription')}
              action={{ label: t('tax.ratesEmptyCta'), onClick: addPendingRate }}
            />
          )}
          {serverRates.length > 0 && (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8, marginBottom: 14 }}>
              {serverRates.map((r) => (
                <div
                  key={r.id ?? r.name}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    gap: 12,
                    padding: '7px 12px',
                    borderRadius: 7,
                    border: '1px solid var(--hub-border)',
                    background: 'var(--hub-hover)',
                  }}
                >
                  <span style={{ fontSize: 13, color: 'var(--hub-text)' }}>{r.name}</span>
                  <span
                    style={{
                      fontSize: 13,
                      color: 'var(--hub-text-muted)',
                      fontVariantNumeric: 'tabular-nums',
                    }}
                  >
                    {r.rate}%
                  </span>
                </div>
              ))}
              <p style={{ margin: 0, fontSize: 11, color: 'var(--hub-text-dim)' }}>
                {t('tax.ratesEditUnsupported')}
              </p>
            </div>
          )}
          <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
            {pendingRates.map((r, i) => (
              <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                <input
                  value={r.name}
                  onChange={(e) => updatePendingRate(i, { name: e.target.value })}
                  aria-label={t('tax.rateNameLabel')}
                  placeholder={t('tax.rateNameLabel')}
                  style={{
                    flex: 1,
                    padding: '7px 12px',
                    borderRadius: 7,
                    border: '1px solid var(--hub-border)',
                    background: 'var(--hub-input-bg)',
                    color: 'var(--hub-text)',
                    fontSize: 16,
                    outline: 'none',
                  }}
                />
                <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                  <input
                    type="number"
                    value={r.rate}
                    step={0.5}
                    min={0}
                    onChange={(e) => {
                      const parsed = parseFloat(e.target.value);
                      updatePendingRate(i, { rate: isNaN(parsed) ? 0 : parsed });
                    }}
                    aria-label={t('tax.rateValueLabel')}
                    style={{
                      width: 72,
                      padding: '7px 12px',
                      borderRadius: 7,
                      border: '1px solid var(--hub-border)',
                      background: 'var(--hub-input-bg)',
                      color: 'var(--hub-text)',
                      fontSize: 16,
                      outline: 'none',
                      textAlign: 'end',
                    }}
                  />
                  <span style={{ fontSize: 13, color: 'var(--hub-text-muted)' }}>%</span>
                </div>
                <button
                  type="button"
                  onClick={() => removePendingRate(i)}
                  className="p-1 rounded hover:bg-red-500/10 text-white/60 hover:text-red-400 transition"
                  title={t('tax.removeRate')}
                  aria-label={t('tax.removeRate')}
                >
                  ✕
                </button>
              </div>
            ))}
          </div>
          <button
            type="button"
            onClick={addPendingRate}
            style={{
              marginTop: 12,
              alignSelf: 'flex-start',
              display: 'flex',
              alignItems: 'center',
              gap: 6,
              padding: '6px 12px',
              borderRadius: 7,
              border: '1px dashed var(--hub-border)',
              background: 'transparent',
              color: ACCENT,
              fontSize: 12,
              fontWeight: 600,
              cursor: 'pointer',
            }}
          >
            <Plus size={13} /> {t('tax.addRate')}
          </button>
          <button
            type="button"
            disabled={saveRates.isPending || pendingRates.length === 0}
            style={{
              marginTop: 14,
              padding: '7px 16px',
              borderRadius: 7,
              background: ACCENT,
              color: 'var(--finance-white)',
              border: 'none',
              fontSize: 13,
              fontWeight: 600,
              cursor: saveRates.isPending || pendingRates.length === 0 ? 'not-allowed' : 'pointer',
              opacity: saveRates.isPending || pendingRates.length === 0 ? 0.6 : 1,
            }}
            onClick={() => void handleSaveRates()}
          >
            {saveRates.isPending ? t('common.saving') : t('tax.saveRates')}
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
                    textAlign: 'start',
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
                    color: tp.status === 'Due' ? 'var(--finance-danger)' : 'var(--hub-text-muted)',
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
