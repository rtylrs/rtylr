import { useCallback, useMemo } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import type { ApiUserSettings } from '../services/flowApi';
import { flowApi } from '../services/flowApi';

export type SidebarIconKey =
  | 'archive'
  | 'bar-chart'
  | 'calendar'
  | 'clock'
  | 'contact'
  | 'eye'
  | 'inbox'
  | 'link'
  | 'layout-grid'
  | 'list-todo'
  | 'map'
  | 'milestone'
  | 'radar'
  | 'settings'
  | 'sun'
  | 'timer'
  | 'users';

export type SidebarItemKind =
  | 'builtin'
  | 'project'
  | 'view'
  | 'report'
  | 'cycle'
  | 'calendar'
  | 'saved-filter'
  | 'custom-link'
  | 'recent';

export type SidebarGroupKind = 'system' | 'custom' | 'recent';

export interface SidebarItemPreference {
  id: string;
  kind: SidebarItemKind;
  label: string;
  route?: string;
  url?: string;
  icon: SidebarIconKey;
  color?: string;
  groupId?: string;
  visible: boolean;
  lockedVisibility?: boolean;
  lockedGroup?: boolean;
  openInNewTab?: boolean;
  sourceId?: string;
}

export interface SidebarGroupPreference {
  id: string;
  kind: SidebarGroupKind;
  label: string;
  icon: SidebarIconKey;
  color?: string;
  itemIds: string[];
  visible: boolean;
  collapsed: boolean;
  lockedVisibility?: boolean;
  lockedGroup?: boolean;
}

export interface SidebarPreferences {
  version: 1;
  topLevelItemIds: string[];
  groupOrder: string[];
  groups: SidebarGroupPreference[];
  items: SidebarItemPreference[];
}

// Keep projects near pinned items; saved group order still takes precedence.
export const SIDEBAR_GROUP_IDS = ['projects', 'plan', 'team', 'time'] as const;

