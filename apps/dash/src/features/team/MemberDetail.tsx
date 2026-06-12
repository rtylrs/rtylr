import React, { useEffect, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useNavigate, useParams } from '@tanstack/react-router';
import {
  Mail,
  Shield,
  Clock,
  UserX,
  X,
  Pencil,
  Save,
  MonitorSpeaker,
  Package,
  UserCheck,
  Users,
  CreditCard,
  Layers,
} from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { Avatar, Button, Input, Label } from '../../shared/components/ui';
import {
  RoleBadge,
  AppBadge,
  SettingSection,
  DangerZone,
  ActivityEntryRow,
  ConfirmDialog,
  EmptyState,
} from '../../shared/components/dash';
import {
  useMember,
  useDeactivateMember,
  useUpdateMember,
} from '../../shared/hooks/queries/useTeam';
import { useActivity } from '../../shared/hooks/queries/useActivity';
import {
  useMySubscription,
  centsToDisplay,
  computeSeatProration,
} from '../../shared/hooks/queries/useSubscription';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import type { GroupSummary } from '../../shared/hooks/queries/useGroups';
import { safe } from '../../shared/lib/api';
import { useUserStore } from '../../shared/stores/userStore';

const ALL_APPS = [
  { key: 'pos', label: 'POS', icon: MonitorSpeaker, color: '#4583C4' },
  { key: 'erp', label: 'ERP', icon: Package, color: '#10b981' },
  { key: 'hr', label: 'HR', icon: UserCheck, color: '#f59e0b' },
  { key: 'crm', label: 'CRM', icon: Users, color: '#8b5cf6' },
  { key: 'finance', label: 'Finance', icon: CreditCard, color: '#ef4444' },
  { key: 'flow', label: 'Flow', icon: Layers, color: '#64C5B9' },
];

const ROLES = [
  { id: 'admin', label: 'Admin' },
  { id: 'operator', label: 'Operator' },
  { id: 'billing', label: 'Billing' },
  { id: 'menu_manager', label: 'Menu Manager' },
  { id: 'staff', label: 'Staff' },
  { id: 'viewer', label: 'Viewer' },
];

function getInitials(name?: string | null): string {
  if (!name) return '?';
  return name
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);
}

