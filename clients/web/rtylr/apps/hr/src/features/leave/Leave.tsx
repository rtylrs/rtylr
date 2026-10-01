import React, { useState } from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { useUserStore } from '@voxire/auth';
import { Link, useNavigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { Plus, Clock } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { useListLeaveRequests, useApproveLeave } from '../../shared/hooks/queries/useLeave';
import {
  Card,
  SectionHeader,
  Button,
  Skeleton,
  QueryError,
  Tabs,
  Pill,
} from '../../shared/components/ui';
import ApprovalButtons from '../../shared/components/ApprovalButtons';
import LeaveBalances from './LeaveBalances';
import { canHrPermission } from '../../shared/lib/hrPermissions';
import { isHrRequestStatus, normalizeHrRequestStatus } from '../../shared/lib/requestStatus';

const STATUS_COLORS: Record<string, string> = {
  pending: 'var(--hr-status-submitted)',
  approved: 'var(--hr-status-approved)',
  rejected: 'var(--hr-status-rejected)',
};

const TAB_KEYS = ['myLeave', 'teamCalendar'] as const;
type TabKey = (typeof TAB_KEYS)[number];

const FILTER_STATUSES = ['All', 'Pending', 'Approved', 'Rejected'] as const;

const LeaveCalendarLazy = React.lazy(() => import('./LeaveCalendar'));

const Leave = (): React.JSX.Element => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [tab, setTab] = useState<TabKey>('myLeave');
  const [statusFilter, setStatusFilter] = useState<(typeof FILTER_STATUSES)[number]>('All');
  const { data: requests, isLoading, isError, isFetching, refetch } = useListLeaveRequests();
  const approveLeave = useApproveLeave();
  const canApprove = canHrPermission('hr:time_off:approve');

  const filteredRequests = (requests ?? []).filter((r) =>
    statusFilter === 'All' ? true : isHrRequestStatus(r.status, statusFilter)
  );

  const tabLabels = TAB_KEYS.map((key) => t(`leave.tabs.${key}`));

  return (
    <>
      <HelmetContainer title={pageTitle(t('nav.leave'))} />
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

            {isError ? (
              <QueryError onRetry={() => refetch()} isRetrying={isFetching} />
            ) : isLoading ? (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                {[1, 2, 3].map((i) => (
                  <Skeleton key={i} height={72} />
                ))}
              </div>
            ) : filteredRequests.length === 0 ? (
              <EmptyState
                icon={Clock}
                title={t('leave.emptyTitle')}
                description={t('leave.emptyDescription')}
                action={
                  statusFilter !== 'All'
                    ? { label: t('common.clearFilters'), onClick: () => setStatusFilter('All') }
                    : {
                        label: t('leave.newRequest'),
                        onClick: () => navigate({ to: '/leave/new' }),
                      }
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
                          {t('leave.employee')} #{req.employeeId}
                        </span>
                        <Pill color={STATUS_COLORS[normalizeHrRequestStatus(req.status)]}>
                          {t(`leave.statuses.${(req.status ?? '').toLowerCase()}`, {
                            defaultValue: req.status || '—',
                          })}
                        </Pill>
                        <span style={{ fontSize: 10, color: 'var(--hr-text-dim)' }}>
                          #{req.leaveTypeId}
                        </span>
                      </div>
                      <div style={{ fontSize: 11, color: 'var(--hr-text-muted)', marginTop: 3 }}>
                        {req.startDate}, {req.endDate} · {req.totalDays} {t('leave.days')}
                      </div>
                      {req.notes && (
                        <div style={{ fontSize: 10, color: 'var(--hr-text-dim)', marginTop: 2 }}>
                          {req.notes}
                        </div>
                      )}
                    </div>
                    {canApprove && isHrRequestStatus(req.status, 'pending') && (
                      <ApprovalButtons
                        allowNote
                        rejectNoteRequired
                        onApprove={() => {
                          if (req.id) {
                            approveLeave.mutate({
                              requestId: req.id,
                              body: { approved: true },
                            });
                          }
                        }}
                        onReject={(reason) => {
                          if (req.id && reason?.trim()) {
                            approveLeave.mutate({
                              requestId: req.id,
                              body: { approved: false, rejectionReason: reason.trim() },
                            });
                          }
                        }}
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
    </>
  );
};

export default Leave;
