import React, { useEffect, useMemo, useRef, useState } from 'react';
import { motion, useReducedMotion } from 'motion/react';
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
import { createDefaultSidebarPreferences, updatePreferencesCopy } from '../lib/sidebarPreferences';
import {
  isWithinSameList,
  placementOver,
  reorderGroupIds,
  samePlacement,
  type ItemPlacement,
} from '../lib/sidebarDrag';

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
  myIssuesCount?: number;
  todayFocusCount?: number;
  projects: ApiProject[];
  projectId?: string;
  myTasksProjectId?: string;
  onNewProject: () => void;
  canCreateProject?: boolean;
  sidebarRef: React.RefObject<HTMLDivElement | null>;
}

const isAvailable = (item: SidebarItemPreference, showPreviewRoutes: boolean): boolean => {
  if (item.id === 'builtin:roadmap' || item.id === 'builtin:timeline') return showPreviewRoutes;
  if (item.id === 'builtin:workload' || item.id === 'builtin:insights') return showPreviewRoutes;
  return true;
};

const builtinLabelsByRoute = new Map(
  createDefaultSidebarPreferences().items.map((item) => [item.route, item.label])
);

const itemLabel = (item: SidebarItemPreference, t: (key: string) => string): string => {
  if (item.kind === 'builtin') return t(item.label);
  if (item.kind === 'recent' && item.route) {
    const label = builtinLabelsByRoute.get(item.route);
    if (label) return t(label);
  }
  return item.label;
};

type SidebarDragState = { kind: 'item'; id: string } | { kind: 'group'; id: string };

