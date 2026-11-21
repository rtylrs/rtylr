import { useUserProjectPref } from '../../shared/lib/userProjectStorage';
import {
  BOARD_VIEW_PREF,
  DEFAULT_BOARD_VIEW_SETTINGS,
  boardViewSettingsFrom,
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
  const [settings, change] = useUserProjectPref(BOARD_VIEW_PREF, userId, projectId, (stored) =>
    boardViewSettingsFrom(stored, projectId)
  );

  return {
    settings,
    update: (patch) => change((current) => ({ ...current, ...patch })),
    reset: () => change(() => DEFAULT_BOARD_VIEW_SETTINGS),
  };
}
