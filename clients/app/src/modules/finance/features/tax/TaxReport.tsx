import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import { ReportServiceClient } from '@voxire/voxie/finance/report/v1/client';
import { Link } from '@tanstack/react-router';
import { pageTitle } from '../../shared/lib/pageTitle';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import { Download } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import PageHeader from '../../shared/components/PageHeader';
import { cardStyle, ghostBtn, ACCENT } from '../../shared/lib/tokens';
import { taxReportSchema } from '../../shared/schemas/tax';
import { todayIso } from '../../shared/lib/date';
import { formatMoney } from '../../shared/lib/money';
import { useSettings } from '../../shared/lib/settings';
import { exportTaxSummaryCsv } from '../../shared/lib/exports';

const reportClient = createOneKitClient(ReportServiceClient, 'finance');

type Period = 'month' | 'quarter' | 'year';
const PERIODS: Period[] = ['month', 'quarter', 'year'];

/**
 * Tax collected on paid invoices and tax owed for the chosen period. There is
 * no filing workflow behind this page, so it reports the figures and says
 * where they come from rather than showing filing statuses it cannot know.
 */
const TaxReport: React.FC = () => {
  const { t } = useTranslation('financeCommon');
  const currency = useSettings((s) => s.currency);
  const [period, setPeriod] = useState<Period>('month');
  const reportDate = todayIso();
  const {
    data: taxReport,
    error,
    isError,
    isFetching,
    isPending,
    refetch,
  } = useQuery({
    queryKey: ['finance', 'tax-report', period, reportDate],
    queryFn: async () => {
      const response = await reportClient.getTaxReport({ period, date: reportDate });
      return taxReportSchema.parse(response);
    },
  });

  const collected = taxReport?.taxCollected ?? 0;
  const owed = taxReport?.taxOwed ?? 0;
  const periodLabel = t(`financeTax.report.periods.${period}`);
  const nothingYet = !!taxReport && collected === 0 && owed === 0;

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title={pageTitle(t('financeCommon.pageTitles.taxSummary'))} />
      <PageHeader
        title={t('financeTax.report.title')}
        subtitle={t('financeTax.report.subtitle')}
        actions={
          <button
            type="button"
            style={ghostBtn}
            disabled={!taxReport}
            onClick={() => exportTaxSummaryCsv([{ period: periodLabel, collected, owed, paid: 0 }])}
            aria-label={t('financeTax.exportAria')}
          >
            <Download size={13} /> {t('financeTax.exportForAccountant')}
          </button>
        }
      />

      <div
        role="tablist"
        style={{
          display: 'flex',
          gap: 4,
          padding: 4,
          background: 'var(--hub-surface)',
          borderRadius: 10,
          border: '1px solid var(--hub-border)',
          width: 'fit-content',
        }}
      >
        {PERIODS.map((p) => (
          <button
            type="button"
            role="tab"
            aria-selected={period === p}
            key={p}
            onClick={() => setPeriod(p)}
            style={{
              padding: '5px 16px',
              borderRadius: 7,
              border: 'none',
              fontSize: 12,
              fontWeight: 500,
              cursor: 'pointer',
              background: period === p ? ACCENT : 'transparent',
              color: period === p ? 'var(--finance-on-accent)' : 'var(--hub-text-muted)',
            }}
          >
            {t(`financeTax.report.periods.${p}`)}
          </button>
        ))}
      </div>

      {isPending ? (
        <div style={{ ...cardStyle, padding: 20, fontSize: 13, color: 'var(--hub-text-muted)' }}>
          {t('financeCommon.loading')}
        </div>
      ) : isError ? (
        <div role="alert" style={{ ...cardStyle, padding: 20 }}>
          <p style={{ margin: '0 0 12px', color: 'var(--hub-text)', fontSize: 13 }}>
            {getOneKitErrorMessage(error, t('financeErrors.generic'))}
          </p>
          <button
            type="button"
            style={ghostBtn}
            disabled={isFetching}
            onClick={() => void refetch()}
          >
            {t('financeCommon.refresh')}
          </button>
        </div>
      ) : (
        <>
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))',
              gap: 14,
            }}
          >
            {[
              {
                id: 'collected',
                label: t('financeTax.report.collected'),
                hint: t('financeTax.report.collectedHint'),
                value: collected,
              },
              {
                id: 'owed',
                label: t('financeTax.report.owed'),
                hint: t('financeTax.report.owedHint'),
                value: owed,
              },
            ].map((card) => (
              <div key={card.id} style={{ ...cardStyle, padding: 20 }}>
                <div style={{ fontSize: 12, color: 'var(--hub-text-muted)' }}>{card.label}</div>
                <div
                  style={{
                    fontSize: 22,
                    fontWeight: 700,
                    color: 'var(--hub-text)',
                    marginTop: 6,
                    fontVariantNumeric: 'tabular-nums',
                  }}
                >
                  {formatMoney(card.value, currency)}
                </div>
                <div style={{ fontSize: 12, color: 'var(--hub-text-dim)', marginTop: 4 }}>
                  {card.hint}
                </div>
              </div>
            ))}
          </div>
          {nothingYet && (
            <div
              style={{
                ...cardStyle,
                padding: 20,
                fontSize: 13,
                color: 'var(--hub-text-muted)',
                lineHeight: 1.5,
              }}
            >
              {t('financeTax.report.nothingYet', { period: periodLabel.toLowerCase() })}{' '}
              <Link to="/finance/tax" style={{ color: 'var(--finance-accent)', fontWeight: 600 }}>
                {t('financeTax.report.manageRates')}
              </Link>
            </div>
          )}
        </>
      )}
    </motion.div>
  );
};

export default TaxReport;
