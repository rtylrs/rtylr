// ── Platform roles (business-level) ────────────────────────────────

export type PlatformRole =
  | 'owner'
  | 'admin'
  | 'operator'
  | 'billing'
  | 'menu_manager'
  | 'store_manager'
  | 'staff'
  | 'viewer';

/**
 * Role hierarchy - higher index = more power.
 *
 * store_manager sits just above staff in the GLOBAL hierarchy on purpose: its
 * elevated powers (managing POS users, voids/refunds) are location-scoped and
 * enforced per-location via pos_member_locations flags, not granted globally.
 * Keeping it low here prevents a global hasRoleLevel check from over-granting.
 */
export const ROLE_HIERARCHY: PlatformRole[] = [
  'viewer',
  'staff',
  'store_manager',
  'menu_manager',
  'billing',
  'operator',
  'admin',
  'owner',
];

/** Returns true if `role` is at least as powerful as `required` */
export function hasRoleLevel(role: PlatformRole, required: PlatformRole): boolean {
  return ROLE_HIERARCHY.indexOf(role) >= ROLE_HIERARCHY.indexOf(required);
}

// ── POS roles (terminal-level) ─────────────────────────────────────

export type POSRole = 'manager' | 'cashier' | 'server' | 'bartender' | 'kitchen';

export const POS_ROLE_HIERARCHY: POSRole[] = [
  'kitchen',
  'bartender',
  'server',
  'cashier',
  'manager',
];

export function hasPOSRoleLevel(role: POSRole, required: POSRole): boolean {
  return POS_ROLE_HIERARCHY.indexOf(role) >= POS_ROLE_HIERARCHY.indexOf(required);
}
