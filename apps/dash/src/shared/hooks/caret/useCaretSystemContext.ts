import { useRouterState } from '@tanstack/react-router';
import { useMemo } from 'react';
import { useUserStore } from '../../stores/userStore';
import { buildDashSystemContext } from '../../lib/caret';

/** Returns the dynamic system-context string for the current dash route. */
export function useCaretSystemContext(): string {
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const businessName = useUserStore((s) => s.user?.businessName);
  return useMemo(
    () => buildDashSystemContext({ pathname: pathname, businessName }),
    [pathname, businessName]
  );
}
