import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { AlertCircle, CalendarDays, Plus, Wallet, X } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { useMyEmployeeId } from '../../shared/hooks/useMyEmployeeId';
import { CardSkeleton } from '../../shared/components/Skeleton';
import { hasQueryFailed } from '../../shared/lib/queryFailureUtils';
import {
  useCancelLeaveRequest,
  useListLeaveTypes,
  useMyLeaveBalances,
  useMyLeaveRequests,
} from './useLeave';
import RequestLeaveDialog from './RequestLeaveDialog';

const cardStyle: React.CSSProperties = {
  background: 'var(--flow-card)',
  border: '1px solid var(--flow-border)',
};
const sectionHeadingClass = 'mb-3 text-body font-semibold uppercase tracking-wider';

const statusTone: Record<string, string> = {
  pending: 'var(--flow-status-pending, var(--flow-text-muted))',
  approved: 'var(--flow-status-approved, var(--accent-fg))',
  rejected: 'var(--flow-status-rejected, var(--flow-danger-fg))',
  cancelled: 'var(--flow-text-muted)',
};

/**
 * Front door onto hr.leave.v1: HR stays the system of record, Flow never
 * stores a leave request. Everything here is scoped to the caller's own
 * hr_employees.id, resolved via useMyEmployeeId - never the hub auth user id.
 */
