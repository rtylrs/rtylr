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
  it('renders Flow settings directly instead of the unavailable-app guard', async () => {
    await open('/settings/flow');
    expect(await screen.findByRole('heading', { name: 'Notification preferences' })).toBeVisible();
    expect(screen.queryByText('Coming soon')).toBeNull();
  });
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
      const router = await open(url + '?section=time');
      expect(
        await screen.findByRole('heading', {
          name: url.startsWith('/flow') ? 'Notification preferences' : 'Report preferences',
        })
      ).toBeVisible();
      expect(router.state.location.pathname).toBe(
        url.startsWith('/flow') ? '/settings/flow' : '/settings/insights'
      );
      if (url.startsWith('/flow'))
        expect(router.state.location.search).toEqual({ section: 'time' });
    }
  );
});
