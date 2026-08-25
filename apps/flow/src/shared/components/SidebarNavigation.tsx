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
  onOpenSettings: () => void;
  sidebarRef: React.RefObject<HTMLDivElement | null>;
}

const isAvailable = (item: SidebarItemPreference, showPreviewRoutes: boolean): boolean => {
  if (item.id === 'builtin:roadmap' || item.id === 'builtin:timeline') return showPreviewRoutes;
  if (item.id === 'builtin:workload' || item.id === 'builtin:insights') return showPreviewRoutes;
  return true;
};

const itemLabel = (item: SidebarItemPreference, t: (key: string) => string): string =>
  item.kind === 'builtin' ? t(item.label) : item.label;

const moveItemBetweenGroups = (
  preferences: SidebarPreferences,
  itemId: string,
  destinationGroupId: string | undefined
): SidebarPreferences => {
  const next = updatePreferencesCopy(preferences);
  next.topLevelItemIds = next.topLevelItemIds.filter((id) => id !== itemId);
  next.groups = next.groups.map((group) => ({
    ...group,
    itemIds: group.itemIds.filter((id) => id !== itemId),
  }));
  if (destinationGroupId) {
    const group = next.groups.find((candidate) => candidate.id === destinationGroupId);
    if (group && !group.lockedGroup) group.itemIds.push(itemId);
  } else {
    next.topLevelItemIds.push(itemId);
  }
  const item = next.items.find((candidate) => candidate.id === itemId);
  if (item && !item.lockedGroup) item.groupId = destinationGroupId;
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
    >
      {!item.lockedVisibility && (
        <button
          type="button"
          className="flex min-h-10 w-full items-center rounded-md px-3 text-start text-[12px] hover:bg-[var(--flow-hover)]"
          style={{ color: 'var(--flow-text)' }}
          onClick={() => {
            update((draft) => hideItem(draft, item.id));
            onClose();
          }}
        >
          {t('flow.sidebar.hideItem')}
        </button>
      )}
      <div
        className="px-3 pb-1 pt-2 text-[10px] font-semibold uppercase tracking-wider"
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
            className="flex min-h-10 w-full items-center rounded-md px-3 text-start text-[12px] hover:bg-[var(--flow-hover)]"
            style={{ color: 'var(--flow-text)' }}
            onClick={() => {
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
          className="flex min-h-10 w-full items-center rounded-md px-3 text-start text-[12px] hover:bg-[var(--flow-hover)]"
          style={{ color: 'var(--flow-text)' }}
          onClick={() => {
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
  onDragStart: (itemId: string) => void;
  onDrop: (itemId: string) => void;
}> = ({ item, label, badge, update, groups, onDragStart, onDrop }) => {
  const { t } = useTranslation();
  const location = useLocation();
  const [actionsOpen, setActionsOpen] = useState(false);
  const active = item.route ? location.pathname === item.route.split('?')[0] : false;
  const Icon = ICONS[item.icon] ?? Circle;
  const commonClass = cn(
    'group relative flex min-h-10 items-center gap-2 rounded-md px-2 text-[12px] font-medium transition-colors lg:min-h-0 lg:py-1.5',
    active ? '' : 'hover:bg-[var(--flow-hover)]'
  );
  const content = (
    <>
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
          className="relative rounded-full px-1.5 text-[10px] font-semibold"
          style={{ background: 'var(--flow-hover)', color: 'var(--flow-text-dim)' }}
        >
          {badge}
        </span>
      )}
      <button
        type="button"
        aria-label={t('flow.sidebar.itemActions', { label })}
        className="relative flex h-8 w-8 shrink-0 items-center justify-center rounded-md hover:bg-[var(--flow-hover)]"
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
  const style = { color: active ? 'var(--flow-text)' : 'var(--flow-text-muted)' };
  const draggableProps = {
    draggable: true,
    onDragStart: () => onDragStart(item.id),
    onDragOver: (event: React.DragEvent) => event.preventDefault(),
    onDrop: () => onDrop(item.id),
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
  onGroupDragStart: (groupId: string) => void;
  onGroupDrop: (groupId: string) => void;
}> = ({
  group,
  preferences,
  update,
  showPreviewRoutes,
  projects,
  projectId,
  myTasksProjectId,
  onNewProject,
  onGroupDragStart,
  onGroupDrop,
}) => {
  const { t } = useTranslation();
  const [groupActionsOpen, setGroupActionsOpen] = useState(false);
  const [draggedItemId, setDraggedItemId] = useState<string>();
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
  const onDrop = (itemId: string): void => {
    if (!draggedItemId || draggedItemId === itemId) return;
    update((draft) => {
      const next = updatePreferencesCopy(draft);
      const ids = next.groups.find((candidate) => candidate.id === group.id)?.itemIds ?? [];
      const from = ids.indexOf(draggedItemId);
      const to = ids.indexOf(itemId);
      if (from >= 0 && to >= 0) {
        ids.splice(from, 1);
        ids.splice(to, 0, draggedItemId);
      }
      return next;
    });
    setDraggedItemId(undefined);
  };
  return (
    <div className="relative flex flex-col gap-0">
      <div
        draggable
        onDragStart={() => onGroupDragStart(group.id)}
        onDragOver={(event) => event.preventDefault()}
        onDrop={() => onGroupDrop(group.id)}
        className="group flex min-h-10 items-center gap-1"
      >
        <GripVertical
          size={14}
          style={{ color: 'var(--flow-text-dim)' }}
          aria-label={t('flow.sidebar.dragHandle')}
        />
        <button
          type="button"
          onClick={() => update((draft) => toggleGroup(draft, group.id))}
          className="flex min-h-10 min-w-0 flex-1 items-center justify-between gap-2 rounded-md px-2 text-start hover:bg-[var(--flow-hover)]"
        >
          <span
            className="truncate text-[12px] font-semibold tracking-wide"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            {group.kind === 'custom' ? group.label : t(group.label)}
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
          className="flex h-8 w-8 shrink-0 items-center justify-center rounded-md hover:bg-[var(--flow-hover)]"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          <MoreHorizontal size={15} />
        </button>
        {groupActionsOpen && (
          <div
            className="absolute end-3 z-30 mt-20 min-w-44 rounded-lg p-1 shadow-xl"
            style={{ background: 'var(--flow-card)', border: '1px solid var(--flow-border)' }}
          >
            {!group.lockedVisibility && (
              <button
                type="button"
                className="flex min-h-10 w-full items-center rounded-md px-3 text-start text-[12px] hover:bg-[var(--flow-hover)]"
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
        <div className="flex flex-col gap-px pt-px">
          {visibleProjects.map((project) => {
            const active = projectId === project.id;
            return (
              <Link
                key={project.id}
                to="/p/$projectId"
                params={{ projectId: project.id }}
                className={cn(
                  'relative flex min-h-10 items-center gap-2 px-3 py-1.5 rounded-md text-[12px]',
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
                <span className="relative text-[10px]" style={{ color: 'var(--flow-text-dim)' }}>
                  {project.taskCount - project.completedCount}
                </span>
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
              onDragStart={setDraggedItemId}
              onDrop={onDrop}
            />
          ))}
        </div>
      )}
      {!group.collapsed && group.id !== 'projects' && (
        <nav className="flex flex-col gap-px pt-px">
          {groupItems.map((item) => (
            <SidebarItemRow
              key={item.id}
              item={item}
              label={itemLabel(item, t)}
              groups={preferences.groups}
              update={update}
              onDragStart={setDraggedItemId}
              onDrop={onDrop}
            />
          ))}
        </nav>
      )}
    </div>
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
  onOpenSettings,
  sidebarRef,
}) => {
  const { t } = useTranslation();
  const [draggedTopItemId, setDraggedTopItemId] = useState<string>();
  const [draggedGroupId, setDraggedGroupId] = useState<string>();
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
  const onTopItemDrop = (itemId: string): void => {
    if (!draggedTopItemId || draggedTopItemId === itemId) return;
    update((draft) => {
      const next = updatePreferencesCopy(draft);
      const from = next.topLevelItemIds.indexOf(draggedTopItemId);
      const to = next.topLevelItemIds.indexOf(itemId);
      if (from >= 0 && to >= 0) {
        next.topLevelItemIds.splice(from, 1);
        next.topLevelItemIds.splice(to, 0, draggedTopItemId);
      }
      return next;
    });
    setDraggedTopItemId(undefined);
  };
  const onGroupDrop = (groupId: string): void => {
    if (!draggedGroupId || draggedGroupId === groupId) return;
    update((draft) => {
      const next = updatePreferencesCopy(draft);
      const from = next.groupOrder.indexOf(draggedGroupId);
      const to = next.groupOrder.indexOf(groupId);
      if (from >= 0 && to >= 0) {
        next.groupOrder.splice(from, 1);
        next.groupOrder.splice(to, 0, draggedGroupId);
      }
      return next;
    });
    setDraggedGroupId(undefined);
  };

  return (
    <div ref={sidebarRef} className="flex min-h-0 flex-1 flex-col gap-2 overflow-y-auto py-2">
      <nav className="flex flex-col gap-px">
        {topItems.map((item) => (
          <SidebarItemRow
            key={item.id}
            item={item}
            label={itemLabel(item, t)}
            badge={badges[item.id]}
            groups={preferences.groups}
            update={update}
            onDragStart={setDraggedTopItemId}
            onDrop={onTopItemDrop}
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
            onGroupDragStart={setDraggedGroupId}
            onGroupDrop={onGroupDrop}
          />
        ))}
      </div>
      <button
        type="button"
        onClick={onOpenSettings}
        className="mt-1 flex min-h-10 items-center gap-2 rounded-md px-3 text-[12px] hover:bg-[var(--flow-hover)]"
        style={{ color: 'var(--flow-text-muted)' }}
      >
        <Settings size={14} />
        {t('flow.sidebar.customize')}
      </button>
    </div>
  );
};
