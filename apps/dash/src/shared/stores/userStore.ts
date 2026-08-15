import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { removeCookie, getDefaultCookieDomain } from '@voxire/auth';

interface Addons {
  feedback: boolean;
  about: boolean;
  contact: boolean;
  landing: boolean;
}

export interface StoredUser {
  id?: number;
  email?: string;
  fullName?: string;
  role?: string;
  businessId?: number;
  businessName?: string;
  businessSlug?: string;
  business_slug?: string;
  logo?: string;
  isOwner?: boolean;
  isConsoleAdmin?: boolean;
  apps?: string[];
  onboarded?: boolean;
  planSlug?: string;
  subscriptionStatus?: string;
  effectiveResources?: Record<string, string>;
}

interface UserState {
  user: StoredUser | null;
  token: string | null;
  isAuthenticated: boolean;
  role: string | null;
  userId: string | null;
}

interface UserActions {
  login: (user: StoredUser) => void;
  logout: () => void;
  setToken: (token: string) => void;
  setUser: (user: StoredUser) => void;
  setRole: (role: string) => void;
  setUserId: (userId: string) => void;
}

const initialState: UserState = {
  user: null,
  token: null,
  isAuthenticated: false,
  role: null,
  userId: null,
};

export const useUserStore = create<UserState & UserActions>()(
  persist(
    (set) => ({
      ...initialState,

      login: (user) => set({ user, isAuthenticated: true }),

      logout: () => {
        removeCookie('token', { domain: getDefaultCookieDomain() });
        removeCookie('token');
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
    }),
    {
      name: 'user-store',
      partialize: (state) => ({
        user: state.user,
        isAuthenticated: state.isAuthenticated,
        role: state.role,
        userId: state.userId,
      }),
    }
  )
);