const builtinItems: SidebarItemPreference[] = [
  {
    id: 'builtin:leave',
    kind: 'builtin',
    label: 'flow.leave.heading',
    route: '/flow/leave',
    icon: 'calendar',
    groupId: 'time',
    visible: true,
  },
  {
    id: 'builtin:inbox',
    kind: 'builtin',
    label: 'flowNav.inbox',
    route: '/flow/inbox',
    icon: 'inbox',
    visible: true,
  },
  {
    id: 'builtin:my-issues',
    kind: 'builtin',
    label: 'flow.shell.myTasks',
    route: '/flow/my-issues',
    icon: 'list-todo',
    visible: true,
  },
  {
    id: 'builtin:today',
    kind: 'builtin',
    label: 'flowNav.focusToday',
    route: '/flow/today',
    icon: 'sun',
    visible: true,
  },
  {
    id: 'builtin:projects-shortcut',
    kind: 'builtin',
    // The title of the screen it opens, not nav.projects. That key also names
    // the Projects group below it, so the sidebar showed two rows called
    // "Projects" pointing at different places.
    label: 'flow.allProjects.title',
    route: '/flow/projects',
    icon: 'layout-grid',
    visible: true,
  },
  {
    id: 'builtin:triage',
    kind: 'builtin',
    label: 'flow.shell.triage',
    route: '/flow/triage',
    icon: 'eye',
    visible: true,
  },
  {
    id: 'builtin:roadmap',
    kind: 'builtin',
    label: 'flow.shell.roadmap',
    route: '/flow/roadmap',
    icon: 'map',
    groupId: 'plan',
    visible: true,
  },
  {
    id: 'builtin:cycles',
    kind: 'builtin',
    label: 'flow.shell.cycles',
    route: '/flow/cycles',
    icon: 'timer',
    groupId: 'plan',
    visible: true,
  },
  {
    id: 'builtin:milestones',
    kind: 'builtin',
    label: 'flow.shell.milestones',
    route: '/flow/milestones',
    icon: 'milestone',
    groupId: 'plan',
    visible: true,
  },
  {
    id: 'builtin:timeline',
    kind: 'builtin',
    label: 'flow.shell.timeline',
    route: '/flow/timeline',
    icon: 'timer',
    groupId: 'plan',
    visible: true,
  },
  {
    id: 'builtin:calendar',
    kind: 'builtin',
    label: 'flow.shell.calendar',
    route: '/flow/calendar',
    icon: 'calendar',
    groupId: 'plan',
    visible: true,
  },
  {
    id: 'builtin:views',
    kind: 'builtin',
    label: 'flow.shell.views',
    route: '/flow/views',
    icon: 'eye',
    groupId: 'plan',
    visible: true,
  },
  {
    id: 'builtin:archive',
    kind: 'builtin',
    label: 'flow.shell.archive',
    route: '/flow/archive',
    icon: 'archive',
    groupId: 'plan',
    visible: true,
  },
  {
    id: 'builtin:workload',
    kind: 'builtin',
    label: 'flow.shell.workload',
    route: '/flow/team',
    icon: 'users',
    groupId: 'team',
    visible: true,
  },
  {
    id: 'builtin:insights',
    kind: 'builtin',
    label: 'flow.shell.insights',
    route: '/flow/insights',
    icon: 'bar-chart',
    groupId: 'team',
    visible: true,
  },
  {
    id: 'builtin:time-day',
    kind: 'builtin',
    label: 'flow.time.myDay',
    route: '/flow/time',
    icon: 'clock',
    groupId: 'time',
    visible: true,
  },
  {
    id: 'builtin:time-team',
    kind: 'builtin',
    label: 'flow.time.team',
    route: '/flow/time/team',
    icon: 'users',
    groupId: 'time',
    visible: true,
  },
  {
    id: 'builtin:time-reports',
    kind: 'builtin',
    label: 'flow.time.reports',
    route: '/flow/time/reports',
    icon: 'bar-chart',
    groupId: 'time',
    visible: true,
  },
  {
    id: 'builtin:directory',
    kind: 'builtin',
    label: 'flow.directory.heading',
    route: '/flow/directory',
    icon: 'contact',
    groupId: 'time',
    visible: true,
  },
  {
    id: 'builtin:presence',
    kind: 'builtin',
    label: 'flow.presence.heading',
    route: '/flow/presence',
    icon: 'radar',
    groupId: 'time',
    visible: true,
  },
  {
    id: 'builtin:time-settings',
    kind: 'builtin',
    // Its own key, not flow.settings.time.section. That string is also the
    // Settings tab label, where "Time" is right; here it sat inside a group
    // already called Time, so the sidebar showed two entries with one name.
    label: 'flow.time.settingsLink',
    route: '/settings/flow?section=time',
    icon: 'settings',
    groupId: 'time',
    visible: true,
  },
];

/**
 * Plan, Team and Time start collapsed; Projects starts open.
 *
 * Expanded, Plan's five items and Time's six pushed the project list off the
 * bottom of a 900px viewport (#1076). Collapsing them costs one extra click to
 * reach Cycles or Views on a fresh profile, and nothing from the second
 * session on, because the toggle is saved with the rest of the preferences.
 */
const builtinGroups: SidebarGroupPreference[] = [
  {
    id: 'plan',
    kind: 'system',
    label: 'flow.shell.plan',
    icon: 'map',
    itemIds: builtinItems.filter((item) => item.groupId === 'plan').map((item) => item.id),
    visible: true,
    collapsed: true,
  },
  {
    id: 'team',
    kind: 'system',
    label: 'flow.shell.team',
    icon: 'users',
    itemIds: builtinItems.filter((item) => item.groupId === 'team').map((item) => item.id),
    visible: true,
    collapsed: true,
  },
  {
    id: 'time',
    kind: 'system',
    label: 'flow.time.heading',
    icon: 'clock',
    itemIds: builtinItems.filter((item) => item.groupId === 'time').map((item) => item.id),
    visible: true,
    collapsed: true,
  },
  {
    id: 'projects',
    kind: 'system',
    label: 'flowNav.projects',
    icon: 'layout-grid',
    itemIds: [],
    visible: true,
    collapsed: false,
  },
];

