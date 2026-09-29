import type { POSUser, UpdatePOSUserRequest } from '@voxire/voxie/pos/user/v1/types';
import { PIN_MAX_LENGTH, PIN_MIN_LENGTH, isNewPin } from '../login/pinPolicy';

// Pure helpers for the staff-management screen.

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

function asRole(value: unknown): StaffRole {
  return ROLES.includes(value as StaffRole) ? (value as StaffRole) : 'cashier';
}

export function mapPOSUsers(users: POSUser[]): StaffUser[] {
  return users.flatMap((u) => {
    const id = positiveInt(u.id);
    const name = (u.name ?? '').trim();
    if (!id || !name) return [];
    return [
      {
        id,
        name,
        role: asRole(u.role),
        isActive: u.isActive ?? false,
        maxDiscountPct: u.maxDiscountPct ?? 0,
        canVoid: u.canVoid ?? false,
        canRefund: u.canRefund ?? false,
        canOpenDrawer: u.canOpenDrawer ?? false,
        canApplyDiscount: u.canApplyDiscount ?? false,
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
 * Builds the generated updatePOSUser payload from edited staff fields. A blank
 * newPin is omitted so an edit that doesn't touch the PIN leaves it unchanged.
 * hourlyRate is omitted too: the list response carries no rate, and sending a
 * hardcoded 0 would zero out the employee's real rate on every role/PIN edit.
 */
export function buildStaffUpdate(
  user: StaffUser,
  edits: StaffEdits
): UpdatePOSUserRequest {
  const pin = (edits.newPin ?? '').trim();
  // Same policy the server enforces on rotation, so a manager is told here
  // rather than after a round trip. A 4-digit PIN cannot be set any more, but
  // one that already exists still verifies: this only ever rejects a new PIN.
  if (pin && !isNewPin(pin)) {
    throw new Error(`PIN must be between ${PIN_MIN_LENGTH} and ${PIN_MAX_LENGTH} digits.`);
  }
  return {
    id: String(user.id),
    name: user.name,
    ...(pin ? { pin } : {}),
    isActive: edits.isActive,
  };
}
