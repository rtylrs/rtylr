import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { createOneKitClient } from '@voxire/api';
import { UserServiceClient } from '@voxire/voxie/pos/user/v1/client';

const userClient = createOneKitClient(UserServiceClient, 'pos');

export interface EmployeeCapabilities {
  canVoid: boolean;
  canRefund: boolean;
  canOpenDrawer: boolean;
  canApplyDiscount: boolean;
  maxDiscountPct: number;
}

export interface Employee {
  /** Team users.id — the unified identity behind the PIN. */
  id: string;
  posUserId: string;
  locationId: number;
  name: string;
  role: string;
  capabilities: EmployeeCapabilities;
}

interface EmployeeSessionState {
  employee: Employee | null;
  isActive: boolean;
  loading: boolean;
  error: string | null;
}

interface EmployeeSessionActions {
  authenticatePin: (locationId: number, userId: number, pin: string) => Promise<void>;
  endSession: () => void;
  clearError: () => void;
}

const INVALID_EMPLOYEE_ERROR = 'Employee login returned invalid employee data';
const BLANK_PIN_ERROR = 'Enter your employee PIN';

function positiveId(value: unknown): string | null {
  const numeric = typeof value === 'string' ? Number(value.trim()) : Number(value);
  if (!Number.isFinite(numeric) || numeric <= 0) return null;
  return String(Math.floor(numeric));
}

function trimmedText(value: unknown): string | null {
  if (typeof value !== 'string') return null;
  const trimmed = value.trim();
  return trimmed || null;
}

function asBool(value: unknown): boolean {
  return value === true;
}

function asNumber(value: unknown): number {
  const n = Number(value);
  return Number.isFinite(n) ? n : 0;
}

// Backend VerifyPINResponse (camelCased by @voxire/api): user_id, pos_user_id,
// location_id, role, display_name + capability flags. The PIN is never echoed,
// so nothing sensitive lands in this persisted store.
function normalizeVerifyResponse(raw: unknown): Employee {
  if (!raw || typeof raw !== 'object') {
    throw new Error(INVALID_EMPLOYEE_ERROR);
  }
  const r = raw as Record<string, unknown>;
  const id = positiveId(r.userId ?? r.user_id);
  const name = trimmedText(r.displayName ?? r.display_name ?? r.name);
  const role = trimmedText(r.role) ?? 'cashier';
  if (!id || !name) {
    throw new Error(INVALID_EMPLOYEE_ERROR);
  }
  return {
    id,
    posUserId: positiveId(r.posUserId ?? r.pos_user_id) ?? '0',
    locationId: asNumber(r.locationId ?? r.location_id),
    name,
    role,
    capabilities: {
      canVoid: asBool(r.canVoid ?? r.can_void),
      canRefund: asBool(r.canRefund ?? r.can_refund),
      canOpenDrawer: asBool(r.canOpenDrawer ?? r.can_open_drawer),
      canApplyDiscount: asBool(r.canApplyDiscount ?? r.can_apply_discount),
      maxDiscountPct: asNumber(r.maxDiscountPct ?? r.max_discount_pct),
    },
  };
}

export const useEmployeeSessionStore = create<EmployeeSessionState & EmployeeSessionActions>()(
  persist(
    (set) => ({
      employee: null,
      isActive: false,
      loading: false,
      error: null,

      authenticatePin: async (locationId: number, userId: number, pin: string) => {
        const normalizedPin = pin.trim();
        if (!normalizedPin) {
          set({ loading: false, error: BLANK_PIN_ERROR });
          throw new Error(BLANK_PIN_ERROR);
        }

        set({ loading: true, error: null });
        try {
          const data = await userClient.verifyPOSUserPIN({
            locationId: String(locationId),
            userId: String(userId),
            pin: normalizedPin,
          });
          const employee = normalizeVerifyResponse(data);
          set({ employee, isActive: true, loading: false, error: null });
        } catch (err) {
          const message =
            err instanceof Error && err.message === INVALID_EMPLOYEE_ERROR
              ? err.message
              : 'Incorrect PIN';
          set({ loading: false, employee: null, isActive: false, error: message });
          throw err;
        }
      },

      endSession: () => {
        set({ employee: null, isActive: false, error: null });
      },

      clearError: () => set({ error: null }),
    }),
    {
      name: 'employee-session',
      partialize: (state) => ({
        employee: state.employee,
        isActive: state.isActive,
      }),
    }
  )
);
