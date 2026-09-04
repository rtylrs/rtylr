// @vitest-environment happy-dom
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

const { getUserSettings, updateUserSettings } = vi.hoisted(() => ({
  getUserSettings: vi.fn(),
  updateUserSettings: vi.fn(),
}));

vi.mock('../../shared/services/flowApi', () => ({
  flowApi: {
    getUserSettings,
    updateUserSettings,
    // FlowSettings imports a wide surface; these are never reached by the
    // notifications section but must exist for the module to evaluate.
    getUserProfile: vi.fn(),
    listLabels: vi.fn(),
    getFlowTeam: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: async (value: unknown) => await value,
  // The module graph reaches shared/lib/clients, which builds OneKit clients at
  // module scope. None of them are called from this section.
  createOneKitClient: () => new Proxy({}, { get: () => vi.fn() }),
  createRestRouteClient: () => new Proxy({}, { get: () => vi.fn() }),
  oneKitResult: vi.fn(),
  getOneKitErrorMessage: (_error: unknown, fallback: string) => fallback,
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
  // shared/services/i18next calls i18next.use(initReactI18next) at import.
  initReactI18next: { type: '3rdParty', init: () => undefined },
  Trans: ({ children }: { children: React.ReactNode }) => children,
}));

vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

import { NotificationsSection } from './FlowSettings';

const ALL_SIX = [
  'taskAssigned',
  'taskComments',
  'mentions',
  'statusChanges',
  'weeklyDigest',
  'emailNotifications',
] as const;

const renderSection = (): void => {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(
    React.createElement(
      QueryClientProvider,
      { client: queryClient },
      React.createElement(NotificationsSection)
    )
  );
};

/**
 * Renders with the settings query held open, then resolves it inside `act`.
 *
 * `waitFor` is the wrong tool for these assertions: it retries until something
 * becomes true, so for a value that is supposed to STAY as it is it passes on
 * the first tick, before the query has resolved and before the effect under
 * test has run. That made an earlier version of these tests pass with the fix
 * reverted. Resolving explicitly makes "the effect has run" a fact.
 */
const renderAndResolveWith = async (settings: unknown): Promise<void> => {
  let resolveSettings: (value: unknown) => void = () => undefined;
  getUserSettings.mockReturnValue(
    new Promise((resolve) => {
      resolveSettings = resolve;
    })
  );
  renderSection();
  await act(async () => {
    resolveSettings(settings);
  });
  // Two macrotask flushes: one for react-query to commit the query result,
  // one for the effect that reads it to run. Without both, an assertion made
  // here is really an assertion about the component's initial defaults.
  for (let i = 0; i < 2; i += 1) {
    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 0));
    });
  }
};

const switchFor = (key: string): HTMLElement =>
  screen.getByRole('switch', {
    name: `flow.settings.notifications.items.${key}.label`,
  });

describe('NotificationsSection sends every preference (#1063)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    updateUserSettings.mockResolvedValue({});
  });

  it('sends all six booleans when one is toggled', async () => {
    await renderAndResolveWith({
      taskAssigned: true,
      taskComments: true,
      mentions: true,
      statusChanges: true,
      weeklyDigest: true,
      emailNotifications: true,
    });
    // statusChanges, not taskAssigned: the component's own default for
    // taskAssigned is already true, so asserting that would not prove the
    // server settings were adopted.
    expect(switchFor('statusChanges')).toHaveAttribute('aria-checked', 'true');

    fireEvent.click(switchFor('taskAssigned'));

    await waitFor(() => expect(updateUserSettings).toHaveBeenCalledTimes(1));
    const sent = updateUserSettings.mock.calls[0][0] as Record<string, unknown>;
    // The reported bug was a body carrying only the changed field, which the
    // server reads as the other five being false.
    for (const key of ALL_SIX) expect(sent).toHaveProperty(key);
    expect(sent).toMatchObject({
      taskAssigned: false,
      taskComments: true,
      mentions: true,
      statusChanges: true,
      weeklyDigest: true,
      emailNotifications: true,
    });
  });

  it('does not adopt an empty settings object, so a toggle still sends six fields', async () => {
    // This is the shape the ['flow-settings'] cache entry used to hold, because
    // the sidebar save cached the save endpoint's empty 200 {} response.
    await renderAndResolveWith({});

    // Adopting {} would leave every switch reading false, because
    // settings[key] would be undefined. Keeping the defaults is what says the
    // empty payload was rejected.
    expect(switchFor('taskAssigned')).toHaveAttribute('aria-checked', 'true');
    expect(switchFor('mentions')).toHaveAttribute('aria-checked', 'true');

    fireEvent.click(switchFor('taskAssigned'));

    await waitFor(() => expect(updateUserSettings).toHaveBeenCalledTimes(1));
    const sent = updateUserSettings.mock.calls[0][0] as Record<string, unknown>;
    for (const key of ALL_SIX) expect(sent).toHaveProperty(key);
    expect(Object.keys(sent).length).toBeGreaterThanOrEqual(ALL_SIX.length);
  });

  it('adopts real server settings rather than keeping its own defaults', async () => {
    // The component's defaults have statusChanges false and weeklyDigest true.
    await renderAndResolveWith({
      taskAssigned: false,
      taskComments: false,
      mentions: false,
      statusChanges: true,
      weeklyDigest: false,
      emailNotifications: false,
    });

    expect(switchFor('statusChanges')).toHaveAttribute('aria-checked', 'true');
    expect(switchFor('weeklyDigest')).toHaveAttribute('aria-checked', 'false');
  });

  it('names every switch, so six controls are not announced anonymously', async () => {
    await renderAndResolveWith({
      taskAssigned: true,
      taskComments: true,
      mentions: true,
      statusChanges: true,
      weeklyDigest: true,
      emailNotifications: true,
    });

    expect(screen.getAllByRole('switch')).toHaveLength(ALL_SIX.length);
    for (const key of ALL_SIX) expect(switchFor(key)).toBeInTheDocument();
  });
});
