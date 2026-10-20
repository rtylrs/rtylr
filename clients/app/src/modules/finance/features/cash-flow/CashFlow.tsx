import React, { useMemo, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { AlertTriangle, RotateCcw } from 'lucide-react';
import PageHeader from '../../shared/components/PageHeader';
import CashFlowChart from '../../shared/components/CashFlowChart';
import PeriodSelector from '../../shared/components/PeriodSelector';
import { SkeletonRow } from '../../shared/components/Skeleton';
import { useCashFlow } from './useCashFlow';
import { bucketCashFlow, type CashFlowPeriod } from './bucketCashFlow';
import { cardStyle, ghostBtn } from '../../shared/lib/tokens';
import { formatMoney } from '../../shared/lib/money';
import { useSettings } from '../../shared/lib/settings';

const CashFlow: React.FC = () => {
  const { t } = useTranslation('financeCommon');
  const currency = useSettings((s) => s.currency);
  const [period, setPeriod] = useState<CashFlowPeriod>('month');
  const { data, isPending, isError, isFetching, refetch } = useCashFlow({ period });
  const points = useMemo(() => data ?? [], [data]);
  const buckets = useMemo(() => bucketCashFlow(points, period), [points, period]);
  const loading = isPending || (!data && !isError);

  const totalIn = points.reduce((s, d) => s + d.inflow, 0);
  const totalOut = points.reduce((s, d) => s + d.outflow, 0);
  const net = totalIn - totalOut;
  const empty = totalIn === 0 && totalOut === 0;
  const periodName = t(`finance.cashFlow.periods.${period}`).toLowerCase();

  const cards = [
    { id: 'in', label: t('finance.cashFlow.totalInflow'), value: formatMoney(totalIn, currency) },
    {
      id: 'out',
      label: t('finance.cashFlow.totalOutflow'),
      value: formatMoney(totalOut, currency),
    },
    {
      id: 'net',
      label: t('finance.cashFlow.netCashFlow'),
      value: `${net > 0 ? '+' : ''}${formatMoney(net, currency)}`,
      color: net < 0 ? 'var(--finance-danger)' : net > 0 ? 'var(--finance-success)' : undefined,
    },
  ];

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title={`${t('finance.cashFlow.title')} · rtylr`} />
      <PageHeader
        title={t('finance.cashFlow.title')}
        subtitle={t('finance.cashFlow.subtitle')}
        breadcrumbs={[
          { label: t('finance.cashFlow.reports') },
          { label: t('finance.cashFlow.title') },
        ]}
        actions={<PeriodSelector period={period} onPeriodChange={setPeriod} />}
      />

      {loading ? (
        <div style={{ ...cardStyle, padding: 20 }}>
          <SkeletonRow />
        </div>
      ) : isError ? (
        <div
          role="alert"
          style={{ ...cardStyle, padding: 20, display: 'flex', flexDirection: 'column', gap: 12 }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <AlertTriangle size={16} style={{ color: 'var(--finance-danger)' }} />
            <span style={{ fontSize: 13, fontWeight: 600, color: 'var(--hub-text)' }}>
              {t('financeErrors.generic')}
            </span>
          </div>
          <button
            type="button"
            style={{ ...ghostBtn, alignSelf: 'flex-start' }}
            disabled={isFetching}
            onClick={() => void refetch()}
          >
            <RotateCcw size={13} /> {t('financeCommon.refresh')}
          </button>
        </div>
      ) : (
        <>
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
              gap: 14,
            }}
          >
            {cards.map((card) => (
              <div key={card.id} style={{ ...cardStyle, padding: 20 }}>
                <div style={{ fontSize: 12, color: 'var(--hub-text-muted)' }}>{card.label}</div>
                <div
                  style={{
                    fontSize: 22,
                    fontWeight: 700,
                    marginTop: 6,
                    color: card.color ?? 'var(--hub-text)',
                    fontVariantNumeric: 'tabular-nums',
                  }}
                >
                  {card.value}
                </div>
              </div>
            ))}
          </div>

          <div style={{ ...cardStyle, padding: 24 }} aria-busy={isFetching}>
            {empty ? (
              <p
                style={{
                  margin: 0,
                  fontSize: 13,
                  color: 'var(--hub-text-muted)',
                  lineHeight: 1.6,
                }}
              >
                {t('finance.cashFlow.empty', { period: periodName })}
              </p>
            ) : (
              <CashFlowChart data={buckets} currency={currency} />
            )}
          </div>
          <p style={{ margin: 0, fontSize: 12, color: 'var(--hub-text-dim)' }}>
            {t('finance.cashFlow.basisNote')}
          </p>
        </>
      )}
    </motion.div>
  );
};

export default CashFlow;
