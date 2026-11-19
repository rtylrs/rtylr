import { useCallback, useEffect, useState } from 'react';
import { useUserStore } from '../../../shared/stores/userStore';
import { loadListPrefs, saveListPrefs, type ListColumnPrefs } from './listColumns';

/**
 * The List's columns, widths and sort for the signed-in user on one project.
 * Kept in this browser (see listPrefsStorageKey); with no user id yet the
 * defaults apply and nothing is written.
 */
export function useListPrefs(
  projectId: string
): [ListColumnPrefs, (update: (prefs: ListColumnPrefs) => ListColumnPrefs) => void] {
  // Read straight from the store rather than through currentUser, whose other
  // hooks pull in the project member queries.
  const userId = useUserStore((state) =>
    state.user?.id === undefined ? null : String(state.user.id)
  );
  const [prefs, setPrefs] = useState<ListColumnPrefs>(() => loadListPrefs(userId, projectId));

  useEffect(() => {
    setPrefs(loadListPrefs(userId, projectId));
  }, [userId, projectId]);

  const update = useCallback(
    (change: (prefs: ListColumnPrefs) => ListColumnPrefs) => {
      setPrefs((current) => {
        const next = change(current);
        if (next !== current) saveListPrefs(userId, projectId, next);
        return next;
      });
    },
    [userId, projectId]
  );

  return [prefs, update];
}
