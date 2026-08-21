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
 * Tier ladder: each entry is strictly more powerful than the one before it.
 *
 * Only genuine tiers belong here. The three lateral roles - billing,
 * menu_manager and store_manager - are specialisations, not power levels, and
 * live in ROLE_PARENT below.
 *
 * Before #795 all eight roles sat in one list and permissions were accumulated
 * by array index, so every role inherited everything declared below it. That is
 * how `billing` came to hold pos:drawer:open, pos:sessions:close and hr:staff:
 * it sat above staff in the list, and a finance role inherited the shop floor.
 */
export const ROLE_TIERS: PlatformRole[] = ['viewer', 'staff', 'operator', 'admin', 'owner'];

/**
 * The single role each role inherits from. null terminates the chain.
 *
 * A lateral role names its base explicitly and inherits from nothing else, so
 * adding a fourth lateral cannot silently grant it a third one's permissions.
 */
export const ROLE_PARENT: Record<PlatformRole, PlatformRole | null> = {
  viewer: null,

  staff: 'viewer',
  operator: 'staff',
  admin: 'operator',
  owner: 'admin',

  // store_manager is a POS floor role, so staff is the right base. It declares
  // no global permissions of its own on purpose: its elevated powers (staff
  // PINs, voids, refunds, reports) are location-scoped and enforced per-location
  // via pos_member_locations plus the capability flags returned at PIN verify.
  store_manager: 'staff',

  // menu_manager keeps staff as its base, which is what it resolved to before
  // #795. Whether someone who manages the menu should also be able to take
  // orders and see HR self-service is a product question, not one to settle
  // while restructuring the model - so today's behaviour is preserved exactly.
  menu_manager: 'staff',

  // billing is a finance role and its base is viewer, NOT staff. This is the
  // #795 fix: nothing about approving an invoice implies opening a cash drawer.
  billing: 'viewer',
};

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
