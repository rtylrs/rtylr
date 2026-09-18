import React, { useState } from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { useTranslation } from 'react-i18next';
import { useNavigate } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import { FileText, ChevronRight } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { useListTimesheets } from '../../shared/hooks/queries/useTimesheets';
import type { TimesheetPayload } from '@voxire/voxie/hr/time_entry/v1/types';
import { SectionHeader, Card, Pill, Skeleton, QueryError, Tabs } from '../../shared/components/ui';

function formatDuration(minutes: number): string {
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  if (m === 0) return `${h}h`;
  return `${h}h ${m}m`;
}

const STATUS_COLORS: Record<string, string> = {
  draft: 'var(--hr-status-draft)',
  submitted: 'var(--hr-status-submitted)',
  approved: 'var(--hr-status-approved)',
  rejected: 'var(--hr-status-rejected)',
  locked: 'var(--hr-status-locked)',
};

const STATUS_TABS = ['all', 'draft', 'submitted', 'approved', 'rejected'] as const;

const TimesheetsList = (): React.JSX.Element => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [statusFilter, setStatusFilter] = useState<string>('all');

  const { data, isLoading, isError, isFetching, refetch } = useListTimesheets(
    statusFilter === 'all' ? {} : { status: statusFilter }
  );

  const timesheets = data?.timesheets ?? [];

  const formatPeriod = (ts: TimesheetPayload): string => {
    const start = new Date(ts.periodStart + 'T00:00:00').toLocaleDateString(undefined, {
      month: 'short',
      day: 'numeric',
    });
    const end = new Date(ts.periodEnd + 'T00:00:00').toLocaleDateString(undefined, {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
    return `${start} – ${end}`;
  };

  return (
    <>
      <HelmetContainer title={pageTitle(t('nav.timesheets'))} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
        <SectionHeader title={t('timesheets.title')} subtitle={t('timesheets.subtitle')} />

        <Tabs
          tabs={STATUS_TABS.map((s) => t(`timesheets.filters.${s}`))}
          active={t(`timesheets.filters.${statusFilter}`)}
          onChange={(label) => {
            const tab = STATUS_TABS.find((s) => t(`timesheets.filters.${s}`) === label);
            if (tab) setStatusFilter(tab);
          }}
        />

        {isError ? (
          <QueryError onRetry={() => refetch()} isRetrying={isFetching} />
        ) : isLoading ? (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            <Skeleton height={72} />
            <Skeleton height={72} />
            <Skeleton height={72} />
          </div>
        ) : timesheets.length === 0 ? (
          <EmptyState
            icon={FileText}
            title={t('timesheets.emptyTitle')}
            description={t('timesheets.emptyDescription')}
            action={
              statusFilter !== 'all'
                ? { label: t('common.clearFilters'), onClick: () => setStatusFilter('all') }
                : { label: t('common.refresh'), onClick: () => refetch() }
            }
          />
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            {timesheets.map((ts, idx) => (
              <motion.div
                key={ts.id}
                initial={{ opacity: 0, y: 6 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: idx * 0.03 }}
              >
                <Card
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 12,
                    padding: '14px 16px',
                    cursor: 'pointer',
                  }}
                  onClick={() => navigate({ to: '/timesheets/$id', params: { id: String(ts.id) } })}
                >
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div
                      style={{
                        fontSize: 13,
                        fontWeight: 600,
                        color: 'var(--hr-text)',
                      }}
                    >
                      {formatPeriod(ts)}
                    </div>
                    <div
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: 8,
                        marginTop: 4,
                        fontSize: 11,
                        color: 'var(--hr-text-muted)',
                      }}
                    >
                      <span>{formatDuration(ts.totalMinutes ?? 0)}</span>
                      <span>·</span>
                      <span>
                        {ts.entries?.length ?? 0} {t('timesheets.entries')}
                      </span>
                    </div>
                  </div>
                  <Pill color={STATUS_COLORS[ts.status ?? ''] ?? 'var(--hr-status-draft)'}>
                    {t(`timesheets.statuses.${ts.status}`)}
                  </Pill>
                  <ChevronRight size={14} style={{ color: 'var(--hr-text-dim)', flexShrink: 0 }} />
                </Card>
              </motion.div>
            ))}
          </div>
        )}
      </div>
    </>
  );
};

export default TimesheetsList;
