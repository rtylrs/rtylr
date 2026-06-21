import { useRouterState } from '@tanstack/react-router';
import { useMemo } from 'react';
import { useUserStore } from '../../stores/userStore';
import { buildDashPageContext, buildDashSystemContext } from '../../lib/caret';
import type { CaretPageContext } from '@voxire/caret';

/** Returns the dynamic system-context string for the current dash route. */
export function useCaretSystemContext(): string {
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const search = useRouterState({ select: (s) => s.location.searchStr });
  const user = useUserStore((s) => s.user);
  return useMemo(
    () =>
      buildDashSystemContext({
        pathname: pathname,
        search,
        businessName: user?.businessName,
        role: user?.role,
        apps: user?.apps,
      }),
    [pathname, search, user?.businessName, user?.role, user?.apps]
  );
}

export function useCaretPageContext(): CaretPageContext {
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const search = useRouterState({ select: (s) => s.location.searchStr });
  const user = useUserStore((s) => s.user);
  return useMemo(
    () =>
      buildDashPageContext({
        pathname: pathname,
        search,
        businessName: user?.businessName,
        role: user?.role,
        apps: user?.apps,
      }),
    [pathname, search, user?.businessName, user?.role, user?.apps]
  );
}
