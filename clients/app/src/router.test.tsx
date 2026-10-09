import path from 'node:path';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import {
  createRouter,
  createMemoryHistory,
  RouterProvider,
  useParams,
} from '@tanstack/react-router';
import { initTestI18nFrom } from '../test/i18nTestInstance';
import { createEffectiveAccess } from '../test/effectiveAccess';
import { useUserStore } from './shared/stores/userStore';
import { router as appRouter } from './router';
import { useSettingsModalStore } from './features/settings/settingsModalStore';
import { modalRouteTree } from './features/settings/SettingsModal';

vi.mock('./shared/utils/BusinessRoute', async () => {
  const { Outlet } = await import('@tanstack/react-router');
  return { BusinessRoute: () => <Outlet /> };
});
vi.mock('./modules/flow/layouts/FlowLayout', () => ({
  default: ({ children }: { children: React.ReactNode }) => (
    <section aria-label="Flow app">{children}</section>
  ),
}));
vi.mock('./modules/flow/features/settings/FlowSettings', () => ({
  default: () => <h1>Notification preferences</h1>,
}));
vi.mock('./modules/insights/features/settings/Settings', () => ({
  default: () => <h1>Report preferences</h1>,
}));
vi.mock('./modules/flow/features/tasks/MyIssues', () => ({ default: () => <h1>My tasks</h1> }));
vi.mock('./modules/flow/features/projects/ProjectBoard', () => ({
  default: () => <h1>Project {useParams({ strict: false }).projectId}</h1>,
}));
vi.mock('./modules/flow/features/Start', () => ({ default: () => <h1>Create in Flow</h1> }));
const locales = path.resolve(import.meta.dirname, 'locales');
beforeEach(async () => {
  await initTestI18nFrom(locales);
  useSettingsModalStore.setState({ path: null });
  useUserStore.setState({
    user: { effectiveAccess: createEffectiveAccess('flow', { isOwner: true }) },
  });
});
async function open(url: string) {
  const router = createRouter({
    routeTree: appRouter.options.routeTree,
    history: createMemoryHistory({ initialEntries: [url] }),
  });
  await router.load();
  render(<RouterProvider router={router} />);
  return router;
}
describe('Portal native app routes', () => {
  it('opens settings in the modal instead of navigating to a settings page', async () => {
    const router = await open('/settings/flow');
    expect(useSettingsModalStore.getState().path).toBe('/settings/flow');
    // A direct visit has no page behind it, so it lands on home.
    expect(router.state.location.pathname).toBe('/');
    expect(screen.queryByRole('heading', { name: 'Notification preferences' })).toBeNull();
  });
  it.each(['/security', '/api-keys', '/webhooks'])(
    'opens %s in the settings modal',
    async (url) => {
      await open(url);
      expect(useSettingsModalStore.getState().path).toBe(url);
    }
  );
  it('keeps Flow project deep links under their native prefix', async () => {
    await open('/flow/p/42');
    expect(await screen.findByRole('heading', { name: 'Project 42' })).toBeVisible();
    expect(screen.getByRole('region', { name: 'Flow app' })).toBeVisible();
  });
  it('lands read-only members in their task list', async () => {
    useUserStore.setState({
      user: {
        effectiveAccess: createEffectiveAccess('flow', { permissions: ['flow:tasks:view'] }),
      },
    });
    const router = await open('/flow');
    expect(await screen.findByRole('heading', { name: 'My tasks' })).toBeVisible();
    expect(router.state.location.pathname).toBe('/flow/my-issues');
  });
  it('lands write-only members in their permitted actions', async () => {
    useUserStore.setState({
      user: {
        effectiveAccess: createEffectiveAccess('flow', { permissions: ['flow:tasks:create'] }),
      },
    });
    const router = await open('/flow');
    expect(await screen.findByRole('heading', { name: 'Create in Flow' })).toBeVisible();
    expect(router.state.location.pathname).toBe('/flow/start');
  });
  it.each(['/flow/settings', '/insights/settings'])(
    'sends the old settings bookmark %s to shared settings',
    async (url) => {
      await open(url + '?section=time');
      expect(useSettingsModalStore.getState().path).toBe(
        url.startsWith('/flow') ? '/settings/flow?section=time' : '/settings/insights'
      );
    }
  );
  // Settings sections look routes up by id (FlowSettings uses
  // getRouteApi('/business/settings/flow')), so both routers must agree on them.
  it('gives every settings route the same id in the settings modal router', () => {
    const modalRouter = createRouter({ routeTree: modalRouteTree });
    const settingsIds = Object.keys(appRouter.routesById).filter((id) =>
      /^\/business\/(settings|security|api-keys|webhooks)(\/|$)/.test(id)
    );
    expect(settingsIds.length).toBeGreaterThan(10);
    expect(settingsIds.filter((id) => !(id in modalRouter.routesById))).toEqual([]);
  });
});
