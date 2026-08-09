/**
 * Pure employee utility helpers used throughout the HR module.
 * All functions are side-effect free and operate only on plain values.
 */

/** Contract type options available for HR employees. */
export type ContractType = 'full_time' | 'part_time' | 'freelancer' | 'intern';

/** Status options representing an employee's current standing. */
export type EmployeeStatus = 'active' | 'probation' | 'inactive' | 'terminated';

export function contractTypeLabel(type: ContractType): string {
  const labels: Record<ContractType, string> = {
    full_time: 'Full-time',
    part_time: 'Part-time',
    freelancer: 'Freelancer',
    intern: 'Intern',
  };
  return labels[type];
}

export function isActive(status: EmployeeStatus): boolean {
  return status === 'active' || status === 'probation';
}

export function isOnProbation(status: EmployeeStatus): boolean {
  return status === 'probation';
}

export function nameInitials(name: string): string {
  const trimmed = name.trim();
  if (!trimmed) return '';
  const parts = trimmed.split(/\s+/);
  if (parts.length === 1) return parts[0][0].toUpperCase();
  return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
}

export function yearsOfService(hireDateIso: string, asOfIso?: string): number {
  const hire = new Date(hireDateIso);
  const asOf = asOfIso ? new Date(asOfIso) : new Date();
  const years = asOf.getFullYear() - hire.getFullYear();
  const hasNotYetReachedAnniversary =
    asOf.getMonth() < hire.getMonth() ||
    (asOf.getMonth() === hire.getMonth() && asOf.getDate() < hire.getDate());
  return Math.max(0, hasNotYetReachedAnniversary ? years - 1 : years);
}

export function filterByDepartment<T extends { department: string }>(
  employees: T[],
  department: string
): T[] {
  if (!department) return employees;
  return employees.filter((e) => e.department === department);
}

export function filterByStatus<T extends { status: string }>(employees: T[], status: string): T[] {
  if (!status) return employees;
  return employees.filter((e) => e.status === status);
}

export function sortByName<T extends { name: string }>(employees: T[]): T[] {
  return [...employees].sort((a, b) => a.name.localeCompare(b.name));
}
