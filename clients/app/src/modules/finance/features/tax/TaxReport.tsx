import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import { ReportServiceClient } from '@voxire/voxie/finance/report/v1/client';
import { pageTitle } from '../../shared/lib/pageTitle';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import { FileSpreadsheet, Download, Landmark } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { EmptyState } from '@voxire/ui';
import PageHeader from '../../shared/components/PageHeader';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import { cardStyle, ghostBtn, ACCENT } from '../../shared/lib/tokens';
import { taxReportSchema } from '../../shared/schemas/tax';
import { todayIso } from '../../shared/lib/date';

const reportClient = createOneKitClient(ReportServiceClient, 'finance');

const TaxReport: React.FC = () => {
  const { t } = useTranslation('financeCommon');
  const reportDate = todayIso();
  const {
    data: taxReport,
    error,
    isError,
    isFetching,
    isPending,
    refetch,
  } = useQuery({
    queryKey: ['finance', 'tax-report', reportDate],
    queryFn: async () => {
      const response = await reportClient.getTaxReport({ period: 'month', date: reportDate });
      return taxReportSchema.parse(response);
    },
  });
  const loading = isPending && isFetching;
  const failed = isError || (isPending && !isFetching);

  const totalCollected = taxReport?.taxCollected ?? 0;
  const totalOwed = taxReport?.taxOwed ?? 0;
  const rows = taxReport
    ? [
        {
          id: taxReport.period || 'current',
          period: taxReport.period || 'Current period',
          collected: totalCollected,
          owed: totalOwed,
          status: totalOwed > 0 ? 'Pending' : 'Filed',
          dueDate: '-',
        },
      ]
    : [];
  const toFile = rows.filter((row) => row.status !== 'Filed').length;

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title={pageTitle(t('financeCommon.pageTitles.taxSummary'))} />
      <PageHeader
        title={t('financeTax.report.title')}
        breadcrumbs={[{ label: t('financeNav.reports') }, { label: t('financeNav.tax') }]}
        actions={
          <button type="button" style={ghostBtn} disabled title={t('financeTax.report.exportSoon')}>
            <Download size={13} /> Export
          </button>
        }
      />

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 14 }}>
        <div style={{ ...cardStyle, padding: 20 }}>
          <div style={{ fontSize: 12, color: 'var(--hub-text-muted)' }}>
            {t('financeTax.report.collected')}
          </div>
          <div style={{ fontSize: 22, fontWeight: 700, color: ACCENT, marginTop: 6 }}>
            <MoneyDisplay amount={totalCollected} />
          </div>
        </div>
        <div style={{ ...cardStyle, padding: 20 }}>
          <div style={{ fontSize: 12, color: 'var(--hub-text-muted)' }}>
            {t('financeTax.report.owed')}
          </div>
          <div
            style={{ fontSize: 22, fontWeight: 700, color: 'var(--finance-warning)', marginTop: 6 }}
          >
            <MoneyDisplay amount={totalOwed} />
          </div>
        </div>
        <div style={{ ...cardStyle, padding: 20 }}>
          <div style={{ fontSize: 12, color: 'var(--hub-text-muted)' }}>
            {t('financeTax.report.periodsToFile')}
          </div>
          <div style={{ fontSize: 22, fontWeight: 700, color: 'var(--hub-text)', marginTop: 6 }}>
            {toFile}
          </div>
        </div>
      </div>

      <div style={{ ...cardStyle, overflow: 'hidden' }}>
        {loading ? (
          <div
            style={{
              padding: '32px 20px',
              textAlign: 'center',
              color: 'var(--hub-text-muted)',
              fontSize: 13,
            }}
          >
            Loading tax periods…
          </div>
        ) : failed ? (
          <div role="alert" style={{ padding: '32px 20px', textAlign: 'center' }}>
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
        ) : rows.length === 0 ? (
          <EmptyState
            icon={Landmark}
            title={t('financeTax.reportEmptyTitle')}
            description={t('financeTax.reportEmptyDescription')}
            action={{
              label: t('financeCommon.refresh'),
              onClick: () => void refetch(),
            }}
          />
        ) : (
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr style={{ borderBottom: '1px solid var(--hub-border)' }}>
                {['Period', 'Collected', 'Owed', 'Status', 'Due date', 'Actions'].map((h) => (
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
              {rows.map((p) => {
                const filed = p.status === 'Filed';
                return (
                  <tr key={p.id} style={{ borderBottom: '1px solid var(--hub-border)' }}>
                    <td style={{ padding: '12px 16px', fontSize: 13, fontWeight: 600 }}>
                      {p.period}
                    </td>
                    <td style={{ padding: '12px 16px', fontSize: 13 }}>
                      <MoneyDisplay amount={p.collected} />
                    </td>
                    <td style={{ padding: '12px 16px', fontSize: 13 }}>
                      <MoneyDisplay amount={p.owed} />
                    </td>
                    <td style={{ padding: '12px 16px' }}>
                      <span
                        style={{
                          fontSize: 11,
                          padding: '3px 10px',
                          borderRadius: 20,
                          fontWeight: 600,
                          background: filed
                            ? 'var(--finance-success-soft)'
                            : 'color-mix(in srgb, var(--finance-warning) 12%, transparent)',
                          color: filed ? 'var(--finance-success)' : 'var(--finance-warning)',
                        }}
                      >
                        {p.status}
                      </span>
                    </td>
                    <td
                      style={{ padding: '12px 16px', fontSize: 12, color: 'var(--hub-text-muted)' }}
                    >
                      {p.dueDate}
                    </td>
                    <td style={{ padding: '12px 16px' }}>
                      <button
                        type="button"
                        style={{ ...ghostBtn, padding: '4px 10px' }}
                        disabled
                        onClick={undefined}
                      >
                        <FileSpreadsheet size={12} /> View
                      </button>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
      </div>
    </motion.div>
  );
};

export default TaxReport;
