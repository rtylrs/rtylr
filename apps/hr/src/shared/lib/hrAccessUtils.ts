export type HrRole = 'admin' | 'manager' | 'hr_staff' | 'employee' | 'viewer';

export function hrRoleLabel(role: HrRole): string {
  const labels: Record<HrRole, string> = {
    admin: 'Administrator',
    manager: 'Manager',
    hr_staff: 'HR Staff',
    employee: 'Employee',
    viewer: 'Viewer',
  };
  return labels[role] ?? role;
}

export function canManagePayroll(role: HrRole): boolean {
  return role === 'admin' || role === 'hr_staff';
}

export function canViewAllEmployees(role: HrRole): boolean {
  return role === 'admin' || role === 'manager' || role === 'hr_staff';
}

export function canApproveLeave(role: HrRole): boolean {
  return role === 'admin' || role === 'manager' || role === 'hr_staff';
}

export function canEditSettings(role: HrRole): boolean {
  return role === 'admin';
}

export function canExportData(role: HrRole): boolean {
  return role === 'admin' || role === 'hr_staff';
}

export function hrRoleRank(role: HrRole): number {
  const ranks: Record<HrRole, number> = {
    admin: 5,
    manager: 4,
    hr_staff: 3,
    employee: 2,
    viewer: 1,
  };
  return ranks[role] ?? 0;
}

export function hasHigherRole(roleA: HrRole, roleB: HrRole): boolean {
  return hrRoleRank(roleA) > hrRoleRank(roleB);
}
