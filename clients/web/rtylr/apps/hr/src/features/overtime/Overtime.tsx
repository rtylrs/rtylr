import React, { useState } from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { useUserStore } from '@voxire/auth';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { Plus, Clock } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import {
  useListOvertimeRequests,
  useApproveOvertime,
} from '../../shared/hooks/queries/useOvertime';
import { Card, SectionHeader, Button, Skeleton, Tabs, Pill } from '../../shared/components/ui';
import ApprovalButtons from '../../shared/components/ApprovalButtons';
import { canHrPermission } from '../../shared/lib/hrPermissions';
import { isHrRequestStatus, normalizeHrRequestStatus } from '../../shared/lib/requestStatus';
import OvertimeRequestForm from './OvertimeRequestForm';

const STATUS_COLORS: Record<string, string> = {
  pending: 'var(--hr-status-submitted)',
  approved: 'var(--hr-status-approved)',
  rejected: 'var(--hr-status-rejected)',
};

const FILTER_STATUSES = ['All', 'Pending', 'Approved', 'Rejected'] as const;

const Overtime = (): React.JSX.Element => {
  const { t } = useTranslation();
  const [statusFilter, setStatusFilter] = useState<(typeof FILTER_STATUSES)[number]>('All');
  const [formOpen, setFormOpen] = useState(false);
  const { data: requests, isLoading } = useListOvertimeRequests();
  const approveOvertime = useApproveOvertime();
  const canApprove = canHrPermission('hr:time_off:approve');

  const filteredRequests = (requests ?? []).filter((r) =>
    statusFilter === 'All' ? true : isHrRequestStatus(r.status, statusFilter)
  );

  const formatMinutes = (mins: number): string => {
    const h = Math.floor(mins / 60);
    const m = mins % 60;
    if (h === 0) return `${m}m`;
    if (m === 0) return `${h}h`;
    return `${h}h ${m}m`;
  };

  return (
    <>
      <HelmetContainer title={pageTitle(t('nav.overtime'))} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <SectionHeader
          title={t('overtime.title')}
          subtitle={t('overtime.subtitle')}
          right={
            <Button variant="primary" icon={<Plus size={13} />} onClick={() => setFormOpen(true)}>
              {t('overtime.newRequest')}
            </Button>
          }
        />

        <Tabs
          tabs={FILTER_STATUSES.map((s) => t(`overtime.statuses.${s.toLowerCase()}`))}
          active={t(`overtime.statuses.${statusFilter.toLowerCase()}`)}
          onChange={(label) => {
            const idx = FILTER_STATUSES.findIndex(
              (s) => t(`overtime.statuses.${s.toLowerCase()}`) === label
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
            icon={Clock}
            title={t('overtime.emptyTitle')}
            description={t('overtime.emptyDescription')}
            action={{ label: t('overtime.newRequest'), onClick: () => setFormOpen(true) }}
          />
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            {filteredRequests.map((req) => (
              <Card key={req.id} style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
                    <span style={{ fontSize: 13, fontWeight: 700, color: 'var(--hr-text)' }}>
                      {t('overtime.employee')} #{req.employeeId}
                    </span>
                    <Pill color={STATUS_COLORS[normalizeHrRequestStatus(req.status)]}>
                      {t(`overtime.statuses.${(req.status ?? '').toLowerCase()}`, {
                        defaultValue: req.status || '—',
                      })}
                    </Pill>
                  </div>
                  <div style={{ fontSize: 11, color: 'var(--hr-text-muted)', marginTop: 3 }}>
                    {req.date} · {formatMinutes(req.plannedExtraMinutes ?? 0)}{' '}
                    {t('overtime.planned')}
                  </div>
                  {req.reason && (
                    <div style={{ fontSize: 10, color: 'var(--hr-text-dim)', marginTop: 2 }}>
                      {req.reason}
                    </div>
                  )}
                </div>
                {canApprove && isHrRequestStatus(req.status, 'pending') && (
                  <ApprovalButtons
                    onApprove={() =>
                      approveOvertime.mutate({
                        overtimeRequestId: req.id,
                        body: {
                          approved: true,
                          multiplierApplied: undefined,
                          rejectionReason: undefined,
                        },
                      })
                    }
                    onReject={() =>
                      approveOvertime.mutate({
                        overtimeRequestId: req.id,
                        body: {
                          approved: false,
                          multiplierApplied: undefined,
                          rejectionReason: undefined,
                        },
                      })
                    }
                  />
                )}
              </Card>
            ))}
          </div>
        )}

        <OvertimeRequestForm open={formOpen} onClose={() => setFormOpen(false)} />
      </div>
    </>
  );
};

export default Overtime;
