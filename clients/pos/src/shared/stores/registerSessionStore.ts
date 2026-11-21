import { create } from 'zustand';

export type RegisterSessionStatus = 'unknown' | 'open' | 'none';

interface RegisterSessionState {
  status: RegisterSessionStatus;
  setStatus: (status: RegisterSessionStatus) => void;
}

// Not persisted: the server owns register sessions, so the status is read
// again after every sign-in instead of trusted from a previous shift.
export const useRegisterSessionStore = create<RegisterSessionState>()((set) => ({
  status: 'unknown',
  setStatus: (status) => set({ status }),
}));
