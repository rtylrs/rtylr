import { useRouterState } from '@tanstack/react-router';
import { useMemo } from 'react';
import { suggestionsFor } from '../../lib/caret';

/** Returns the suggestion array to show in Caret's empty state. */
export function useCaretSuggestions(): string[] {
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  return useMemo(() => suggestionsFor(pathname), [pathname]);
}
