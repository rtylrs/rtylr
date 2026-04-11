import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface Addons {
  feedback: boolean;
  about: boolean;
  contact: boolean;
  landing: boolean;
}

interface UserState {
  user: any | null;
  token: string | null;
  isAuthenticated: boolean;
  role: string | null;
  userId: string | null;
  addons: Addons;
  visits: any[];
}

interface UserActions {
  login: (user: any) => void;
  logout: () => void;
  setToken: (token: string) => void;
  setUser: (user: any) => void;
  setRole: (role: string) => void;
  setUserId: (userId: string) => void;
  setAddons: (addons: Addons) => void;
  setVisits: (visits: any[]) => void;
}

const initialState: UserState = {
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

export const useUserStore = create<UserState & UserActions>()(
  persist(
    (set, get) => ({
      ...initialState,

      login: (user) => set({ user, isAuthenticated: true }),

      logout: () => {
        document.cookie = 'token=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
        set({
          user: null,
          isAuthenticated: false,
          role: null,
          userId: null,
          token: null,
        });
      },

      setToken: (token) => set({ token, isAuthenticated: !!token }),

      setUser: (user) => set({ user }),

      setRole: (role) => set({ role }),

      setUserId: (userId) => set({ userId }),

      setAddons: (addons) => set({ addons }),

      setVisits: (visits) => set({ visits }),
    }),
    {
      name: 'user-store',
      partialize: (state) => ({
        user: state.user,
        isAuthenticated: state.isAuthenticated,
        role: state.role,
        userId: state.userId,
        addons: state.addons,
        visits: state.visits,
      }),
    }
  )
);
