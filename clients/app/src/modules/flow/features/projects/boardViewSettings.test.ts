import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import {
  DEFAULT_BOARD_VIEW_SETTINGS,
  DEFAULT_CARD_FIELDS,
  boardViewStorageKey,
  loadBoardViewSettings,
  parseBoardViewSettings,
  saveBoardViewSettings,
} from './boardViewSettings';

beforeEach(() => {
  localStorage.clear();
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe('board view defaults', () => {
  it('reproduce the board as it was: no grouping, 14 days, fixed columns', () => {
    expect(DEFAULT_BOARD_VIEW_SETTINGS.groupBy).toBe('none');
    expect(DEFAULT_BOARD_VIEW_SETTINGS.hideDoneAfterDays).toBe(14);
    expect(DEFAULT_BOARD_VIEW_SETTINGS.columnWidth).toBe('fixed');
  });

  it('show the fields the card always showed and none it did not', () => {
    expect(DEFAULT_CARD_FIELDS).toEqual({
      key: true,
      priority: true,
      labels: true,
      assignee: true,
      subtasks: true,
      dueDate: true,
      estimate: false,
      cycle: false,
      milestone: false,
      // New with task relations, on so a blocked card says so out of the box.
      blocked: true,
    });
  });

  it('turns the blocked badge on for boards saved before it existed', () => {
    const parsed = parseBoardViewSettings({ cardFields: { estimate: true } });
    expect(parsed.cardFields.blocked).toBe(true);
  });
});

describe('board view storage', () => {
  it('keeps one person one project apart from everyone else', () => {
    const mine = { ...DEFAULT_BOARD_VIEW_SETTINGS, hideDoneAfterDays: 1 as const };
    saveBoardViewSettings('1', '2', mine);

    expect(loadBoardViewSettings('1', '2').hideDoneAfterDays).toBe(1);
    expect(loadBoardViewSettings('5', '2').hideDoneAfterDays).toBe(14);
    expect(loadBoardViewSettings('1', '3').hideDoneAfterDays).toBe(14);
    expect(localStorage.getItem(boardViewStorageKey('1', '2'))).not.toBeNull();
  });

  it('saves nothing before the user is known', () => {
    saveBoardViewSettings(null, '2', DEFAULT_BOARD_VIEW_SETTINGS);
    expect(localStorage.length).toBe(0);
  });

  it('carries over the group-by this browser kept per project', () => {
    localStorage.setItem('flow:board:groupBy:2', 'priority');
    expect(loadBoardViewSettings('1', '2').groupBy).toBe('priority');
    expect(loadBoardViewSettings(null, '2').groupBy).toBe('priority');
  });

  it('prefers the person own settings over the old group-by', () => {
    localStorage.setItem('flow:board:groupBy:2', 'priority');
    saveBoardViewSettings('1', '2', { ...DEFAULT_BOARD_VIEW_SETTINGS, groupBy: 'label' });
    expect(loadBoardViewSettings('1', '2').groupBy).toBe('label');
  });

  it('falls back to the defaults when storage throws', () => {
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('blocked');
    });
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('blocked');
    });
    expect(loadBoardViewSettings('1', '2')).toEqual(DEFAULT_BOARD_VIEW_SETTINGS);
    expect(() => saveBoardViewSettings('1', '2', DEFAULT_BOARD_VIEW_SETTINGS)).not.toThrow();
  });

  it('falls back to the defaults on a value that is not JSON', () => {
    localStorage.setItem(boardViewStorageKey('1', '2'), '{nope');
    expect(loadBoardViewSettings('1', '2')).toEqual(DEFAULT_BOARD_VIEW_SETTINGS);
  });
});

describe('parseBoardViewSettings', () => {
  it('keeps each valid field and defaults each invalid one', () => {
    const parsed = parseBoardViewSettings({
      groupBy: 'sprint',
      hideDoneAfterDays: 3,
      columnWidth: 'flexible',
      cardFields: { estimate: true, key: 'yes' },
    });
    expect(parsed.groupBy).toBe('none');
    expect(parsed.hideDoneAfterDays).toBe(14);
    expect(parsed.columnWidth).toBe('flexible');
    expect(parsed.cardFields).toEqual({ ...DEFAULT_CARD_FIELDS, estimate: true });
  });

  it('keeps never as a hide choice', () => {
    expect(parseBoardViewSettings({ hideDoneAfterDays: null }).hideDoneAfterDays).toBeNull();
  });

  it('keeps a custom field grouping and drops a malformed one', () => {
    expect(parseBoardViewSettings({ groupBy: 'field:12' }).groupBy).toBe('field:12');
    expect(parseBoardViewSettings({ groupBy: 'field:abc' }).groupBy).toBe('none');
  });
});
