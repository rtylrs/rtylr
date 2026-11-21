import { useUserStore } from '../../../shared/stores/userStore';
import { useUserProjectPref } from '../../../shared/lib/userProjectStorage';
import { LIST_COLUMNS_PREF, normalizeListPrefs, type ListColumnPrefs } from './listColumns';

const fromStored = (stored: unknown): ListColumnPrefs => normalizeListPrefs(stored ?? null);

/**
 * The List's columns, widths and sort for the signed-in user on one project.
 * Kept in this browser (see userProjectStorage); with no user id yet the
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
  return useUserProjectPref(LIST_COLUMNS_PREF, userId, projectId, fromStored);
}
