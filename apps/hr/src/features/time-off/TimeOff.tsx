import React, { useState } from 'react';
import { useUserStore } from '@voxire/auth';
import { Link } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { toast } from 'sonner';
import { Plus, Calendar } from 'lucide-react';
import { useTimeOff, useApproveTimeOff } from '../../shared/hooks/useHrExtras';
import {
  Card,
  SectionHeader,
  Button,
  EmptyState,
  Skeleton,
  Tabs,
} from '../../shared/components/ui';
import TimeOffBadge from '../../shared/components/TimeOffBadge';
import ApprovalButtons from '../../shared/components/ApprovalButtons';
import { canHrPermission } from '../../shared/lib/hrPermissions';
import { isHrRequestStatus } from '../../shared/lib/requestStatus';

const STATUSES = ['All', 'Pending', 'Approved', 'Rejected'] as const;

const TimeOff: React.FC = () => {
  const { data, isLoading } = useTimeOff();
  const approve = useApproveTimeOff();
  const role = useUserStore((state) => state.user?.role ?? state.role);
  const canApprove = canHrPermission(role, 'hr:time_off:approve');
  const [tab, setTab] = useState<(typeof STATUSES)[number]>('All');

  const filtered = (data ?? []).filter((t) =>
    tab === 'All' ? true : isHrRequestStatus(t.status, tab)
  );

  return (
    <>
      <HelmetContainer title="Time off · rtylr" />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <SectionHeader
          title="Time Off"
          subtitle="Review pending requests and track approved absences"
          right={
            <Link to="/time-off/new" style={{ textDecoration: 'none' }}>
              <Button variant="primary" icon={<Plus size={13} />}>
                New request
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
            icon={<Calendar size={20} />}
            title="No time off requests"
            description="Requests will appear here once they're submitted."
            action={
              <Link
                to="/time-off/new"
                style={{ color: '#8b5cf6', textDecoration: 'none', fontSize: 12 }}
              >
                Submit the first one
              </Link>
            }
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
                        { onSuccess: () => toast.success('Approved') }
                      )
                    }
                    onReject={() =>
                      approve.mutate(
                        { id: req.id, approved: false },
                        { onSuccess: () => toast.success('Rejected') }
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