const MemberDetail: React.FC = (): React.JSX.Element => {
  const { id } = useParams({ strict: false });
  const navigate = useNavigate();
  const { data: member, isLoading } = useMember(id);
  const { data: activity = [] } = useActivity(10);
  const deactivate = useDeactivateMember();
  const updateMember = useUpdateMember();
  const currentUser = useUserStore((state) => state.user);
  const { data: sub } = useMySubscription();
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [nameDraft, setNameDraft] = useState('');
  const [roleDraft, setRoleDraft] = useState('staff');
  const [appsDraft, setAppsDraft] = useState<string[]>([]);

  const { data: memberGroups = [] } = useQuery<GroupSummary[]>({
    queryKey: ['user-groups', id],
    queryFn: async () => {
      const res = await unwrap(safe.get(`/team/users/${id}/groups`));
      return (res as { groups?: GroupSummary[] })?.groups ?? [];
    },
    enabled: !!id,
  });

  const qc = useQueryClient();
  const removeFromGroup = useMutation({
    mutationFn: async ({ groupId, userId }: { groupId: number; userId: string }) => {
      await unwrap(safe.delete(`/team/groups/${groupId}/members/${userId}`));
    },
    onSuccess: () => {
      toast.success('Removed from group');
      qc.invalidateQueries({ queryKey: ['user-groups', id] });
      qc.invalidateQueries({ queryKey: ['team'] });
    },
    onError: () => toast.error('Failed to remove from group'),
  });

  useEffect(() => {
    if (!member) return;
    setNameDraft(member.fullName || member.name || '');
    setRoleDraft(member.role || 'staff');
    setAppsDraft(member.apps ?? []);
  }, [member]);

  if (isLoading) {
    return (
      <div className="max-w-3xl">
        <p className="text-sm" style={{ color: 'var(--dash-text-dim)' }}>
          Loading member...
        </p>
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
          onAction={() => navigate({ to: '/team' })}
        />
      </div>
    );
  }

  const memberActivity = activity.filter((a) => a.actor === member.name);
  const currentRole = currentUser?.role;
  const canEditMember =
    (currentRole === 'owner' || currentRole === 'admin' || currentUser?.isOwner === true) &&
    member.role !== 'owner';

  const toggleApp = (appKey: string): void => {
    setAppsDraft((prev) =>
      prev.includes(appKey) ? prev.filter((key) => key !== appKey) : [...prev, appKey]
    );
  };

  const saveMember = async (): Promise<void> => {
    if (!nameDraft.trim()) {
      toast.error('Name is required');
      return;
    }
    try {
      await updateMember.mutateAsync({
        id: member.id,
        patch: {
          fullName: nameDraft.trim(),
          name: nameDraft.trim(),
          role: roleDraft,
          apps: appsDraft,
        },
      });
      setIsEditing(false);
    } catch {
      // onError in hook handles toast
    }
  };

  return (
    <>
      <HelmetContainer title="Member detail · rtylr" />
      <div className="space-y-5 max-w-4xl">
        <div className="flex items-center gap-4">
          <Avatar
            initials={getInitials(member.name)}
            size="lg"
            className="shrink-0"
            style={{ background: 'rgba(69,131,196,0.15)', color: '#4583C4' }}
          />
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <h1 className="text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
                {member.name}
              </h1>
              <RoleBadge role={member.role} />
            </div>
            <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
              {member.email}
            </p>
          </div>
          {canEditMember && (
            <div className="ml-auto flex items-center gap-2">
              {isEditing ? (
                <>
                  <Button
                    variant="outline"
                    onClick={() => {
                      setIsEditing(false);
                      setNameDraft(member.fullName || member.name || '');
                      setRoleDraft(member.role || 'staff');
                      setAppsDraft(member.apps ?? []);
                    }}
                  >
                    <X size={14} />
                    Cancel
                  </Button>
                  <Button onClick={saveMember} disabled={updateMember.isPending}>
                    <Save size={14} />
                    Save
                  </Button>
                </>
              ) : (
                <Button variant="outline" onClick={() => setIsEditing(true)}>
                  <Pencil size={14} />
                  Edit
                </Button>
              )}
            </div>
          )}
        </div>

        {isEditing && (
          <SettingSection
            title="Member details"
            description="Edit this member's profile and access."
          >
            <div className="grid gap-4 md:grid-cols-2">
              <div className="space-y-1.5">
                <Label htmlFor="member-name">Name</Label>
                <Input
                  id="member-name"
                  value={nameDraft}
                  onChange={(event) => setNameDraft(event.target.value)}
                />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="member-role">Role</Label>
                <select
                  id="member-role"
                  value={roleDraft}
                  onChange={(event) => setRoleDraft(event.target.value)}
                  className="w-full rounded-lg px-3 py-2 text-sm"
                  style={{
                    background: 'var(--dash-bg)',
                    border: '1px solid var(--dash-border)',
                    color: 'var(--dash-text)',
                  }}
                >
                  {ROLES.map((role) => (
                    <option key={role.id} value={role.id}>
                      {role.label}
                    </option>
                  ))}
                </select>
              </div>
            </div>
            <div className="mt-4 grid grid-cols-2 gap-2 md:grid-cols-3">
              {ALL_APPS.map((app) => {
                const Icon = app.icon;
                const selected = appsDraft.includes(app.key);
                return (
                  <button
                    key={app.key}
                    type="button"
                    onClick={() => toggleApp(app.key)}
                    className="flex items-center gap-2 rounded-lg px-3 py-2 text-left transition-colors"
                    style={{
                      background: selected ? `${app.color}10` : 'var(--dash-bg)',
                      border: `1px solid ${selected ? `${app.color}45` : 'var(--dash-border)'}`,
                      color: selected ? app.color : 'var(--dash-text-muted)',
                    }}
                  >
                    <Icon size={15} style={{ color: app.color }} />
                    <span className="text-xs font-semibold">{app.label}</span>
                  </button>
                );
              })}
            </div>
          </SettingSection>
        )}

        <div className="grid grid-cols-3 gap-3">
          <div
            className="rounded-xl p-4"
            style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
          >
            <div
              className="flex items-center gap-2 text-[11px] font-semibold uppercase tracking-wider mb-2"
              style={{ color: 'var(--dash-text-dim)' }}
            >
              <Mail size={11} /> Email
            </div>
            <p className="text-xs font-medium truncate" style={{ color: 'var(--dash-text)' }}>
              {member.email}
            </p>
          </div>
          <div
            className="rounded-xl p-4"
            style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
          >
            <div
              className="flex items-center gap-2 text-[11px] font-semibold uppercase tracking-wider mb-2"
              style={{ color: 'var(--dash-text-dim)' }}
            >
              <Shield size={11} /> Role
            </div>
            <p className="text-xs font-medium capitalize" style={{ color: 'var(--dash-text)' }}>
              {member.role}
            </p>
          </div>
          <div
            className="rounded-xl p-4"
            style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
          >
            <div
              className="flex items-center gap-2 text-[11px] font-semibold uppercase tracking-wider mb-2"
              style={{ color: 'var(--dash-text-dim)' }}
            >
              <Clock size={11} /> Status
            </div>
            <p className="text-xs font-medium capitalize" style={{ color: 'var(--dash-text)' }}>
              {member.status}
            </p>
          </div>
        </div>

        <SettingSection title="App access" description="Apps this member can launch.">
          <div className="flex flex-wrap gap-2">
            {(member.apps ?? []).length === 0 ? (
              <p className="text-xs" style={{ color: 'var(--dash-text-dim)' }}>
                No apps assigned.
              </p>
            ) : (
              (member.apps ?? []).map((a: string) => <AppBadge key={a} app={a} size="sm" />)
            )}
          </div>
        </SettingSection>

        <SettingSection
          title="Recent activity"
          description="Recent actions performed by this member."
        >
          {memberActivity.length === 0 ? (
            <p className="text-xs text-center py-6" style={{ color: 'var(--dash-text-dim)' }}>
              No recent activity
            </p>
          ) : (
            <div className="-mx-5">
              {memberActivity.map((a, i) => (
                <ActivityEntryRow key={a.id} entry={a} index={i} />
              ))}
            </div>
          )}
        </SettingSection>

        <div
          className="rounded-xl"
          style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
        >
          <div className="px-5 py-4" style={{ borderBottom: '1px solid var(--dash-border)' }}>
            <div className="flex items-center justify-between">
              <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
                Groups
              </p>
              <button
                onClick={() => navigate({ to: '/team/groups' })}
                className="text-xs"
                style={{ color: '#4583C4' }}
              >
                Manage groups
              </button>
            </div>
            <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
              Groups this member belongs to.
            </p>
          </div>
          <div className="px-5 py-4">
            {memberGroups.length === 0 ? (
              <p className="text-xs" style={{ color: 'var(--dash-text-dim)' }}>
                Not in any group.
              </p>
            ) : (
              <div className="flex flex-wrap gap-2">
                {memberGroups.map((g) => (
                  <div
                    key={g.id}
                    className="flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium"
                    style={{
                      background: `${g.color}18`,
                      border: `1px solid ${g.color}30`,
                      color: g.color,
                    }}
                  >
                    <span>{g.name}</span>
                    <button
                      aria-label={`Remove from ${g.name}`}
                      onClick={() => removeFromGroup.mutate({ groupId: g.id, userId: member.id })}
                      className="hover:opacity-70 transition-opacity"
                    >
                      <X size={11} />
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>

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
          description={
            <span className="flex flex-col gap-1.5">
              <span>They will lose access to all apps immediately.</span>
              {sub &&
                sub.pricePerSeatCents > 0 &&
                (() => {
                  const { daysRemaining, daysInMonth, proratedCents, nextMonthName } =
                    computeSeatProration(sub.pricePerSeatCents);
                  return (
                    <span style={{ color: 'var(--dash-text-muted)', fontSize: '12px' }}>
                      Saves{' '}
                      <span className="font-semibold" style={{ color: 'var(--dash-text)' }}>
                        {centsToDisplay(proratedCents)}
                      </span>{' '}
                      this month ({daysRemaining}/{daysInMonth} days), then{' '}
                      <span className="font-semibold" style={{ color: 'var(--dash-text)' }}>
                        {centsToDisplay(sub.pricePerSeatCents)}/mo
                      </span>{' '}
                      from {nextMonthName} 1st.
                    </span>
                  );
                })()}
            </span>
          }
          confirmLabel="Deactivate"
          destructive
          isLoading={deactivate.isPending}
          onConfirm={async () => {
            try {
              await deactivate.mutateAsync(member.id);
              navigate({ to: '/team' });
            } catch {
              // onError in hook handles toast
            }
          }}
        />
      </div>
    </>
  );
};

export default MemberDetail;
