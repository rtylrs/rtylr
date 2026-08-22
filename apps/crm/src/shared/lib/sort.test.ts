import { describe, it, expect } from 'vitest';
import {
  compareStrings,
  compareNumbers,
  compareDates,
  compareDatesDescending,
  sortByKey,
} from './sort';

describe('compareStrings', () => {
  it('returns 0 for equal strings', () => {
    expect(compareStrings('apple', 'apple')).toBe(0);
  });
  it('is case-insensitive', () => {
    expect(compareStrings('Apple', 'apple')).toBe(0);
  });
  it('returns negative when a comes before b', () => {
    expect(compareStrings('apple', 'banana')).toBeLessThan(0);
  });
  it('treats null as an empty string', () => {
    expect(compareStrings(null, 'apple')).toBeLessThan(0);
  });
});

describe('compareNumbers', () => {
  it('returns 0 for equal numbers', () => {
    expect(compareNumbers(5, 5)).toBe(0);
  });
  it('returns negative when a < b', () => {
    expect(compareNumbers(1, 2)).toBeLessThan(0);
  });
  it('returns positive when a > b', () => {
    expect(compareNumbers(3, 2)).toBeGreaterThan(0);
  });
});

describe('compareDates', () => {
  it('returns 0 for identical date strings', () => {
    expect(compareDates('2024-01-01', '2024-01-01')).toBe(0);
  });
  it('returns negative when a is earlier than b', () => {
    expect(compareDates('2023-01-01', '2024-01-01')).toBeLessThan(0);
  });
  it('sorts null dates last', () => {
    expect(compareDates(null, '2024-01-01')).toBeGreaterThan(0);
  });
  it('returns 0 when both are null', () => {
    expect(compareDates(null, null)).toBe(0);
  });
  it('puts malformed dates after valid dates', () => {
    expect(compareDates('not-a-date', '2024-01-01')).toBeGreaterThan(0);
    expect(compareDates('2024-01-01', 'not-a-date')).toBeLessThan(0);
  });
  it('rejects impossible calendar dates', () => {
    expect(compareDates('2024-02-30', '2024-03-01')).toBeGreaterThan(0);
  });

  it('sorts descending without moving malformed dates ahead', () => {
    expect(compareDatesDescending('2024-05-10', '2024-05-01')).toBeLessThan(0);
    expect(compareDatesDescending('not-a-date', '2024-05-01')).toBeGreaterThan(0);
  });
});

describe('sortByKey', () => {
  interface Item {
    name: string;
    score: number;
  }
  const items: Item[] = [
    { name: 'Charlie', score: 30 },
    { name: 'Alice', score: 10 },
    { name: 'Bob', score: 20 },
  ];
  it('sorts by a string key ascending', () => {
    const sorted = [...items].sort(sortByKey('name'));
    expect(sorted.map((i) => i.name)).toEqual(['Alice', 'Bob', 'Charlie']);
  });
  it('sorts by a string key descending', () => {
    const sorted = [...items].sort(sortByKey('name', 'desc'));
    expect(sorted.map((i) => i.name)).toEqual(['Charlie', 'Bob', 'Alice']);
  });
  it('sorts by a number key ascending', () => {
    const sorted = [...items].sort(sortByKey('score'));
    expect(sorted.map((i) => i.score)).toEqual([10, 20, 30]);
  });
});
