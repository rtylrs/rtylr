import { useCallback, useEffect, useMemo, useState } from 'react';
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

export interface SidebarPreferencesUpdateOptions {
  showToast?: boolean;
}

export const SIDEBAR_GROUP_IDS = ['plan', 'team', 'time', 'projects'] as const;

const builtinItems: SidebarItemPreference[] = [
  {
    id: 'builtin:inbox',
    kind: 'builtin',
    label: 'nav.inbox',
    route: '/inbox',
    icon: 'inbox',
    visible: true,
  },
  {
    id: 'builtin:my-issues',
    kind: 'builtin',
    label: 'flow.shell.myTasks',
    route: '/my-issues',
    icon: 'list-todo',
    visible: true,
  },
  {
    id: 'builtin:today',
    kind: 'builtin',
    label: 'nav.focusToday',
    route: '/today',
    icon: 'sun',
    visible: true,
  },
  {
    id: 'builtin:projects-shortcut',
    kind: 'builtin',
    label: 'nav.projects',
    route: '/projects',
    icon: 'layout-grid',
    visible: true,
  },
  {
    id: 'builtin:triage',
    kind: 'builtin',
    label: 'flow.shell.triage',
    route: '/triage',
    icon: 'eye',
    visible: true,
  },
  {
    id: 'builtin:roadmap',
    kind: 'builtin',
    label: 'flow.shell.roadmap',
    route: '/roadmap',
    icon: 'map',
    groupId: 'plan',
    visible: true,
  },
  {
    id: 'builtin:cycles',
    kind: 'builtin',
    label: 'flow.shell.cycles',
    route: '/cycles',
    icon: 'timer',
    groupId: 'plan',
    visible: true,
  },
  {
    id: 'builtin:milestones',
    kind: 'builtin',
    label: 'flow.shell.milestones',
    route: '/milestones',
    icon: 'milestone',
    groupId: 'plan',
    visible: true,
  },
  {
    id: 'builtin:timeline',
    kind: 'builtin',
    label: 'flow.shell.timeline',
    route: '/timeline',
    icon: 'timer',
    groupId: 'plan',
    visible: true,
  },
  {
    id: 'builtin:calendar',
    kind: 'builtin',
    label: 'flow.shell.calendar',
    route: '/calendar',
    icon: 'calendar',
    groupId: 'plan',
    visible: true,
  },
  {
    id: 'builtin:views',
    kind: 'builtin',
    label: 'flow.shell.views',
    route: '/views',
    icon: 'eye',
    groupId: 'plan',
    visible: true,
  },
  {
    id: 'builtin:archive',
    kind: 'builtin',
    label: 'flow.shell.archive',
    route: '/archive',
    icon: 'archive',
    groupId: 'plan',
    visible: true,
  },
  {
    id: 'builtin:workload',
    kind: 'builtin',
    label: 'flow.shell.workload',
    route: '/team',
    icon: 'users',
    groupId: 'team',
    visible: true,
  },
  {
    id: 'builtin:insights',
    kind: 'builtin',
    label: 'flow.shell.insights',
    route: '/insights',
    icon: 'bar-chart',
    groupId: 'team',
    visible: true,
  },
  {
    id: 'builtin:time-day',
    kind: 'builtin',
    label: 'flow.time.myDay',
    route: '/time',
    icon: 'clock',
    groupId: 'time',
    visible: true,
  },
  {
    id: 'builtin:time-team',
    kind: 'builtin',
    label: 'flow.time.team',
    route: '/time/team',
    icon: 'users',
    groupId: 'time',
    visible: true,
  },
  {
    id: 'builtin:time-reports',
    kind: 'builtin',
    label: 'flow.time.reports',
    route: '/time/reports',
    icon: 'bar-chart',
    groupId: 'time',
    visible: true,
  },
  {
    id: 'builtin:directory',
    kind: 'builtin',
    label: 'flow.directory.heading',
    route: '/directory',
    icon: 'contact',
    groupId: 'time',
    visible: true,
  },
  {
    id: 'builtin:presence',
    kind: 'builtin',
    label: 'flow.presence.heading',
    route: '/presence',
    icon: 'radar',
    groupId: 'time',
    visible: true,
  },
  {
    id: 'builtin:time-settings',
    kind: 'builtin',
    label: 'flow.settings.time.section',
    route: '/settings?section=time',
    icon: 'settings',
    groupId: 'time',
    visible: true,
  },
];

const builtinGroups: SidebarGroupPreference[] = [
  {
    id: 'plan',
    kind: 'system',
    label: 'flow.shell.plan',
    icon: 'map',
    itemIds: builtinItems.filter((item) => item.groupId === 'plan').map((item) => item.id),
    visible: true,
    collapsed: false,
  },
  {
    id: 'team',
    kind: 'system',
    label: 'flow.shell.team',
    icon: 'users',
    itemIds: builtinItems.filter((item) => item.groupId === 'team').map((item) => item.id),
    visible: true,
    collapsed: false,
  },
  {
    id: 'time',
    kind: 'system',
    label: 'flow.time.heading',
    icon: 'clock',
    itemIds: builtinItems.filter((item) => item.groupId === 'time').map((item) => item.id),
    visible: true,
    collapsed: false,
  },
  {
    id: 'projects',
    kind: 'system',
    label: 'nav.projects',
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

    const mergedItems = defaults.items.map((item) => ({
      ...item,
      ...(itemsById.get(item.id) ?? {}),
    }));
    parsedItems.forEach((item) => {
      if (!mergedItems.some((candidate) => candidate.id === item.id)) mergedItems.push(item);
    });
    const mergedGroups = defaults.groups.map((group) => ({
      ...group,
      ...(groupsById.get(group.id) ?? {}),
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

export const moveId = (ids: string[], id: string, destination: number): string[] => {
  const currentIndex = ids.indexOf(id);
  if (currentIndex < 0) return ids;
  const next = ids.filter((entry) => entry !== id);
  next.splice(Math.max(0, Math.min(destination, next.length)), 0, id);
  return next;
};

export const recordRecentSidebarPage = (
  current: SidebarPreferences,
  route: string,
  label: string
): SidebarPreferences => {
  if (!route || route === '/settings') return current;
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
  const [preferences, setPreferences] = useState<SidebarPreferences>(() =>
    parseSidebarPreferences(settings?.sidebarPreferences)
  );

  useEffect(() => {
    if (settings) setPreferences(parseSidebarPreferences(settings.sidebarPreferences));
  }, [settings]);

  const save = useMutation({
    mutationFn: async ({
      preferences: next,
    }: {
      preferences: SidebarPreferences;
      showToast: boolean;
    }) => {
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
    onSuccess: (sent, variables) => {
      queryClient.setQueryData(sidebarSettingsQueryKey, sent);
      if (variables.showToast) toast.success(t('flow.sidebar.saved'));
    },
    onError: (_error, variables) => {
      if (variables.showToast) toast.error(t('flow.sidebar.saveFailed'));
    },
  });

  const update = useCallback(
    (
      mutator: (draft: SidebarPreferences) => SidebarPreferences,
      options: SidebarPreferencesUpdateOptions = {}
    ) => {
      const next = updateSidebarPreferences(preferences, mutator);
      setPreferences(next);
      save.mutate(
        { preferences: next, showToast: options.showToast !== false },
        {
          onError: () => setPreferences(preferences),
        }
      );
    },
    [preferences, save]
  );

  return useMemo(
    () => ({ preferences, update, isLoading, isSaving: save.isPending }),
    [preferences, update, isLoading, save.isPending]
  );
}
