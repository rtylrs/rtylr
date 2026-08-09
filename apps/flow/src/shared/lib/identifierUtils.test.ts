import { describe, it, expect } from 'vitest';

import {
  makeTaskIdentifier,
  identifierToNumber,
  sortIdentifiers,
  isValidIdentifier,
  normalizeIdentifier,
} from './identifierUtils';

describe('makeTaskIdentifier', () => {
  it('creates identifier from project name and count', () => {
    expect(makeTaskIdentifier('Flow', 42)).toBe('FLOW-42');
  });

  it('truncates project name to 4 chars', () => {
    expect(makeTaskIdentifier('LongProjectName', 1)).toBe('LONG-1');
  });

  it('falls back to TASK for empty name', () => {
    expect(makeTaskIdentifier('', 1)).toBe('TASK-1');
  });
});

describe('identifierToNumber', () => {
  it('extracts the numeric part', () => {
    expect(identifierToNumber('FLOW-42')).toBe(42);
  });

  it('handles multi-segment identifiers', () => {
    expect(identifierToNumber('TASK-123')).toBe(123);
  });

  it('returns 0 for non-numeric suffix', () => {
    expect(identifierToNumber('FLOW-abc')).toBe(0);
  });
});

describe('sortIdentifiers', () => {
  it('sorts numerically, not lexicographically', () => {
    const ids = ['FLOW-10', 'FLOW-2', 'FLOW-1', 'FLOW-9'];
    expect(sortIdentifiers(ids)).toEqual(['FLOW-1', 'FLOW-2', 'FLOW-9', 'FLOW-10']);
  });

  it('does not mutate the original array', () => {
    const ids = ['FLOW-2', 'FLOW-1'];
    const original = [...ids];
    sortIdentifiers(ids);
    expect(ids).toEqual(original);
  });
});

describe('isValidIdentifier', () => {
  it('returns true for valid identifiers', () => {
    expect(isValidIdentifier('FLOW-1')).toBe(true);
    expect(isValidIdentifier('TASK-999')).toBe(true);
    expect(isValidIdentifier('AB-10')).toBe(true);
  });

  it('returns false for invalid identifiers', () => {
    expect(isValidIdentifier('flow-1')).toBe(false);
    expect(isValidIdentifier('FLOW')).toBe(false);
    expect(isValidIdentifier('1-FLOW')).toBe(false);
    expect(isValidIdentifier('')).toBe(false);
  });
});

describe('normalizeIdentifier', () => {
  it('uppercases and trims', () => {
    expect(normalizeIdentifier('  flow-42  ')).toBe('FLOW-42');
  });

  it('handles already-normalized input', () => {
    expect(normalizeIdentifier('FLOW-42')).toBe('FLOW-42');
  });
});
