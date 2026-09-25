import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface TerminalContextState {
  locationId: number | null;
  terminalId: number | null;
}

interface TerminalContextActions {
  setTerminalContext: (locationId: number, terminalId: number) => void;
  clearTerminalContext: () => void;
}

function normalizeId(id: number): number | null {
  if (!Number.isFinite(id) || id <= 0) return null;
  return Math.floor(id);
}

export const useTerminalContextStore = create<TerminalContextState & TerminalContextActions>()(
  persist(
    (set) => ({
      locationId: null,
      terminalId: null,
      setTerminalContext: (locationId, terminalId) => {
        const normalizedLocationId = normalizeId(locationId);
        const normalizedTerminalId = normalizeId(terminalId);
        if (!normalizedLocationId || !normalizedTerminalId) {
          set({ locationId: null, terminalId: null });
          return;
        }
        set({ locationId: normalizedLocationId, terminalId: normalizedTerminalId });
      },
      clearTerminalContext: () => set({ locationId: null, terminalId: null }),
    }),
    { name: 'pos-terminal-context' }
  )
);
