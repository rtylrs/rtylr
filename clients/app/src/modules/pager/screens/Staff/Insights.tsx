import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { Alert } from '@voxire/ui';
import type { GetStatsResponse } from '@voxire/voxie/pager/v1/types';
import { ArrowLeft } from 'lucide-react';
import { Button } from '../../../../shared/components/ui';
import { PAGER_BOARD_PATH } from '../../pagerPaths';
import { errorStatus } from '../../shared/lib/apiError';
import { staffPagerClient } from '../../shared/lib/clients';

const PERIODS = [1, 7, 30] as const;
type Period = (typeof PERIODS)[number];

function hourLabel(hour: number): string {
  return `${String(hour).padStart(2, '0')}:00`;
}

interface BarsProps {
  values: number[];
  labels: string[];
  /** Which labels to print under the bars; every bar keeps its own name. */
  tickEvery: number;
  barName: (index: number) => string;
  summary: string;
}

/** A row of bars scaled to the largest value, with a sentence for screen readers. */
function Bars({ values, labels, tickEvery, barName, summary }: BarsProps): React.JSX.Element {
  const max = Math.max(1, ...values);
  return (
    <figure className="m-0 space-y-2">
      <div className="flex h-36 items-end gap-[3px]" role="img" aria-label={summary} dir="ltr">
        {values.map((value, index) => (
          <div
            key={labels[index]}
            title={barName(index)}
            className="min-w-0 flex-1 rounded-t-sm"
            style={{
              height: `${Math.max(value > 0 ? 6 : 2, (value / max) * 100)}%`,
              background: value > 0 ? 'var(--dash-text)' : 'var(--dash-border)',
            }}
          />
        ))}
      </div>
      <div
        className="flex gap-[3px] text-[11px]"
        style={{ color: 'var(--dash-text-dim)' }}
        aria-hidden="true"
        dir="ltr"
      >
        {labels.map((label, index) => (
          <span key={label} className="min-w-0 flex-1 overflow-visible whitespace-nowrap">
            {index % tickEvery === 0 ? label : ''}
          </span>
        ))}
      </div>
    </figure>
  );
}

/**
 * How the venue's orders went: counts, waits from scan to Ready, and when the
 * rush comes, in the venue's own time zone.
 */
