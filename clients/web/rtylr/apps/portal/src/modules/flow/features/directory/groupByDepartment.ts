import type { DirectoryMember } from './useDirectory';

export interface DepartmentGroup {
  /** Stable key for React. `''` for the members with no department. */
  key: string;
  /** The department's name, or null for the unassigned group. */
  name: string | null;
  members: DirectoryMember[];
}

/**
 * The directory grouped by department, per S-21.
 *
 * Departments come back in whatever order the staff list happens to be in, so
 * they are sorted by name and the members inside each by their own name: a
 * directory that reorders itself between loads is not a directory.
 *
 * Anyone without a department goes into one group at the end rather than being
 * dropped, because a person missing an attribute is still a colleague you
 * might be looking for.
 */
export function groupByDepartment(members: DirectoryMember[]): DepartmentGroup[] {
  const byDepartment = new Map<string, DirectoryMember[]>();
  for (const member of members) {
    const key = member.departmentName?.trim() ?? '';
    const group = byDepartment.get(key) ?? [];
    group.push(member);
    byDepartment.set(key, group);
  }

  const fullName = (member: DirectoryMember): string =>
    `${member.firstName ?? ''} ${member.lastName ?? ''}`.trim().toLocaleLowerCase();

  return [...byDepartment.entries()]
    .map(([key, groupMembers]) => ({
      key,
      name: key === '' ? null : key,
      members: [...groupMembers].sort((a, b) => fullName(a).localeCompare(fullName(b))),
    }))
    .sort((a, b) => {
      // The unassigned group last, whatever it would sort as.
      if (a.name === null) return 1;
      if (b.name === null) return -1;
      return a.name.localeCompare(b.name);
    });
}
