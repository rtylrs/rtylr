import { describe, expect, it } from 'vitest';
import {
  sortByBooleanDesc,
  sortByDate,
  sortByNumber,
  sortByString,
  stableSort,
} from './hrSortUtils';

describe('sortByString', () => {
  const employees = [
    { name: 'Charlie', dept: 'HR' },
    { name: 'Alice', dept: 'Eng' },
    { name: 'Bob', dept: 'Sales' },
  ];

  it('sorts ascending by default', () => {
    const result = sortByString(employees, 'name');
    expect(result[0].name).toBe('Alice');
    expect(result[2].name).toBe('Charlie');
  });

  it('sorts descending', () => {
    const result = sortByString(employees, 'name', 'desc');
    expect(result[0].name).toBe('Charlie');
    expect(result[2].name).toBe('Alice');
  });

  it('does not mutate original array', () => {
    sortByString(employees, 'name');
    expect(employees[0].name).toBe('Charlie');
  });
});

describe('sortByNumber', () => {
  const records = [
    { id: 1, salary: 75000 },
    { id: 2, salary: 50000 },
    { id: 3, salary: 90000 },
  ];

  it('sorts ascending by default', () => {
    const result = sortByNumber(records, 'salary');
    expect(result[0].salary).toBe(50000);
    expect(result[2].salary).toBe(90000);
  });

  it('sorts descending', () => {
    const result = sortByNumber(records, 'salary', 'desc');
    expect(result[0].salary).toBe(90000);
  });
});

describe('sortByDate', () => {
  const items = [
    { name: 'C', date: '2024-03-01' },
    { name: 'A', date: '2024-01-01' },
    { name: 'B', date: '2024-02-01' },
  ];

  it('sorts ascending by default', () => {
    const result = sortByDate(items, 'date');
    expect(result[0].name).toBe('A');
    expect(result[2].name).toBe('C');
  });

  it('sorts descending', () => {
    const result = sortByDate(items, 'date', 'desc');
    expect(result[0].name).toBe('C');
  });
});

describe('sortByBooleanDesc', () => {
  const items = [
    { name: 'A', active: false },
    { name: 'B', active: true },
    { name: 'C', active: false },
    { name: 'D', active: true },
  ];

  it('puts true values first', () => {
    const result = sortByBooleanDesc(items, 'active');
    expect(result[0].active).toBe(true);
    expect(result[1].active).toBe(true);
    expect(result[2].active).toBe(false);
  });
});

describe('stableSort', () => {
  it('maintains relative order of equal elements', () => {
    const items = [
      { priority: 1, name: 'First' },
      { priority: 2, name: 'Second A' },
      { priority: 2, name: 'Second B' },
      { priority: 3, name: 'Third' },
    ];
    const result = stableSort(items, (a, b) => a.priority - b.priority);
    expect(result[1].name).toBe('Second A');
    expect(result[2].name).toBe('Second B');
  });
});
