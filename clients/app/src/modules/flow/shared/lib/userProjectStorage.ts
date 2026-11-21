import { useCallback, useRef, useState } from 'react';

/**
 * One person's view preference for one project, kept in this browser.
 *
 * Flow has no server store for a personal, per project view preference (the
 * user settings contract carries notification switches and a sidebar blob
 * only), so views such as the board and the List keep theirs here, keyed by
 * the signed in user and the project. One person's choice never changes
 * anyone else's view, and nothing is written before the user is known.
 *
 * `view` names the preference, e.g. `board:view` or `list:columns`.
 */
export function userProjectStorageKey(view: string, userId: string, projectId: string): string {
  return `flow:${view}:${userId}:${projectId}`;
}

/**
 * The stored value, parsed. Undefined when the user or project is not known
 * yet, nothing is stored, storage is blocked or the value is not JSON: the
 * caller's defaults stand in every one of those cases.
 */
export function readUserProjectPref(
  view: string,
  userId: string | null,
  projectId: string
): unknown {
  if (!userId || !projectId) return undefined;
  try {
    const raw = localStorage.getItem(userProjectStorageKey(view, userId, projectId));
    return raw === null ? undefined : (JSON.parse(raw) as unknown);
  } catch {
    return undefined;
  }
}

export function writeUserProjectPref(
  view: string,
  userId: string | null,
  projectId: string,
  value: unknown
): void {
  if (!userId || !projectId) return;
  try {
    localStorage.setItem(userProjectStorageKey(view, userId, projectId), JSON.stringify(value));
  } catch {
    // Storage full or blocked: the choice still applies until the page reloads.
  }
}

/**
 * A preference as React state. `load` turns what is stored (undefined when
 * nothing is) into the preference, so each view keeps its own validation and
 * defaults. When the user or project changes the new scope is read during
 * render, so the view never paints a frame with the previous one. A change is
 * saved at once; one that returns the current value writes nothing.
 */
export function useUserProjectPref<T>(
  view: string,
  userId: string | null,
  projectId: string,
  load: (stored: unknown) => T
): [T, (change: (current: T) => T) => void] {
  const scope = `${userId ?? ''}:${projectId}`;
  const [state, setState] = useState(() => ({
    scope,
    value: load(readUserProjectPref(view, userId, projectId)),
  }));

  let value = state.value;
  if (state.scope !== scope) {
    value = load(readUserProjectPref(view, userId, projectId));
    setState({ scope, value });
  }

  // Several changes in one event build on each other, not on the value the
  // last render saw.
  const latest = useRef(value);
  latest.current = value;

  const update = useCallback(
    (change: (current: T) => T) => {
      const current = latest.current;
      const next = change(current);
      if (Object.is(next, current)) return;
      latest.current = next;
      setState({ scope, value: next });
      writeUserProjectPref(view, userId, projectId, next);
    },
    [view, scope, userId, projectId]
  );

  return [value, update];
}
