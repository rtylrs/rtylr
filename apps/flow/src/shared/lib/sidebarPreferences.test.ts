import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { act, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

const { getUserSettings, updateUserSettings } = vi.hoisted(() => ({
  getUserSettings: vi.fn(),
  updateUserSettings: vi.fn(),
}));

vi.mock('../services/flowApi', () => ({
  flowApi: { getUserSettings, updateUserSettings },
}));

// The real unwrap takes a neverthrow ResultAsync and reads isOk()/value. The
// hook's cache behaviour is what is under test here, not that plumbing, so
// unwrap becomes an await and the flowApi mocks resolve plain objects.
vi.mock('@voxire/api', () => ({
  unwrap: async (value: unknown) => await value,
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

import {
  SIDEBAR_GROUP_IDS,
  createDefaultSidebarPreferences,
  parseSidebarPreferences,
  recordRecentSidebarPage,
  serializeSidebarPreferences,
  sidebarSettingsQueryKey,
  updatePreferencesCopy,
  useSidebarPreferences,
} from './sidebarPreferences';

describe('sidebar preferences', () => {
  it('starts with the system groups and top navigation', () => {
    const preferences = createDefaultSidebarPreferences();

    // Projects first, deliberately. With it last, three projects put P-Alpha at
    // y=1080 in a 900px viewport, below a fold with no scrollbar and no fade to
    // say the region scrolled (#1076).
    expect(preferences.groupOrder).toEqual(['projects', 'plan', 'team', 'time']);
    expect(preferences.topLevelItemIds).toEqual([
      'builtin:inbox',
      'builtin:my-issues',
      'builtin:today',
      'builtin:projects-shortcut',
      'builtin:triage',
    ]);
  });

  it('keeps personal changes while adding new system items', () => {
    const current = createDefaultSidebarPreferences();
    current.items.find((item) => item.id === 'builtin:inbox')!.visible = false;
    const parsed = parseSidebarPreferences(serializeSidebarPreferences(current));

    expect(parsed.items.find((item) => item.id === 'builtin:inbox')?.visible).toBe(false);
    expect(parsed.items.some((item) => item.id === 'builtin:archive')).toBe(true);
  });

  it('limits recent pages to five and keeps the newest first', () => {
    let preferences = createDefaultSidebarPreferences();
    for (let index = 0; index < 6; index += 1) {
      preferences = recordRecentSidebarPage(preferences, `/page-${index}`, `Page ${index}`);
    }

    const recent = preferences.groups.find((group) => group.id === 'recent');
    expect(recent?.itemIds).toHaveLength(5);
    expect(recent?.itemIds[0]).toBe('recent:/page-5');
    expect(preferences.items.some((item) => item.id === 'recent:/page-0')).toBe(false);
  });

  it('does not mutate the current preference document while copying it', () => {
    const current = createDefaultSidebarPreferences();
    const copy = updatePreferencesCopy(current);
    copy.groups[0].itemIds.reverse();
    copy.items[0].visible = false;

    expect(current.groups[0].itemIds).not.toEqual(copy.groups[0].itemIds);
    expect(current.items[0].visible).toBe(true);
  });
});

describe('useSidebarPreferences save does not clobber the cache (#1059)', () => {
  // The endpoint really does answer `200 {}`: UpdateUserSettingsResponse
  // carries no fields. Caching that response as the settings object is the
  // defect, so the mock resolves an empty object rather than echoing.
  beforeEach(() => {
    vi.clearAllMocks();
    updateUserSettings.mockImplementation(async () => ({}));
  });

  const withRecentGroup = (): string => {
    const prefs = createDefaultSidebarPreferences();
    prefs.groupOrder = [...prefs.groupOrder, 'recent'];
    prefs.groups = [
      ...prefs.groups,
      {
        id: 'recent',
        kind: 'recent',
        label: 'flow.sidebar.recent',
        icon: 'clock',
        itemIds: [],
        visible: true,
        collapsed: false,
      },
    ];
    return serializeSidebarPreferences(prefs);
  };

  const renderPrefs = async () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const wrapper = ({ children }: { children: React.ReactNode }): React.JSX.Element =>
      React.createElement(QueryClientProvider, { client: queryClient }, children);
    const view = renderHook(() => useSidebarPreferences(), { wrapper });
    await waitFor(() => expect(view.result.current.isLoading).toBe(false));
    return { ...view, queryClient };
  };

  it('keeps the custom groups after a successful save', async () => {
    getUserSettings.mockResolvedValue({
      taskAssigned: true,
      taskComments: true,
      mentions: true,
      statusChanges: false,
      weeklyDigest: true,
      emailNotifications: false,
      sidebarPreferences: withRecentGroup(),
    });

    const { result, queryClient } = await renderPrefs();
    expect(result.current.preferences.groupOrder).toContain('recent');

    await act(async () => {
      result.current.update((draft) => draft);
    });
    await waitFor(() => expect(updateUserSettings).toHaveBeenCalledTimes(1));

    // The cache entry must still describe the settings, not the empty response.
    const cached = queryClient.getQueryData(sidebarSettingsQueryKey) as
      | { sidebarPreferences?: string }
      | undefined;
    expect(cached?.sidebarPreferences).toBeTruthy();
    expect(Object.keys(cached ?? {})).not.toEqual([]);

    // And the hook, which re-derives from that entry, must not fall back.
    await waitFor(() => expect(result.current.preferences.groupOrder).toContain('recent'));
  });

  it('keeps the notification booleans, so a later save cannot send a partial body', async () => {
    getUserSettings.mockResolvedValue({
      taskAssigned: true,
      taskComments: true,
      mentions: true,
      statusChanges: true,
      weeklyDigest: true,
      emailNotifications: true,
      sidebarPreferences: withRecentGroup(),
    });

    const { result, queryClient } = await renderPrefs();
    await act(async () => {
      result.current.update((draft) => draft);
    });
    await waitFor(() => expect(updateUserSettings).toHaveBeenCalledTimes(1));

    const cached = queryClient.getQueryData(sidebarSettingsQueryKey) as
      | Record<string, unknown>
      | undefined;
    expect(cached).toMatchObject({
      taskAssigned: true,
      taskComments: true,
      mentions: true,
      statusChanges: true,
      weeklyDigest: true,
      emailNotifications: true,
    });
  });

  it('sends the settings it already holds rather than the defaults', async () => {
    getUserSettings.mockResolvedValue({
      taskAssigned: false,
      taskComments: false,
      mentions: false,
      statusChanges: false,
      weeklyDigest: false,
      emailNotifications: false,
      sidebarPreferences: withRecentGroup(),
    });

    const { result } = await renderPrefs();
    await act(async () => {
      result.current.update((draft) => draft);
    });
    await waitFor(() => expect(updateUserSettings).toHaveBeenCalledTimes(1));

    // defaultSettings has taskAssigned/mentions/weeklyDigest true, so if the
    // payload came from the defaults rather than the loaded settings these
    // would flip back on behind the user.
    expect(updateUserSettings.mock.calls[0][0]).toMatchObject({
      taskAssigned: false,
      mentions: false,
      weeklyDigest: false,
    });
  });
});

describe('sidebar defaults keep projects above the fold (#1076)', () => {
  it('collapses the long groups and leaves projects open', () => {
    // Expanded, Plan's five items and Time's six pushed the project list off
    // the bottom of the viewport.
    const byId = new Map(createDefaultSidebarPreferences().groups.map((g) => [g.id, g]));
    expect(byId.get('plan')?.collapsed).toBe(true);
    expect(byId.get('team')?.collapsed).toBe(true);
    expect(byId.get('time')?.collapsed).toBe(true);
    expect(byId.get('projects')?.collapsed).toBe(false);
  });

  it('keeps a group order the user saved, rather than reimposing the default', () => {
    const stored = JSON.stringify({
      version: 1,
      groupOrder: ['time', 'projects', 'plan', 'team'],
      groups: [],
      items: [],
      topLevelItemIds: [],
    });
    expect(parseSidebarPreferences(stored).groupOrder).toEqual([
      'time',
      'projects',
      'plan',
      'team',
    ]);
  });

  it('keeps a group the user expanded expanded', () => {
    const stored = JSON.stringify({
      version: 1,
      groupOrder: [...SIDEBAR_GROUP_IDS],
      // icon is required by normalizeGroup: without it the group is dropped and
      // the default wins, which is not what a real serialized payload looks like.
      groups: [
        {
          id: 'plan',
          kind: 'system',
          label: 'flow.shell.plan',
          icon: 'map',
          itemIds: [],
          visible: true,
          collapsed: false,
        },
      ],
      items: [],
      topLevelItemIds: [],
    });
    const plan = parseSidebarPreferences(stored).groups.find((g) => g.id === 'plan');
    expect(plan?.collapsed).toBe(false);
  });

  it('appends a group missing from a stored order rather than dropping it', () => {
    const stored = JSON.stringify({
      version: 1,
      groupOrder: ['plan'],
      groups: [],
      items: [],
      topLevelItemIds: [],
    });
    const order = parseSidebarPreferences(stored).groupOrder;
    expect(order[0]).toBe('plan');
    expect([...order].sort()).toEqual([...SIDEBAR_GROUP_IDS].sort());
  });
});
