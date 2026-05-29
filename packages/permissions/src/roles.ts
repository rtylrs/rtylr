// ── Platform roles (business-level) ────────────────────────────────

export type PlatformRole =
  | 'owner'
  | 'admin'
  | 'operator'
  | 'billing'
  | 'menu_manager'
  | 'staff'
  | 'viewer';

/** Role hierarchy - higher index = more power */
export const ROLE_HIERARCHY: PlatformRole[] = [
  'viewer',
  'staff',
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