export const createDefaultSidebarPreferences = (): SidebarPreferences => ({
  version: 1,
  topLevelItemIds: builtinItems.filter((item) => item.groupId === undefined).map((item) => item.id),
  groupOrder: [...SIDEBAR_GROUP_IDS],
  groups: builtinGroups.map((group) => ({ ...group, itemIds: [...group.itemIds] })),
  items: builtinItems.map((item) => ({ ...item })),
});

const asRecord = (value: unknown): Record<string, unknown> | undefined =>
  typeof value === 'object' && value !== null ? (value as Record<string, unknown>) : undefined;

const stringArray = (value: unknown): string[] =>
  Array.isArray(value) ? value.filter((entry): entry is string => typeof entry === 'string') : [];

const normalizeItem = (value: unknown): SidebarItemPreference | undefined => {
  const record = asRecord(value);
  if (!record || typeof record.id !== 'string' || typeof record.kind !== 'string') return undefined;
  if (typeof record.label !== 'string' || typeof record.icon !== 'string') return undefined;
  return {
    id: record.id,
    kind: record.kind as SidebarItemKind,
    label: record.label,
    route: typeof record.route === 'string' ? record.route : undefined,
    url: typeof record.url === 'string' ? record.url : undefined,
    icon: record.icon as SidebarIconKey,
    color: typeof record.color === 'string' ? record.color : undefined,
    groupId: typeof record.groupId === 'string' ? record.groupId : undefined,
    visible: record.visible !== false,
    lockedVisibility: record.lockedVisibility === true,
    lockedGroup: record.lockedGroup === true,
    openInNewTab: record.openInNewTab === true,
    sourceId: typeof record.sourceId === 'string' ? record.sourceId : undefined,
  };
};

const normalizeGroup = (value: unknown): SidebarGroupPreference | undefined => {
  const record = asRecord(value);
  if (!record || typeof record.id !== 'string' || typeof record.kind !== 'string') return undefined;
  if (typeof record.label !== 'string' || typeof record.icon !== 'string') return undefined;
  return {
    id: record.id,
    kind: record.kind as SidebarGroupKind,
    label: record.label,
    icon: record.icon as SidebarIconKey,
    color: typeof record.color === 'string' ? record.color : undefined,
    itemIds: stringArray(record.itemIds),
    visible: record.visible !== false,
    collapsed: record.collapsed === true,
    lockedVisibility: record.lockedVisibility === true,
    lockedGroup: record.lockedGroup === true,
  };
};

