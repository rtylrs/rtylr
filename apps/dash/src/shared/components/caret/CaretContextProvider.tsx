import React, { useMemo } from 'react';
import { useRouterState } from '@tanstack/react-router';
import type { CaretPageContext } from '@voxire/caret';
import { useUserStore } from '../../stores/userStore';
import { buildDashPageContext, buildDashSystemContext, suggestionsFor } from '../../lib/caret';

interface CaretContextValue {
  systemContext: string;
  pageContext: CaretPageContext;
  suggestions: string[];
  userId: number;
  businessName: string | undefined;
}

const CaretContext = React.createContext<CaretContextValue | null>(null);

export const CaretContextProvider: React.FC<React.PropsWithChildren> = ({ children }) => {
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const search = useRouterState({ select: (s) => s.location.searchStr });
  const { user } = useUserStore();

  const value = useMemo<CaretContextValue>(() => {
    const businessName = user?.businessName;
    return {
      systemContext: buildDashSystemContext({
        pathname: pathname,
        search,
        businessName,
        role: user?.role,
        apps: user?.apps,
      }),
      pageContext: buildDashPageContext({
        pathname: pathname,
        search,
        businessName,
        role: user?.role,
        apps: user?.apps,
      }),
      suggestions: suggestionsFor(pathname),
      userId: Number(user?.id) || 0,
      businessName,
    };
  }, [pathname, search, user?.apps, user?.businessName, user?.id, user?.role]);

  return <CaretContext.Provider value={value}>{children}</CaretContext.Provider>;
};

export function useCaretContext(): CaretContextValue {
  const ctx = React.useContext(CaretContext);
  if (!ctx) {
    throw new Error('useCaretContext must be used within CaretContextProvider');
  }
  return ctx;
}
