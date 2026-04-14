import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Plus, Calendar, Clock } from 'lucide-react';
import { useListLeaveRequests, useApproveLeave } from '../../shared/hooks/queries/useLeave';
import {
  Card,
  SectionHeader,
  Button,
  EmptyState,
  Skeleton,
  Tabs,
  Pill,
} from '../../shared/components/ui';
import ApprovalButtons from '../../shared/components/ApprovalButtons';
import LeaveBalances from './LeaveBalances';

const STATUS_COLORS: Record<string, string> = {
  Pending: '#f59e0b',
  Approved: '#10b981',
  Rejected: '#ef4444',
};

const TAB_KEYS = ['myLeave', 'teamCalendar'] as const;
type TabKey = (typeof TAB_KEYS)[number];

const FILTER_STATUSES = ['All', 'Pending', 'Approved', 'Rejected'] as const;

const LeaveCalendarLazy = React.lazy(() => import('./LeaveCalendar'));

const Leave = (): React.JSX.Element => {
  const { t } = useTranslation();
  const [tab, setTab] = useState<TabKey>('myLeave');
  const [statusFilter, setStatusFilter] = useState<(typeof FILTER_STATUSES)[number]>('All');
  const { data: requests, isLoading } = useListLeaveRequests();
  const approveLeave = useApproveLeave();

  const filteredRequests = (requests ?? []).filter((r) =>
    statusFilter === 'All' ? true : r.status === statusFilter
  );

  const tabLabels = TAB_KEYS.map((key) => t(`leave.tabs.${key}`));

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <SectionHeader
        title={t('leave.title')}
        subtitle={t('leave.subtitle')}
        right={
          <Link to="/leave/new" style={{ textDecoration: 'none' }}>
            <Button variant="primary" icon={<Plus size={13} />}>
              {t('leave.newRequest')}
            </Button>
          </Link>
        }
      />

      <Tabs
        tabs={tabLabels}
        active={t(`leave.tabs.${tab}`)}
        onChange={(label) => {
          const idx = tabLabels.indexOf(label);
          if (idx >= 0) setTab(TAB_KEYS[idx]);
        }}
      />

      {tab === 'myLeave' && (
        <>
          <LeaveBalances />

          <Tabs
            tabs={FILTER_STATUSES.map((s) => t(`leave.statuses.${s.toLowerCase()}`))}
            active={t(`leave.statuses.${statusFilter.toLowerCase()}`)}
            onChange={(label) => {
              const idx = FILTER_STATUSES.findIndex(
                (s) => t(`leave.statuses.${s.toLowerCase()}`) === label
              );
              if (idx >= 0) setStatusFilter(FILTER_STATUSES[idx]);
            }}
          />

          {isLoading ? (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              {[1, 2, 3].map((i) => (
                <Skeleton key={i} height={72} />
              ))}
            </div>
          ) : filteredRequests.length === 0 ? (
            <EmptyState
              icon={<Clock size={20} />}
              title={t('leave.emptyTitle')}
              description={t('leave.emptyDescription')}
              action={
                <Link
                  to="/leave/new"
                  style={{ color: '#8b5cf6', textDecoration: 'none', fontSize: 12 }}
                >
                  {t('leave.newRequest')}
                </Link>
              }
            />
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              {filteredRequests.map((req) => (
                <Card key={req.id} style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div
                      style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}
                    >
                      <span style={{ fontSize: 13, fontWeight: 700, color: 'var(--hr-text)' }}>
                        {req.staffName}
                      </span>
                      <Pill color={STATUS_COLORS[req.status]}>
                        {t(`leave.statuses.${req.status.toLowerCase()}`)}
                      </Pill>
                      <span style={{ fontSize: 10, color: 'var(--hr-text-dim)' }}>
                        {req.leaveTypeName}
                      </span>
                    </div>
                    <div style={{ fontSize: 11, color: 'var(--hr-text-muted)', marginTop: 3 }}>
                      {req.startDate} — {req.endDate} · {req.daysCount} {t('leave.days')}
                    </div>
                    {req.reason && (
                      <div style={{ fontSize: 10, color: 'var(--hr-text-dim)', marginTop: 2 }}>
                        {req.reason}
                      </div>
                    )}
                  </div>
                  {req.status === 'Pending' && (
                    <ApprovalButtons
                      onApprove={() => approveLeave.mutate({ id: req.id, status: 'Approved' })}
                      onReject={() => approveLeave.mutate({ id: req.id, status: 'Rejected' })}
                    />
                  )}
                </Card>
              ))}
            </div>
          )}
        </>
      )}

      {tab === 'teamCalendar' && (
        <React.Suspense fallback={<Skeleton height={400} />}>
          <LeaveCalendarLazy />
        </React.Suspense>
      )}
    </div>
  );
};

export default Leave;