export const parseSidebarPreferences = (raw?: string): SidebarPreferences => {
  const defaults = createDefaultSidebarPreferences();
  if (!raw) return defaults;
  try {
    const record = asRecord(JSON.parse(raw));
    if (!record || record.version !== 1) return defaults;
    const parsedItems = Array.isArray(record.items)
      ? record.items
          .map(normalizeItem)
          .filter((item): item is SidebarItemPreference => item !== undefined)
      : [];
    const parsedGroups = Array.isArray(record.groups)
      ? record.groups
          .map(normalizeGroup)
          .filter((group): group is SidebarGroupPreference => group !== undefined)
      : [];
    const itemsById = new Map(parsedItems.map((item) => [item.id, item]));
    const groupsById = new Map(parsedGroups.map((group) => [group.id, group]));

    const mergedItems: SidebarItemPreference[] = defaults.items.map((item) => ({
      ...item,
      ...(itemsById.get(item.id) ?? {}),
      kind: item.kind,
      label: item.label,
      route: item.route,
    }));
    parsedItems.forEach((item) => {
      if (item.route && !item.route.startsWith('/flow/')) return;
      if (!mergedItems.some((candidate) => candidate.id === item.id)) mergedItems.push(item);
    });
    const mergedGroups = defaults.groups.map((group) => ({
      ...group,
      ...(groupsById.get(group.id) ?? {}),
      kind: group.kind,
      label: group.label,
      itemIds: groupsById.get(group.id)?.itemIds ?? group.itemIds,
    }));
    parsedGroups.forEach((group) => {
      if (!mergedGroups.some((candidate) => candidate.id === group.id)) mergedGroups.push(group);
    });
    const validGroupIds = new Set(mergedGroups.map((group) => group.id));
    const validItemIds = new Set(mergedItems.map((item) => item.id));
    const cleanGroup = (group: SidebarGroupPreference): SidebarGroupPreference => ({
      ...group,
      itemIds: group.itemIds.filter((id) => validItemIds.has(id)),
    });
    const usedIds = new Set(mergedGroups.flatMap((group) => group.itemIds));
    mergedItems.forEach((item) => {
      if (item.groupId && validGroupIds.has(item.groupId) && !usedIds.has(item.id)) {
        const group = mergedGroups.find((candidate) => candidate.id === item.groupId);
        group?.itemIds.push(item.id);
      }
    });
    const topLevelItemIds = stringArray(record.topLevelItemIds).filter(
      (id) => validItemIds.has(id) && !usedIds.has(id)
    );
    defaults.topLevelItemIds.forEach((id) => {
      if (
        !topLevelItemIds.includes(id) &&
        !mergedGroups.some((group) => group.itemIds.includes(id))
      ) {
        topLevelItemIds.push(id);
      }
    });
    const groupOrder = stringArray(record.groupOrder).filter((id) => validGroupIds.has(id));
    mergedGroups.forEach((group) => {
      if (!groupOrder.includes(group.id)) groupOrder.push(group.id);
    });
    return {
      version: 1,
      topLevelItemIds,
      groupOrder,
      groups: mergedGroups.map(cleanGroup),
      items: mergedItems,
    };
  } catch {
    return defaults;
  }
};

export const serializeSidebarPreferences = (preferences: SidebarPreferences): string =>
  JSON.stringify(preferences);

export function updateSidebarPreferences(
  current: SidebarPreferences,
  update: (draft: SidebarPreferences) => SidebarPreferences
): SidebarPreferences {
  return update(updatePreferencesCopy(current));
}

export const updatePreferencesCopy = (current: SidebarPreferences): SidebarPreferences => ({
  version: 1,
  topLevelItemIds: [...current.topLevelItemIds],
  groupOrder: [...current.groupOrder],
  groups: current.groups.map((group) => ({ ...group, itemIds: [...group.itemIds] })),
  items: current.items.map((item) => ({ ...item })),
});

export const recordRecentSidebarPage = (
  current: SidebarPreferences,
  route: string,
  label: string
): SidebarPreferences => {
  if (!route || route === '/settings/flow') return current;
  const next = updatePreferencesCopy(current);
  let group = next.groups.find((candidate) => candidate.id === 'recent');
  if (!group) {
    group = {
      id: 'recent',
      kind: 'recent',
      label: 'flow.sidebar.recent',
      icon: 'clock',
      itemIds: [],
      visible: true,
      collapsed: false,
    };
    next.groups.push(group);
    next.groupOrder.push('recent');
  }
  const itemId = `recent:${route}`;
  group.itemIds = group.itemIds.filter((id) => id !== itemId);
  next.items = next.items.filter((item) => item.id !== itemId);
  next.items.push({
    id: itemId,
    kind: 'recent',
    label,
    route,
    icon: 'clock',
    groupId: 'recent',
    visible: true,
  });
  group.itemIds.unshift(itemId);
  const recentIds = group.itemIds.slice(0, 5);
  const removedIds = group.itemIds.slice(5);
  group.itemIds = recentIds;
  next.items = next.items.filter((item) => !removedIds.includes(item.id));
  return next;
};

