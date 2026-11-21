import React from 'react';
import { useRouterState, type AnyRouter } from '@tanstack/react-router';
import { CookieBanner } from '@voxire/consent';
import { isPagerPublicPath } from '../../modules/pager/pagerPaths';

/**
 * The Portal's consent banner, left out on the pager's customer pages: it
 * covered the bottom of a page the customer keeps open while waiting, and its
 * copy is English only. Nothing is written there, so a visitor without a choice
 * stays undecided and `bootObservability` keeps optional analytics off; a
 * choice made earlier on another page still applies.
 *
 * It renders beside `RouterProvider`, not inside it, so it reads the router
 * it is handed.
 */
export function PortalCookieBanner({ router }: { router: AnyRouter }): React.JSX.Element | null {
  const pathname = useRouterState({ router, select: (state) => state.location.pathname });
  return isPagerPublicPath(pathname) ? null : <CookieBanner />;
}
