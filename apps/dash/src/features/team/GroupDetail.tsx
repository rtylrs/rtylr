import React, { useEffect, useState } from 'react';
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
  type ResourcePermission,
  type ResourcePermissionLevel,
} from '../../shared/hooks/queries/useGroups';
import { useTeam } from '../../shared/hooks/queries/useTeam';
import { APP_RESOURCES, type PermissionLevel } from '../../shared/config/appResources';

type LocalPermissionLevel = PermissionLevel; // 'none' | 'viewer' | 'editor'

const LEVELS: { value: LocalPermissionLevel; tKey: string }[] = [
  { value: 'none', tKey: 'groupDetail.levelNone' },
  { value: 'viewer', tKey: 'groupDetail.levelViewer' },
  { value: 'editor', tKey: 'groupDetail.levelEditor' },
];

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

const GroupDetail = (): React.JSX.Element => {
  const { t } = useTranslation();
  const { id } = useParams({ strict: false });
  const navigate = useNavigate();
  const isNew = id === 'new';

  const { data: group, isLoading } = useGroup(isNew ? undefined : id);
  const createGroup = useCreateGroup();
  const updateGroup = useUpdateGroup(id);
  const addMembers = useAddGroupMembers(id);
  const removeMember = useRemoveGroupMember(id);
  const { data: teamData } = useTeam({ pageSize: 50 });
  const allMembers = teamData?.members ?? [];

  const [tab, setTab] = useState<'apps' | 'members'>('apps');
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [color, setColor] = useState('#6366f1');
  const [resourcePerms, setResourcePerms] = useState<Map<string, LocalPermissionLevel>>(new Map());
  const [showAddMembers, setShowAddMembers] = useState(false);
  const [selectedToAdd, setSelectedToAdd] = useState<number[]>([]);
  const [dirty, setDirty] = useState(false);

  useEffect(() => {
    if (group) {
      setName(group.name);
      setDescription(group.description ?? '');
      setColor(group.color ?? '#6366f1');
      const perms = new Map<string, LocalPermissionLevel>();
      for (const p of group.resourcePermissions) {
        perms.set(p.resourceKey, p.permissionLevel);
      }
      setResourcePerms(perms);
      setDirty(false);
    } else if (isNew) {
      setResourcePerms(new Map());
    }
  }, [group, isNew]);

  const setLevel = (resourceKey: string, level: LocalPermissionLevel): void => {
    setResourcePerms((prev) => {
      const next = new Map(prev);
      if (level === 'none') {
        next.delete(resourceKey);
      } else {
        next.set(resourceKey, level);
      }
      return next;
    });
    setDirty(true);
  };

  const buildPermissions = (): ResourcePermission[] => {
    const result: ResourcePermission[] = [];
    resourcePerms.forEach((level, resourceKey) => {
      result.push({ resourceKey, permissionLevel: level as ResourcePermissionLevel });
    });
    return result;
  };

  const activeSectionCount = resourcePerms.size;

  const handleSave = async (): Promise<void> => {
    const resourcePermissions = buildPermissions();
    try {
      if (isNew) {
        const created = await createGroup.mutateAsync({
          name,
          description,
          color,
          resourcePermissions,
        });
        navigate({ to: `/team/groups/${created.id}`, replace: true });
      } else {
        await updateGroup.mutateAsync({ name, description, color, resourcePermissions });
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

  const currentMemberIds = new Set((group?.members ?? []).map((m) => m.userId));
  const addableMembersList = allMembers.filter((m) => !currentMemberIds.has(Number(m.id)));

  const isSaving = createGroup.isPending || updateGroup.isPending;

  if (isLoading && !isNew) {
    return (
      <div className="p-4">
        <SkeletonList rows={6} />
      </div>
    );
  }

  return (
    <>
      <HelmetContainer title="Group detail · rtylr" />
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
                  ? t('team.groupDetail.tabApps', { count: activeSectionCount })
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
            {/* Column headers */}
            <div
              className="grid grid-cols-12 gap-2 px-5 py-2.5 text-[10px] font-semibold uppercase tracking-wider"
              style={{
                color: 'var(--dash-text-dim)',
                borderBottom: '1px solid var(--dash-border)',
              }}
            >
              <div className="col-span-6">{t('team.groupDetail.columnSection')}</div>
              {LEVELS.map((l) => (
                <div key={l.value} className="col-span-2 text-center">
                  {t(l.tKey)}
                </div>
              ))}
            </div>

            {APP_RESOURCES.map((app) => (
              <React.Fragment key={app.id}>
                {/* App header row */}
                <div
                  className="px-5 py-2.5"
                  style={{
                    background: 'var(--dash-hover)',
                    borderBottom: '1px solid var(--dash-border)',
                  }}
                >
                  <span
                    className="text-xs font-semibold uppercase tracking-wider"
                    style={{ color: 'var(--dash-text-muted)' }}
                  >
                    {app.label}
                  </span>
                </div>

                {/* Section rows */}
                {app.sections.map((section) => {
                  const current: LocalPermissionLevel = resourcePerms.get(section.key) ?? 'none';
                  return (
                    <div
                      key={section.key}
                      className="grid grid-cols-12 gap-2 items-center px-5 py-3 transition-colors hover:bg-[var(--dash-hover)]"
                      style={{ borderBottom: '1px solid var(--dash-border)' }}
                    >
                      <div className="col-span-6 pl-3">
                        <span className="text-sm font-medium" style={{ color: 'var(--dash-text)' }}>
                          {section.label}
                        </span>
                      </div>
                      {LEVELS.map((l) => (
                        <div key={l.value} className="col-span-2 flex justify-center">
                          <button
                            type="button"
                            aria-label={t('team.groupDetail.radioAriaLabel', {
                              section: section.label,
                              level: t(l.tKey),
                            })}
                            onClick={() => setLevel(section.key, l.value)}
                            className="h-5 w-5 rounded-full border-2 flex items-center justify-center transition-all"
                            style={{
                              borderColor: current === l.value ? color : 'var(--dash-border)',
                              background: current === l.value ? color : 'transparent',
                            }}
                          >
                            {current === l.value && (
                              <div className="h-1.5 w-1.5 rounded-full bg-white" />
                            )}
                          </button>
                        </div>
                      ))}
                    </div>
                  );
                })}
              </React.Fragment>
            ))}
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
                onClick={() => setShowAddMembers(true)}
                className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-colors"
                style={{
                  background: 'var(--accent-muted)',
                  color: 'var(--accent)',
                  border: '1px solid var(--accent-border)',
                }}
              >
                <Plus size={12} /> {t('team.groupDetail.addMembersButton')}
              </button>
            </div>

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
                      onClick: () => setShowAddMembers(true),
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
                        style={{ background: 'var(--accent-muted)', color: 'var(--accent)' }}
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
                {addableMembersList.length === 0 ? (
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
                            <p className="text-[11px]" style={{ color: 'var(--dash-text-dim)' }}>
                              {m.email}
                            </p>
                          </div>
                        </label>
                      );
                    })}
                  </div>
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
                    {t('groupDetail.addNButton', { count: selectedToAdd.length })}
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
