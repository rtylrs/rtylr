import React, { useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { ArrowLeft, Mail, Shield, Clock, UserX } from 'lucide-react';
import { Avatar, Button } from '../../../components/ui';
import {
  Breadcrumbs,
  RoleBadge,
  AppBadge,
  SettingSection,
  DangerZone,
  ActivityEntryRow,
  ConfirmDialog,
  EmptyState,
} from '../../../components/hub';
import { useMember, useDeactivateMember } from '../../../hooks/queries/useTeam';
import { useActivity } from '../../../hooks/queries/useActivity';

function getInitials(name?: string | null): string {
  if (!name) return '?';
  return name.split(' ').map((n) => n[0]).join('').toUpperCase().slice(0, 2);
}

const MemberDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data: member, isLoading } = useMember(id);
  const { data: activity = [] } = useActivity(10);
  const deactivate = useDeactivateMember();
  const [confirmOpen, setConfirmOpen] = useState(false);

  if (isLoading) {
    return (
      <div className="max-w-3xl">
        <p className="text-sm" style={{ color: 'var(--hub-text-dim)' }}>Loading member...</p>
      </div>
    );
  }

  if (!member) {
    return (
      <div className="max-w-3xl space-y-4">
        
        <EmptyState
          icon={UserX}
          title="Member not found"
          description="This member may have been deactivated or removed."
          actionLabel="Back to team"
          onAction={() => navigate('/team')}
        />
      </div>
    );
  }

  const memberActivity = activity.filter((a) => a.actor === member.name);

  return (
    <div className="space-y-5 max-w-4xl">

      <div className="flex items-center gap-4">
        <Avatar
          initials={getInitials(member.name)}
          size="lg"
          className="shrink-0"
          style={{ background: 'rgba(69,131,196,0.15)', color: '#4583C4' }}
        />
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-lg font-semibold" style={{ color: 'var(--hub-text)' }}>{member.name}</h1>
            <RoleBadge role={member.role} />
          </div>
          <p className="text-xs mt-0.5" style={{ color: 'var(--hub-text-dim)' }}>{member.email}</p>
        </div>
      </div>

      <div className="grid grid-cols-3 gap-3">
        <div className="rounded-xl p-4" style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}>
          <div className="flex items-center gap-2 text-[11px] font-semibold uppercase tracking-wider mb-2" style={{ color: 'var(--hub-text-dim)' }}>
            <Mail size={11} /> Email
          </div>
          <p className="text-xs font-medium truncate" style={{ color: 'var(--hub-text)' }}>{member.email}</p>
        </div>
        <div className="rounded-xl p-4" style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}>
          <div className="flex items-center gap-2 text-[11px] font-semibold uppercase tracking-wider mb-2" style={{ color: 'var(--hub-text-dim)' }}>
            <Shield size={11} /> Role
          </div>
          <p className="text-xs font-medium capitalize" style={{ color: 'var(--hub-text)' }}>{member.role}</p>
        </div>
        <div className="rounded-xl p-4" style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}>
          <div className="flex items-center gap-2 text-[11px] font-semibold uppercase tracking-wider mb-2" style={{ color: 'var(--hub-text-dim)' }}>
            <Clock size={11} /> Status
          </div>
          <p className="text-xs font-medium capitalize" style={{ color: 'var(--hub-text)' }}>{member.status}</p>
        </div>
      </div>

      <SettingSection title="App access" description="Apps this member can launch.">
        <div className="flex flex-wrap gap-2">
          {member.apps.length === 0 ? (
            <p className="text-xs" style={{ color: 'var(--hub-text-dim)' }}>No apps assigned.</p>
          ) : (
            member.apps.map((a) => <AppBadge key={a} app={a} size="sm" />)
          )}
        </div>
      </SettingSection>

      <SettingSection title="Recent activity" description="Recent actions performed by this member.">
        {memberActivity.length === 0 ? (
          <p className="text-xs text-center py-6" style={{ color: 'var(--hub-text-dim)' }}>
            No recent activity
          </p>
        ) : (
          <div className="-mx-5">
            {memberActivity.map((a) => <ActivityEntryRow key={a.id} entry={a} />)}
          </div>
        )}
      </SettingSection>

      {member.role !== 'owner' && (
        <DangerZone
          actions={[
            {
              label: 'Deactivate member',
              description: 'Immediately revoke all access. Can be undone later.',
              buttonLabel: 'Deactivate',
              onClick: () => setConfirmOpen(true),
            },
          ]}
        />
      )}

      <ConfirmDialog
        open={confirmOpen}
        onOpenChange={setConfirmOpen}
        title={`Deactivate ${member.name}?`}
        description="They will lose access to all apps immediately."
        confirmLabel="Deactivate"
        destructive
        isLoading={deactivate.isPending}
        onConfirm={async () => {
          await deactivate.mutateAsync(member.id);
          navigate('/team');
        }}
      />
    </div>
  );
};

export default MemberDetail;