type SidebarDropTarget =
  { kind: 'item'; id: string; groupId?: string } | { kind: 'group'; id: string } | { kind: 'top' };

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
          className="flex min-h-10 w-full items-center rounded-md px-3 text-start text-flow-meta hover:bg-[var(--flow-hover)]"
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
        className="px-3 pb-1 pt-2 text-flow-meta font-semibold uppercase tracking-wider"
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
            className="flex min-h-10 w-full items-center rounded-md px-3 text-start text-flow-meta hover:bg-[var(--flow-hover)]"
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
          className="flex min-h-10 w-full items-center rounded-md px-3 text-start text-flow-meta hover:bg-[var(--flow-hover)]"
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
  /** `after` is true when the pointer is in the lower half of the row. */
  onDragOver: (itemId: string, after: boolean) => void;
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
  const content = (
    <>
      <Icon size={17} aria-hidden />
      <span className="min-w-0 flex-1 truncate">{label}</span>
      {badge !== undefined && badge > 0 && (
        <span className="text-flow-meta tabular-nums">{badge}</span>
      )}
    </>
  );
  const linkProps = {
    className: cn('portal-nav-link flow-nav-row-link flow-focus-inset', active && 'is-active'),
    'aria-current': active ? ('page' as const) : undefined,
  };
  const reduceMotion = useReducedMotion();
  return (
    // The wrapper animates the row to its new place when a drag reorders the list.
    // It is separate from the draggable element because motion components take
    // over the `onDrag*` props for their own drag gesture, and native HTML5
    // drag events must reach the inner element.
    <motion.div layout="position" transition={{ duration: reduceMotion ? 0 : 0.16 }}>
      <div
        className={cn(
          'group relative rounded-lg',
          isDragging && 'opacity-45',
          isDropTarget && 'ring-1 ring-[var(--flow-border-focus)]'
        )}
        draggable
        onDragStart={(event) => {
          event.dataTransfer.effectAllowed = 'move';
          event.dataTransfer.setData('text/plain', item.id);
          onDragStart(item.id);
        }}
        onDragOver={(event) => {
          event.preventDefault();
          event.stopPropagation();
          event.dataTransfer.dropEffect = 'move';
          const box = event.currentTarget.getBoundingClientRect();
          onDragOver(item.id, event.clientY > box.top + box.height / 2);
        }}
        onDrop={(event) => {
          event.preventDefault();
          event.stopPropagation();
          onDrop(item.id);
        }}
        onDragEnd={onDragEnd}
      >
        {item.url ? (
          <a
            {...linkProps}
            href={item.url}
            target={item.openInNewTab ? '_blank' : undefined}
            rel={item.openInNewTab ? 'noreferrer' : undefined}
          >
            {content}
          </a>
        ) : (
          <Link {...linkProps} to={item.route ?? '/flow'}>
            {content}
          </Link>
        )}
        <button
          type="button"
          aria-label={t('flow.sidebar.itemActions', { label })}
          className="flow-reveal-on-hover absolute end-1 top-1 flex h-7 w-7 items-center justify-center rounded-md hover:bg-[var(--flow-hover)]"
          style={{ color: 'var(--flow-text-muted)' }}
          onClick={() => setActionsOpen((open) => !open)}
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
      </div>
    </motion.div>
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
  canCreateProject?: boolean;
  dragState?: SidebarDragState;
  dropTarget?: SidebarDropTarget;
  onItemDragStart: (itemId: string) => void;
  onItemDragOver: (groupId: string, itemId?: string, after?: boolean) => void;
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
  canCreateProject = true,
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
  const reduceMotion = useReducedMotion();
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
  if (group.id !== 'projects' && groupItems.length === 0) return null;
  const groupLabel = group.kind === 'custom' ? group.label : t(group.label);
  const isGroupDragging = dragState?.kind === 'group' && dragState.id === group.id;
  const isGroupDropTarget = dropTarget?.kind === 'group' && dropTarget.id === group.id;
  const isItemDropTarget = dropTarget?.kind === 'item' && dropTarget.groupId === group.id;
  // No scale transform on a dragged group: the section is moved by a layout
  // animation, which owns `transform`.
  const groupStyle = {
    background: isGroupDropTarget ? 'var(--accent-muted)' : undefined,
    opacity: isGroupDragging ? 0.5 : 1,
    transition: 'background 160ms ease, opacity 160ms ease',
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
    <motion.section
      layout="position"
      transition={{ duration: reduceMotion ? 0 : 0.16 }}
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
        className="group relative flex min-h-8 items-center gap-0.5 rounded-md pe-1 transition-colors hover:bg-[var(--flow-hover)] focus-within:bg-[var(--flow-hover)]"
      >
        <button
          type="button"
          onClick={() => update((draft) => toggleGroup(draft, group.id))}
          aria-expanded={!group.collapsed}
          className="flex min-h-8 min-w-0 flex-1 items-center gap-3 rounded-md px-3 py-1.5 text-start"
        >
          <span
            className="min-w-0 flex-1 truncate text-flow-meta font-semibold leading-tight"
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
        <GripVertical
          size={14}
          className="flow-reveal-on-hover shrink-0 cursor-grab active:cursor-grabbing"
          style={{ color: isGroupDropTarget ? 'var(--accent)' : 'var(--flow-text-dim)' }}
          aria-label={t('flow.sidebar.dragHandle')}
        />
        {group.id === 'projects' && canCreateProject && (
          <button
            type="button"
            title={t('flow.shell.newProject')}
            aria-label={t('flow.shell.newProject')}
            onClick={onNewProject}
            className="flex h-6 w-6 shrink-0 items-center justify-center rounded-md hover:bg-[var(--flow-hover)]"
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
          aria-expanded={groupActionsOpen}
          onClick={() => setGroupActionsOpen((open) => !open)}
          className="flow-reveal-on-hover flex h-6 w-6 shrink-0 items-center justify-center rounded-md hover:bg-[var(--flow-hover)]"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          <MoreHorizontal size={15} />
        </button>
        {groupActionsOpen && (
          <div
            className="absolute end-0 top-full z-30 mt-1 min-w-44 rounded-lg p-1 shadow-xl"
            style={{ background: 'var(--flow-card)', border: '1px solid var(--flow-border)' }}
          >
            {!group.lockedVisibility && (
              <button
                type="button"
                className="flex min-h-10 w-full items-center rounded-md px-3 text-start text-flow-meta hover:bg-[var(--flow-hover)]"
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
                to="/flow/p/$projectId"
                params={{ projectId: project.id }}
                className={cn(
                  'relative flex items-center gap-2 rounded-lg px-3 py-1.5 text-flow-meta',
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
                    className="relative text-flow-meta tabular-nums"
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
              onDragOver={(itemId, after) => onItemDragOver(group.id, itemId, after)}
              onDrop={(itemId) => onItemDrop(group.id, itemId)}
              onDragEnd={onDragEnd}
            />
          ))}
          {isGroupDropTarget && (
            <div
              className="mx-2 my-1 flex min-h-8 items-center justify-center rounded-md border border-dashed text-flow-meta font-semibold"
              style={{ borderColor: 'var(--accent-border)', color: 'var(--flow-text)' }}
            >
              {t('flow.sidebar.dropHere')}
            </div>
          )}
        </div>
      )}
      {!group.collapsed && group.id !== 'projects' && (
        <nav
          className={cn(
            'flex flex-col gap-px pt-1',
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
              onDragOver={(itemId, after) => onItemDragOver(group.id, itemId, after)}
              onDrop={(itemId) => onItemDrop(group.id, itemId)}
              onDragEnd={onDragEnd}
            />
          ))}
          {isGroupDropTarget && (
            <div
              className="mx-2 my-1 flex min-h-8 items-center justify-center rounded-md border border-dashed text-flow-meta font-semibold"
              style={{ borderColor: 'var(--accent-border)', color: 'var(--flow-text)' }}
            >
              {t('flow.sidebar.dropHere')}
            </div>
          )}
        </nav>
      )}
    </motion.section>
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
  canCreateProject = true,
  sidebarRef,
}) => {
  const { t } = useTranslation();
  const [dragState, setDragState] = useState<SidebarDragState>();
  const [dropTarget, setDropTarget] = useState<SidebarDropTarget>();
  // Where the dragged row would land (an item drag) or which group it would
  // take the place of (a group drag). Both are previewed live, see `shown`.
  const [placement, setPlacement] = useState<ItemPlacement>();
  const [groupTargetId, setGroupTargetId] = useState<string>();
  // A row that has just slid is still animating under the pointer; reading the
  // pointer against it straight away would flip the order back and forth.
  const settledAt = useRef(0);

  // What is drawn: the stored preferences with the drag applied, so rows and
  // groups sit in their new place (and slide there) while the pointer is still
  // down. Nothing is saved until the drop.
  const shown = useMemo((): SidebarPreferences => {
    if (dragState?.kind === 'item' && placement) {
      const source = preferences.items.find((item) => item.id === dragState.id)?.groupId;
      // A row that changes list is unmounted and mounted again by React, and a
      // browser drag does not survive its source node being replaced, so only
      // moves within the same list are previewed. The rest keep the drop ring.
      if (!isWithinSameList(source, placement)) return preferences;
      return moveItemBetweenGroups(
        preferences,
        dragState.id,
        placement.groupId,
        placement.beforeItemId
      );
    }
    if (dragState?.kind === 'group' && groupTargetId) {
      const next = updatePreferencesCopy(preferences);
      next.groupOrder = reorderGroupIds(next.groupOrder, dragState.id, groupTargetId);
      return next;
    }
    return preferences;
  }, [preferences, dragState, placement, groupTargetId]);
  const previewing = shown !== preferences;
  const topItems = shown.topLevelItemIds
    .map((id) => shown.items.find((item) => item.id === id))
    .filter((item): item is SidebarItemPreference =>
      Boolean(item && item.visible && isAvailable(item, showPreviewRoutes))
    );
  const badges: Record<string, number | undefined> = {
    'builtin:inbox': unreadNotifications,
    'builtin:my-issues': myIssuesCount,
    'builtin:today': todayFocusCount,
  };
  const groupsById = new Map<string, SidebarGroupPreference>(
    shown.groups.map((group): [string, SidebarGroupPreference] => [group.id, group])
  );
  const groups = shown.groupOrder
    .map((id) => groupsById.get(id))
    .filter((group): group is SidebarGroupPreference => Boolean(group && group.visible));
  const clearDrag = (): void => {
    setDragState(undefined);
    setDropTarget(undefined);
    setPlacement(undefined);
    setGroupTargetId(undefined);
  };

  // A drag that ends where the browser sends no `dragend` (the source row was
  // replaced) would otherwise leave the preview stuck on screen. Pointer events
  // are suppressed for the length of a native drag and resume when it is over.
  useEffect(() => {
    if (!dragState) return undefined;
    const end = (): void => clearDrag();
    window.addEventListener('dragend', end);
    window.addEventListener('pointermove', end);
    return () => {
      window.removeEventListener('dragend', end);
      window.removeEventListener('pointermove', end);
    };
  }, [dragState]);

  const commitDrag = (): void => {
    if (dragState?.kind === 'item' && placement) {
      update((draft) =>
        moveItemBetweenGroups(draft, dragState.id, placement.groupId, placement.beforeItemId)
      );
    } else if (dragState?.kind === 'group' && groupTargetId) {
      update((draft) => {
        const next = updatePreferencesCopy(draft);
        next.groupOrder = reorderGroupIds(next.groupOrder, dragState.id, groupTargetId);
        return next;
      });
    }
    clearDrag();
  };
  const place = (next: ItemPlacement): void => {
    if (samePlacement(placement, next)) return;
    if (performance.now() < settledAt.current) return;
    settledAt.current = performance.now() + 140;
    setPlacement(next);
  };
  const onItemDragStart = (itemId: string): void => {
    setDragState({ kind: 'item', id: itemId });
    setDropTarget(undefined);
  };
  const onItemDragOver = (groupId: string, itemId?: string, after = false): void => {
    if (dragState?.kind !== 'item' || dragState.id === itemId) return;
    if (itemId) {
      const group = preferences.groups.find((candidate) => candidate.id === groupId);
      place(placementOver(group?.itemIds ?? [], dragState.id, itemId, after, groupId));
      setDropTarget({ kind: 'item', id: itemId, groupId });
    } else {
      place({ groupId });
      setDropTarget({ kind: 'group', id: groupId });
    }
  };
  const onItemDrop = (_groupId: string, _itemId?: string): void => commitDrag();
  const onTopItemDragOver = (itemId?: string, after = false): void => {
    if (dragState?.kind !== 'item' || dragState.id === itemId) return;
    if (itemId) {
      place(placementOver(preferences.topLevelItemIds, dragState.id, itemId, after));
      setDropTarget({ kind: 'item', id: itemId });
    } else {
      place({});
      setDropTarget({ kind: 'top' });
    }
  };
  const onTopItemDrop = (_itemId?: string): void => commitDrag();
  const onGroupDragStart = (groupId: string): void => {
    setDragState({ kind: 'group', id: groupId });
    setDropTarget(undefined);
  };
  const onGroupDragOver = (groupId: string): void => {
    if (dragState?.kind !== 'group' || dragState.id === groupId) return;
    if (performance.now() < settledAt.current) return;
    settledAt.current = performance.now() + 140;
    setGroupTargetId(groupId);
    setDropTarget({ kind: 'group', id: groupId });
  };
  const onGroupDrop = (groupId: string): void => {
    // An item dropped on a group header has no row to sit before: it joins the end.
    if (dragState?.kind === 'item' && !placement) {
      update((draft) => moveItemBetweenGroups(draft, dragState.id, groupId));
      clearDrag();
      return;
    }
    commitDrag();
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
            isDropTarget={!previewing && dropTarget?.kind === 'item' && dropTarget.id === item.id}
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
            preferences={shown}
            update={update}
            showPreviewRoutes={showPreviewRoutes}
            projects={projects}
            projectId={projectId}
            myTasksProjectId={myTasksProjectId}
            onNewProject={onNewProject}
            canCreateProject={canCreateProject}
            dragState={dragState}
            dropTarget={previewing && dropTarget?.kind === 'item' ? undefined : dropTarget}
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
