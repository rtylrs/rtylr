import React, { useMemo, useState } from 'react';
import {
  AlertCircle,
  Check,
  ChevronDown,
  Eye,
  EyeOff,
  GripVertical,
  Link as LinkIcon,
  PanelLeft,
  Pencil,
  Plus,
  RotateCcw,
  Trash2,
  X,
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { useProjects } from '../../projects/useProjects';
import { ACCENT_DEFAULT } from '../../../shared/lib/colorUtils';
import { cn } from '../../../shared/lib/utils';
import {
  createDefaultSidebarPreferences,
  updatePreferencesCopy,
  useSidebarPreferences,
  type SidebarGroupPreference,
  type SidebarIconKey,
  type SidebarItemPreference,
  type SidebarPreferences,
} from '../../../shared/lib/sidebarPreferences';

const INTERNAL_ROUTES = [
  ['/inbox', 'nav.inbox'],
  ['/my-issues', 'flow.shell.myTasks'],
  ['/today', 'nav.focusToday'],
  ['/projects', 'nav.projects'],
  ['/triage', 'flow.shell.triage'],
  ['/roadmap', 'flow.shell.roadmap'],
  ['/cycles', 'flow.shell.cycles'],
  ['/milestones', 'flow.shell.milestones'],
  ['/timeline', 'flow.shell.timeline'],
  ['/calendar', 'flow.shell.calendar'],
  ['/views', 'flow.shell.views'],
  ['/archive', 'flow.shell.archive'],
  ['/team', 'flow.shell.workload'],
  ['/insights', 'flow.shell.insights'],
  ['/time', 'flow.time.myDay'],
  ['/time/team', 'flow.time.team'],
  ['/time/reports', 'flow.time.reports'],
  ['/directory', 'flow.directory.heading'],
  ['/presence', 'flow.presence.heading'],
  ['/settings', 'nav.settings'],
] as const;

const removeId = (ids: string[], id: string): string[] => ids.filter((entry) => entry !== id);
const iconOptions: SidebarIconKey[] = [
  'layout-grid',
  'list-todo',
  'calendar',
  'clock',
  'eye',
  'map',
  'users',
];

const moveGroupInPreferences = (
  current: SidebarPreferences,
  groupId: string,
  offset: number
): SidebarPreferences => {
  const next = updatePreferencesCopy(current);
  const index = next.groupOrder.indexOf(groupId);
  const destination = index + offset;
  if (index < 0 || destination < 0 || destination >= next.groupOrder.length) return next;
  [next.groupOrder[index], next.groupOrder[destination]] = [
    next.groupOrder[destination],
    next.groupOrder[index],
  ];
  return next;
};

const itemName = (item: SidebarItemPreference, t: (key: string) => string): string =>
  item.kind === 'builtin' ? t(item.label) : item.label;

const GroupEditor: React.FC<{
  group: SidebarGroupPreference;
  preferences: SidebarPreferences;
  update: (mutator: (draft: SidebarPreferences) => SidebarPreferences) => void;
  onMove: (offset: number) => void;
}> = ({ group, preferences, update, onMove }) => {
  const { t } = useTranslation();
  const [open, setOpen] = useState(true);
  const [editing, setEditing] = useState(false);
  const [deleteConfirm, setDeleteConfirm] = useState(false);
  const [editName, setEditName] = useState(group.label);
  const itemsById = useMemo(
    () =>
      new Map(preferences.items.map((item): [string, SidebarItemPreference] => [item.id, item])),
    [preferences.items]
  );
  const items = group.itemIds
    .map((id) => itemsById.get(id))
    .filter((item): item is SidebarItemPreference => Boolean(item));
  return (
    <section
      className="rounded-xl border p-3"
      style={{ borderColor: 'var(--flow-border)', background: 'var(--flow-surface)' }}
    >
      <div className="flex min-h-11 items-center gap-2">
        <GripVertical size={16} style={{ color: 'var(--flow-text-dim)' }} />
        <button
          type="button"
          onClick={() => setOpen((value) => !value)}
          className="flex min-h-11 min-w-0 flex-1 items-center gap-2 text-start"
        >
          <PanelLeft size={15} style={{ color: 'var(--accent)' }} />
          {editing ? (
            <input
              autoFocus
              value={editName}
              onChange={(event) => setEditName(event.target.value)}
              onBlur={() => {
                const label = editName.trim();
                if (label)
                  update((draft) => {
                    const next = updatePreferencesCopy(draft);
                    const candidate = next.groups.find((entry) => entry.id === group.id);
                    if (candidate) candidate.label = label;
                    return next;
                  });
                setEditing(false);
              }}
              onKeyDown={(event) => {
                if (event.key === 'Enter') event.currentTarget.blur();
                if (event.key === 'Escape') setEditing(false);
              }}
              className="min-w-0 flex-1 rounded-md border px-2 py-1 text-[16px] outline-none"
              style={{
                background: 'var(--flow-input-bg)',
                borderColor: 'var(--flow-border)',
                color: 'var(--flow-text)',
              }}
            />
          ) : (
            <span
              className="min-w-0 flex-1 truncate text-[13px] font-semibold"
              style={{ color: 'var(--flow-text)' }}
            >
              {group.kind === 'custom' ? group.label : t(group.label)}
            </span>
          )}
          <ChevronDown
            size={15}
            className={cn('transition-transform', !open && '-rotate-90')}
            style={{ color: 'var(--flow-text-muted)' }}
          />
        </button>
        <button
          type="button"
          aria-label={t('flow.sidebar.moveGroupUp')}
          onClick={() => onMove(-1)}
          className="flex h-10 w-10 items-center justify-center rounded-md hover:bg-[var(--flow-hover)]"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          ↑
        </button>
        {group.kind === 'custom' && (
          <button
            type="button"
            aria-label={t('flow.sidebar.editGroup')}
            onClick={() => setEditing(true)}
            className="flex h-10 w-10 items-center justify-center rounded-md hover:bg-[var(--flow-hover)]"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            <Pencil size={15} />
          </button>
        )}
        <button
          type="button"
          aria-label={t('flow.sidebar.moveGroupDown')}
          onClick={() => onMove(1)}
          className="flex h-10 w-10 items-center justify-center rounded-md hover:bg-[var(--flow-hover)]"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          ↓
        </button>
        <button
          type="button"
          aria-label={group.visible ? t('flow.sidebar.hideGroup') : t('flow.sidebar.showGroup')}
          disabled={group.lockedVisibility}
          onClick={() =>
            update((draft) => {
              const next = updatePreferencesCopy(draft);
              const candidate = next.groups.find((entry) => entry.id === group.id);
              if (candidate && !candidate.lockedVisibility) candidate.visible = !candidate.visible;
              return next;
            })
          }
          className="flex h-10 w-10 items-center justify-center rounded-md hover:bg-[var(--flow-hover)] disabled:cursor-not-allowed disabled:opacity-35"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          {group.visible ? <Eye size={15} /> : <EyeOff size={15} />}
        </button>
        {group.kind === 'custom' && (
          <button
            type="button"
            aria-label={t('flow.sidebar.deleteGroup')}
            onClick={() => setDeleteConfirm((value) => !value)}
            className="flex h-10 w-10 items-center justify-center rounded-md hover:bg-[var(--flow-hover)]"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            <Trash2 size={15} />
          </button>
        )}
      </div>
      {deleteConfirm && group.kind === 'custom' && (
        <div
          className="mt-2 flex flex-wrap items-center gap-2 rounded-lg border p-2"
          style={{ borderColor: 'var(--flow-border)' }}
        >
          <span className="me-auto text-[11px]" style={{ color: 'var(--flow-text-muted)' }}>
            {t('flow.sidebar.deleteGroup')}
          </span>
          <button
            type="button"
            onClick={() => {
              update((draft) => {
                const next = updatePreferencesCopy(draft);
                next.groups = next.groups.filter((entry) => entry.id !== group.id);
                next.groupOrder = removeId(next.groupOrder, group.id);
                next.items = next.items.map((item) =>
                  item.groupId === group.id ? { ...item, groupId: undefined } : item
                );
                next.topLevelItemIds.push(
                  ...group.itemIds.filter((id) => !next.topLevelItemIds.includes(id))
                );
                return next;
              });
              setDeleteConfirm(false);
            }}
            className="min-h-10 rounded-md px-2 text-[11px] hover:bg-[var(--flow-hover)]"
            style={{ color: 'var(--flow-text)' }}
          >
            {t('flow.sidebar.moveItemsToTop')}
          </button>
          <button
            type="button"
            onClick={() => {
              update((draft) => {
                const next = updatePreferencesCopy(draft);
                const removed = new Set(group.itemIds);
                next.groups = next.groups.filter((entry) => entry.id !== group.id);
                next.groupOrder = removeId(next.groupOrder, group.id);
                next.items = next.items.filter((item) => !removed.has(item.id));
                next.topLevelItemIds = next.topLevelItemIds.filter((id) => !removed.has(id));
                return next;
              });
              setDeleteConfirm(false);
            }}
            className="min-h-10 rounded-md px-2 text-[11px] hover:bg-[var(--flow-hover)]"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            {t('flow.sidebar.deleteGroupAndItems')}
          </button>
        </div>
      )}
      {editing && group.kind === 'custom' && (
        <div
          className="mt-2 flex flex-wrap items-center gap-2 border-t pt-2"
          style={{ borderColor: 'var(--flow-border)' }}
        >
          <label
            className="flex min-h-10 items-center gap-2 text-[11px]"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            <input
              type="color"
              value={group.color ?? ACCENT_DEFAULT}
              onChange={(event) =>
                update((draft) => {
                  const next = updatePreferencesCopy(draft);
                  const candidate = next.groups.find((entry) => entry.id === group.id);
                  if (candidate) candidate.color = event.target.value;
                  return next;
                })
              }
              className="h-8 w-8 rounded-md border-0 p-0"
            />
            {t('flow.sidebar.groupColor')}
          </label>
          <label
            className="flex min-h-10 items-center gap-2 text-[11px]"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            {t('flow.sidebar.groupIcon')}
            <select
              value={group.icon}
              onChange={(event) =>
                update((draft) => {
                  const next = updatePreferencesCopy(draft);
                  const candidate = next.groups.find((entry) => entry.id === group.id);
                  if (candidate) candidate.icon = event.target.value as SidebarIconKey;
                  return next;
                })
              }
              className="min-h-10 rounded-md border px-2 text-[16px]"
              style={{
                background: 'var(--flow-input-bg)',
                borderColor: 'var(--flow-border)',
                color: 'var(--flow-text)',
              }}
            >
              {iconOptions.map((icon) => (
                <option key={icon} value={icon}>
                  {icon}
                </option>
              ))}
            </select>
          </label>
        </div>
      )}
      {open && (
        <div className="mt-2 space-y-1 border-t pt-2" style={{ borderColor: 'var(--flow-border)' }}>
          {items.map((item) => (
            <div
              key={item.id}
              className="flex min-h-11 items-center gap-2 rounded-lg px-2 hover:bg-[var(--flow-hover)]"
            >
              <GripVertical size={14} style={{ color: 'var(--flow-text-dim)' }} />
              <span
                className="min-w-0 flex-1 truncate text-[12px]"
                style={{ color: item.visible ? 'var(--flow-text)' : 'var(--flow-text-dim)' }}
              >
                {itemName(item, t)}
              </span>
              <button
                type="button"
                aria-label={item.visible ? t('flow.sidebar.hideItem') : t('flow.sidebar.showItem')}
                disabled={item.lockedVisibility}
                onClick={() =>
                  update((draft) => {
                    const next = updatePreferencesCopy(draft);
                    const candidate = next.items.find((entry) => entry.id === item.id);
                    if (candidate && !candidate.lockedVisibility)
                      candidate.visible = !candidate.visible;
                    return next;
                  })
                }
                className="flex h-10 w-10 items-center justify-center rounded-md hover:bg-[var(--flow-hover)] disabled:cursor-not-allowed disabled:opacity-35"
                style={{ color: 'var(--flow-text-muted)' }}
              >
                {item.visible ? <Eye size={15} /> : <EyeOff size={15} />}
              </button>
              {item.lockedVisibility && <Check size={14} style={{ color: 'var(--accent)' }} />}
            </div>
          ))}
          {items.length === 0 && (
            <p className="px-2 py-2 text-[12px]" style={{ color: 'var(--flow-text-dim)' }}>
              {t('flow.sidebar.emptyGroup')}
            </p>
          )}
        </div>
      )}
    </section>
  );
};

const SidebarSettingsSection: React.FC = () => {
  const { t } = useTranslation();
  const { preferences, update, isLoading } = useSidebarPreferences();
  const { data: projects = [] } = useProjects();
  const [newGroupName, setNewGroupName] = useState('');
  const [linkLabel, setLinkLabel] = useState('');
  const [linkUrl, setLinkUrl] = useState('');
  const [selectedRoute, setSelectedRoute] = useState<string>(INTERNAL_ROUTES[0][0]);
  const [linkMode, setLinkMode] = useState<'route' | 'external'>('route');
  const [linkError, setLinkError] = useState<string | null>(null);
  const [dialog, setDialog] = useState<'group' | 'link' | null>(null);

  const groups = preferences.groupOrder
    .map((id) => preferences.groups.find((group) => group.id === id))
    .filter((group): group is SidebarGroupPreference => Boolean(group));
  const createGroup = (): void => {
    const label = newGroupName.trim();
    if (!label) return;
    update((draft) => {
      const next = updatePreferencesCopy(draft);
      const id = `custom:${crypto.randomUUID()}`;
      next.groups.push({
        id,
        kind: 'custom',
        label,
        icon: 'layout-grid',
        itemIds: [],
        visible: true,
        collapsed: false,
      });
      next.groupOrder.push(id);
      return next;
    });
    setNewGroupName('');
    setDialog(null);
  };
  const createLink = (): void => {
    const label = linkLabel.trim();
    const url = linkMode === 'route' ? selectedRoute : linkUrl.trim();
    if (!label) {
      setLinkError(t('flow.sidebar.shortcutNameRequired'));
      return;
    }
    if (!url) {
      setLinkError(t('flow.sidebar.shortcutDestinationRequired'));
      return;
    }
    if (linkMode === 'external' && !/^https:\/\//.test(url)) {
      setLinkError(t('flow.sidebar.externalLinkInvalid'));
      return;
    }
    setLinkError(null);
    update((draft) => {
      const next = updatePreferencesCopy(draft);
      const id = `custom:link:${crypto.randomUUID()}`;
      next.items.push({
        id,
        kind: 'custom-link',
        label,
        icon: 'link' as SidebarIconKey,
        route: linkMode === 'route' ? url : undefined,
        url: linkMode === 'external' ? url : undefined,
        openInNewTab: linkMode === 'external',
        visible: true,
      });
      next.topLevelItemIds.push(id);
      return next;
    });
    setLinkLabel('');
    setLinkUrl('');
    setDialog(null);
  };
  const pinProject = (projectId: string): void =>
    update((draft) => {
      const next = updatePreferencesCopy(draft);
      const project = projects.find((entry) => entry.id === projectId);
      if (!project || next.items.some((item) => item.id === `project:${project.id}`)) return next;
      next.items.push({
        id: `project:${project.id}`,
        kind: 'project',
        label: project.name,
        route: `/p/${project.id}`,
        icon: 'layout-grid',
        color: project.color,
        sourceId: project.id,
        visible: true,
      });
      next.topLevelItemIds.push(`project:${project.id}`);
      return next;
    });

  if (isLoading)
    return (
      <div className="p-6 text-[12px]" style={{ color: 'var(--flow-text-muted)' }}>
        {t('flow.shell.loading')}
      </div>
    );
  return (
    <div className="flex flex-col gap-6 p-4 sm:p-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 className="text-[18px] font-semibold" style={{ color: 'var(--flow-text)' }}>
            {t('flow.sidebar.title')}
          </h2>
          <p className="mt-1 max-w-2xl text-[12px]" style={{ color: 'var(--flow-text-muted)' }}>
            {t('flow.sidebar.subtitle')}
          </p>
        </div>
        <button
          type="button"
          onClick={() => update(() => createDefaultSidebarPreferences())}
          className="flex min-h-11 items-center gap-2 rounded-lg border px-3 text-[12px] font-medium hover:bg-[var(--flow-hover)]"
          style={{ borderColor: 'var(--flow-border)', color: 'var(--flow-text)' }}
        >
          <RotateCcw size={14} />
          {t('flow.sidebar.reset')}
        </button>
      </div>

      <div className="order-1 flex flex-wrap gap-2">
        <button
          type="button"
          onClick={() => setDialog('group')}
          className="flex min-h-11 items-center gap-2 rounded-lg px-3 text-[12px] font-semibold"
          style={{ background: 'var(--accent)', color: 'var(--flow-on-accent)' }}
        >
          <Plus size={15} />
          {t('flow.sidebar.newGroup')}
        </button>
        <button
          type="button"
          onClick={() => {
            setLinkError(null);
            setDialog('link');
          }}
          className="flex min-h-11 items-center gap-2 rounded-lg border px-3 text-[12px] font-semibold"
          style={{ borderColor: 'var(--flow-border)', color: 'var(--flow-text)' }}
        >
          <LinkIcon size={15} />
          {t('flow.sidebar.addShortcut')}
        </button>
      </div>

      <section
        className="order-3 rounded-xl border p-3"
        style={{ borderColor: 'var(--flow-border)', background: 'var(--flow-surface)' }}
      >
        <div className="mb-2 flex items-center gap-2">
          <GripVertical size={16} style={{ color: 'var(--flow-text-dim)' }} />
          <h3 className="text-[13px] font-semibold" style={{ color: 'var(--flow-text)' }}>
            {t('flow.sidebar.topItems')}
          </h3>
        </div>
        <div className="space-y-1">
          {preferences.topLevelItemIds
            .map((id) => preferences.items.find((item) => item.id === id))
            .filter((item): item is SidebarItemPreference => Boolean(item))
            .map((item) => (
              <div
                key={item.id}
                className="flex min-h-11 items-center gap-2 rounded-lg px-2 hover:bg-[var(--flow-hover)]"
              >
                <GripVertical size={14} style={{ color: 'var(--flow-text-dim)' }} />
                <span
                  className="min-w-0 flex-1 truncate text-[12px]"
                  style={{ color: item.visible ? 'var(--flow-text)' : 'var(--flow-text-dim)' }}
                >
                  {itemName(item, t)}
                </span>
                <button
                  type="button"
                  aria-label={
                    item.visible ? t('flow.sidebar.hideItem') : t('flow.sidebar.showItem')
                  }
                  disabled={item.lockedVisibility}
                  onClick={() =>
                    update((draft) => {
                      const next = updatePreferencesCopy(draft);
                      const candidate = next.items.find((entry) => entry.id === item.id);
                      if (candidate && !candidate.lockedVisibility)
                        candidate.visible = !candidate.visible;
                      return next;
                    })
                  }
                  className="flex h-10 w-10 items-center justify-center rounded-md hover:bg-[var(--flow-hover)] disabled:cursor-not-allowed disabled:opacity-35"
                  style={{ color: 'var(--flow-text-muted)' }}
                >
                  {item.visible ? <Eye size={15} /> : <EyeOff size={15} />}
                </button>
              </div>
            ))}
        </div>
      </section>

      <div className="order-4 space-y-3">
        {groups.map((group) => (
          <GroupEditor
            key={group.id}
            group={group}
            preferences={preferences}
            update={update}
            onMove={(offset) => update((draft) => moveGroupInPreferences(draft, group.id, offset))}
          />
        ))}
      </div>

      {projects.length > 0 && (
        <section
          className="order-5 rounded-xl border p-4"
          style={{ borderColor: 'var(--flow-border)' }}
        >
          <h3 className="mb-1 text-[13px] font-semibold" style={{ color: 'var(--flow-text)' }}>
            {t('flow.sidebar.pinItems')}
          </h3>
          <p className="mb-3 text-[12px]" style={{ color: 'var(--flow-text-muted)' }}>
            {t('flow.sidebar.pinItemsHint')}
          </p>
          <div className="grid gap-2 sm:grid-cols-2">
            {projects.slice(0, 12).map((project) => (
              <button
                key={project.id}
                type="button"
                onClick={() => pinProject(project.id)}
                disabled={preferences.items.some((item) => item.id === `project:${project.id}`)}
                className="flex min-h-11 items-center gap-2 rounded-lg border px-3 text-start text-[12px] hover:bg-[var(--flow-hover)] disabled:cursor-not-allowed disabled:opacity-45"
                style={{ borderColor: 'var(--flow-border)', color: 'var(--flow-text)' }}
              >
                <Plus size={14} />
                {project.name}
              </button>
            ))}
          </div>
        </section>
      )}

      {dialog === 'group' && (
        <div
          className="order-2 rounded-xl border p-4"
          style={{ borderColor: 'var(--accent)', background: 'var(--flow-surface)' }}
        >
          <label
            className="mb-2 block text-[12px] font-medium"
            style={{ color: 'var(--flow-text)' }}
          >
            {t('flow.sidebar.groupName')}
          </label>
          <div className="flex gap-2">
            <input
              autoFocus
              value={newGroupName}
              onChange={(event) => setNewGroupName(event.target.value)}
              onKeyDown={(event) => event.key === 'Enter' && createGroup()}
              className="min-h-11 min-w-0 flex-1 rounded-lg border px-3 text-[16px] outline-none"
              style={{
                background: 'var(--flow-input-bg)',
                borderColor: 'var(--flow-border)',
                color: 'var(--flow-text)',
              }}
            />
            <button
              type="button"
              onClick={createGroup}
              className="min-h-11 rounded-lg px-4 text-[12px] font-semibold"
              style={{ background: 'var(--accent)', color: 'var(--flow-on-accent)' }}
            >
              {t('common.create')}
            </button>
          </div>
        </div>
      )}

      {dialog === 'link' && (
        <form
          onSubmit={(event) => {
            event.preventDefault();
            createLink();
          }}
          className="order-2 space-y-3 rounded-xl border p-4"
          style={{ borderColor: 'var(--accent)', background: 'var(--flow-surface)' }}
        >
          <div className="flex items-start justify-between gap-3">
            <div>
              <h3 className="text-[13px] font-semibold" style={{ color: 'var(--flow-text)' }}>
                {t('flow.sidebar.addShortcut')}
              </h3>
              <p className="mt-1 text-[11px]" style={{ color: 'var(--flow-text-muted)' }}>
                {t('flow.sidebar.addShortcutHint')}
              </p>
            </div>
            <button
              type="button"
              aria-label={t('common.close')}
              onClick={() => {
                setLinkError(null);
                setDialog(null);
              }}
              className="flex h-10 w-10 shrink-0 items-center justify-center rounded-md hover:bg-[var(--flow-hover)]"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              <X size={15} />
            </button>
          </div>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={() => {
                setLinkError(null);
                setLinkMode('route');
              }}
              className={cn(
                'min-h-10 rounded-lg px-3 text-[12px]',
                linkMode === 'route' && 'font-semibold'
              )}
              style={{
                background: linkMode === 'route' ? 'var(--flow-active)' : 'transparent',
                color: 'var(--flow-text)',
              }}
            >
              {t('flow.sidebar.internalRoute')}
            </button>
            <button
              type="button"
              onClick={() => {
                setLinkError(null);
                setLinkMode('external');
              }}
              className={cn(
                'min-h-10 rounded-lg px-3 text-[12px]',
                linkMode === 'external' && 'font-semibold'
              )}
              style={{
                background: linkMode === 'external' ? 'var(--flow-active)' : 'transparent',
                color: 'var(--flow-text)',
              }}
            >
              {t('flow.sidebar.externalLink')}
            </button>
          </div>
          <label
            htmlFor="sidebar-shortcut-name"
            className="block text-[11px] font-medium"
            style={{ color: 'var(--flow-text)' }}
          >
            {t('flow.sidebar.shortcutName')}
            <input
              id="sidebar-shortcut-name"
              value={linkLabel}
              onChange={(event) => {
                setLinkError(null);
                setLinkLabel(event.target.value);
              }}
              placeholder={t('flow.sidebar.shortcutName')}
              aria-required="true"
              aria-invalid={linkError !== null && linkLabel.trim().length === 0}
              className="mt-1 min-h-11 w-full rounded-lg border px-3 text-[16px] font-normal outline-none"
              style={{
                background: 'var(--flow-input-bg)',
                borderColor:
                  linkError !== null && linkLabel.trim().length === 0
                    ? 'var(--flow-danger-fg)'
                    : 'var(--flow-border)',
                color: 'var(--flow-text)',
              }}
            />
          </label>
          <label
            htmlFor="sidebar-shortcut-destination"
            className="block text-[11px] font-medium"
            style={{ color: 'var(--flow-text)' }}
          >
            {linkMode === 'route'
              ? t('flow.sidebar.internalRoute')
              : t('flow.sidebar.externalLink')}
            {linkMode === 'route' ? (
              <select
                id="sidebar-shortcut-destination"
                value={selectedRoute}
                onChange={(event) => {
                  setLinkError(null);
                  setSelectedRoute(event.target.value);
                }}
                className="mt-1 min-h-11 w-full rounded-lg border px-3 text-[16px] font-normal"
                style={{
                  background: 'var(--flow-input-bg)',
                  borderColor: 'var(--flow-border)',
                  color: 'var(--flow-text)',
                }}
              >
                {INTERNAL_ROUTES.map(([route, label]) => (
                  <option key={route} value={route}>
                    {t(label)} · {route}
                  </option>
                ))}
              </select>
            ) : (
              <input
                id="sidebar-shortcut-destination"
                value={linkUrl}
                onChange={(event) => {
                  setLinkError(null);
                  setLinkUrl(event.target.value);
                }}
                placeholder={t('flow.sidebar.externalUrlPlaceholder')}
                aria-required="true"
                aria-invalid={linkError !== null && linkMode === 'external'}
                className="mt-1 min-h-11 w-full rounded-lg border px-3 text-[16px] font-normal outline-none"
                style={{
                  background: 'var(--flow-input-bg)',
                  borderColor:
                    linkError !== null && linkMode === 'external'
                      ? 'var(--flow-danger-fg)'
                      : 'var(--flow-border)',
                  color: 'var(--flow-text)',
                }}
              />
            )}
          </label>
          {linkError && (
            <p
              className="flex items-center gap-1.5 text-[11px]"
              style={{ color: 'var(--flow-danger-fg)' }}
              role="alert"
            >
              <AlertCircle size={14} aria-hidden />
              {linkError}
            </p>
          )}
          <div className="flex justify-end">
            <button
              type="submit"
              className="min-h-11 rounded-lg px-4 text-[12px] font-semibold"
              style={{ background: 'var(--accent)', color: 'var(--flow-on-accent)' }}
            >
              {t('flow.sidebar.addShortcut')}
            </button>
          </div>
        </form>
      )}
    </div>
  );
};

export default SidebarSettingsSection;
