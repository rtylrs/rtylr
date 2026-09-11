import React, { useState } from 'react';
import { pageTitle } from '../../../shared/lib/pageTitle';
import { useUserStore } from '@voxire/auth';
import { useTranslation } from 'react-i18next';
import { motion } from 'motion/react';
import { HelmetContainer } from '@voxire/components';
import { toast } from 'sonner';
import { CheckCircle, XCircle, Users } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import {
  useListTimesheets,
  useApproveTimesheet,
  useRejectTimesheet,
} from '../../../shared/hooks/queries/useTimesheets';
import type { TimesheetPayload } from '@voxire/voxie/hr/time_entry/v1/types';
import {
  SectionHeader,
  Card,
  Button,
  Pill,
  Skeleton,
  Modal,
  Textarea,
  Label,
  ACCENT,
} from '../../../shared/components/ui';
import { canHrPermission } from '../../../shared/lib/hrPermissions';

function formatDuration(minutes: number): string {
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  if (m === 0) return `${h}h`;
  return `${h}h ${m}m`;
}

const TimesheetApprovals = (): React.JSX.Element => {
  const { t } = useTranslation();
  const { data, isLoading, refetch } = useListTimesheets({ status: 'submitted' });
  const approveMutation = useApproveTimesheet();
  const rejectMutation = useRejectTimesheet();
  const role = useUserStore((state) => state.user?.role ?? state.role);
  const canApprove = canHrPermission(role, 'hr:time_off:approve');

  const [rejectTarget, setRejectTarget] = useState<TimesheetPayload | null>(null);
  const [rejectionReason, setRejectionReason] = useState('');

  const handleApprove = (timesheetId: number): void => {
    approveMutation.mutate(
      { timesheetId, approved: true },
      {
        onSuccess: () => toast.success(t('timesheets.approved')),
        onError: () => toast.error(t('errors.approveTimesheet')),
      }
    );
  };

  const handleReject = (): void => {
    if (!rejectTarget) return;
    rejectMutation.mutate(
      {
        timesheetId: Number(rejectTarget.id ?? 0),
        rejectionReason: rejectionReason || undefined,
      },
      {
        onSuccess: () => {
          setRejectTarget(null);
          setRejectionReason('');
        },
      }
    );
  };

  const timesheets = data?.timesheets ?? [];

  return (
    <>
      <HelmetContainer title={pageTitle(t('common.pageTitles.timesheetApprovals'))} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
        <SectionHeader
          title={t('timesheets.approvals.title')}
          subtitle={t('timesheets.approvals.subtitle', { count: timesheets.length })}
        />

        {isLoading ? (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            <Skeleton height={80} />
            <Skeleton height={80} />
          </div>
        ) : timesheets.length === 0 ? (
          <EmptyState
            icon={CheckCircle}
            title={t('timesheets.approvals.emptyTitle')}
            description={t('timesheets.approvals.emptyDescription')}
            action={{ label: t('common.refresh'), onClick: () => refetch() }}
          />
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            {timesheets.map((ts, idx) => {
              const period = `${new Date(ts.periodStart + 'T00:00:00').toLocaleDateString(
                undefined,
                {
                  month: 'short',
                  day: 'numeric',
                }
              )} – ${new Date(ts.periodEnd + 'T00:00:00').toLocaleDateString(undefined, {
                month: 'short',
                day: 'numeric',
              })}`;

              return (
                <motion.div
                  key={ts.id}
                  initial={{ opacity: 0, y: 6 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: idx * 0.04 }}
                >
                  <Card
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 12,
                      padding: '14px 16px',
                      flexWrap: 'wrap',
                    }}
                  >
                    <div
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        width: 32,
                        height: 32,
                        borderRadius: 8,
                        background: `${ACCENT}15`,
                        color: ACCENT,
                        flexShrink: 0,
                      }}
                    >
                      <Users size={15} />
                    </div>
                    <div style={{ flex: 1, minWidth: 140 }}>
                      <div style={{ fontSize: 13, fontWeight: 600, color: 'var(--hr-text)' }}>
                        {t('timesheets.approvals.employeeId', { id: ts.employeeId })}
                      </div>
                      <div
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          gap: 8,
                          fontSize: 11,
                          color: 'var(--hr-text-muted)',
                          marginTop: 3,
                        }}
                      >
                        <span>{period}</span>
                        <span>·</span>
                        <span style={{ fontWeight: 600, color: ACCENT }}>
                          {formatDuration(ts.totalMinutes ?? 0)}
                        </span>
                        <span>·</span>
                        <span>
                          {ts.entries?.length ?? 0} {t('timesheets.entries')}
                        </span>
                      </div>
                    </div>
                    <Pill color="var(--hr-status-submitted)">
                      {t('timesheets.statuses.submitted')}
                    </Pill>
                    {canApprove && (
                      <div style={{ display: 'flex', gap: 6 }}>
                        <Button
                          variant="primary"
                          style={{ padding: '6px 12px', fontSize: 11 }}
                          onClick={() => handleApprove(Number(ts.id ?? 0))}
                          disabled={approveMutation.isPending}
                          icon={<CheckCircle size={13} />}
                        >
                          {t('timesheets.approvals.approve')}
                        </Button>
                        <Button
                          variant="danger"
                          style={{ padding: '6px 12px', fontSize: 11 }}
                          onClick={() => setRejectTarget(ts)}
                          disabled={rejectMutation.isPending}
                          icon={<XCircle size={13} />}
                        >
                          {t('timesheets.approvals.reject')}
                        </Button>
                      </div>
                    )}
                  </Card>
                </motion.div>
              );
            })}
          </div>
        )}

        {/* Rejection reason modal */}
        <Modal
          open={!!rejectTarget}
          onClose={() => {
            setRejectTarget(null);
            setRejectionReason('');
          }}
          title={t('timesheets.approvals.rejectTitle')}
          width={400}
        >
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
            <div>
              <Label>{t('timesheets.approvals.reasonLabel')}</Label>
              <Textarea
                value={rejectionReason}
                onChange={(e) => setRejectionReason(e.target.value)}
                placeholder={t('timesheets.approvals.reasonPlaceholder')}
                rows={3}
              />
            </div>
            <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
              <Button
                variant="ghost"
                onClick={() => {
                  setRejectTarget(null);
                  setRejectionReason('');
                }}
              >
                {t('common.cancel')}
              </Button>
              <Button variant="danger" onClick={handleReject} disabled={rejectMutation.isPending}>
                {t('timesheets.approvals.confirmReject')}
              </Button>
            </div>
          </div>
        </Modal>
      </div>
    </>
  );
};

export default TimesheetApprovals;
