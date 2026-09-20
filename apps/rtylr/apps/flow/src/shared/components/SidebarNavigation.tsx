import React, { useMemo, useState } from 'react';
import { Link, useLocation } from '@tanstack/react-router';
import {
  Archive,
  BarChart3,
  Calendar,
  ChevronDown,
  Circle,
  Clock,
  Contact,
  Eye,
  GripVertical,
  Inbox,
  Link as LinkIcon,
  LayoutGrid,
  ListTodo,
  Map as MapIcon,
  MoreHorizontal,
  Plus,
  Radar,
  Settings,
  Sun,
  Timer,
  Users,
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import type { ApiProject } from '../services/flowApi';
import { cn } from '../lib/utils';
import type {
  SidebarGroupPreference,
  SidebarIconKey,
  SidebarItemPreference,
  SidebarPreferences,
} from '../lib/sidebarPreferences';
import { updatePreferencesCopy } from '../lib/sidebarPreferences';

const ICONS: Record<SidebarIconKey, React.ComponentType<{ size?: number; className?: string }>> = {
  archive: Archive,
  'bar-chart': BarChart3,
  calendar: Calendar,
  clock: Clock,
  contact: Contact,
  eye: Eye,
  inbox: Inbox,
  link: LinkIcon,
  'layout-grid': LayoutGrid,
  'list-todo': ListTodo,
  map: MapIcon,
  milestone: Timer,
  radar: Radar,
  settings: Settings,
  sun: Sun,
  timer: Timer,
  users: Users,
};

interface SidebarNavigationProps {
  preferences: SidebarPreferences;
  update: (mutator: (draft: SidebarPreferences) => SidebarPreferences) => void;
  showPreviewRoutes: boolean;
  unreadNotifications: number;
  myIssuesCount: number;
  todayFocusCount: number;
  projects: ApiProject[];
  projectId?: string;
  myTasksProjectId?: string;
  onNewProject: () => void;
  sidebarRef: React.RefObject<HTMLDivElement | null>;
}

const isAvailable = (item: SidebarItemPreference, showPreviewRoutes: boolean): boolean => {
  if (item.id === 'builtin:roadmap' || item.id === 'builtin:timeline') return showPreviewRoutes;
  if (item.id === 'builtin:workload' || item.id === 'builtin:insights') return showPreviewRoutes;
  return true;
};

const itemLabel = (item: SidebarItemPreference, t: (key: string) => string): string =>
  item.kind === 'builtin' ? t(item.label) : item.label;

type SidebarDragState = { kind: 'item'; id: string } | { kind: 'group'; id: string };

type SidebarDropTarget =
  | { kind: 'item'; id: string; groupId?: string }
  | { kind: 'group'; id: string }
  | { kind: 'top' };

const moveItemBetweenGroups = (
  preferences: SidebarPreferences,
  itemId: string,
  destinationGroupId: string | undefined,
  beforeItemId?: string
): SidebarPreferences => {
  const next = updatePreferencesCopy(preferences);
  const item = next.items.find((candidate) => candidate.id === itemId);
  if (!item || item.lockedGroup) return next;

  const destinationGroup = destinationGroupId
    ? next.groups.find((candidate) => candidate.id === destinationGroupId)
    : undefined;
  if (destinationGroupId && (!destinationGroup || destinationGroup.lockedGroup)) return next;

  next.topLevelItemIds = next.topLevelItemIds.filter((id) => id !== itemId);
  next.groups = next.groups.map((group) => ({
    ...group,
    itemIds: group.itemIds.filter((id) => id !== itemId),
  }));

  const destinationIds = destinationGroupId
    ? next.groups.find((candidate) => candidate.id === destinationGroupId)?.itemIds
    : next.topLevelItemIds;
  if (!destinationIds) return next;

  const insertionIndex = beforeItemId ? destinationIds.indexOf(beforeItemId) : -1;
  if (insertionIndex >= 0) destinationIds.splice(insertionIndex, 0, itemId);
  else destinationIds.push(itemId);
  item.groupId = destinationGroupId;
  return next;
};

const hideItem = (preferences: SidebarPreferences, itemId: string): SidebarPreferences => {
  const next = updatePreferencesCopy(preferences);
  const item = next.items.find((candidate) => candidate.id === itemId);
  if (item && !item.lockedVisibility) item.visible = false;
  return next;
};

const toggleGroup = (preferences: SidebarPreferences, groupId: string): SidebarPreferences => {
  const next = updatePreferencesCopy(preferences);
  const group = next.groups.find((candidate) => candidate.id === groupId);
  if (group) group.collapsed = !group.collapsed;
  return next;
};

const SidebarItemActions: React.FC<{
  item: SidebarItemPreference;
  groups: SidebarGroupPreference[];
  update: SidebarNavigationProps['update'];
  onClose: () => void;
}> = ({ item, groups, update, onClose }) => {
  const { t } = useTranslation();
  return (
    <div
      className="absolute end-1 top-full z-30 mt-1 min-w-44 rounded-lg p-1 shadow-xl"
      style={{ background: 'var(--flow-card)', border: '1px solid var(--flow-border)' }}
      role="menu"
      onClick={(event) => event.stopPropagation()}
    >
      {!item.lockedVisibility && (
        <button
          type="button"
          className="flex min-h-10 w-full items-center rounded-md px-3 text-start text-meta hover:bg-[var(--flow-hover)]"
          style={{ color: 'var(--flow-text)' }}
          onClick={(event) => {
            event.preventDefault();
            update((draft) => hideItem(draft, item.id));
            onClose();
          }}
        >
          {t('flow.sidebar.hideItem')}
        </button>
      )}
      <div
        className="px-3 pb-1 pt-2 text-meta font-semibold uppercase tracking-wider"
        style={{ color: 'var(--flow-text-dim)' }}
      >
        {t('flow.sidebar.moveTo')}
      </div>
      {groups
        .filter((group) => !group.lockedGroup && group.id !== item.groupId)
        .map((group) => (
          <button
            type="button"
            key={group.id}
            className="flex min-h-10 w-full items-center rounded-md px-3 text-start text-meta hover:bg-[var(--flow-hover)]"
            style={{ color: 'var(--flow-text)' }}
            onClick={(event) => {
              event.preventDefault();
              update((draft) => moveItemBetweenGroups(draft, item.id, group.id));
              onClose();
            }}
          >
            {group.kind === 'custom' ? group.label : t(group.label)}
          </button>
        ))}
      {!item.lockedGroup && item.groupId && (
        <button
          type="button"
          className="flex min-h-10 w-full items-center rounded-md px-3 text-start text-meta hover:bg-[var(--flow-hover)]"
          style={{ color: 'var(--flow-text)' }}
          onClick={(event) => {
            event.preventDefault();
            update((draft) => moveItemBetweenGroups(draft, item.id, undefined));
            onClose();
          }}
        >
          {t('flow.sidebar.moveToTop')}
        </button>
      )}
    </div>
  );
};

const SidebarItemRow: React.FC<{
  item: SidebarItemPreference;
  label: string;
  badge?: number;
  update: SidebarNavigationProps['update'];
  groups: SidebarGroupPreference[];
  isDragging: boolean;
  isDropTarget: boolean;
  onDragStart: (itemId: string) => void;
  onDragOver: (itemId: string) => void;
  onDrop: (itemId: string) => void;
  onDragEnd: () => void;
}> = ({
  item,
  label,
  badge,
  update,
  groups,
  isDragging,
  isDropTarget,
  onDragStart,
  onDragOver,
  onDrop,
  onDragEnd,
}) => {
  const { t } = useTranslation();
  const location = useLocation();
  const [actionsOpen, setActionsOpen] = useState(false);
  const active = item.route ? location.pathname === item.route.split('?')[0] : false;
  const Icon = ICONS[item.icon] ?? Circle;
  const commonClass = cn(
    // flow-focus-inset, not the default ring: the sidebar drawer needs
    // overflow:hidden for its slide transition, which clips a ring drawn at a
    // positive offset on every row.
    'flow-focus-inset',
    'group relative flex min-h-10 items-center gap-3 rounded-lg px-3 py-2 text-meta font-medium transition-[background,box-shadow,opacity,transform] duration-150 lg:min-h-0',
    active ? '' : 'hover:bg-[var(--flow-hover)]',
    isDragging && 'scale-[0.98] opacity-45',
    isDropTarget && 'z-10'
  );
  const content = (
    <>
      {isDropTarget && (
        <span
          className="pointer-events-none absolute -top-1 start-1 end-1 h-0.5 rounded-full"
          style={{ background: 'var(--accent)', boxShadow: '0 0 0 3px var(--accent-muted)' }}
          aria-hidden
        />
      )}
      {active && (
        <span
          className="absolute inset-0 rounded-md"
          style={{ background: 'var(--flow-active)' }}
        />
      )}
      <GripVertical
        size={13}
        className="relative shrink-0"
        style={{ color: 'var(--flow-text-dim)' }}
        aria-hidden
      />
      <Icon size={14} className="relative shrink-0" />
      <span className="relative min-w-0 flex-1 truncate">{label}</span>
      {badge !== undefined && badge > 0 && (
        <span
          className="relative rounded-full px-1.5 text-meta font-semibold"
          style={{ background: 'var(--flow-hover)', color: 'var(--flow-text-dim)' }}
        >
          {badge}
        </span>
      )}
      <button
        type="button"
        aria-label={t('flow.sidebar.itemActions', { label })}
        className="flow-reveal-on-hover relative flex h-8 w-8 shrink-0 items-center justify-center rounded-md hover:bg-[var(--flow-hover)]"
        style={{ color: 'var(--flow-text-muted)' }}
        onClick={(event) => {
          event.preventDefault();
          event.stopPropagation();
          setActionsOpen((open) => !open);
        }}
      >
        <MoreHorizontal size={15} />
      </button>
      {actionsOpen && (
        <SidebarItemActions
          item={item}
          groups={groups}
          update={update}
          onClose={() => setActionsOpen(false)}
        />
      )}
    </>
  );
  const style = {
    color: active ? 'var(--flow-text)' : 'var(--flow-text-muted)',
    background: isDropTarget ? 'var(--accent-muted)' : undefined,
    boxShadow: isDropTarget ? 'inset 0 0 0 1px var(--accent-border)' : undefined,
  };
  const draggableProps = {
    draggable: true,
    onDragStart: (event: React.DragEvent) => {
      event.dataTransfer.effectAllowed = 'move';
      event.dataTransfer.setData('text/plain', item.id);
      onDragStart(item.id);
    },
    onDragOver: (event: React.DragEvent) => {
      event.preventDefault();
      event.stopPropagation();
      event.dataTransfer.dropEffect = 'move';
      onDragOver(item.id);
    },
    onDrop: (event: React.DragEvent) => {
      event.preventDefault();
      event.stopPropagation();
      onDrop(item.id);
    },
    onDragEnd,
  };
  if (item.url) {
    return (
      <a
        {...draggableProps}
        href={item.url}
        target={item.openInNewTab ? '_blank' : undefined}
        rel={item.openInNewTab ? 'noreferrer' : undefined}
        className={commonClass}
        style={style}
      >
        {content}
      </a>
    );
  }
  return (
    <Link {...draggableProps} to={item.route ?? '/'} className={commonClass} style={style}>
      {content}
    </Link>
  );
};

const SidebarGroup: React.FC<{
  group: SidebarGroupPreference;
  preferences: SidebarPreferences;
  update: SidebarNavigationProps['update'];
  showPreviewRoutes: boolean;
  projects: ApiProject[];
  projectId?: string;
  myTasksProjectId?: string;
  onNewProject: () => void;
  dragState?: SidebarDragState;
  dropTarget?: SidebarDropTarget;
  onItemDragStart: (itemId: string) => void;
  onItemDragOver: (groupId: string, itemId?: string) => void;
  onItemDrop: (groupId: string, itemId?: string) => void;
  onGroupDragStart: (groupId: string) => void;
  onGroupDragOver: (groupId: string) => void;
  onGroupDrop: (groupId: string) => void;
  onDragEnd: () => void;
}> = ({
  group,
  preferences,
  update,
  showPreviewRoutes,
  projects,
  projectId,
  myTasksProjectId,
  onNewProject,
  dragState,
  dropTarget,
  onItemDragStart,
  onItemDragOver,
  onItemDrop,
  onGroupDragStart,
  onGroupDragOver,
  onGroupDrop,
  onDragEnd,
}) => {
  const { t } = useTranslation();
  const [groupActionsOpen, setGroupActionsOpen] = useState(false);
  const itemsById = useMemo(
    () =>
      new Map<string, SidebarItemPreference>(
        preferences.items.map((item): [string, SidebarItemPreference] => [item.id, item])
      ),
    [preferences.items]
  );
  const groupItems = group.itemIds
    .map((id) => itemsById.get(id))
    .filter((item): item is SidebarItemPreference =>
      Boolean(item && item.visible && isAvailable(item, showPreviewRoutes))
    );
  const visibleProjects = projects.filter(
    (project) =>
      !project.archivedAt &&
      project.id !== myTasksProjectId &&
      !(project.name === 'My Tasks' && project.icon === '✅') &&
      !(project.name === 'Inbox' && project.icon === '📥')
  );
  const groupLabel = group.kind === 'custom' ? group.label : t(group.label);
  const isGroupDragging = dragState?.kind === 'group' && dragState.id === group.id;
  const isGroupDropTarget = dropTarget?.kind === 'group' && dropTarget.id === group.id;
  const isItemDropTarget = dropTarget?.kind === 'item' && dropTarget.groupId === group.id;
  const groupStyle = {
    background: isGroupDropTarget ? 'var(--accent-muted)' : undefined,
    opacity: isGroupDragging ? 0.5 : 1,
    transform: isGroupDragging ? 'scale(0.98)' : undefined,
    transition: 'background 160ms ease, opacity 160ms ease, transform 160ms ease',
  };
  const handleGroupDragOver = (event: React.DragEvent): void => {
    if (!dragState) return;
    event.preventDefault();
    event.dataTransfer.dropEffect = 'move';
    if (dragState.kind === 'group') onGroupDragOver(group.id);
    else onItemDragOver(group.id);
  };
  const handleGroupDrop = (event: React.DragEvent): void => {
    event.preventDefault();
    event.stopPropagation();
    onGroupDrop(group.id);
  };
  return (
    <section
      className={cn(
        'relative flex flex-col gap-1',
        isGroupDropTarget && 'rounded-md shadow-[0_0_0_3px_var(--accent-muted)]'
      )}
      style={groupStyle}
      aria-label={groupLabel}
    >
      <div
        draggable
        onDragStart={(event) => {
          event.dataTransfer.effectAllowed = 'move';
          event.dataTransfer.setData('text/plain', group.id);
          onGroupDragStart(group.id);
        }}
        onDragOver={handleGroupDragOver}
        onDrop={handleGroupDrop}
        onDragEnd={onDragEnd}
        className="group flex min-h-10 items-center gap-1"
      >
        <GripVertical
          size={15}
          className="shrink-0"
          style={{ color: isGroupDropTarget ? 'var(--accent)' : 'var(--flow-text-dim)' }}
          aria-label={t('flow.sidebar.dragHandle')}
        />
        <button
          type="button"
          onClick={() => update((draft) => toggleGroup(draft, group.id))}
          aria-expanded={!group.collapsed}
          className="flex min-h-10 min-w-0 flex-1 items-center gap-3 rounded-lg px-3 py-2 text-start hover:bg-[var(--flow-hover)] lg:min-h-0"
        >
          <span
            className="min-w-0 flex-1 truncate text-meta font-semibold leading-tight"
            style={{ color: 'var(--flow-text)' }}
          >
            {groupLabel}
          </span>
          <ChevronDown
            size={14}
            className={cn('shrink-0 transition-transform', group.collapsed && '-rotate-90')}
            style={{ color: 'var(--flow-text-muted)' }}
          />
        </button>
        {group.id === 'projects' && (
          <button
            type="button"
            title={t('flow.shell.newProject')}
            aria-label={t('flow.shell.newProject')}
            onClick={onNewProject}
            className="flex h-8 w-8 shrink-0 items-center justify-center rounded-md hover:bg-[var(--flow-hover)]"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            <Plus size={15} />
          </button>
        )}
        <button
          type="button"
          aria-label={t('flow.sidebar.groupActions', {
            label: group.kind === 'custom' ? group.label : t(group.label),
          })}
          onClick={() => setGroupActionsOpen((open) => !open)}
          className="flow-reveal-on-hover flex h-8 w-8 shrink-0 items-center justify-center rounded-md hover:bg-[var(--flow-hover)]"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          <MoreHorizontal size={15} />
        </button>
        {groupActionsOpen && (
          <div
            className="absolute end-3 top-12 z-30 min-w-44 rounded-lg p-1 shadow-xl"
            style={{ background: 'var(--flow-card)', border: '1px solid var(--flow-border)' }}
          >
            {!group.lockedVisibility && (
              <button
                type="button"
                className="flex min-h-10 w-full items-center rounded-md px-3 text-start text-meta hover:bg-[var(--flow-hover)]"
                style={{ color: 'var(--flow-text)' }}
                onClick={() => {
                  update((draft) => {
                    const next = updatePreferencesCopy(draft);
                    const candidate = next.groups.find((entry) => entry.id === group.id);
                    if (candidate) candidate.visible = false;
                    return next;
                  });
                  setGroupActionsOpen(false);
                }}
              >
                {t('flow.sidebar.hideGroup')}
              </button>
            )}
          </div>
        )}
      </div>
      {!group.collapsed && group.id === 'projects' && (
        <div
          className={cn(
            'ms-5 flex flex-col gap-px border-s ps-2 pt-px',
            isGroupDropTarget && 'rounded-md border-dashed'
          )}
          style={{ borderInlineStartColor: 'var(--flow-border)' }}
          onDragOver={handleGroupDragOver}
          onDrop={(event) => {
            event.preventDefault();
            event.stopPropagation();
            onGroupDrop(group.id);
          }}
        >
          {visibleProjects.map((project) => {
            const active = projectId === project.id;
            return (
              <Link
                key={project.id}
                to="/p/$projectId"
                params={{ projectId: project.id }}
                className={cn(
                  'relative flex min-h-10 items-center gap-2 rounded-lg px-3 py-2 text-meta lg:min-h-0',
                  active ? '' : 'hover:bg-[var(--flow-hover)]'
                )}
                style={{ color: active ? 'var(--flow-text)' : 'var(--flow-text-muted)' }}
              >
                <Circle
                  size={8}
                  className="relative shrink-0 fill-current"
                  style={{ color: project.color }}
                />
                <span className="relative min-w-0 flex-1 truncate font-medium">{project.name}</span>
                {(project.taskCount ?? 0) - (project.completedCount ?? 0) > 0 && (
                  <span
                    // Hidden at zero, and tabular so the column does not shift
                    // as digits change. A finished project showed a bare "0",
                    // which reads as a count worth noting rather than as
                    // nothing left to do (#1097, D-058 and G-T5).
                    className="relative text-meta tabular-nums"
                    style={{ color: 'var(--flow-text-dim)' }}
                  >
                    {(project.taskCount ?? 0) - (project.completedCount ?? 0)}
                  </span>
                )}
              </Link>
            );
          })}
          {groupItems.map((item) => (
            <SidebarItemRow
              key={item.id}
              item={item}
              label={itemLabel(item, t)}
              groups={preferences.groups}
              update={update}
              isDragging={dragState?.kind === 'item' && dragState.id === item.id}
              isDropTarget={
                isItemDropTarget && dropTarget?.kind === 'item' && dropTarget.id === item.id
              }
              onDragStart={onItemDragStart}
              onDragOver={(itemId) => onItemDragOver(group.id, itemId)}
              onDrop={(itemId) => onItemDrop(group.id, itemId)}
              onDragEnd={onDragEnd}
            />
          ))}
          {isGroupDropTarget && (
            <div
              className="mx-2 my-1 flex min-h-8 items-center justify-center rounded-md border border-dashed text-meta font-semibold"
              style={{ borderColor: 'var(--accent-border)', color: 'var(--accent-fg)' }}
            >
              {t('flow.sidebar.dropHere')}
            </div>
          )}
        </div>
      )}
      {!group.collapsed && group.id !== 'projects' && (
        <nav
          className={cn(
            'ms-7 flex flex-col gap-px border-s ps-1 pt-px',
            isGroupDropTarget && 'rounded-md border-dashed'
          )}
          style={{ borderInlineStartColor: 'var(--flow-border)' }}
          onDragOver={handleGroupDragOver}
          onDrop={(event) => {
            event.preventDefault();
            event.stopPropagation();
            onGroupDrop(group.id);
          }}
        >
          {groupItems.map((item) => (
            <SidebarItemRow
              key={item.id}
              item={item}
              label={itemLabel(item, t)}
              groups={preferences.groups}
              update={update}
              isDragging={dragState?.kind === 'item' && dragState.id === item.id}
              isDropTarget={
                isItemDropTarget && dropTarget?.kind === 'item' && dropTarget.id === item.id
              }
              onDragStart={onItemDragStart}
              onDragOver={(itemId) => onItemDragOver(group.id, itemId)}
              onDrop={(itemId) => onItemDrop(group.id, itemId)}
              onDragEnd={onDragEnd}
            />
          ))}
          {isGroupDropTarget && (
            <div
              className="mx-2 my-1 flex min-h-8 items-center justify-center rounded-md border border-dashed text-meta font-semibold"
              style={{ borderColor: 'var(--accent-border)', color: 'var(--accent-fg)' }}
            >
              {t('flow.sidebar.dropHere')}
            </div>
          )}
        </nav>
      )}
    </section>
  );
};

export const SidebarNavigation: React.FC<SidebarNavigationProps> = ({
  preferences,
  update,
  showPreviewRoutes,
  unreadNotifications,
  myIssuesCount,
  todayFocusCount,
  projects,
  projectId,
  myTasksProjectId,
  onNewProject,
  sidebarRef,
}) => {
  const { t } = useTranslation();
  const [dragState, setDragState] = useState<SidebarDragState>();
  const [dropTarget, setDropTarget] = useState<SidebarDropTarget>();
  const topItems = preferences.topLevelItemIds
    .map((id) => preferences.items.find((item) => item.id === id))
    .filter((item): item is SidebarItemPreference =>
      Boolean(item && item.visible && isAvailable(item, showPreviewRoutes))
    );
  const badges: Record<string, number> = {
    'builtin:inbox': unreadNotifications,
    'builtin:my-issues': myIssuesCount,
    'builtin:today': todayFocusCount,
  };
  const groupsById = new Map<string, SidebarGroupPreference>(
    preferences.groups.map((group): [string, SidebarGroupPreference] => [group.id, group])
  );
  const groups = preferences.groupOrder
    .map((id) => groupsById.get(id))
    .filter((group): group is SidebarGroupPreference => Boolean(group && group.visible));
  const clearDrag = (): void => {
    setDragState(undefined);
    setDropTarget(undefined);
  };
  const onItemDragStart = (itemId: string): void => {
    setDragState({ kind: 'item', id: itemId });
    setDropTarget(undefined);
  };
  const onItemDragOver = (groupId: string, itemId?: string): void => {
    if (dragState?.kind !== 'item' || dragState.id === itemId) return;
    setDropTarget(itemId ? { kind: 'item', id: itemId, groupId } : { kind: 'group', id: groupId });
  };
  const onItemDrop = (groupId: string, itemId?: string): void => {
    if (dragState?.kind !== 'item' || dragState.id === itemId) {
      clearDrag();
      return;
    }
    update((draft) => {
      return moveItemBetweenGroups(draft, dragState.id, groupId, itemId);
    });
    clearDrag();
  };
  const onTopItemDragOver = (itemId?: string): void => {
    if (dragState?.kind !== 'item' || dragState.id === itemId) return;
    setDropTarget(itemId ? { kind: 'item', id: itemId } : { kind: 'top' });
  };
  const onTopItemDrop = (itemId?: string): void => {
    if (dragState?.kind !== 'item' || dragState.id === itemId) {
      clearDrag();
      return;
    }
    update((draft) => moveItemBetweenGroups(draft, dragState.id, undefined, itemId));
    clearDrag();
  };
  const onGroupDragStart = (groupId: string): void => {
    setDragState({ kind: 'group', id: groupId });
    setDropTarget(undefined);
  };
  const onGroupDragOver = (groupId: string): void => {
    if (dragState?.kind !== 'group' || dragState.id === groupId) return;
    setDropTarget({ kind: 'group', id: groupId });
  };
  const onGroupDrop = (groupId: string): void => {
    if (!dragState || dragState.id === groupId) {
      clearDrag();
      return;
    }
    if (dragState.kind === 'item') {
      update((draft) => moveItemBetweenGroups(draft, dragState.id, groupId));
      clearDrag();
      return;
    }
    update((draft) => {
      const next = updatePreferencesCopy(draft);
      const from = next.groupOrder.indexOf(dragState.id);
      const to = next.groupOrder.indexOf(groupId);
      if (from >= 0 && to >= 0) {
        next.groupOrder.splice(from, 1);
        next.groupOrder.splice(to, 0, dragState.id);
      }
      return next;
    });
    clearDrag();
  };

  return (
    <div ref={sidebarRef} className="flex min-h-0 flex-1 flex-col gap-2 overflow-y-auto px-2 py-2">
      <nav
        className={cn(
          'flex flex-col gap-px rounded-lg transition-[background,box-shadow] duration-150',
          dropTarget?.kind === 'top' &&
            'bg-[var(--accent-muted)] shadow-[inset_0_0_0_1px_var(--accent-border)]'
        )}
        onDragOver={(event) => {
          if (dragState?.kind !== 'item') return;
          event.preventDefault();
          event.dataTransfer.dropEffect = 'move';
          onTopItemDragOver();
        }}
        onDrop={(event) => {
          event.preventDefault();
          onTopItemDrop();
        }}
      >
        {topItems.map((item) => (
          <SidebarItemRow
            key={item.id}
            item={item}
            label={itemLabel(item, t)}
            badge={badges[item.id]}
            groups={preferences.groups}
            update={update}
            isDragging={dragState?.kind === 'item' && dragState.id === item.id}
            isDropTarget={dropTarget?.kind === 'item' && dropTarget.id === item.id}
            onDragStart={onItemDragStart}
            onDragOver={onTopItemDragOver}
            onDrop={onTopItemDrop}
            onDragEnd={clearDrag}
          />
        ))}
      </nav>
      <div className="flex flex-col gap-2">
        {groups.map((group) => (
          <SidebarGroup
            key={group.id}
            group={group}
            preferences={preferences}
            update={update}
            showPreviewRoutes={showPreviewRoutes}
            projects={projects}
            projectId={projectId}
            myTasksProjectId={myTasksProjectId}
            onNewProject={onNewProject}
            dragState={dragState}
            dropTarget={dropTarget}
            onItemDragStart={onItemDragStart}
            onItemDragOver={onItemDragOver}
            onItemDrop={onItemDrop}
            onGroupDragStart={onGroupDragStart}
            onGroupDragOver={onGroupDragOver}
            onGroupDrop={onGroupDrop}
            onDragEnd={clearDrag}
          />
        ))}
      </div>
    </div>
  );
};
