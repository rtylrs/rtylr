import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { createMemoryHistory, createRouter, RouterProvider } from '@tanstack/react-router';
import { routeTree } from './router';

const gate = vi.hoisted(() => ({ portalPaths: [] as string[] }));

// The account checks have their own suite; here the gate only has to wrap the
// staff pages and hand over the Portal path each one stands for.
vi.mock('../../shared/utils/BusinessRoute', async () => {
  const { Outlet, useRouterState } = await import('@tanstack/react-router');
  return {
    BusinessRoute: ({
      standalone,
    }: {
      standalone?: { portalPath: (pathname: string) => string };
    }) => {
      const pathname = useRouterState({ select: (state) => state.location.pathname });
      if (standalone) gate.portalPaths.push(standalone.portalPath(pathname));
      return <Outlet />;
    },
  };
});
vi.mock('./screens/Public/Landing', () => ({ default: () => <h1>Rennili landing</h1> }));
vi.mock('./screens/Public/HowItWorksPage', () => ({
  default: () => <h1>Rennili how it works</h1>,
}));
vi.mock('./screens/Public/FeaturesPage', () => ({ default: () => <h1>Rennili features</h1> }));
vi.mock('./screens/Public/FaqPage', () => ({ default: () => <h1>Rennili FAQ</h1> }));
vi.mock('./screens/Public/RtylrServicesPage', () => ({ default: () => <h1>RTYLR services</h1> }));
vi.mock('./screens/Public/Scan', () => ({ default: () => <h1>Pager scan</h1> }));
vi.mock('./screens/Public/Ticket', () => ({ default: () => <h1>Pager ticket</h1> }));
vi.mock('./screens/Public/Try', () => ({ default: () => <h1>Pager demo</h1> }));
vi.mock('./screens/Staff/Board', () => ({ default: () => <h1>Pager board</h1> }));
vi.mock('./screens/Staff/Poster', () => ({ default: () => <h1>Pager poster</h1> }));
vi.mock('./screens/Staff/Display', () => ({ default: () => <h1>Pager display</h1> }));
vi.mock('./screens/Staff/Setup', () => ({ default: () => <h1>Pager setup</h1> }));
vi.mock('./screens/Staff/Insights', () => ({ default: () => <h1>Pager insights</h1> }));

async function open(url: string) {
  const router = createRouter({
    routeTree,
    history: createMemoryHistory({ initialEntries: [url] }),
  });
  await router.load();
  render(<RouterProvider router={router} />);
  return router;
}

describe('Rennili routes', () => {
  it.each([
    ['/', 'Rennili landing'],
    ['/how-it-works', 'Rennili how it works'],
    ['/features', 'Rennili features'],
    ['/faq', 'Rennili FAQ'],
    ['/rtylr', 'RTYLR services'],
    ['/acme', 'Pager scan'],
    ['/t/tok-1', 'Pager ticket'],
    ['/try', 'Pager demo'],
  ])('serves %s with no sign in', async (url, heading) => {
    gate.portalPaths = [];
    const router = await open(url);
    expect(await screen.findByRole('heading', { name: heading })).toBeVisible();
    expect(router.state.location.pathname).toBe(url);
    expect(gate.portalPaths).toEqual([]);
  });

  it.each([
    ['/board', 'Pager board', '/pager'],
    ['/display', 'Pager display', '/pager/display'],
    ['/poster', 'Pager poster', '/pager/poster'],
    ['/setup', 'Pager setup', '/pager/setup'],
    ['/insights', 'Pager insights', '/pager/insights'],
  ])('serves %s behind the account checks', async (url, heading, portalPath) => {
    gate.portalPaths = [];
    await open(url);
    // The first of these loads the lazy account checks cold, which can take
    // longer than findBy's default second on a busy machine.
    expect(await screen.findByRole('heading', { name: heading }, { timeout: 5_000 })).toBeVisible();
    expect(gate.portalPaths).toContain(portalPath);
  });
});
