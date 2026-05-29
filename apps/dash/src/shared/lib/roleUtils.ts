/**
 * Role hierarchy and display helpers for the Dash app.
 * All functions are pure with no side-effects.
 */

export type StaffRole = 'owner' | 'manager' | 'cashier' | 'server' | 'kitchen' | 'host' | 'runner';

const ROLE_ORDER: StaffRole[] = [
  'owner',
  'manager',
  'cashier',
  'server',
  'kitchen',
  'host',
  'runner',
];

/** Return a human-readable label for a staff role. */
export function roleLabel(role: StaffRole): string {
  switch (role) {
    case 'owner':
      return 'Owner';
    case 'manager':
      return 'Manager';
    case 'cashier':
      return 'Cashier';
    case 'server':
      return 'Server';
    case 'kitchen':
      return 'Kitchen Staff';
    case 'host':
      return 'Host';
    case 'runner':
      return 'Runner';
  }
}

/** Return a short description for a staff role. */
export function roleDescription(role: StaffRole): string {
  switch (role) {
    case 'owner':
      return 'Full access to all settings and data';
    case 'manager':
      return 'Manage staff, view reports, and configure settings';
    case 'cashier':
      return 'Process orders and payments';
    case 'server':
      return 'Take orders and manage tables';
    case 'kitchen':
      return 'View and manage kitchen orders';
    case 'host':
      return 'Manage reservations and seating';
    case 'runner':
      return 'Deliver orders to tables';
  }
}

/** Return true if roleA outranks roleB (lower index = higher rank). */
export function outranks(a: StaffRole, b: StaffRole): boolean {
  return ROLE_ORDER.indexOf(a) < ROLE_ORDER.indexOf(b);
}

/** Return true if roleA has at least as much seniority as roleB. */
export function atLeastRole(a: StaffRole, b: StaffRole): boolean {
  return ROLE_ORDER.indexOf(a) <= ROLE_ORDER.indexOf(b);
}

/** Return all roles in order from highest to lowest. */
export function allRoles(): StaffRole[] {
  return [...ROLE_ORDER];
}

/** Return roles that can manage a given role (all roles that outrank it). */
export function managedBy(role: StaffRole): StaffRole[] {
  const idx = ROLE_ORDER.indexOf(role);
  return ROLE_ORDER.slice(0, idx);
}

/** Return the badge color for a staff role. */
export function roleBadgeColor(role: StaffRole): string {
  switch (role) {
    case 'owner':
      return 'purple';
    case 'manager':
      return 'blue';
    case 'cashier':
      return 'green';
    case 'server':
      return 'teal';
    case 'kitchen':
      return 'orange';
    case 'host':
      return 'yellow';
    case 'runner':
      return 'gray';
  }
}
