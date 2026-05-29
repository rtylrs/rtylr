export type EmploymentStatus = 'Active' | 'On Leave' | 'Suspended' | 'Terminated' | 'Probation';

export type ContractType = 'Full-time' | 'Part-time' | 'Contract' | 'Intern' | 'Freelance';

export function employmentStatusLabel(status: EmploymentStatus): string {
  const labels: Record<EmploymentStatus, string> = {
    Active: 'Active',
    'On Leave': 'On Leave',
    Suspended: 'Suspended',
    Terminated: 'Terminated',
    Probation: 'Probation',
  };
  return labels[status] ?? status;
}

export function employmentStatusColor(status: EmploymentStatus): string {
  const colors: Record<EmploymentStatus, string> = {
    Active: 'green',
    'On Leave': 'yellow',
    Suspended: 'orange',
    Terminated: 'red',
    Probation: 'blue',
  };
  return colors[status] ?? 'gray';
}

export function isActiveEmployee(status: EmploymentStatus): boolean {
  return status === 'Active' || status === 'Probation';
}

export function isOnLeave(status: EmploymentStatus): boolean {
  return status === 'On Leave';
}

export function isTerminated(status: EmploymentStatus): boolean {
  return status === 'Terminated';
}

export function contractTypeLabel(type: ContractType): string {
  const labels: Record<ContractType, string> = {
    'Full-time': 'Full-time',
    'Part-time': 'Part-time',
    Contract: 'Contract',
    Intern: 'Intern',
    Freelance: 'Freelance',
  };
  return labels[type] ?? type;
}

export function isPermanentEmployee(type: ContractType): boolean {
  return type === 'Full-time' || type === 'Part-time';
}

export function isTemporaryWorker(type: ContractType): boolean {
  return type === 'Contract' || type === 'Intern' || type === 'Freelance';
}

export function headcountByStatus(
  employees: Array<{ status: EmploymentStatus }>
): Record<EmploymentStatus, number> {
  const counts: Record<EmploymentStatus, number> = {
    Active: 0,
    'On Leave': 0,
    Suspended: 0,
    Terminated: 0,
    Probation: 0,
  };
  for (const emp of employees) {
    counts[emp.status] = (counts[emp.status] ?? 0) + 1;
  }
  return counts;
}
