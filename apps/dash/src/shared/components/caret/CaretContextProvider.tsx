import React, { useMemo } from 'react';
import { useRouterState } from '@tanstack/react-router';
import { useUserStore } from '../../stores/userStore';
import { buildDashSystemContext, suggestionsFor } from '../../lib/caret';

interface CaretContextValue {
  systemContext: string;
  suggestions: string[];
  userId: number;
  businessName: string | undefined;
}

const CaretContext = React.createContext<CaretContextValue | null>(null);

export const CaretContextProvider: React.FC<React.PropsWithChildren> = ({ children }) => {
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const { user } = useUserStore();

  const value = useMemo<CaretContextValue>(() => {
    const businessName = user?.businessName;
    return {
      systemContext: buildDashSystemContext({
        pathname: pathname,
        businessName,
      }),
      suggestions: suggestionsFor(pathname),
      userId: Number(user?.id) || 0,
      businessName,
    };
  }, [pathname, user?.businessName, user?.id]);

  return <CaretContext.Provider value={value}>{children}</CaretContext.Provider>;
};

export function useCaretContext(): CaretContextValue {
  const ctx = React.useContext(CaretContext);
  if (!ctx) {
    throw new Error('useCaretContext must be used within CaretContextProvider');
  }
  return ctx;
}
