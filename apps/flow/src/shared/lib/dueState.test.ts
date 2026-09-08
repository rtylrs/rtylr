import { describe, it, expect, beforeAll, afterAll, vi } from 'vitest';
import { contrastRatio, compositeOver } from '@voxire/utils';
import { dueState, isOverdue } from './dateUtils';

/**
 * Three components each hand-rolled `new Date(task.dueDate) < new Date()`,
 * which parses a bare `YYYY-MM-DD` as UTC midnight and compares it to the
 * current instant. A task due **today** therefore read as late from 00:00 UTC,
 * which is the whole working day at UTC+3 (rtylr#1077).
 */
describe('dueState', () => {
  beforeAll(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-05-02T12:00:00.000Z'));
  });
  afterAll(() => {
    vi.useRealTimers();
  });

  it('separates late, today and future', () => {
    expect(dueState('2026-05-01')).toBe('late');
    expect(dueState('2026-05-02')).toBe('today');
    expect(dueState('2026-05-03')).toBe('future');
  });

  it('treats an absent due date as its own state, not as future', () => {
    expect(dueState('')).toBe('none');
    expect(dueState(null)).toBe('none');
    expect(dueState(undefined)).toBe('none');
  });

  it('reads a timestamped value as the day it names', () => {
    expect(dueState('2026-05-02T23:59:00Z')).toBe('today');
    expect(dueState('2026-05-02T00:00:00Z')).toBe('today');
  });

  it('shows the comparison that was wrong', () => {
    // Not an assertion about the helper, an assertion about why it exists. At
    // this pinned instant the old idiom calls a task due today overdue.
    expect(new Date('2026-05-02') < new Date()).toBe(true);
    expect(dueState('2026-05-02')).toBe('today');
  });
});

describe('isOverdue', () => {
  beforeAll(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-05-02T12:00:00.000Z'));
  });
  afterAll(() => {
    vi.useRealTimers();
  });

  it('is late and unfinished', () => {
    expect(isOverdue('2026-05-01', 'todo')).toBe(true);
    expect(isOverdue('2026-05-02', 'todo')).toBe(false);
  });

  it('is never true for finished work, whatever the date', () => {
    // The three call sites each decided separately what finished meant, and one
    // of them counted only `done`.
    expect(isOverdue('2026-01-01', 'done')).toBe(false);
    expect(isOverdue('2026-01-01', 'cancelled')).toBe(false);
  });
});

describe('the tokens a due date is painted with', () => {
  const CARD = { dark: '#161616', light: '#ffffff' };

  /** `--flow-text-dim` as it lands, composited over each mode's card. */
  const DIM = {
    dark: compositeOver('#f0f0ee', 0.28, CARD.dark),
    light: compositeOver('#09090b', 0.35, CARD.light),
  };
  const MUTED = {
    dark: compositeOver('#f0f0ee', 0.58, CARD.dark),
    light: compositeOver('#09090b', 0.62, CARD.light),
  };

  it('measures why the dim token cannot carry a due date', () => {
    // What TaskRow used for every non-overdue date. globals.css says in as many
    // words that this token never carries text, and it is right.
    for (const mode of ['dark', 'light'] as const) {
      expect(contrastRatio(DIM[mode], CARD[mode]), `text-dim on the ${mode} card`).toBeLessThan(
        4.5
      );
    }
  });

  it('confirms the muted token does carry it', () => {
    for (const mode of ['dark', 'light'] as const) {
      expect(
        contrastRatio(MUTED[mode], CARD[mode]),
        `text-muted on the ${mode} card`
      ).toBeGreaterThanOrEqual(4.5);
    }
  });
});
