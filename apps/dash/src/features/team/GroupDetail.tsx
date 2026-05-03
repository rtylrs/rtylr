import React, { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  ArrowLeft,
  Save,
  Users,
  Plus,
  X,
  MonitorSpeaker,
  Package,
  UserCheck,
  CreditCard,
  Layers,
} from 'lucide-react';
import { Users as UsersIcon } from 'lucide-react';
import { motion } from 'motion/react';
import { Button, Input } from '../../shared/components/ui';
import { EmptyState, SkeletonList } from '../../shared/components/dash';
import {
  useGroup,
  useCreateGroup,
  useUpdateGroup,
  useAddGroupMembers,
  useRemoveGroupMember,
  type AppPermission,
  type PermissionLevel,
} from '../../shared/hooks/queries/useGroups';
import { useTeam } from '../../shared/hooks/queries/useTeam';

const ALL_APPS = [
  { key: 'pos', label: 'POS', icon: MonitorSpeaker, color: '#4583C4' },
  { key: 'erp', label: 'ERP', icon: Package, color: '#10b981' },
  { key: 'hr', label: 'HR', icon: UserCheck, color: '#f59e0b' },
  { key: 'crm', label: 'CRM', icon: UsersIcon, color: '#8b5cf6' },
  { key: 'finance', label: 'Finance', icon: CreditCard, color: '#ef4444' },
  { key: 'flow', label: 'Flow', icon: Layers, color: '#64C5B9' },
];

