import { useLocation } from 'react-router-dom';
import { useMemo } from 'react';
import { suggestionsFor } from '../../lib/caret';

/** Returns the suggestion array to show in Caret's empty state. */
export function useCaretSuggestions(): string[] {
  const location = useLocation();
  return useMemo(() => suggestionsFor(location.pathname), [location.pathname]);
}
