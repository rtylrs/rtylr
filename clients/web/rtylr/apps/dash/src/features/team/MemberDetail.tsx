import { PermissionProfilePicker } from './PermissionProfilePicker';
import React, { useEffect, useRef, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
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
  UsersRound,
  BarChart3,
  Utensils,
  BriefcaseBusiness,
} from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import { EmptyState } from '@voxire/ui';
import { Avatar, Button, Input, Label } from '../../shared/components/ui';
import {
  RoleBadge,
  AppBadge,
  SettingSection,
  DangerZone,
  ConfirmDialog,
} from '../../shared/components/dash';
import {
  useMember,
  useDeactivateMember,
  useUpdateMember,
} from '../../shared/hooks/queries/useTeam';
import { useMySubscription } from '../../shared/hooks/queries/useSubscription';
import { SeatSavingsNote } from '../../shared/components/SeatSavingsNote';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import {
  usePOSLocations,
  useMemberLocations,
  useAssignPOSLocation,
  useRemovePOSLocation,
  type MemberLocation,
} from '../../shared/hooks/queries/usePOSAccess';
import { toast } from 'sonner';
import type { GroupSummary } from '../../shared/hooks/queries/useGroups';
import { groupsClient, permissionClient } from '../../shared/lib/clients';
import { permissionProfileSchema } from '../../shared/schemas/permissions';
import { useUserStore } from '../../shared/stores/userStore';
import { APP_NAMES } from '../../shared/config/appNames';
import {
  buildPermissionProfileAssignments,
  getAssignedProfileIdsByApp,
} from './authorizationDraft';

const ALL_APPS = [
  { key: 'pos', labelKey: APP_NAMES.pos.short, icon: MonitorSpeaker, color: 'var(--app-pos)' },
  { key: 'erp', labelKey: APP_NAMES.erp.short, icon: Package, color: 'var(--app-erp)' },
  { key: 'hr', labelKey: APP_NAMES.hr.short, icon: UserCheck, color: 'var(--app-hr)' },
  { key: 'crm', labelKey: APP_NAMES.crm.short, icon: Users, color: 'var(--app-crm)' },
  {
    key: 'finance',
    labelKey: APP_NAMES.finance.short,
    icon: CreditCard,
    color: 'var(--app-finance)',
  },
  { key: 'flow', labelKey: APP_NAMES.flow.short, icon: Layers, color: 'var(--app-flow)' },
  {
    key: 'insights',
    labelKey: APP_NAMES.insights.short,
    icon: BarChart3,
    color: 'var(--app-insights)',
  },
  { key: 'menu', labelKey: APP_NAMES.menu.short, icon: Utensils, color: 'var(--app-menu)' },
  {
    key: 'recruit',
    labelKey: APP_NAMES.recruit.short,
    icon: BriefcaseBusiness,
    color: 'var(--app-recruit)',
  },
];

const MEMBER_TYPES = ['admin', 'member'] as const;

