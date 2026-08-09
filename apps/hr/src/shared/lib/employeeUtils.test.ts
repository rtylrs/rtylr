import { describe, it, expect } from 'vitest';
import {
  contractTypeLabel,
  isActive,
  isOnProbation,
  nameInitials,
  yearsOfService,
  filterByDepartment,
  filterByStatus,
  sortByName,
} from './employeeUtils';

describe('contractTypeLabel', () => {
  it('returns Full-time for full_time', () => {
    expect(contractTypeLabel('full_time')).toBe('Full-time');
  });

  it('returns Part-time for part_time', () => {
    expect(contractTypeLabel('part_time')).toBe('Part-time');
  });

  it('returns Freelancer for freelancer', () => {
    expect(contractTypeLabel('freelancer')).toBe('Freelancer');
  });

  it('returns Intern for intern', () => {
    expect(contractTypeLabel('intern')).toBe('Intern');
  });
});

describe('isActive', () => {
  it('returns true for active status', () => {
    expect(isActive('active')).toBe(true);
  });

  it('returns true for probation status', () => {
    expect(isActive('probation')).toBe(true);
  });

  it('returns false for inactive status', () => {
    expect(isActive('inactive')).toBe(false);
  });

  it('returns false for terminated status', () => {
    expect(isActive('terminated')).toBe(false);
  });
});

describe('isOnProbation', () => {
  it('returns true for probation', () => {
    expect(isOnProbation('probation')).toBe(true);
  });

  it('returns false for active', () => {
    expect(isOnProbation('active')).toBe(false);
  });
});

describe('nameInitials', () => {
  it('returns initials for two-word name', () => {
    expect(nameInitials('Jane Doe')).toBe('JD');
  });

  it('returns first letter for single word', () => {
    expect(nameInitials('Madonna')).toBe('M');
  });

  it('uses first and last word for multi-word names', () => {
    expect(nameInitials('Mary Jane Watson')).toBe('MW');
  });

  it('returns empty string for empty input', () => {
    expect(nameInitials('')).toBe('');
  });

  it('handles leading and trailing whitespace', () => {
    expect(nameInitials('  Jane Doe  ')).toBe('JD');
  });

  it('uppercases initials', () => {
    expect(nameInitials('john doe')).toBe('JD');
  });
});

describe('yearsOfService', () => {
  it('returns correct years on exact anniversary', () => {
    expect(yearsOfService('2020-01-01', '2025-01-01')).toBe(5);
  });

  it('returns one less when anniversary has not yet passed', () => {
    expect(yearsOfService('2020-06-15', '2025-06-14')).toBe(4);
  });

  it('returns correct years on the anniversary day', () => {
    expect(yearsOfService('2020-06-15', '2025-06-15')).toBe(5);
  });

  it('returns 0 for same day hire', () => {
    expect(yearsOfService('2025-01-01', '2025-01-01')).toBe(0);
  });

  it('returns 0 for hire in the future', () => {
    expect(yearsOfService('2026-01-01', '2025-01-01')).toBe(0);
  });
});

describe('filterByDepartment', () => {
  const employees = [
    { name: 'Alice', department: 'Engineering', status: 'active' },
    { name: 'Bob', department: 'Sales', status: 'active' },
    { name: 'Carol', department: 'Engineering', status: 'inactive' },
  ];

  it('filters to matching department', () => {
    const result = filterByDepartment(employees, 'Engineering');
    expect(result).toHaveLength(2);
  });

  it('returns all employees for empty department string', () => {
    expect(filterByDepartment(employees, '')).toHaveLength(3);
  });

  it('returns empty array for non-existent department', () => {
    expect(filterByDepartment(employees, 'Marketing')).toHaveLength(0);
  });
});

describe('filterByStatus', () => {
  const employees = [
    { name: 'Alice', department: 'Engineering', status: 'active' },
    { name: 'Bob', department: 'Sales', status: 'probation' },
    { name: 'Carol', department: 'Engineering', status: 'inactive' },
  ];

  it('filters to matching status', () => {
    const result = filterByStatus(employees, 'active');
    expect(result).toHaveLength(1);
    expect(result[0].name).toBe('Alice');
  });

  it('returns all for empty status string', () => {
    expect(filterByStatus(employees, '')).toHaveLength(3);
  });
});

describe('sortByName', () => {
  it('sorts alphabetically ascending', () => {
    const employees = [
      { name: 'Carol', status: 'active' },
      { name: 'Alice', status: 'active' },
      { name: 'Bob', status: 'active' },
    ];
    const sorted = sortByName(employees);
    expect(sorted.map((e) => e.name)).toEqual(['Alice', 'Bob', 'Carol']);
  });

  it('does not mutate the original array', () => {
    const employees = [
      { name: 'Zara', status: 'active' },
      { name: 'Anna', status: 'active' },
    ];
    const original = [...employees];
    sortByName(employees);
    expect(employees[0].name).toBe(original[0].name);
  });

  it('returns empty array for empty input', () => {
    expect(sortByName([])).toEqual([]);
  });
});
