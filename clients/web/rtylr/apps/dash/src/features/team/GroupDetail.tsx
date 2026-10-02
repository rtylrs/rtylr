import { PermissionProfilePicker } from './PermissionProfilePicker';
import React, { useEffect, useRef, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useNavigate, useParams } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { ArrowLeft, Save, Users, Plus, UserCheck, X } from 'lucide-react';
import { motion } from 'motion/react';
import { EmptyState } from '@voxire/ui';
import { Button, Input } from '../../shared/components/ui';
import { SkeletonList } from '../../shared/components/dash';
import {
  useGroup,
  useCreateGroup,
  useUpdateGroup,
  useAddGroupMembers,
  useRemoveGroupMember,
} from '../../shared/hooks/queries/useGroups';
import { useTeam } from '../../shared/hooks/queries/useTeam';
import { usePermissionProfiles } from '../../shared/hooks/queries/usePermissions';
import { APP_NAMES } from '../../shared/config/appNames';
import {
  buildPermissionProfileAssignments,
  getAssignedProfileIdsByApp,
} from './authorizationDraft';

const APP_IDS = ['pos', 'erp', 'hr', 'crm', 'finance', 'flow', 'insights', 'menu', 'recruit'];

const PRESET_COLORS = [
  '#6366f1',
  '#4583C4',
  '#10b981',
  '#f59e0b',
  '#8b5cf6',
  '#ef4444',
  '#64C5B9',
  '#ec4899',
  '#f97316',
];

const GROUP_MEMBER_DISPLAY_LIMIT = 500;

