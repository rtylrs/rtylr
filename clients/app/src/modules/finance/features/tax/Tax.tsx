import React, { useState } from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { HelmetContainer } from '@voxire/components';
import { useUserStore } from '../../../../shared/stores/userStore';
import { motion } from 'motion/react';
import { Download, Settings, CheckCircle, Clock, AlertTriangle, Landmark } from 'lucide-react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { getOneKitErrorMessage } from '@voxire/api';
import { EmptyState } from '@voxire/ui';
import { TaxPeriod } from '../../shared/services/financeApi';
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
interface TaxConfig {
  name: string;
  rate: number;
}

const Tax: React.FC = () => {
  const { t } = useTranslation('financeCommon');
  const [showConfig, setShowConfig] = useState(false);
  const { data: taxConfigData } = useTaxConfig();
  const saveRates = useCreateTaxConfig();
  const canManageTax = canFinancePermission('finance:tax:configure');

  // Derive editable rates from real data; fall back to empty list while loading
  const serverRates: TaxConfig[] = (taxConfigData ?? []).map((c) => ({
    name: c.name,
    rate: c.rate,
  }));
  const [rates, setRates] = useState<TaxConfig[] | null>(null);
  // Sync server rates into local state once loaded (only on first load)
  const editableRates: TaxConfig[] = rates ?? serverRates;

  const { data: taxReport } = useTaxReport();

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
      <HelmetContainer title={pageTitle(t('financeNav.tax'))} />
      {/* Header */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div>
          <h2 style={{ margin: 0, fontSize: 18, fontWeight: 700, color: 'var(--hub-text)' }}>
            {t('financeTax.title')}
          </h2>
          <p style={{ margin: '2px 0 0', fontSize: 13, color: 'var(--hub-text-muted)' }}>
            {t('financeTax.subtitle')}
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
              {t('financeTax.rates')}
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
            aria-label={t('financeTax.exportAria')}
          >
            <Download size={14} />
            {t('financeTax.exportForAccountant')}
          </button>
        </div>
      </div>

      {/* VAT Summary cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 16 }}>
        {[
          {
            id: 'collected',

            label: t('financeTax.collected'),
            value: taxCollected,
            color: 'var(--finance-accent)',
            note: 'From invoices',
          },
          {
            id: 'owed',

            label: t('financeTax.owed'),
            value: taxOwed,
            color: 'var(--finance-warning)',
            note: 'Net payable',
          },
        ].map(({ id, label, value, color, note }) => (
          <div
            key={id}
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
            {t('financeTax.rateConfiguration')}
          </h3>
          {editableRates.length === 0 && (
            <EmptyState
              size="compact"
              icon={Landmark}
              title={t('financeTax.ratesEmptyTitle')}
              description={t('financeTax.ratesEmptyDescription')}
              action={{
                label: t('financeTax.ratesEmptyCta'),
                onClick: () => setRates([...editableRates, { name: '', rate: 0 }]),
              }}
            />
          )}
          <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
            {editableRates.map((r, i) => (
              <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                <input
                  value={r.name}
                  onChange={(e) => {
                    const next = [...editableRates];
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
                      const parsed = parseFloat(e.target.value);
                      if (isNaN(parsed)) return;
                      const next = [...editableRates];
                      next[i] = { ...next[i], rate: parsed };
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
                      textAlign: 'end',
                    }}
                  />
                  <span style={{ fontSize: 13, color: 'var(--hub-text-muted)' }}>%</span>
                </div>
              </div>
            ))}
          </div>
          <button
            type="button"
            disabled={saveRates.isPending}
            style={{
              marginTop: 14,
              padding: '7px 16px',
              borderRadius: 7,
              background: ACCENT,
              color: 'var(--finance-white)',
              border: 'none',
              fontSize: 13,
              fontWeight: 600,
              cursor: saveRates.isPending ? 'not-allowed' : 'pointer',
              opacity: saveRates.isPending ? 0.6 : 1,
            }}
            onClick={async () => {
              const invalid = editableRates.find((r) => isNaN(r.rate) || r.rate < 0);
              if (invalid) {
                toast.error(t('financeTax.ratesInvalid'));
                return;
              }
              try {
                for (const r of editableRates) {
                  await saveRates.mutateAsync(r);
                }
                toast.success(t('financeTax.ratesSaved'));
                setShowConfig(false);
              } catch (error) {
                toast.error(getOneKitErrorMessage(error, 'Failed to save tax rates'));
              }
            }}
          >
            {saveRates.isPending ? 'Saving…' : 'Save Rates'}
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
            {t('financeTax.filingPeriods')}
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