// Stable identity: useMemberLocations returns undefined while disabled, and a
// fresh fallback array on every render would retrigger the seeding effect.
const EMPTY_MEMBER_LOCATIONS: MemberLocation[] = [];

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
  const { t } = useTranslation();
  const { id } = useParams({ strict: false });
  const navigate = useNavigate();
  const { data: member, isLoading, refetch: refetchMember } = useMember(id);
  const deactivate = useDeactivateMember();
  const updateMember = useUpdateMember();
  const currentUser = useUserStore((state) => state.user);
  const { data: sub } = useMySubscription();
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [nameDraft, setNameDraft] = useState('');
  const [memberTypeDraft, setMemberTypeDraft] = useState<'member' | 'admin'>('member');
  const [licensedAppsDraft, setLicensedAppsDraft] = useState<string[]>([]);
  const [profileIdsDraft, setProfileIdsDraft] = useState<Record<string, string[]>>({});
  const [locationsDraft, setLocationsDraft] = useState<number[]>([]);

  const posSelected = licensedAppsDraft.includes('pos');
  const assignLocation = useAssignPOSLocation();
  const removeLocation = useRemovePOSLocation();
  const { data: allLocations = [] } = usePOSLocations(isEditing && posSelected);
  const memberLocationsQuery = useMemberLocations(id, isEditing && posSelected);
  const memberLocations = memberLocationsQuery.data ?? EMPTY_MEMBER_LOCATIONS;
  const seededLocationsRef = useRef(false);

  const { data: memberGroups = [] } = useQuery<GroupSummary[]>({
    queryKey: ['user-groups', id],
    queryFn: async () => {
      const res = await groupsClient.getUserGroups({ userId: String(id) });
      return (res as { groups?: GroupSummary[] })?.groups ?? [];
    },
    enabled: !!id,
  });

  const { data: customRoles = [] } = useQuery({
    queryKey: ['custom-roles', 'assignable'],
    queryFn: async () => {
      const res = await permissionClient.listRoles({ pageSize: 100 });
      return z.array(permissionProfileSchema).parse(res.roles ?? []);
    },
    enabled: isEditing,
  });

  const qc = useQueryClient();
  const removeFromGroup = useMutation({
    mutationFn: async ({ groupId, userId }: { groupId: number; userId: string }) => {
      await groupsClient.removeGroupMember({ id: String(groupId), userId: String(userId) });
    },
    onSuccess: () => {
      toast.success(t('team.memberDetail.removedFromGroup'));
      qc.invalidateQueries({ queryKey: ['user-groups', id] });
      qc.invalidateQueries({ queryKey: ['team'] });
    },
    onError: () => toast.error(t('team.memberDetail.removeFromGroupFailed')),
  });

  useEffect(() => {
    if (!member) return;
    setNameDraft(member.fullName || member.name || '');
    setMemberTypeDraft(member.memberType === 'admin' ? 'admin' : 'member');
    setLicensedAppsDraft(member.licensedAppIds);
    setProfileIdsDraft(getAssignedProfileIdsByApp(member.permissionProfileAssignments, 'explicit'));
  }, [member]);

  // Seed the location selection from query data once per edit session, when it
  // first arrives. Re-seeding on later refetches would discard in-progress edits.
  useEffect(() => {
    if (!isEditing) {
      seededLocationsRef.current = false;
      return;
    }
    if (seededLocationsRef.current || memberLocationsQuery.data === undefined) return;
    seededLocationsRef.current = true;
    setLocationsDraft(memberLocationsQuery.data.map((l) => l.locationId));
  }, [isEditing, memberLocationsQuery.data]);

  if (isLoading) {
    return (
      <div className="max-w-3xl">
        <p className="text-sm" style={{ color: 'var(--dash-text-dim)' }}>
          {t('team.memberDetail.loading')}
        </p>
      </div>
    );
  }

  if (!member) {
    return (
      <div className="max-w-3xl space-y-4">
        <EmptyState
          icon={UserX}
          title={t('team.memberDetail.notFoundTitle')}
          description={t('team.memberDetail.notFoundDescription')}
          action={{
            label: t('team.memberDetail.backToTeam'),
            onClick: () => navigate({ to: '/team' }),
          }}
        />
      </div>
    );
  }

  const canEditMember =
    currentUser?.effectiveAccess?.dashPermissions.includes('hub:team:edit') === true &&
    !member.isOwner;
  const visibleLicensedApps =
    member.effectiveAccess?.apps.filter((app) => app.licensed).map((app) => app.appId) ??
    member.licensedAppIds;

  const toggleApp = (appKey: string): void => {
    setLicensedAppsDraft((prev) =>
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
    const toAdd = locationsDraft.filter((locId) => !currentIds.has(locId));
    const toRemove = current.filter((l) => !draftIds.has(l.locationId));
    // Kept locations are re-sent so the assignment survives the save.
    const toUpdate = current.filter((l) => draftIds.has(l.locationId));

    await Promise.allSettled([
      // A newly assigned location starts with no elevated capabilities. There
      // is no prior assignment to carry over, and least privilege is the safe
      // default; granting them is a separate, deliberate action.
      ...toAdd.map((locationId) =>
        assignLocation.mutateAsync({
          userId: Number(member.id),
          locationId,
          canManagePosUsers: false,
          canManageTerminals: false,
          canViewReports: false,
        })
      ),
      // Carry the existing flags across. These were hardcoded to false, so
      // saving any member with POS locations cleared every capability on every
      // location they held, whatever their role, and the save gave no sign of
      // it. The old comment claimed a role change updated these flags; nothing
      // read the role, so it only ever revoked.
      ...toUpdate.map((l) =>
        assignLocation.mutateAsync({
          userId: Number(member.id),
          locationId: l.locationId,
          canManagePosUsers: l.canManagePosUsers,
          canManageTerminals: l.canManageTerminals,
          canViewReports: l.canViewReports,
        })
      ),
      ...toRemove.map((l) => removeLocation.mutateAsync(l.id)),
    ]);
    qc.invalidateQueries({ queryKey: ['pos-member-locations', member.id] });
  };

  const saveMember = async (): Promise<void> => {
    if (!nameDraft.trim()) {
      toast.error(t('team.memberDetail.nameRequired'));
      return;
    }
    try {
      await updateMember.mutateAsync({
        id: member.id,
        patch: {
          fullName: nameDraft.trim(),
          name: nameDraft.trim(),
          memberType: memberTypeDraft,
          licensedAppIds: memberTypeDraft === 'admin' ? [] : licensedAppsDraft,
          permissionProfileAssignments:
            memberTypeDraft === 'admin'
              ? []
              : buildPermissionProfileAssignments(
                  licensedAppsDraft,
                  profileIdsDraft,
                  'explicit',
                  member.permissionProfileAssignments
                ),
          groupIds: memberGroups.map((group) => group.id),
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
      <HelmetContainer title={`${t('team.memberDetail.pageTitle')} · rtylr`} />
      <div className="space-y-5 max-w-4xl">
        <div className="flex items-center gap-4">
          <Avatar
            initials={getInitials(member.name)}
            size="lg"
            className="shrink-0"
            style={{ background: 'var(--accent-muted)', color: 'var(--accent-on-muted)' }}
          />
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <h1 className="text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
                {member.name}
              </h1>
              <RoleBadge role={member.isOwner ? 'owner' : member.memberType} />
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
                      setMemberTypeDraft(member.memberType === 'admin' ? 'admin' : 'member');
                      setLicensedAppsDraft(member.licensedAppIds);
                      setProfileIdsDraft(
                        Object.fromEntries(
                          member.permissionProfileAssignments.map((assignment) => [
                            assignment.appId,
                            assignment.permissionProfileIds,
                          ])
                        )
                      );
                      setLocationsDraft(memberLocations.map((l) => l.locationId));
                    }}
                  >
                    <X size={14} />
                    {t('common.cancel')}
                  </Button>
                  <Button onClick={saveMember} disabled={updateMember.isPending}>
                    <Save size={14} />
                    {t('common.save')}
                  </Button>
                </>
              ) : (
                <Button variant="outline" onClick={() => setIsEditing(true)}>
                  <Pencil size={14} />
                  {t('common.edit')}
                </Button>
              )}
            </div>
          )}
        </div>

        {isEditing && (
          <SettingSection
            title={t('team.memberDetail.detailsTitle')}
            description={t('team.memberDetail.detailsDescription')}
          >
            <div className="grid gap-4 md:grid-cols-2">
              <div className="space-y-1.5">
                <Label htmlFor="member-name">{t('team.memberDetail.nameLabel')}</Label>
                <Input
                  id="member-name"
                  value={nameDraft}
                  onChange={(event) => setNameDraft(event.target.value)}
                />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="member-type">{t('team.memberDetail.memberTypeLabel')}</Label>
                <select
                  id="member-type"
                  value={memberTypeDraft}
                  onChange={(event) => setMemberTypeDraft(event.target.value as 'member' | 'admin')}
                  className="w-full rounded-lg px-3 py-2 text-sm"
                  style={{
                    background: 'var(--dash-bg)',
                    border: '1px solid var(--dash-border)',
                    color: 'var(--dash-text)',
                  }}
                >
                  {MEMBER_TYPES.map((type) => (
                    <option key={type} value={type}>
                      {t(`common.roleLabels.${type}`)}
                    </option>
                  ))}
                </select>
              </div>
            </div>
            <div className="mt-4 grid grid-cols-2 gap-2 md:grid-cols-3">
              {ALL_APPS.map((app) => {
                const Icon = app.icon;
                const selected = licensedAppsDraft.includes(app.key);
                return (
                  <button
                    key={app.key}
                    type="button"
                    onClick={() => memberTypeDraft === 'member' && toggleApp(app.key)}
                    disabled={memberTypeDraft === 'admin'}
                    className="flex items-center gap-2 rounded-lg px-3 py-2 text-left transition-colors"
                    style={{
                      background: selected ? `${app.color}10` : 'var(--dash-bg)',
                      border: `1px solid ${selected ? `${app.color}45` : 'var(--dash-border)'}`,
                      color: selected ? app.color : 'var(--dash-text-muted)',
                    }}
                  >
                    <Icon size={15} style={{ color: app.color }} />
                    <span className="text-xs font-semibold">{t(app.labelKey)}</span>
                  </button>
                );
              })}
            </div>

            {memberTypeDraft === 'admin' ? (
              <p className="mt-4 text-xs" style={{ color: 'var(--dash-text-muted)' }}>
                {t('team.memberDetail.adminFullAccess')}
              </p>
            ) : (
              <div className="mt-4 space-y-3">
                <Label>{t('team.memberDetail.permissionProfiles')}</Label>
                {licensedAppsDraft.map((appKey) => {
                  const options = customRoles.filter((profile) => profile.appId === appKey);
                  if (options.length === 0) return null;
                  return (
                    <PermissionProfilePicker
                      key={appKey}
                      label={appKey.toUpperCase()}
                      profiles={options}
                      selectedIds={profileIdsDraft[appKey] ?? []}
                      onChange={(ids) =>
                        setProfileIdsDraft((current) => ({ ...current, [appKey]: ids }))
                      }
                    />
                  );
                })}
              </div>
            )}

            {posSelected && (
              <div className="mt-4 space-y-2">
                <Label>{t('team.memberDetail.posLocationsLabel')}</Label>
                <p className="text-[11px]" style={{ color: 'var(--dash-text-dim)' }}>
                  {t('team.memberDetail.posLocationsMemberHint')}
                </p>
                {allLocations.length === 0 ? (
                  <EmptyState
                    icon={MapPin}
                    title={t('team.memberDetail.noPosLocationsTitle')}
                    description={t('team.memberDetail.noPosLocationsDescription')}
                    action={{
                      label: t('team.memberDetail.removePosAccess'),
                      onClick: () => toggleApp('pos'),
                    }}
                    size="compact"
                  />
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
              <Shield size={11} /> {t('team.memberDetail.memberTypeLabel')}
            </div>
            <p className="text-xs font-medium capitalize" style={{ color: 'var(--dash-text)' }}>
              {member.isOwner ? 'owner' : member.memberType}
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

        <SettingSection
          title={t('team.memberDetail.appAccessTitle')}
          description={t('team.memberDetail.appAccessDescription')}
        >
          <div className="flex flex-wrap gap-2">
            {visibleLicensedApps.length === 0 ? (
              <EmptyState
                icon={Package}
                title={t('team.memberDetail.noAppsAssignedTitle')}
                description={t('team.memberDetail.noAppsAssignedDescription')}
                action={
                  canEditMember
                    ? { label: t('team.memberDetail.assignApp'), onClick: () => setIsEditing(true) }
                    : { label: t('common.refresh'), onClick: () => void refetchMember() }
                }
                size="compact"
              />
            ) : (
              visibleLicensedApps.map((appId) => <AppBadge key={appId} app={appId} size="sm" />)
            )}
          </div>
        </SettingSection>

        <SettingSection
          title={t('team.memberDetail.effectiveAccessTitle')}
          description={t('team.memberDetail.effectiveAccessDescription')}
        >
          <p className="mb-3 text-[11px]" style={{ color: 'var(--dash-text-dim)' }}>
            {t('team.memberDetail.ownershipNote')}
          </p>
          {(member.effectiveAccess?.apps ?? []).filter((app) => app.licensed && app.hasAccess)
            .length === 0 ? (
            <div
              className="rounded-lg p-4"
              style={{ background: 'var(--dash-bg)', border: '1px solid var(--dash-border)' }}
            >
              <p className="text-xs font-semibold" style={{ color: 'var(--dash-text)' }}>
                {t('team.memberDetail.noEffectiveAccessTitle')}
              </p>
              <p className="mt-1 text-[11px]" style={{ color: 'var(--dash-text-dim)' }}>
                {t('team.memberDetail.noEffectiveAccessDescription')}
              </p>
            </div>
          ) : (
            <div className="grid gap-3 sm:grid-cols-2">
              {(member.effectiveAccess?.apps ?? [])
                .filter((app) => app.licensed && app.hasAccess)
                .map((app) => (
                  <div
                    key={app.appId}
                    className="rounded-lg p-3"
                    style={{ background: 'var(--dash-bg)', border: '1px solid var(--dash-border)' }}
                  >
                    <div className="flex items-center justify-between gap-2">
                      <span
                        className="text-xs font-semibold uppercase"
                        style={{ color: 'var(--dash-text)' }}
                      >
                        {app.appId}
                      </span>
                      <span className="text-[10px]" style={{ color: 'var(--dash-text-dim)' }}>
                        {app.sources.join(', ') || t('team.memberDetail.directSource')}
                      </span>
                    </div>
                    <div
                      className="mt-2 flex flex-wrap gap-1.5 text-[10px]"
                      style={{ color: 'var(--dash-text-dim)' }}
                    >
                      <span>
                        {t('team.memberDetail.permissionCount', { count: app.permissions.length })}
                      </span>
                      <span>•</span>
                      <span>
                        {t('team.memberDetail.scopeCount', {
                          count: Object.values(app.permissionScopes).reduce(
                            (count, scopes) => count + scopes.length,
                            0
                          ),
                        })}
                      </span>
                    </div>
                  </div>
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
                {t('team.memberDetail.groupsTitle')}
              </p>
              <button
                type="button"
                onClick={() => navigate({ to: '/team/groups' })}
                className="text-xs"
                style={{ color: '#4583C4' }}
              >
                {t('team.memberDetail.manageGroups')}
              </button>
            </div>
            <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
              {t('team.memberDetail.groupsDescription')}
            </p>
          </div>
          <div className="px-5 py-4">
            {memberGroups.length === 0 ? (
              <EmptyState
                icon={UsersRound}
                title={t('team.memberDetail.notInGroupTitle')}
                description={t('team.memberDetail.notInGroupDescription')}
                action={{
                  label: t('team.memberDetail.addToGroup'),
                  onClick: () => navigate({ to: '/team/groups' }),
                }}
                size="compact"
              />
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
                      type="button"
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

        {!member.isOwner && (
          <DangerZone
            actions={[
              {
                label: t('team.memberDetail.deactivateLabel'),
                description: t('team.memberDetail.deactivateDescription'),
                buttonLabel: t('team.memberDetail.deactivateButton'),
                onClick: () => setConfirmOpen(true),
              },
            ]}
          />
        )}

        <ConfirmDialog
          open={confirmOpen}
          onOpenChange={setConfirmOpen}
          title={t('team.memberDetail.confirmDeactivateTitle', { name: member.name })}
          description={
            <span className="flex flex-col gap-1.5">
              <span>{t('team.memberDetail.confirmDeactivateDescription')}</span>
              {sub && <SeatSavingsNote pricePerSeatCents={sub.pricePerSeatCents} />}
            </span>
          }
          confirmLabel={t('team.memberDetail.deactivateButton')}
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