export const sidebarSettingsQueryKey = ['flow-settings'] as const;

const defaultSettings: ApiUserSettings = {
  taskAssigned: true,
  taskComments: true,
  mentions: true,
  statusChanges: false,
  weeklyDigest: true,
  emailNotifications: false,
};

export function useSidebarPreferences() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const { data: settings, isLoading } = useQuery({
    queryKey: sidebarSettingsQueryKey,
    queryFn: () => unwrap(flowApi.getUserSettings()),
  });

  // Derived from the shared cache rather than copied into per-instance state.
  // FlowLayout and SidebarSettingsSection each mount this hook, and a private
  // copy of the document is what let one instance autosave a snapshot taken
  // before the other's just-made change.
  const preferences = useMemo(
    () => parseSidebarPreferences(settings?.sidebarPreferences),
    [settings]
  );

  const save = useMutation({
    // The payload is a whole document, so writes from every mounted instance
    // are serialized: an older response must not land after a newer one.
    scope: { id: 'flow-sidebar-preferences' },
    mutationFn: async ({ preferences: next }: { preferences: SidebarPreferences }) => {
      const latest = queryClient.getQueryData<ApiUserSettings>(sidebarSettingsQueryKey);
      const sent: ApiUserSettings = {
        ...(latest ?? defaultSettings),
        sidebarPreferences: serializeSidebarPreferences(next),
      };
      await unwrap(flowApi.updateUserSettings(sent));
      // Return what was sent, not what came back. UpdateUserSettingsResponse
      // carries no fields, so the endpoint answers `200 {}`, and caching that
      // response as the settings object made `settings.sidebarPreferences`
      // undefined. The hook re-derives its state from this cache entry, so
      // every group, pin, shortcut and recent entry reverted to the defaults
      // immediately after a successful save.
      return sent;
    },
    onSuccess: (sent) => {
      const latest = queryClient.getQueryData<ApiUserSettings>(sidebarSettingsQueryKey);
      // A newer optimistic document wins: writing this response would revert a
      // change the user made while the request was in flight.
      if (latest?.sidebarPreferences === sent.sidebarPreferences) {
        queryClient.setQueryData(sidebarSettingsQueryKey, sent);
      }
    },
    onError: () => toast.error(t('flow.sidebar.saveFailed')),
  });

  const update = useCallback(
    (mutator: (draft: SidebarPreferences) => SidebarPreferences) => {
      const cached = queryClient.getQueryData<ApiUserSettings>(sidebarSettingsQueryKey);
      const next = updateSidebarPreferences(
        parseSidebarPreferences(cached?.sidebarPreferences),
        mutator
      );
      const optimistic: ApiUserSettings = {
        ...(cached ?? defaultSettings),
        sidebarPreferences: serializeSidebarPreferences(next),
      };
      // Written before the request so every instance reads the new document
      // immediately, including the one that did not make this change.
      queryClient.setQueryData(sidebarSettingsQueryKey, optimistic);
      save.mutate(
        { preferences: next },
        {
          onError: () => {
            // Restore the document this update replaced, and only if nothing
            // newer has written since: rolling back over a newer change would
            // lose it the way this fix exists to prevent.
            const latest = queryClient.getQueryData<ApiUserSettings>(sidebarSettingsQueryKey);
            if (latest?.sidebarPreferences === optimistic.sidebarPreferences) {
              queryClient.setQueryData<ApiUserSettings | undefined>(
                sidebarSettingsQueryKey,
                cached
              );
            }
          },
        }
      );
    },
    [queryClient, save]
  );

  return useMemo(
    () => ({ preferences, update, isLoading, isSaving: save.isPending }),
    [preferences, update, isLoading, save.isPending]
  );
}
