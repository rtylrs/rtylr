// Pure helpers for the staff-management screen. The list normalization
// tolerates both snake_case and camelCase (the API layer camelCases responses)
// and is unit-tested independently of the screen.

export type StaffRole = 'manager' | 'cashier' | 'server' | 'bartender' | 'kitchen';

export interface StaffUser {
  id: number;
  name: string;
  role: StaffRole;
  isActive: boolean;
  maxDiscountPct: number;
  canVoid: boolean;
  canRefund: boolean;
  canOpenDrawer: boolean;
  canApplyDiscount: boolean;
}

const ROLES: StaffRole[] = ['manager', 'cashier', 'server', 'bartender', 'kitchen'];

function positiveInt(value: unknown): number {
  const n = typeof value === 'string' ? Number(value.trim()) : Number(value);
  return Number.isFinite(n) && n > 0 ? Math.floor(n) : 0;
}

function asBool(value: unknown): boolean {
  return value === true;
}

function asRole(value: unknown): StaffRole {
  return ROLES.includes(value as StaffRole) ? (value as StaffRole) : 'cashier';
}

export function mapPOSUsers(data: unknown): StaffUser[] {
  const raw = Array.isArray(data)
    ? data
    : data && typeof data === 'object' && Array.isArray((data as { users?: unknown[] }).users)
      ? (data as { users: unknown[] }).users
      : [];
  return raw.flatMap((entry) => {
    const u = entry as Record<string, unknown>;
    const id = positiveInt(u.id);
    const name = typeof u.name === 'string' ? u.name.trim() : '';
    if (!id || !name) return [];
    return [
      {
        id,
        name,
        role: asRole(u.role),
        // POSUser.isActive is a boolean; treat anything but an explicit false as active.
        isActive: (u.isActive ?? u.is_active) !== false,
        maxDiscountPct: Number(u.maxDiscountPct ?? u.max_discount_pct ?? 0) || 0,
        canVoid: asBool(u.canVoid ?? u.can_void),
        canRefund: asBool(u.canRefund ?? u.can_refund),
        canOpenDrawer: asBool(u.canOpenDrawer ?? u.can_open_drawer),
        canApplyDiscount: asBool(u.canApplyDiscount ?? u.can_apply_discount),
      },
    ];
  });
}

export interface StaffEdits {
  role: StaffRole;
  isActive: boolean;
  maxDiscountPct: number;
  canVoid: boolean;
  canRefund: boolean;
  canOpenDrawer: boolean;
  canApplyDiscount: boolean;
  newPin?: string;
}

/**
 * Builds the snake_case updatePOSUser payload from edited staff fields. A blank
 * newPin is omitted so an edit that doesn't touch the PIN leaves it unchanged.
 */
export function buildStaffUpdate(
  user: StaffUser,
  edits: StaffEdits
): Record<string, unknown> & { id: number } {
  const pin = (edits.newPin ?? '').trim();
  return {
    id: user.id,
    name: user.name,
    role: edits.role,
    is_active: edits.isActive,
    max_discount_pct: edits.maxDiscountPct,
    can_void: edits.canVoid,
    can_refund: edits.canRefund,
    can_open_drawer: edits.canOpenDrawer,
    can_apply_discount: edits.canApplyDiscount,
    ...(pin ? { pin } : {}),
  };
}
