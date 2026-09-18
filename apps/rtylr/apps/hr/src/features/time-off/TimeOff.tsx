import React, { useState } from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { useUserStore } from '@voxire/auth';
import { Link, useNavigate } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { Plus, Calendar } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { useTimeOff, useApproveTimeOff } from '../../shared/hooks/useHrExtras';
import { Card, SectionHeader, Button, Skeleton, Tabs } from '../../shared/components/ui';
import TimeOffBadge from '../../shared/components/TimeOffBadge';
import ApprovalButtons from '../../shared/components/ApprovalButtons';
import { canHrPermission } from '../../shared/lib/hrPermissions';
import { isHrRequestStatus } from '../../shared/lib/requestStatus';

const STATUSES = ['All', 'Pending', 'Approved', 'Rejected'] as const;

const TimeOff: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { data, isLoading } = useTimeOff();
  const approve = useApproveTimeOff();
  const role = useUserStore((state) => state.user?.role ?? state.role);
  const canApprove = canHrPermission(role, 'hr:time_off:approve');
  const [tab, setTab] = useState<(typeof STATUSES)[number]>('All');

  const filtered = (data ?? []).filter((req) =>
    tab === 'All' ? true : isHrRequestStatus(req.status, tab)
  );

  return (
    <>
      <HelmetContainer title={pageTitle(t('nav.timeOff'))} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <SectionHeader
          title={t('timeOff.page.title')}
          subtitle={t('timeOff.page.subtitle')}
          right={
            <Link to="/time-off/new" style={{ textDecoration: 'none' }}>
              <Button variant="primary" icon={<Plus size={13} />}>
                {t('timeOff.page.newRequest')}
              </Button>
            </Link>
          }
        />

        <Tabs tabs={[...STATUSES]} active={tab} onChange={(t) => setTab(t as typeof tab)} />

        {isLoading ? (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            {[1, 2, 3].map((i) => (
              <Skeleton key={i} height={72} />
            ))}
          </div>
        ) : filtered.length === 0 ? (
          <EmptyState
            icon={Calendar}
            title={t('timeOff.emptyTitle')}
            description={t('timeOff.emptyDescription')}
            action={{
              label: t('timeOff.submitFirstOne'),
              onClick: () => navigate({ to: '/time-off/new' }),
            }}
          />
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            {filtered.map((req) => (
              <Card key={req.id} style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                    <span style={{ fontSize: 13, fontWeight: 700, color: 'var(--hr-text)' }}>
                      {req.staffName}
                    </span>
                    <TimeOffBadge status={req.status} />
                    <span style={{ fontSize: 10, color: 'var(--hr-text-dim)' }}>• {req.type}</span>
                  </div>
                  <div style={{ fontSize: 11, color: 'var(--hr-text-muted)', marginTop: 3 }}>
                    {req.startDate} to {req.endDate} • {req.reason}
                  </div>
                  {req.note && (
                    <div style={{ fontSize: 10, color: 'var(--hr-text-dim)', marginTop: 2 }}>
                      Note: {req.note}
                    </div>
                  )}
                </div>
                {canApprove && isHrRequestStatus(req.status, 'pending') && (
                  <ApprovalButtons
                    onApprove={() =>
                      approve.mutate(
                        { id: req.id, approved: true },
                        { onSuccess: () => toast.success(t('timeOff.page.approved')) }
                      )
                    }
                    onReject={() =>
                      approve.mutate(
                        { id: req.id, approved: false },
                        { onSuccess: () => toast.success(t('timeOff.page.rejected')) }
                      )
                    }
                  />
                )}
              </Card>
            ))}
          </div>
        )}
      </div>
    </>
  );
};

export default TimeOff;
