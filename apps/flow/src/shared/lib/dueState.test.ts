import { describe, it, expect, beforeAll, afterAll, vi } from 'vitest';
import fs from 'node:fs';
import path from 'node:path';
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

describe('nothing compares a stored day against the current instant', () => {
  /**
   * The source check the three earlier fixes did not have.
   *
   * `new Date(storedDay) < new Date()` parses a date-only string as UTC
   * midnight and compares it to an instant, so a value due **today** reads as
   * past from 00:00 UTC onward (rtylr#1077). Three components had it, and a
   * fourth and fifth survived the sweep that fixed them: `DueDatePicker` because
   * its variable is called `value` rather than `dueDate`, and `ProjectTimeline`
   * because it parsed into a `Date` on one line and compared on another.
   *
   * A grep for the exact expression could not have found either. This looks for
   * the shape instead: a comparison against a bare `new Date()`.
   */
  const SRC = path.resolve(__dirname, '../..');

  const walk = (dir: string): string[] =>
    fs.readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
      const full = path.join(dir, entry.name);
      if (entry.isDirectory()) {
        return entry.name === 'node_modules' || entry.name === 'locales' ? [] : walk(full);
      }
      return /\.tsx?$/.test(entry.name) && !/\.test\.tsx?$/.test(entry.name) ? [full] : [];
    });

  it('leaves no comparison of a date against a bare new Date()', () => {
    const offenders: string[] = [];

    for (const file of walk(SRC)) {
      const lines = fs.readFileSync(file, 'utf8').split('\n');
      lines.forEach((line, i) => {
        const trimmed = line.trim();
        if (trimmed.startsWith('*') || trimmed.startsWith('//')) return;
        // `< new Date()` or `> new Date()`, whatever is on the left.
        //
        // The lookbehind excludes `=`, because an arrow function contains `>`
        // and `useState(() => new Date())` matched without it. Same trap as the
        // motion guard in rtylr#1257, where `onClick={(e) => …}` truncated a
        // tag capture.
        //
        // `dateUtils` itself is allowed: it is where the correct comparison
        // lives, and its own comments quote the wrong one.
        if (!/(?<!=)[<>]=?\s*new Date\(\s*\)/.test(line)) return;
        if (file.endsWith('dateUtils.ts')) return;
        offenders.push(`${path.relative(SRC, file)}:${i + 1}`);
      });
    }

    expect(offenders, 'use isOverdue or dueState, which compare stored days').toEqual([]);
  });
});