const Insights = (): React.JSX.Element => {
  const { t, i18n } = useTranslation('pager');
  const [days, setDays] = useState<Period>(7);
  const stats = useQuery({
    queryKey: ['pager', 'stats', days],
    queryFn: () => staffPagerClient.getStats({ days }),
    retry: false,
  });

  const number = new Intl.NumberFormat(i18n.language, { maximumFractionDigits: 1 });
  const duration = (seconds: number | undefined): string => {
    const value = seconds ?? 0;
    return value < 60
      ? t('pager.insights.seconds', { value: number.format(value) })
      : t('pager.insights.minutes', { value: number.format(value / 60) });
  };
  const dayLabel = new Intl.DateTimeFormat(i18n.language, { day: 'numeric', month: 'short' });

  const content = (data: GetStatsResponse): React.ReactNode => {
    const orders = data.orders ?? 0;
    if (orders === 0) {
      return (
        <p className="text-sm" style={{ color: 'var(--dash-text-dim)' }}>
          {t('pager.insights.empty')}
        </p>
      );
    }
    const byHour = data.byHour ?? [];
    const busiest = byHour.reduce((best, value, hour) => (value > byHour[best] ? hour : best), 0);
    const byDay = data.byDay ?? [];
    const cards = [
      { label: t('pager.insights.orders'), value: number.format(orders) },
      { label: t('pager.insights.avgWait'), value: duration(data.avgWaitSeconds) },
      { label: t('pager.insights.medianWait'), value: duration(data.medianWaitSeconds) },
      { label: t('pager.insights.longestWait'), value: duration(data.longestWaitSeconds) },
    ];

    return (
      <div className="space-y-5">
        <dl className="grid grid-cols-2 gap-3 @2xl:grid-cols-4">
          {cards.map((card) => (
            <div
              key={card.label}
              className="rounded-xl p-4"
              style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
            >
              <dt className="text-xs font-semibold" style={{ color: 'var(--dash-text-muted)' }}>
                {card.label}
              </dt>
              <dd className="mt-1 text-2xl font-bold" style={{ color: 'var(--dash-text)' }}>
                {card.value}
              </dd>
            </div>
          ))}
        </dl>
        <p className="text-xs" style={{ color: 'var(--dash-text-dim)' }}>
          {t('pager.insights.readyOf', { ready: data.readyOrders ?? 0, count: orders })}
        </p>

        <section
          className="space-y-3 rounded-xl p-4"
          style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
        >
          <h2 className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
            {t('pager.insights.busiestHours')}
          </h2>
          <Bars
            values={byHour}
            labels={byHour.map((_, hour) => hourLabel(hour))}
            tickEvery={6}
            barName={(hour) =>
              t('pager.insights.barHour', { hour: hourLabel(hour), count: byHour[hour] })
            }
            summary={t('pager.insights.busiestSummary', {
              hour: hourLabel(busiest),
              count: byHour[busiest],
            })}
          />
        </section>

        {byDay.length > 1 && (
          <section
            className="space-y-3 rounded-xl p-4"
            style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
          >
            <h2 className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
              {t('pager.insights.ordersByDay')}
            </h2>
            <Bars
              values={byDay.map((day) => day.orders ?? 0)}
              labels={byDay.map((day) => dayLabel.format(new Date(`${day.date}T12:00:00`)))}
              tickEvery={byDay.length > 7 ? 7 : 1}
              barName={(index) =>
                t('pager.insights.barDay', {
                  day: dayLabel.format(new Date(`${byDay[index].date}T12:00:00`)),
                  count: byDay[index].orders ?? 0,
                })
              }
              summary={t('pager.insights.daySummary', { count: orders, days: byDay.length })}
            />
          </section>
        )}

        {data.timezone && (
          <p className="text-xs" style={{ color: 'var(--dash-text-dim)' }}>
            {t('pager.insights.timezone', { zone: data.timezone })}
          </p>
        )}
      </div>
    );
  };

  return (
    <>
      <HelmetContainer title={t('pager.insights.pageTitle')} />
      <div className="@container w-full min-w-0 space-y-5">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div className="min-w-0">
            <Link
              to={PAGER_BOARD_PATH}
              className="mb-2 inline-flex items-center gap-1 text-xs font-semibold"
              style={{ color: 'var(--dash-text-muted)' }}
            >
              <ArrowLeft size={14} aria-hidden="true" className="rtl:rotate-180" />
              {t('pager.insights.backToBoard')}
            </Link>
            <h1 className="text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
              {t('pager.insights.heading')}
            </h1>
            <p className="mt-1 text-xs" style={{ color: 'var(--dash-text-dim)' }}>
              {t('pager.insights.subheading')}
            </p>
          </div>
          <div className="flex gap-1" role="group" aria-label={t('pager.insights.periodLabel')}>
            {PERIODS.map((period) => (
              <Button
                key={period}
                size="sm"
                variant={days === period ? 'default' : 'outline'}
                aria-pressed={days === period}
                onClick={() => setDays(period)}
              >
                {period === 1
                  ? t('pager.insights.today')
                  : t('pager.insights.lastDays', { count: period })}
              </Button>
            ))}
          </div>
        </div>

        {stats.isError ? (
          errorStatus(stats.error) === 403 ? (
            <Alert variant="warning">{t('pager.board.noAccess')}</Alert>
          ) : (
            <div className="space-y-2">
              <Alert variant="destructive">{t('pager.insights.loadFailed')}</Alert>
              <Button size="sm" variant="outline" onClick={() => void stats.refetch()}>
                {t('pager.board.retry')}
              </Button>
            </div>
          )
        ) : stats.data ? (
          content(stats.data)
        ) : (
          <p role="status" className="text-sm" style={{ color: 'var(--dash-text-dim)' }}>
            {t('pager.board.loading')}
          </p>
        )}
      </div>
    </>
  );
};

export default Insights;
