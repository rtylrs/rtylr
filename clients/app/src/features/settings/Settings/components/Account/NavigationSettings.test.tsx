import path from 'node:path';
import { cleanup, render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import { initTestI18nFrom } from '../../../../../../test/i18nTestInstance';
import NavigationSettings from './NavigationSettings';

vi.mock('../../../../../shared/stores/userStore', async () => {
  const { createEffectiveAccess } = await import('../../../../../../test/effectiveAccess');
  const user = {
    id: 1,
    effectiveAccess: createEffectiveAccess(undefined, { appIds: ['crm', 'hr'], isOwner: true }),
  };
  return { useUserStore: (selector: (state: unknown) => unknown) => selector({ user }) };
});
vi.mock('../../../../../shared/hooks/queries/useAddons', () => ({
  useAddons: () => ({ data: [] }),
  isAddonEnabled: () => false,
}));
const api = vi.hoisted(() => ({
  save: vi.fn(),
  reset: vi.fn(),
  layout: {
    rail: {
      top: ['home', 'insights', 'flow', 'crm', 'erp', 'finance', 'hr', 'pos', 'automations'],
      bottom: ['caret'],
      hidden: [] as string[],
      customized: false,
    },
  },
}));
vi.mock('../../../../../shell/portalLayout', async (original) => ({
  ...(await original<typeof import('../../../../../shell/portalLayout')>()),
  usePortalLayout: () => api.layout,
  useSavePortalLayout: () => ({ mutate: api.save, isPending: false }),
  useResetPortalLayout: () => ({ mutate: api.reset, isPending: false }),
}));

beforeAll(() => initTestI18nFrom(path.resolve(import.meta.dirname, '../../../../../locales')));
afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

function renderSettings() {
  render(
    <QueryClientProvider client={new QueryClient()}>
      <NavigationSettings />
    </QueryClientProvider>
  );
  return within(screen.getByRole('region', { name: 'App bar' }));
}

describe('NavigationSettings', () => {
  it('moves Caret from the bottom of the app bar to the end of the apps', async () => {
    const appBar = renderSettings();
    expect(appBar.getByText('Bottom of the bar')).toBeInTheDocument();
    await userEvent.click(appBar.getByRole('button', { name: 'Move Caret up' }));
    expect(api.save).toHaveBeenCalledWith(
      {
        rail: {
          top: ['home', 'insights', 'flow', 'crm', 'erp', 'finance', 'hr', 'pos', 'automations', 'caret'],
          bottom: [],
          hidden: [],
        },
      },
      expect.anything()
    );
  });

  it('hides an app but never Home, and restores only a customized bar', async () => {
    const appBar = renderSettings();
    expect(appBar.getByRole('button', { name: 'Hide Home' })).toBeDisabled();
    expect(appBar.getByRole('button', { name: 'Restore default' })).toBeDisabled();
    await userEvent.click(appBar.getByRole('button', { name: 'Hide HR' }));
    expect(api.save.mock.calls[0][0].rail.hidden).toEqual(['hr']);
  });
});
