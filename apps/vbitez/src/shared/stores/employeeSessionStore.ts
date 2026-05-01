import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import api from '../lib/api';

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

export const useEmployeeSessionStore = create<EmployeeSessionState & EmployeeSessionActions>()(
  persist(
    (set) => ({
      employee: null,
      isActive: false,
      loading: false,
      error: null,

      authenticatePin: async (pin: string) => {
        set({ loading: true, error: null });
        try {
          const response = await api.post('/employees/pin', { pin });
          // Strip any PIN-shaped field the backend returns before
          // putting the employee in a persisted store. Earlier shape
          // included `pin?: string` which would land in localStorage —
          // a security gap (audit P0 #7).
          const raw = response.data.employee as Employee & {
            pin?: unknown;
            pin_hash?: unknown;
          };
          const { pin: _stripPin, pin_hash: _stripHash, ...safeEmployee } = raw;
          void _stripPin;
          void _stripHash;
          set({
            employee: safeEmployee as Employee,
            isActive: true,
            loading: false,
            error: null,
          });
        } catch (err) {
          const e = err as { response?: { data?: { message?: string } } };
          set({
            loading: false,
            error: e.response?.data?.message || 'Incorrect PIN',
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