const GroupDetail = (): React.JSX.Element => {
  const { t } = useTranslation();
  const { id } = useParams({ strict: false });
  const navigate = useNavigate();
  const isNew = id === 'new';

  const { data: group, isLoading, isError, refetch } = useGroup(isNew ? undefined : id);
  const createGroup = useCreateGroup();
  const updateGroup = useUpdateGroup(id);
  const addMembers = useAddGroupMembers(id);
  const removeMember = useRemoveGroupMember(id);
  const { data: permissionProfiles = [] } = usePermissionProfiles();
  const [teamPage, setTeamPage] = useState(1);
  const {
    data: teamData,
    isLoading: teamLoading,
    isError: teamError,
    refetch: refetchTeam,
  } = useTeam({ page: teamPage, pageSize: 50 });
  const allMembers = teamData?.members ?? [];

  const [tab, setTab] = useState<'apps' | 'members'>('apps');
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [color, setColor] = useState('#6366f1');
  const [licensedAppIds, setLicensedAppIds] = useState<string[]>([]);
  const [profileIdsByApp, setProfileIdsByApp] = useState<Record<string, string[]>>({});
  const [showAddMembers, setShowAddMembers] = useState(false);
  const [selectedToAdd, setSelectedToAdd] = useState<number[]>([]);
  const [dirty, setDirty] = useState(false);
  const initializedGroupId = useRef<string | null>(null);

  useEffect(() => {
    if (group && !isNew) {
      const groupKey = String(group.id);
      if (initializedGroupId.current === groupKey) return;
      initializedGroupId.current = groupKey;
      setName(group.name);
      setDescription(group.description ?? '');
      setColor(group.color ?? '#6366f1');
      setLicensedAppIds(group.licensedAppIds);
      setProfileIdsByApp(getAssignedProfileIdsByApp(group.permissionProfileAssignments, 'group'));
      setDirty(false);
    } else if (isNew && initializedGroupId.current !== 'new') {
      initializedGroupId.current = 'new';
      setLicensedAppIds([]);
      setProfileIdsByApp({});
    }
  }, [group, isNew]);

  const toggleLicensedApp = (appId: string): void => {
    setLicensedAppIds((prev) =>
      prev.includes(appId) ? prev.filter((id) => id !== appId) : [...prev, appId]
    );
    setDirty(true);
  };

  const activeAppCount = licensedAppIds.length;

  const handleSave = async (): Promise<void> => {
    const permissionProfileAssignments = buildPermissionProfileAssignments(
      licensedAppIds,
      profileIdsByApp,
      'group',
      group?.permissionProfileAssignments
    );
    try {
      if (isNew) {
        const created = await createGroup.mutateAsync({
          name,
          description,
          color,
          licensedAppIds,
          permissionProfileAssignments,
        });
        navigate({ to: `/team/groups/${created.id}`, replace: true });
      } else {
        await updateGroup.mutateAsync({
          name,
          description,
          color,
          licensedAppIds,
          permissionProfileAssignments,
        });
      }
      setDirty(false);
    } catch {
      // onError in hooks handles toast
    }
  };

  const handleAddMembers = async (): Promise<void> => {
    if (selectedToAdd.length === 0) return;
    try {
      await addMembers.mutateAsync(selectedToAdd);
      setSelectedToAdd([]);
      setShowAddMembers(false);
    } catch {
      // onError in hook handles toast
    }
  };

  const openAddMembers = () => {
    setTeamPage(1);
    setShowAddMembers(true);
  };

  const groupMembers = group?.members ?? [];
  const currentMemberIds = new Set(groupMembers.map((m) => m.userId));
  const addableMembersList = allMembers.filter((m) => !currentMemberIds.has(Number(m.id)));
  const teamHasNextPage = teamData?.hasMore ?? false;
  const teamTotalPages = teamData?.total
    ? Math.max(1, Math.ceil(teamData.total / teamData.pageSize))
    : undefined;

  const isSaving = createGroup.isPending || updateGroup.isPending;

  if (isLoading && !isNew) {
    return (
      <div className="p-4">
        <SkeletonList rows={6} />
      </div>
    );
  }

  if (isError && !isNew) {
    return (
      <div className="p-4">
        <EmptyState
          icon={Users}
          title={t('team.groupDetail.loadFailed', 'Unable to load this group')}
          description={t(
            'team.groupDetail.loadFailedDescription',
            'The group could not be loaded. Try again before editing its members.'
          )}
          action={{ label: t('common.refresh'), onClick: () => void refetch() }}
        />
      </div>
    );
  }

  return (
    <>
      <HelmetContainer title={`${t('team.groupDetail.pageTitle')} · rtylr`} />
      <div className="space-y-5 max-w-3xl">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <button
              type="button"
              onClick={() => navigate({ to: '/team/groups' })}
              aria-label={t('team.groupDetail.backAriaLabel')}
              className="h-8 w-8 rounded-lg flex items-center justify-center hover:bg-[var(--dash-hover)] transition-colors"
              style={{ color: 'var(--dash-text-muted)' }}
            >
              <ArrowLeft size={16} />
            </button>
            <div>
              <h1 className="text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
                {isNew ? t('team.groupDetail.newGroupTitle') : name}
              </h1>
              {!isNew && (
                <p className="text-[11px] mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
                  {t(
                    (group?.memberCount ?? 0) === 1
                      ? 'groupDetail.membersCountSingular'
                      : 'groupDetail.membersCountPlural',
                    { count: group?.memberCount ?? 0 }
                  )}
                </p>
              )}
            </div>
          </div>
          <Button onClick={handleSave} disabled={isSaving || (!dirty && !isNew)}>
            <Save size={14} />
            {isSaving ? t('team.groupDetail.savingButton') : t('team.groupDetail.saveButton')}
          </Button>
        </div>

        {/* General info */}
        <div
          className="rounded-xl p-5 space-y-4"
          style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
        >
          <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
            {t('team.groupDetail.generalSection')}
          </p>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <label className="text-xs font-medium" style={{ color: 'var(--dash-text-muted)' }}>
                {t('team.groupDetail.labelName')}
              </label>
              <Input
                value={name}
                onChange={(e) => {
                  setName(e.target.value);
                  setDirty(true);
                }}
                placeholder={t('team.groupDetail.namePlaceholder')}
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium" style={{ color: 'var(--dash-text-muted)' }}>
                {t('team.groupDetail.labelDescription')}
              </label>
              <Input
                value={description}
                onChange={(e) => {
                  setDescription(e.target.value);
                  setDirty(true);
                }}
                placeholder={t('team.groupDetail.descriptionPlaceholder')}
              />
            </div>
          </div>
          <div className="space-y-1.5">
            <label className="text-xs font-medium" style={{ color: 'var(--dash-text-muted)' }}>
              {t('team.groupDetail.labelColor')}
            </label>
            <div className="flex items-center gap-2 flex-wrap">
              {PRESET_COLORS.map((c) => (
                <button
                  type="button"
                  key={c}
                  aria-label={t('team.groupDetail.colorAriaLabel', { color: c })}
                  onClick={() => {
                    setColor(c);
                    setDirty(true);
                  }}
                  className="h-6 w-6 rounded-full transition-transform hover:scale-110"
                  style={{
                    background: c,
                    outline: color === c ? `2px solid ${c}` : 'none',
                    outlineOffset: '2px',
                  }}
                />
              ))}
            </div>
          </div>
        </div>

        {/* Tabs */}
        {!isNew && (
          <div
            className="flex items-center gap-1"
            style={{ borderBottom: '1px solid var(--dash-border)' }}
          >
            {(['apps', 'members'] as const).map((tabKey) => (
              <button
                type="button"
                key={tabKey}
                onClick={() => setTab(tabKey)}
                className="px-4 py-2 text-sm font-medium transition-colors"
                style={{
                  color: tab === tabKey ? 'var(--dash-text)' : 'var(--dash-text-muted)',
                  borderBottom:
                    tab === tabKey ? '2px solid var(--accent)' : '2px solid transparent',
                  marginBottom: '-1px',
                }}
              >
                {tabKey === 'apps'
                  ? t('team.groupDetail.tabApps', { count: activeAppCount })
                  : t('team.groupDetail.tabMembers', { count: group?.members?.length ?? 0 })}
              </button>
            ))}
          </div>
        )}

        {/* Apps tab */}
        {(isNew || tab === 'apps') && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="rounded-xl overflow-hidden"
            style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
          >
            {/* Application licenses and permission profiles */}
            <div
              className="grid grid-cols-12 gap-2 px-5 py-2.5 text-[10px] font-semibold uppercase tracking-wider"
              style={{
                color: 'var(--dash-text-dim)',
                borderBottom: '1px solid var(--dash-border)',
              }}
            >
              <div className="col-span-4">{t('team.groupDetail.columnApp')}</div>
              <div className="col-span-2 text-center">{t('team.groupDetail.columnLicense')}</div>
              <div className="col-span-6">{t('team.groupDetail.columnPermissionProfile')}</div>
            </div>

            {APP_IDS.map((appId) => {
              const appName = APP_NAMES[appId as keyof typeof APP_NAMES];
              const appProfiles = permissionProfiles.filter((profile) => profile.appId === appId);
              const licensed = licensedAppIds.includes(appId);
              return (
                <div
                  key={appId}
                  className="grid grid-cols-12 gap-2 items-center px-5 py-3 transition-colors hover:bg-[var(--dash-hover)]"
                  style={{ borderBottom: '1px solid var(--dash-border)' }}
                >
                  <div className="col-span-4">
                    <span className="text-sm font-medium" style={{ color: 'var(--dash-text)' }}>
                      {t(appName.short)}
                    </span>
                  </div>
                  <div className="col-span-2 flex justify-center">
                    <input
                      type="checkbox"
                      aria-label={t('team.groupDetail.licenseAppAriaLabel', {
                        app: t(appName.short),
                      })}
                      checked={licensed}
                      onChange={() => toggleLicensedApp(appId)}
                      style={{ accentColor: color }}
                    />
                  </div>
                  <div className="col-span-6">
                    <PermissionProfilePicker
                      label={t('team.groupDetail.permissionProfileForApp', {
                        app: t(appName.short),
                      })}
                      profiles={appProfiles}
                      selectedIds={profileIdsByApp[appId] ?? []}
                      disabled={!licensed}
                      onChange={(ids) => {
                        setProfileIdsByApp((current) => ({ ...current, [appId]: ids }));
                        setDirty(true);
                      }}
                    />
                  </div>
                </div>
              );
            })}
          </motion.div>
        )}

        {/* Members tab */}
        {!isNew && tab === 'members' && (
          <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="space-y-3">
            <div className="flex items-center justify-between">
              <p className="text-sm font-medium" style={{ color: 'var(--dash-text)' }}>
                {t(
                  (group?.members?.length ?? 0) === 1
                    ? 'groupDetail.membersCountSingular'
                    : 'groupDetail.membersCountPlural',
                  { count: group?.members?.length ?? 0 }
                )}
              </p>
              <button
                type="button"
                onClick={openAddMembers}
                className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-colors"
                style={{
                  background: 'var(--accent-muted)',
                  color: 'var(--accent-on-muted)',
                  border: '1px solid var(--accent-border)',
                }}
              >
                <Plus size={12} /> {t('team.groupDetail.addMembersButton')}
              </button>
            </div>

            {groupMembers.length >= GROUP_MEMBER_DISPLAY_LIMIT && (
              <p
                role="status"
                className="rounded-lg px-3 py-2 text-xs"
                style={{
                  background: 'var(--dash-hover)',
                  color: 'var(--dash-text-muted)',
                }}
              >
                {t(
                  'team.groupDetail.membersTruncated',
                  'Showing the first {{count}} members. Additional members require a paginated group-member API.',
                  { count: GROUP_MEMBER_DISPLAY_LIMIT }
                )}
              </p>
            )}

            <div
              className="rounded-xl overflow-hidden"
              style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
            >
              {(group?.members?.length ?? 0) === 0 ? (
                <div className="p-4">
                  <EmptyState
                    icon={Users}
                    title={t('team.groupDetail.noMembersTitle')}
                    description={t('team.groupDetail.noMembersDescription')}
                    action={{
                      label: t('team.groupDetail.addMembersButton'),
                      onClick: openAddMembers,
                    }}
                  />
                </div>
              ) : (
                group?.members?.map((m) => (
                  <div
                    key={m.userId}
                    className="flex items-center justify-between px-5 py-3 hover:bg-[var(--dash-hover)] transition-colors"
                    style={{ borderBottom: '1px solid var(--dash-border)' }}
                  >
                    <div className="flex items-center gap-3">
                      <div
                        className="h-7 w-7 rounded-full flex items-center justify-center text-[10px] font-bold shrink-0"
                        style={{
                          background: 'var(--accent-muted)',
                          color: 'var(--accent-on-muted)',
                        }}
                      >
                        {(m.fullName || m.email || '?').charAt(0).toUpperCase()}
                      </div>
                      <div>
                        <p className="text-sm font-medium" style={{ color: 'var(--dash-text)' }}>
                          {m.fullName || m.email}
                        </p>
                        <p className="text-[11px]" style={{ color: 'var(--dash-text-dim)' }}>
                          {m.email}
                        </p>
                      </div>
                    </div>
                    <button
                      type="button"
                      aria-label={t('team.groupDetail.removeMemberAriaLabel', {
                        name: m.fullName || m.email,
                      })}
                      onClick={() => removeMember.mutate(m.userId)}
                      className="h-7 w-7 rounded-lg flex items-center justify-center hover:bg-[var(--dash-hover)] transition-colors"
                      style={{ color: 'var(--dash-text-muted)' }}
                    >
                      <X size={13} />
                    </button>
                  </div>
                ))
              )}
            </div>

            {/* Add members panel */}
            {showAddMembers && (
              <div
                className="rounded-xl p-4 space-y-3"
                style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
              >
                <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
                  {t('team.groupDetail.addMembersTitle')}
                </p>
                {teamLoading ? (
                  <SkeletonList rows={3} withAvatar={false} />
                ) : teamError ? (
                  <EmptyState
                    icon={Users}
                    title={t('team.groupDetail.loadMembersFailed', 'Unable to load team members')}
                    description={t(
                      'team.groupDetail.loadMembersFailedDescription',
                      'Refresh the candidate list before adding members.'
                    )}
                    action={{ label: t('common.refresh'), onClick: () => void refetchTeam() }}
                    size="compact"
                  />
                ) : (
                  <>
                    {addableMembersList.length === 0 && !teamHasNextPage ? (
                      <EmptyState
                        icon={UserCheck}
                        title={t('team.groupDetail.allMembersInGroup')}
                        action={{
                          label: t('common.close'),
                          onClick: () => setShowAddMembers(false),
                        }}
                        size="compact"
                      />
                    ) : (
                      <div className="space-y-1 max-h-48 overflow-y-auto">
                        {addableMembersList.map((m) => {
                          const mid = Number(m.id);
                          const checked = selectedToAdd.includes(mid);
                          return (
                            <label
                              key={m.id}
                              className="flex items-center gap-3 px-3 py-2 rounded-lg cursor-pointer hover:bg-[var(--dash-hover)] transition-colors"
                            >
                              <input
                                type="checkbox"
                                checked={checked}
                                onChange={() =>
                                  setSelectedToAdd((prev) =>
                                    checked ? prev.filter((x) => x !== mid) : [...prev, mid]
                                  )
                                }
                                className="rounded"
                                style={{ accentColor: 'var(--accent)' }}
                              />
                              <div>
                                <p
                                  className="text-sm font-medium"
                                  style={{ color: 'var(--dash-text)' }}
                                >
                                  {m.name || m.email}
                                </p>
                                <p
                                  className="text-[11px]"
                                  style={{ color: 'var(--dash-text-dim)' }}
                                >
                                  {m.email}
                                </p>
                              </div>
                            </label>
                          );
                        })}
                      </div>
                    )}
                    {(teamHasNextPage || teamPage > 1) && (
                      <div
                        className="flex items-center justify-between gap-2 pt-2"
                        role="navigation"
                        aria-label={t('team.groupDetail.candidatePagination', 'Member pages')}
                      >
                        <span className="text-xs" style={{ color: 'var(--dash-text-dim)' }}>
                          {t('team.groupDetail.candidatePage', 'Page {{page}}', {
                            page: teamPage,
                            total: teamTotalPages,
                          })}
                        </span>
                        <div className="flex items-center gap-2">
                          <Button
                            variant="outline"
                            size="sm"
                            disabled={teamPage <= 1 || teamLoading}
                            onClick={() => setTeamPage((page) => Math.max(1, page - 1))}
                            aria-label={t('common.pagination.previous', 'Previous')}
                          >
                            {t('common.pagination.previous', 'Previous')}
                          </Button>
                          <Button
                            variant="outline"
                            size="sm"
                            disabled={!teamHasNextPage || teamLoading}
                            onClick={() => setTeamPage((page) => page + 1)}
                            aria-label={t('common.pagination.next', 'Next')}
                          >
                            {t('common.pagination.next', 'Next')}
                          </Button>
                        </div>
                      </div>
                    )}
                  </>
                )}
                <div className="flex items-center justify-end gap-2">
                  <button
                    type="button"
                    onClick={() => {
                      setShowAddMembers(false);
                      setSelectedToAdd([]);
                    }}
                    className="px-3 py-1.5 rounded-lg text-sm"
                    style={{ color: 'var(--dash-text-muted)' }}
                  >
                    {t('team.groupDetail.cancelButton')}
                  </button>
                  <Button
                    onClick={handleAddMembers}
                    disabled={selectedToAdd.length === 0 || addMembers.isPending}
                  >
                    {t('team.groupDetail.addNButton', { count: selectedToAdd.length })}
                  </Button>
                </div>
              </div>
            )}
          </motion.div>
        )}
      </div>
    </>
  );
};

export default GroupDetail;
