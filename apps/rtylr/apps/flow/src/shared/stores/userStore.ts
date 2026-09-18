import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export interface FlowUser {
  id: string;
  name: string;
  email: string;
  businessName?: string;
  avatar?: string;
  /** From API when `name` is assembled from parts */
  firstName?: string;
}

interface UserState {
  user: FlowUser | null;
  token: string | null;
  isAuthenticated: boolean;
}

interface UserActions {
  login: (user: FlowUser, token: string) => void;
  logout: () => void;
  setUser: (user: FlowUser) => void;
}

export const useUserStore = create<UserState & UserActions>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      isAuthenticated: false,
      login: (user, token) => set({ user, token, isAuthenticated: true }),
      logout: () => set({ user: null, token: null, isAuthenticated: false }),
      setUser: (user) => set({ user }),
    }),
    { name: 'flow-user' }
  )
);
