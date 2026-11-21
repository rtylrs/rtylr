import type { BoardGroupKey } from '../../shared/lib/boardUtils';
import {
  readUserProjectPref,
  userProjectStorageKey,
  writeUserProjectPref,
} from '../../shared/lib/userProjectStorage';
import { CLOSED_WINDOW_DAYS } from './statusUtils';

/**
 * One person's view of one project board: how it groups, what each card shows,
 * how long finished cards stay on it and how wide its columns are.
 *
 * These are personal, kept in this browser under the user id and the project
 * id (see userProjectStorage). Someone hiding done cards after a day leaves
 * everyone else's board as it was.
 */
export interface BoardViewSettings {
  groupBy: BoardGroupKey;
  cardFields: CardFields;
  /** Days a done or cancelled card stays on the board, or null for never hidden. */
  hideDoneAfterDays: HideDoneAfter;
  columnWidth: ColumnWidth;
}

/** Every grouping, in the order the pickers list them. */
export const BOARD_GROUP_KEYS: readonly BoardGroupKey[] = [
  'none',
  'assignee',
  'priority',
  'due',
  'label',
  'cycle',
  'milestone',
  'epic',
];

/** Optional card fields, in the order the settings list them. */
export const CARD_FIELDS = [
  'key',
  'priority',
  'labels',
  'assignee',
  'subtasks',
  'dueDate',
  'estimate',
  'cycle',
  'milestone',
] as const;

export type CardField = (typeof CARD_FIELDS)[number];
export type CardFields = Record<CardField, boolean>;

/** The card as it looked before any of this was a choice. */
export const DEFAULT_CARD_FIELDS: CardFields = {
  key: true,
  priority: true,
  labels: true,
  assignee: true,
  subtasks: true,
  dueDate: true,
  estimate: false,
  cycle: false,
  milestone: false,
};

export const HIDE_DONE_CHOICES = [1, 7, 14, 30, null] as const;
export type HideDoneAfter = (typeof HIDE_DONE_CHOICES)[number];

export const COLUMN_WIDTHS = ['fixed', 'flexible'] as const;
export type ColumnWidth = (typeof COLUMN_WIDTHS)[number];

export const DEFAULT_BOARD_VIEW_SETTINGS: BoardViewSettings = {
  groupBy: 'none',
  cardFields: DEFAULT_CARD_FIELDS,
  // The board has hidden closed cards older than this since columns became
  // statuses; the setting starts where the board already was.
  hideDoneAfterDays: CLOSED_WINDOW_DAYS,
  columnWidth: 'fixed',
};

/** The name these settings are kept under (see userProjectStorage). */
export const BOARD_VIEW_PREF = 'board:view';

export function boardViewStorageKey(userId: string, projectId: string): string {
  return userProjectStorageKey(BOARD_VIEW_PREF, userId, projectId);
}

/** Group-by was kept per project, for whoever used the browser, before this. */
const legacyGroupByKey = (projectId: string): string => `flow:board:groupBy:${projectId}`;

const isGroupKey = (value: unknown): value is BoardGroupKey =>
  typeof value === 'string' && (BOARD_GROUP_KEYS as readonly string[]).includes(value);

/**
 * Settings from whatever was stored, field by field. A value that is missing or
 * no longer valid falls back to its default instead of discarding the rest, so
 * adding a field later does not reset anyone's board.
 */
export function parseBoardViewSettings(raw: unknown): BoardViewSettings {
  const d = DEFAULT_BOARD_VIEW_SETTINGS;
  if (raw === null || typeof raw !== 'object') return d;
  const r = raw as Record<string, unknown>;
  const storedFields =
    r.cardFields !== null && typeof r.cardFields === 'object'
      ? (r.cardFields as Record<string, unknown>)
      : {};
  const cardFields = { ...DEFAULT_CARD_FIELDS };
  for (const field of CARD_FIELDS) {
    if (typeof storedFields[field] === 'boolean') cardFields[field] = storedFields[field];
  }
  const hide = r.hideDoneAfterDays;
  return {
    groupBy: isGroupKey(r.groupBy) ? r.groupBy : d.groupBy,
    cardFields,
    hideDoneAfterDays: (HIDE_DONE_CHOICES as readonly unknown[]).includes(hide)
      ? (hide as HideDoneAfter)
      : d.hideDoneAfterDays,
    columnWidth:
      r.columnWidth === 'flexible' || r.columnWidth === 'fixed' ? r.columnWidth : d.columnWidth,
  };
}

/**
 * Settings from what this person has stored for the project (undefined when
 * nothing). With nothing stored, or before the user is known, the board shows
 * the defaults with any group-by this browser remembered for the project.
 */
export function boardViewSettingsFrom(stored: unknown, projectId: string): BoardViewSettings {
  if (!projectId) return DEFAULT_BOARD_VIEW_SETTINGS;
  if (stored !== undefined) return parseBoardViewSettings(stored);
  try {
    const legacy = localStorage.getItem(legacyGroupByKey(projectId));
    if (isGroupKey(legacy)) return { ...DEFAULT_BOARD_VIEW_SETTINGS, groupBy: legacy };
  } catch {
    // Storage blocked: the defaults stand.
  }
  return DEFAULT_BOARD_VIEW_SETTINGS;
}

/** This person's settings for this project. */
export function loadBoardViewSettings(userId: string | null, projectId: string): BoardViewSettings {
  return boardViewSettingsFrom(readUserProjectPref(BOARD_VIEW_PREF, userId, projectId), projectId);
}

export function saveBoardViewSettings(
  userId: string | null,
  projectId: string,
  settings: BoardViewSettings
): void {
  writeUserProjectPref(BOARD_VIEW_PREF, userId, projectId, settings);
}
