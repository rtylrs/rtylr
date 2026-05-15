import React, { useMemo } from 'react';
import { useLocation } from 'react-router-dom';
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
  const location = useLocation();
  const { user } = useUserStore();

  const value = useMemo<CaretContextValue>(() => {
    const businessName = user?.businessName;
    return {
      systemContext: buildDashSystemContext({
        pathname: location.pathname,
        businessName,
      }),
      suggestions: suggestionsFor(location.pathname),
      userId: Number(user?.id) || 0,
      businessName,
    };
  }, [location.pathname, user?.businessName, user?.id]);

  return <CaretContext.Provider value={value}>{children}</CaretContext.Provider>;
};

export function useCaretContext(): CaretContextValue {
  const ctx = React.useContext(CaretContext);
  if (!ctx) {
    throw new Error('useCaretContext must be used within CaretContextProvider');
  }
  return ctx;
}