const LeaveScreen = (): React.JSX.Element => {
  const { t } = useTranslation();
  const { data: employeeId, isLoading: employeeLoading } = useMyEmployeeId();
  const { data: leaveTypes = [] } = useListLeaveTypes();
  const balancesQuery = useMyLeaveBalances(employeeId);
  const requestsQuery = useMyLeaveRequests(employeeId);
  const balances = balancesQuery.data ?? [];
  const requests = requestsQuery.data ?? [];
  const balancesFailed = hasQueryFailed(balancesQuery);
  const requestsFailed = hasQueryFailed(requestsQuery);
  const cancelRequest = useCancelLeaveRequest();
  const [dialogOpen, setDialogOpen] = useState(false);

  const typeName = (id?: string): string => leaveTypes.find((lt) => lt.id === id)?.name ?? id ?? '';

  if (employeeLoading) {
    return <CardSkeleton />;
  }

  return (
    <>
      <HelmetContainer title={`${t('flow.leave.heading')} · Flow`} />
      <div className="mx-auto max-w-2xl p-4 sm:p-6">
        <div className="mb-6 flex items-center justify-between">
          <h1 className="text-lg font-semibold" style={{ color: 'var(--flow-text)' }}>
            {t('flow.leave.heading')}
          </h1>
          <button
            type="button"
            onClick={() => setDialogOpen(true)}
            className="flow-focus inline-flex min-h-[44px] items-center justify-center gap-1.5 rounded-sm px-3 text-body font-semibold md:min-h-0 md:py-2"
            style={{ background: 'var(--accent)', color: 'var(--flow-on-accent)' }}
          >
            <Plus size={14} aria-hidden="true" />
            {t('flow.leave.newRequest')}
          </button>
        </div>

        <section className="mb-4 rounded-lg p-4" style={cardStyle}>
          <h2 className={sectionHeadingClass} style={{ color: 'var(--flow-text-muted)' }}>
            {t('flow.leave.balanceHeading')}
          </h2>
          {balancesFailed ? (
            <EmptyState
              icon={AlertCircle}
              title={t('flow.leave.loadBalancesFailedTitle')}
              description={t('flow.leave.loadBalancesFailedDescription')}
              action={{
                label: t('flow.leave.retryCta'),
                onClick: () => void balancesQuery.refetch(),
              }}
            />
          ) : balances.length === 0 ? (
            <EmptyState
              icon={Wallet}
              title={t('flow.leave.noBalances')}
              action={{
                label: t('flow.leave.newRequest'),
                onClick: () => setDialogOpen(true),
              }}
            />
          ) : (
            <div className="grid gap-3 sm:grid-cols-2">
              {balances.map((balance) => (
                <div
                  key={balance.id}
                  className="rounded-md p-3"
                  style={{ background: 'var(--flow-hover)' }}
                >
                  <div
                    className="mb-1 text-body font-semibold"
                    style={{ color: 'var(--flow-text)' }}
                  >
                    {typeName(balance.leaveTypeId)}
                  </div>
                  <div className="flex items-baseline gap-1.5">
                    <span
                      className="text-xl font-bold tabular-nums"
                      style={{ color: 'var(--accent-fg)' }}
                    >
                      {balance.remainingDays}
                    </span>
                    <span className="text-meta" style={{ color: 'var(--flow-text-muted)' }}>
                      {t('flow.leave.remaining')}
                    </span>
                  </div>
                  <div
                    className="mt-2 flex flex-wrap gap-x-3 gap-y-0.5 text-meta"
                    style={{ color: 'var(--flow-text-muted)' }}
                  >
                    <span>
                      {t('flow.leave.entitled')} {balance.entitledDays}
                    </span>
                    <span>
                      {t('flow.leave.used')} {balance.usedDays}
                    </span>
                    <span>
                      {t('flow.leave.pending')} {balance.pendingDays}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </section>

        <section className="rounded-lg p-4" style={cardStyle}>
          <h2 className={sectionHeadingClass} style={{ color: 'var(--flow-text-muted)' }}>
            {t('flow.leave.requestsHeading')}
          </h2>
          {requestsFailed ? (
            <EmptyState
              icon={AlertCircle}
              title={t('flow.leave.loadRequestsFailedTitle')}
              description={t('flow.leave.loadRequestsFailedDescription')}
              action={{
                label: t('flow.leave.retryCta'),
                onClick: () => void requestsQuery.refetch(),
              }}
            />
          ) : requests.length === 0 ? (
            <EmptyState
              icon={CalendarDays}
              title={t('flow.leave.noRequests')}
              action={{
                label: t('flow.leave.newRequest'),
                onClick: () => setDialogOpen(true),
              }}
            />
          ) : (
            <ul className="space-y-1.5">
              {requests.map((request) => (
                <li
                  key={request.id}
                  className="flex items-center gap-3 rounded-md p-2"
                  style={{ background: 'var(--flow-hover)' }}
                >
                  <span
                    className="min-w-0 flex-1 text-body font-medium"
                    style={{ color: 'var(--flow-text)' }}
                  >
                    {typeName(request.leaveTypeId)}
                  </span>
                  <span className="text-body" style={{ color: 'var(--flow-text-muted)' }}>
                    {request.startDate}
                  </span>
                  <span style={{ color: 'var(--flow-text-muted)' }}>→</span>
                  <span className="text-body" style={{ color: 'var(--flow-text-muted)' }}>
                    {request.endDate}
                  </span>
                  <span
                    className="rounded-sm px-1.5 py-0.5 text-meta font-semibold"
                    style={{ color: statusTone[request.status ?? ''] ?? 'var(--flow-text-muted)' }}
                  >
                    {t(`flow.leave.status.${request.status}`)}
                  </span>
                  {request.status === 'pending' && request.id && (
                    <button
                      type="button"
                      onClick={() => cancelRequest.mutate(request.id!)}
                      aria-label={t('flow.leave.cancelRequestLabel')}
                      className="flow-focus flex h-[44px] w-[44px] shrink-0 items-center justify-center rounded-sm md:h-8 md:w-8"
                      style={{ color: 'var(--flow-text-muted)' }}
                    >
                      <X size={14} aria-hidden="true" />
                    </button>
                  )}
                </li>
              ))}
            </ul>
          )}
        </section>
      </div>

      <RequestLeaveDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        employeeId={employeeId ?? ''}
        leaveTypes={leaveTypes}
      />
    </>
  );
};

export default LeaveScreen;
