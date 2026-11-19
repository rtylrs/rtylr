import { useState } from 'react';
import {
  DEFAULT_BOARD_VIEW_SETTINGS,
  loadBoardViewSettings,
  saveBoardViewSettings,
  type BoardViewSettings,
} from './boardViewSettings';

export interface BoardViewSettingsState {
  settings: BoardViewSettings;
  /** Change some settings; the rest stay as they are. Saved at once. */
  update: (patch: Partial<BoardViewSettings>) => void;
  /** Back to the default board. */
  reset: () => void;
}

/**
 * The signed-in person's board settings for one project, read from and saved
 * to this browser. The user id is passed in rather than read here, because the
 * board already has it from its filters and tests stub that one source.
 */
export function useBoardViewSettings(
  projectId: string,
  userId: string | null
): BoardViewSettingsState {
  const scope = `${userId ?? ''}:${projectId}`;
  const [state, setState] = useState(() => ({
    scope,
    settings: loadBoardViewSettings(userId, projectId),
  }));

  // Another project, or the user became known: read that scope's settings
  // during render, so the board never paints one frame with the old ones.
  let settings = state.settings;
  if (state.scope !== scope) {
    settings = loadBoardViewSettings(userId, projectId);
    setState({ scope, settings });
  }

  const write = (next: BoardViewSettings): void => {
    setState({ scope, settings: next });
    saveBoardViewSettings(userId, projectId, next);
  };

  return {
    settings,
    update: (patch) => write({ ...settings, ...patch }),
    reset: () => write(DEFAULT_BOARD_VIEW_SETTINGS),
  };
}
