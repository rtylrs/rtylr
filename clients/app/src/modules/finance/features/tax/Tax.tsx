import React, { useState } from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import { Plus } from 'lucide-react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { getOneKitErrorMessage } from '@voxire/api';
import { useTaxConfig, useCreateTaxConfig } from '../../shared/hooks/useTax';
import { canFinancePermission } from '../../shared/lib/financePermissions';
import PageHeader from '../../shared/components/PageHeader';
import { SkeletonRow } from '../../shared/components/Skeleton';
import { cardStyle, ghostBtn, inputStyle, labelStyle, primaryBtn } from '../../shared/lib/tokens';

/**
 * Tax rates the business charges on invoices and quotes. The service can
 * add a rate but not edit or remove one, so existing rates are listed as they
 * are and only the form below writes. It used to re-save every listed rate on
 * each save, creating a copy of each one every time.
 */
const Tax: React.FC = () => {
  const { t } = useTranslation('financeCommon');
  const { data: rates = [], isLoading, isError, refetch } = useTaxConfig();
  const createRate = useCreateTaxConfig();
  const canManageTax = canFinancePermission('finance:tax:configure');
  const [name, setName] = useState('');
  const [rate, setRate] = useState('');

  const addRate = async (): Promise<void> => {
    const parsed = Number(rate);
    if (!name.trim() || rate.trim() === '' || !Number.isFinite(parsed) || parsed < 0) {
      toast.error(t('financeTax.ratesInvalid'));
      return;
    }
    try {
      await createRate.mutateAsync({ name: name.trim(), rate: parsed });
      toast.success(t('financeTax.rateAdded'));
      setName('');
      setRate('');
    } catch (error) {
      toast.error(getOneKitErrorMessage(error, t('financeErrors.generic')));
    }
  };

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.25 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title={pageTitle(t('financeTax.ratesTitle'))} />
      <PageHeader title={t('financeTax.ratesTitle')} subtitle={t('financeTax.ratesSubtitle')} />

      <div style={{ ...cardStyle, overflow: 'hidden' }}>
        {isLoading ? (
          <div style={{ padding: 20 }}>
            <SkeletonRow />
          </div>
        ) : isError ? (
          <div
            role="alert"
            style={{ padding: 20, display: 'flex', flexDirection: 'column', gap: 12 }}
          >
            <p style={{ margin: 0, fontSize: 13, color: 'var(--hub-text)' }}>
              {t('financeErrors.generic')}
            </p>
            <button
              type="button"
              style={{ ...ghostBtn, alignSelf: 'flex-start' }}
              onClick={() => void refetch()}
            >
              {t('financeCommon.refresh')}
            </button>
          </div>
        ) : rates.length === 0 ? (
          <div style={{ padding: 20 }}>
            <h3 style={{ margin: 0, fontSize: 14, fontWeight: 600, color: 'var(--hub-text)' }}>
              {t('financeTax.ratesEmptyTitle')}
            </h3>
            <p style={{ margin: '6px 0 0', fontSize: 13, color: 'var(--hub-text-muted)' }}>
              {t('financeTax.ratesEmptyDescription')}
            </p>
          </div>
        ) : (
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr style={{ borderBottom: '1px solid var(--hub-border)' }}>
                <th style={th}>{t('financeTax.rateName')}</th>
                <th style={{ ...th, textAlign: 'end' }}>{t('financeTax.rateValue')}</th>
              </tr>
            </thead>
            <tbody>
              {rates.map((r, i) => (
                <tr
                  key={r.id ?? `${r.name}-${i}`}
                  style={{
                    borderBottom: i < rates.length - 1 ? '1px solid var(--hub-border)' : 'none',
                  }}
                >
                  <td style={{ padding: '12px 16px', fontSize: 13, color: 'var(--hub-text)' }}>
                    {r.name}
                  </td>
                  <td
                    style={{
                      padding: '12px 16px',
                      fontSize: 13,
                      fontWeight: 600,
                      textAlign: 'end',
                      color: 'var(--hub-text)',
                      fontVariantNumeric: 'tabular-nums',
                    }}
                  >
                    {r.rate}%
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {canManageTax && (
        <form
          onSubmit={(e) => {
            e.preventDefault();
            void addRate();
          }}
          style={{
            ...cardStyle,
            padding: 20,
            display: 'flex',
            alignItems: 'flex-end',
            gap: 12,
            flexWrap: 'wrap',
          }}
        >
          <div style={{ flex: '1 1 220px' }}>
            <label style={labelStyle} htmlFor="tax-rate-name">
              {t('financeTax.rateName')}
            </label>
            <input
              id="tax-rate-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder={t('financeTax.rateNamePlaceholder')}
              style={{ ...inputStyle, marginTop: 6 }}
            />
          </div>
          <div style={{ width: 120 }}>
            <label style={labelStyle} htmlFor="tax-rate-value">
              {t('financeTax.rateValue')}
            </label>
            <input
              id="tax-rate-value"
              type="number"
              min={0}
              step={0.01}
              value={rate}
              onChange={(e) => setRate(e.target.value)}
              placeholder="11"
              style={{ ...inputStyle, marginTop: 6, textAlign: 'end' }}
            />
          </div>
          <button type="submit" style={primaryBtn} disabled={createRate.isPending}>
            <Plus size={13} /> {t('financeTax.ratesEmptyCta')}
          </button>
        </form>
      )}
    </motion.div>
  );
};

const th: React.CSSProperties = {
  padding: '10px 16px',
  textAlign: 'start',
  fontSize: 11,
  fontWeight: 600,
  textTransform: 'uppercase',
  letterSpacing: '0.07em',
  color: 'var(--hub-text-dim)',
};

export default Tax;
