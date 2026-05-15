import { useLocation } from 'react-router-dom';
import { useMemo } from 'react';
import { useUserStore } from '../../stores/userStore';
import { buildDashSystemContext } from '../../lib/caret';

/** Returns the dynamic system-context string for the current dash route. */
export function useCaretSystemContext(): string {
  const location = useLocation();
  const businessName = useUserStore((s) => s.user?.businessName);
  return useMemo(
    () => buildDashSystemContext({ pathname: location.pathname, businessName }),
    [location.pathname, businessName]
  );
}
