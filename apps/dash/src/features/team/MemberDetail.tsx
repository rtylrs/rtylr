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
  MapPin,
} from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { Avatar, Button, Input, Label } from '../../shared/components/ui';
import {
  RoleBadge,
  AppBadge,
  SettingSection,
  DangerZone,
  ConfirmDialog,
  EmptyState,
} from '../../shared/components/dash';
import {
  useMember,
  useDeactivateMember,
  useUpdateMember,
} from '../../shared/hooks/queries/useTeam';
import {
  useMySubscription,
  centsToDisplay,
  computeSeatProration,
} from '../../shared/hooks/queries/useSubscription';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import {
  usePOSLocations,
  useMemberLocations,
  useAssignPOSLocation,
  useRemovePOSLocation,
  locationFlagsForRole,
} from '../../shared/hooks/queries/usePOSAccess';
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
  { id: 'store_manager', label: 'Store Manager' },
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
  const deactivate = useDeactivateMember();
  const updateMember = useUpdateMember();
  const currentUser = useUserStore((state) => state.user);
  const { data: sub } = useMySubscription();
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [nameDraft, setNameDraft] = useState('');
  const [roleDraft, setRoleDraft] = useState('staff');
  const [appsDraft, setAppsDraft] = useState<string[]>([]);
  const [locationsDraft, setLocationsDraft] = useState<number[]>([]);

  const posSelected = appsDraft.includes('pos');
  const assignLocation = useAssignPOSLocation();
  const removeLocation = useRemovePOSLocation();
  const { data: allLocations = [] } = usePOSLocations(isEditing && posSelected);
  const { data: memberLocations = [] } = useMemberLocations(id, isEditing && posSelected);

  const { data: memberGroups = [] } = useQuery<GroupSummary[]>({
    queryKey: ['user-groups', id],
    queryFn: async () => {
      const res = await unwrap(safe.get(`/hub/team/groups/v1/users/${id}/groups`));
      return (res as { groups?: GroupSummary[] })?.groups ?? [];
    },
    enabled: !!id,
  });

  const qc = useQueryClient();
  const removeFromGroup = useMutation({
    mutationFn: async ({ groupId, userId }: { groupId: number; userId: string }) => {
      await unwrap(safe.delete(`/hub/team/groups/v1/groups/${groupId}/members/${userId}`));
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

  // Seed the location selection from the member's current assignments when they load.
  useEffect(() => {
    setLocationsDraft(memberLocations.map((l) => l.locationId));
  }, [memberLocations]);

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

  const currentRole = currentUser?.role;
  const canEditMember =
    (currentRole === 'owner' || currentRole === 'admin' || currentUser?.isOwner === true) &&
    member.role !== 'owner';

  const toggleApp = (appKey: string): void => {
    setAppsDraft((prev) =>
      prev.includes(appKey) ? prev.filter((key) => key !== appKey) : [...prev, appKey]
    );
  };

  const toggleLocation = (locationId: number): void => {
    setLocationsDraft((prev) =>
      prev.includes(locationId) ? prev.filter((l) => l !== locationId) : [...prev, locationId]
    );
  };

  // Diff the location selection against current assignments: assign newly-added,
  // remove unchecked. Only relevant when POS is one of the member's apps.
  const syncLocations = async (): Promise<void> => {
    if (!posSelected) return;
    const current = memberLocations;
    const currentIds = new Set(current.map((l) => l.locationId));
    const draftIds = new Set(locationsDraft);
    const flags = locationFlagsForRole(roleDraft);

    const toAdd = locationsDraft.filter((locId) => !currentIds.has(locId));
    const toRemove = current.filter((l) => !draftIds.has(l.locationId));
    // Re-assign kept locations too, so a role change updates their manage flags.
    const toUpdate = current.filter((l) => draftIds.has(l.locationId));

    await Promise.allSettled([
      ...toAdd.map((locationId) =>
        assignLocation.mutateAsync({ userId: Number(member.id), locationId, ...flags })
      ),
      ...toUpdate.map((l) =>
        assignLocation.mutateAsync({
          userId: Number(member.id),
          locationId: l.locationId,
          ...flags,
        })
      ),
      ...toRemove.map((l) => removeLocation.mutateAsync(l.id)),
    ]);
    qc.invalidateQueries({ queryKey: ['pos-member-locations', member.id] });
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
      await syncLocations();
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
                      setLocationsDraft(memberLocations.map((l) => l.locationId));
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

            {posSelected && (
              <div className="mt-4 space-y-2">
                <Label>POS locations</Label>
                <p className="text-[11px]" style={{ color: 'var(--dash-text-dim)' }}>
                  {roleDraft === 'store_manager'
                    ? 'Locations this manager runs (staff PINs, voids, refunds, reports)'
                    : 'Locations this member can work at'}
                </p>
                {allLocations.length === 0 ? (
                  <p className="text-[11px]" style={{ color: 'var(--dash-text-dim)' }}>
                    No POS locations yet.
                  </p>
                ) : (
                  <div className="grid grid-cols-2 gap-2">
                    {allLocations.map((loc) => {
                      const checked = locationsDraft.includes(loc.id);
                      return (
                        <button
                          key={loc.id}
                          type="button"
                          onClick={() => toggleLocation(loc.id)}
                          className="flex items-center gap-2.5 rounded-lg px-3 py-2.5 text-left transition-colors"
                          style={{
                            background: checked ? 'rgba(69,131,196,0.08)' : 'var(--dash-bg)',
                            border: `1px solid ${checked ? 'rgba(69,131,196,0.4)' : 'var(--dash-border)'}`,
                          }}
                        >
                          <MapPin
                            size={14}
                            style={{ color: checked ? '#4583C4' : 'var(--dash-text-dim)' }}
                          />
                          <span
                            className="text-xs font-medium truncate"
                            style={{ color: checked ? '#4583C4' : 'var(--dash-text)' }}
                          >
                            {loc.name}
                          </span>
                        </button>
                      );
                    })}
                  </div>
                )}
              </div>
            )}
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
