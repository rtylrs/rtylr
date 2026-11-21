import React from 'react';
import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
} from '@tanstack/react-router';
import { act, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import { readConsent } from '@voxire/consent';

import { PortalCookieBanner } from './PortalCookieBanner';

function routerAt(path: string) {
  const rootRoute = createRootRoute();
  const page = (routePath: string) =>
    createRoute({ getParentRoute: () => rootRoute, path: routePath, component: () => null });
  return createRouter({
    routeTree: rootRoute.addChildren([
      page('/'),
      page('/pager'),
      page('/$slug/pager'),
      page('/$slug/menu'),
      page('/pager/t/$token'),
    ]),
    history: createMemoryHistory({ initialEntries: [path] }),
  });
}

async function renderAt(path: string) {
  const router = routerAt(path);
  await act(() => router.load());
  render(<PortalCookieBanner router={router} />);
  return router;
}

const banner = () => screen.queryByRole('dialog', { name: 'Cookie consent' });

afterEach(() => {
  document.cookie = 'rtylr-consent=; Max-Age=0; Path=/';
});

describe('Portal cookie banner', () => {
  it.each(['/acme/pager', '/pager/t/abc123'])(
    'stays hidden on the public pager page %s',
    async (path) => {
      await renderAt(path);
      // The banner mounts hidden and decides in an effect; let it settle.
      await act(async () => undefined);

      expect(banner()).not.toBeInTheDocument();
      // No choice is recorded for the visitor, so optional analytics stay off.
      expect(readConsent()).toBe('undecided');
    }
  );

  it.each(['/', '/pager', '/acme/menu'])('still asks on %s', async (path) => {
    await renderAt(path);

    expect(await screen.findByRole('dialog', { name: 'Cookie consent' })).toBeInTheDocument();
  });

  it('comes back when the visitor leaves the pager', async () => {
    const router = await renderAt('/acme/pager');
    await act(async () => undefined);
    expect(banner()).not.toBeInTheDocument();

    await act(() => router.navigate({ to: '/$slug/menu', params: { slug: 'acme' } }));

    expect(await screen.findByRole('dialog', { name: 'Cookie consent' })).toBeInTheDocument();
  });
});
