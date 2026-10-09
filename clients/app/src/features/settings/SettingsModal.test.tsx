import path from 'node:path';
import { act, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  RouterProvider,
} from '@tanstack/react-router';
import { initTestI18nFrom } from '../../../test/i18nTestInstance';
import { createEffectiveAccess } from '../../../test/effectiveAccess';
import { useUserStore } from '../../shared/stores/userStore';
import SettingsModal from './SettingsModal';
import { useSettingsModalStore } from './settingsModalStore';

vi.mock('../../modules/flow/features/settings/FlowSettings', () => ({
  default: () => <h1>Notification preferences</h1>,
}));
vi.mock('../../shared/hooks/queries/useAddons', () => ({
  useAddons: () => ({ data: [] }),
  isAddonEnabled: () => false,
}));

const locales = path.resolve(import.meta.dirname, '../../locales');

// The page behind the modal: a minimal app router with one page.
async function renderOverPage() {
  const root = createRootRoute({ component: () => <SettingsModal /> });
  const home = createRoute({ getParentRoute: () => root, path: '/', component: () => null });
  const team = createRoute({ getParentRoute: () => root, path: '/team', component: () => null });
  const app = createRouter({
    routeTree: root.addChildren([home, team]),
    history: createMemoryHistory({ initialEntries: ['/'] }),
  });
  await app.load();
  render(<RouterProvider router={app} />);
  return app;
}

describe('SettingsModal', () => {
  beforeEach(async () => {
    await initTestI18nFrom(locales);
    useUserStore.setState({
      user: { effectiveAccess: createEffectiveAccess('flow', { isOwner: true }) },
    });
    useSettingsModalStore.setState({ path: null });
  });

  it('renders the requested settings section in a dialog over the page', async () => {
    const app = await renderOverPage();
    act(() => useSettingsModalStore.getState().open('/settings/flow?section=time'));

    const dialog = await screen.findByRole('dialog');
    expect(await screen.findByRole('heading', { name: 'Notification preferences' })).toBeVisible();
    expect(dialog).toContainElement(
      screen.getByRole('heading', { name: 'Notification preferences' })
    );
    expect(app.state.location.pathname).toBe('/');
  });

  it('closes and hands non-settings locations to the app router', async () => {
    const app = await renderOverPage();
    act(() => useSettingsModalStore.getState().open('/team'));

    await vi.waitFor(() => expect(useSettingsModalStore.getState().path).toBeNull());
    await vi.waitFor(() => expect(app.state.location.pathname).toBe('/team'));
    expect(screen.queryByRole('dialog')).toBeNull();
  });
});
