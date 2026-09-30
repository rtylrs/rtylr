import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';

interface Addons {
  feedback: boolean;
  about: boolean;
  contact: boolean;
  landing: boolean;
}

interface UserState {
  user: Record<string, unknown> | null;
  token: string | null;
  isAuthenticated: boolean;
  role: string | null;
  userId: string | null;
  addons: Addons;
  visits: unknown[];

  // Actions
  setUser: (user: Record<string, unknown>) => void;
  setToken: (token: string | null) => void;
  setAuthenticated: (isAuthenticated: boolean) => void;
  setRole: (role: string | null) => void;
  setUserId: (userId: string | null) => void;
  setAddons: (addons: Addons) => void;
  setVisits: (visits: unknown[]) => void;
  logout: () => void;
  reset: () => void;
}

const initialState = {
  user: null,
  token: null,
  isAuthenticated: false,
  role: null,
  userId: null,
  addons: {
    feedback: false,
    about: false,
    contact: false,
    landing: false,
  },
  visits: [],
};

export const useUserStore = create<UserState>()(
  persist(
    (set) => ({
      ...initialState,

      setUser: (user) => set({ user }),
      setToken: (token) => set({ token }),
      setAuthenticated: (isAuthenticated) => set({ isAuthenticated }),
      setRole: (role) => set({ role }),
      setUserId: (userId) => set({ userId }),
      setAddons: (addons) => set({ addons }),
      setVisits: (visits) => set({ visits }),
      logout: () =>
        set({
          ...initialState,
          token: null,
          isAuthenticated: false,
        }),
      reset: () => set(initialState),
    }),
    {
      name: 'rtylr-user',
      storage: createJSONStorage(() => localStorage),
    }
  )
);
