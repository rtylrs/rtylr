import { describe, expect, it } from 'vitest';
import { groupByDepartment } from './groupByDepartment';
import type { DirectoryMember } from './useDirectory';

function member(over: Partial<DirectoryMember> = {}): DirectoryMember {
  return {
    id: '1',
    firstName: 'Ada',
    lastName: 'Amouneh',
    email: 'ada@example.com',
    role: 'Foreman',
    departmentName: 'Site',
    ...over,
  } as DirectoryMember;
}

describe('groupByDepartment', () => {
  it('returns nothing for nobody', () => {
    expect(groupByDepartment([])).toEqual([]);
  });

  it('puts each person under their department', () => {
    const groups = groupByDepartment([
      member({ id: '1', departmentName: 'Site' }),
      member({ id: '2', departmentName: 'Office' }),
      member({ id: '3', departmentName: 'Site' }),
    ]);

    expect(groups.map((g) => g.name)).toEqual(['Office', 'Site']);
    expect(groups.find((g) => g.name === 'Site')?.members).toHaveLength(2);
  });

  it('sorts departments by name, so the list does not reorder between loads', () => {
    const groups = groupByDepartment([
      member({ id: '1', departmentName: 'Warehouse' }),
      member({ id: '2', departmentName: 'Accounts' }),
      member({ id: '3', departmentName: 'Site' }),
    ]);
    expect(groups.map((g) => g.name)).toEqual(['Accounts', 'Site', 'Warehouse']);
  });

  it('sorts people inside a department by name', () => {
    const groups = groupByDepartment([
      member({ id: '1', firstName: 'Zaid', lastName: 'B' }),
      member({ id: '2', firstName: 'Ada', lastName: 'A' }),
    ]);
    expect(groups[0].members.map((m) => m.firstName)).toEqual(['Ada', 'Zaid']);
  });

  it('keeps someone with no department rather than dropping them', () => {
    const groups = groupByDepartment([
      member({ id: '1', departmentName: 'Site' }),
      member({ id: '2', firstName: 'Sam', departmentName: undefined }),
    ]);

    expect(groups).toHaveLength(2);
    const unassigned = groups.find((g) => g.name === null);
    expect(unassigned?.members.map((m) => m.firstName)).toEqual(['Sam']);
  });

  it('puts the unassigned group last, whatever it would sort as', () => {
    const groups = groupByDepartment([
      member({ id: '1', departmentName: undefined }),
      member({ id: '2', departmentName: 'Warehouse' }),
      member({ id: '3', departmentName: 'Accounts' }),
    ]);
    expect(groups.map((g) => g.name)).toEqual(['Accounts', 'Warehouse', null]);
  });

  it('treats a blank department the same as none', () => {
    const groups = groupByDepartment([
      member({ id: '1', departmentName: '   ' }),
      member({ id: '2', departmentName: 'Site' }),
    ]);
    expect(groups.map((g) => g.name)).toEqual(['Site', null]);
  });

  it('loses nobody', () => {
    const people = [
      member({ id: '1', departmentName: 'Site' }),
      member({ id: '2', departmentName: undefined }),
      member({ id: '3', departmentName: 'Office' }),
      member({ id: '4', departmentName: 'Site' }),
    ];
    const total = groupByDepartment(people).reduce((n, g) => n + g.members.length, 0);
    expect(total).toBe(people.length);
  });
});
