import { useEmployeeSessionStore } from '../stores/employeeSessionStore';

export interface Capabilities {
  canVoid: boolean;
  canRefund: boolean;
  canOpenDrawer: boolean;
  canApplyDiscount: boolean;
  maxDiscountPct: number;
  role: string;
}

const EMPTY: Capabilities = {
  canVoid: false,
  canRefund: false,
  canOpenDrawer: false,
  canApplyDiscount: false,
  maxDiscountPct: 0,
  role: '',
};

/**
 * Reads the signed-in operator's effective POS capabilities from the session.
 * These flags come from verify-pin (role-defaulted, admin-overridable) and are
 * the source of truth for gating manager-only actions. There is no separate
 * "manager PIN" — a manager simply signs in via the roster.
 */
export function useCapabilities(): Capabilities {
  const employee = useEmployeeSessionStore((s) => s.employee);
  if (!employee) return EMPTY;
  return {
    canVoid: employee.capabilities.canVoid,
    canRefund: employee.capabilities.canRefund,
    canOpenDrawer: employee.capabilities.canOpenDrawer,
    canApplyDiscount: employee.capabilities.canApplyDiscount,
    maxDiscountPct: employee.capabilities.maxDiscountPct,
    role: employee.role,
  };
}
