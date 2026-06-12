import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import api from '../services/api/api';

export interface Employee {
  id: string;
  name: string;
  role: string;
}

interface EmployeeSessionState {
  employee: Employee | null;
  isActive: boolean;
  loading: boolean;
  error: string | null;
}

interface EmployeeSessionActions {
  authenticatePin: (pin: string) => Promise<void>;
  endSession: () => void;
  clearError: () => void;
}

const INVALID_EMPLOYEE_ERROR = 'Employee login returned invalid employee data';
const BLANK_PIN_ERROR = 'Enter your employee PIN';

function positiveEmployeeId(value: unknown): string | null {
  const numeric = typeof value === 'string' ? Number(value.trim()) : Number(value);
  if (!Number.isFinite(numeric) || numeric <= 0) return null;
  return String(Math.floor(numeric));
}

function trimmedText(value: unknown): string | null {
  if (typeof value !== 'string') return null;
  const trimmed = value.trim();
  return trimmed || null;
}

function normalizeEmployee(raw: unknown): Employee {
  if (!raw || typeof raw !== 'object') {
    throw new Error(INVALID_EMPLOYEE_ERROR);
  }

  const record = raw as Record<string, unknown>;
  const id = positiveEmployeeId(record.id);
  const name = trimmedText(record.name);
  const role = trimmedText(record.role) ?? 'cashier';
  if (!id || !name) {
    throw new Error(INVALID_EMPLOYEE_ERROR);
  }

  return { id, name, role };
}

export const useEmployeeSessionStore = create<EmployeeSessionState & EmployeeSessionActions>()(
  persist(
    (set) => ({
      employee: null,
      isActive: false,
      loading: false,
      error: null,

      authenticatePin: async (pin: string) => {
        const normalizedPin = pin.trim();
        if (!normalizedPin) {
          set({ loading: false, error: BLANK_PIN_ERROR });
          throw new Error(BLANK_PIN_ERROR);
        }

        set({ loading: true, error: null });
        try {
          const response = await api.post('/employees/pin', { pin: normalizedPin });
          // Strip any PIN-shaped field the backend returns before
          // putting the employee in a persisted store. Earlier shape
          // included `pin?: string` which would land in localStorage -
          // a security gap (audit P0 #7).
          const raw = response.data.employee as Record<string, unknown> & {
            pin?: unknown;
            pin_hash?: unknown;
          };
          const { pin: _stripPin, pin_hash: _stripHash, ...safeEmployee } = raw;
          void _stripPin;
          void _stripHash;
          const employee = normalizeEmployee(safeEmployee);
          set({
            employee,
            isActive: true,
            loading: false,
            error: null,
          });
        } catch (err) {
          const e = err as { response?: { data?: { message?: string } } };
          const message =
            err instanceof Error && err.message === INVALID_EMPLOYEE_ERROR
              ? err.message
              : e.response?.data?.message || 'Incorrect PIN';
          set({
            loading: false,
            employee: null,
            isActive: false,
            error: message,
          });
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