const LEVELS: { value: PermissionLevel | 'none'; label: string }[] = [
  { value: 'none', label: 'None' },
  { value: 'use', label: 'Use' },
  { value: 'edit', label: 'Edit' },
  { value: 'own', label: 'Own' },
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

const GroupDetail: React.FC = (): React.JSX.Element => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const isNew = id === 'new';

  const { data: group, isLoading } = useGroup(isNew ? undefined : id);
  const createGroup = useCreateGroup();
  const updateGroup = useUpdateGroup(id);
  const addMembers = useAddGroupMembers(id);
  const removeMember = useRemoveGroupMember(id);
  const { data: allMembers = [] } = useTeam();

  const [tab, setTab] = useState<'apps' | 'members'>('apps');
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [color, setColor] = useState('#6366f1');
  const [appPerms, setAppPerms] = useState<Record<string, PermissionLevel | 'none'>>({});
  const [showAddMembers, setShowAddMembers] = useState(false);
  const [selectedToAdd, setSelectedToAdd] = useState<number[]>([]);
  const [dirty, setDirty] = useState(false);

  useEffect(() => {
    if (group) {
      setName(group.name);
      setDescription(group.description ?? '');
      setColor(group.color ?? '#6366f1');
      const perms: Record<string, PermissionLevel | 'none'> = {};
      for (const app of ALL_APPS) {
        perms[app.key] = 'none';
      }
      for (const p of group.appPermissions) {
        perms[p.appId] = p.permissionLevel;
      }
      setAppPerms(perms);
      setDirty(false);
    } else if (isNew) {
      const perms: Record<string, PermissionLevel | 'none'> = {};
      for (const app of ALL_APPS) {
        perms[app.key] = 'none';
      }
      setAppPerms(perms);
    }
  }, [group, isNew]);

  const setLevel = (appKey: string, level: PermissionLevel | 'none'): void => {
    setAppPerms((prev) => ({ ...prev, [appKey]: level }));
    setDirty(true);
  };

  const buildPermissions = (): AppPermission[] =>
    Object.entries(appPerms)
      .filter(([, level]) => level !== 'none')
      .map(([appId, level]) => ({ appId, permissionLevel: level as PermissionLevel }));

  const handleSave = async (): Promise<void> => {
    const appPermissions = buildPermissions();
    if (isNew) {
      const created = await createGroup.mutateAsync({ name, description, color, appPermissions });
      navigate(`/team/groups/${created.id}`, { replace: true });
    } else {
      await updateGroup.mutateAsync({ name, description, color, appPermissions });
    }
    setDirty(false);
  };

  const handleAddMembers = async (): Promise<void> => {
    if (selectedToAdd.length === 0) return;
    await addMembers.mutateAsync(selectedToAdd);
    setSelectedToAdd([]);
    setShowAddMembers(false);
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
    <div className="space-y-5 max-w-3xl">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <button
            onClick={() => navigate('/team/groups')}
            aria-label="Back to groups"
            className="h-8 w-8 rounded-lg flex items-center justify-center hover:bg-[var(--dash-hover)] transition-colors"
            style={{ color: 'var(--dash-text-muted)' }}
          >
            <ArrowLeft size={16} />
          </button>
          <div>
            <h1 className="text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
              {isNew ? 'New group' : name}
            </h1>
            {!isNew && (
              <p className="text-[11px] mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
                {group?.memberCount ?? 0} member{(group?.memberCount ?? 0) !== 1 ? 's' : ''}
              </p>
            )}
          </div>
        </div>
        <Button onClick={handleSave} disabled={isSaving || (!dirty && !isNew)}>
          <Save size={14} />
          {isSaving ? 'Saving...' : 'Save'}
        </Button>
      </div>

      {/* General info */}
      <div
        className="rounded-xl p-5 space-y-4"
        style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
      >
        <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
          General
        </p>
        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-1.5">
            <label className="text-xs font-medium" style={{ color: 'var(--dash-text-muted)' }}>
              Group name
            </label>
            <Input
              value={name}
              onChange={(e) => {
                setName(e.target.value);
                setDirty(true);
              }}
              placeholder="e.g. Store Managers"
            />
          </div>
          <div className="space-y-1.5">
            <label className="text-xs font-medium" style={{ color: 'var(--dash-text-muted)' }}>
              Description
            </label>
            <Input
              value={description}
              onChange={(e) => {
                setDescription(e.target.value);
                setDirty(true);
              }}
              placeholder="Optional"
            />
          </div>
        </div>
        <div className="space-y-1.5">
          <label className="text-xs font-medium" style={{ color: 'var(--dash-text-muted)' }}>
            Color
          </label>
          <div className="flex items-center gap-2 flex-wrap">
            {PRESET_COLORS.map((c) => (
              <button
                key={c}
                aria-label={`Select color ${c}`}
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
          {(['apps', 'members'] as const).map((t) => (
            <button
              key={t}
              onClick={() => setTab(t)}
              className="px-4 py-2 text-sm font-medium capitalize transition-colors"
              style={{
                color: tab === t ? 'var(--dash-text)' : 'var(--dash-text-muted)',
                borderBottom: tab === t ? '2px solid #4583C4' : '2px solid transparent',
                marginBottom: '-1px',
              }}
            >
              {t === 'apps'
                ? `Apps (${buildPermissions().length})`
                : `Members (${group?.members?.length ?? 0})`}
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
            <div className="col-span-4">App</div>
            {LEVELS.map((l) => (
              <div key={l.value} className="col-span-2 text-center">
                {l.label}
              </div>
            ))}
          </div>
          {ALL_APPS.map((app) => {
            const Icon = app.icon;
            const current = appPerms[app.key] ?? 'none';
            return (
              <div
                key={app.key}
                className="grid grid-cols-12 gap-2 items-center px-5 py-3 transition-colors hover:bg-[var(--dash-hover)]"
                style={{ borderBottom: '1px solid var(--dash-border)' }}
              >
                <div className="col-span-4 flex items-center gap-2.5">
                  <div
                    className="h-7 w-7 rounded-lg flex items-center justify-center shrink-0"
                    style={{ background: `${app.color}12` }}
                  >
                    <Icon size={14} style={{ color: app.color }} />
                  </div>
                  <span className="text-sm font-medium" style={{ color: 'var(--dash-text)' }}>
                    {app.label}
                  </span>
                </div>
                {LEVELS.map((l) => (
                  <div key={l.value} className="col-span-2 flex justify-center">
                    <button
                      aria-label={`Set ${app.label} to ${l.label}`}
                      onClick={() => setLevel(app.key, l.value)}
                      className="h-5 w-5 rounded-full border-2 flex items-center justify-center transition-all"
                      style={{
                        borderColor: current === l.value ? '#4583C4' : 'var(--dash-border)',
                        background: current === l.value ? '#4583C4' : 'transparent',
                      }}
                    >
                      {current === l.value && <div className="h-1.5 w-1.5 rounded-full bg-white" />}
                    </button>
                  </div>
                ))}
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
              {group?.members?.length ?? 0} member
              {(group?.members?.length ?? 0) !== 1 ? 's' : ''}
            </p>
            <button
              onClick={() => setShowAddMembers(true)}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-colors"
              style={{
                background: 'rgba(69,131,196,0.1)',
                color: '#4583C4',
                border: '1px solid rgba(69,131,196,0.2)',
              }}
            >
              <Plus size={12} /> Add members
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
                  title="No members yet"
                  description="Add team members to this group."
                  actionLabel="Add members"
                  onAction={() => setShowAddMembers(true)}
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
                      style={{ background: 'rgba(69,131,196,0.12)', color: '#4583C4' }}
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
                    aria-label={`Remove ${m.fullName || m.email} from group`}
                    onClick={() => removeMember.mutateAsync(m.userId)}
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
                Add members
              </p>
              {addableMembersList.length === 0 ? (
                <p className="text-sm" style={{ color: 'var(--dash-text-muted)' }}>
                  All team members are already in this group.
                </p>
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
                          style={{ accentColor: '#4583C4' }}
                        />
                        <div>
                          <p className="text-sm font-medium" style={{ color: 'var(--dash-text)' }}>
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
                  onClick={() => {
                    setShowAddMembers(false);
                    setSelectedToAdd([]);
                  }}
                  className="px-3 py-1.5 rounded-lg text-sm"
                  style={{ color: 'var(--dash-text-muted)' }}
                >
                  Cancel
                </button>
                <Button
                  onClick={handleAddMembers}
                  disabled={selectedToAdd.length === 0 || addMembers.isPending}
                >
                  Add {selectedToAdd.length > 0 ? selectedToAdd.length : ''} member
                  {selectedToAdd.length !== 1 ? 's' : ''}
                </Button>
              </div>
            </div>
          )}
        </motion.div>
      )}
    </div>
  );
};

export default GroupDetail;
