import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { removeCookie, getDefaultCookieDomain } from '@voxire/auth';
import type { EffectiveAccess, ProvisioningStatus } from '@voxire/types';

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
  businessId?: number;
  businessName?: string;
  businessSlug?: string;
  business_slug?: string;
  logo?: string;
  isConsoleAdmin?: boolean;
  onboarded?: boolean;
  planSlug?: string;
  subscriptionStatus?: string;
  effectiveAccess?: EffectiveAccess;
  provisioningStatus?: ProvisioningStatus;
}

interface UserState {
  user: StoredUser | null;
  token: string | null;
  isAuthenticated: boolean;
  userId: string | null;
  addons: Addons;
  visits: unknown[];
}

interface UserActions {
  login: (user: StoredUser) => void;
  logout: () => void;
  setToken: (token: string) => void;
  setUser: (user: StoredUser) => void;
  setUserId: (userId: string) => void;
  setAddons: (addons: Addons) => void;
  setVisits: (visits: unknown[]) => void;
}

const initialState: UserState = {
  user: null,
  token: null,
  isAuthenticated: false,
  userId: null,
  addons: { feedback: false, about: false, contact: false, landing: false },
  visits: [],
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
          ...initialState,
        });
      },

      setToken: (token) => set({ token, isAuthenticated: !!token }),

      setUser: (user) => set({ user }),

      setUserId: (userId) => set({ userId }),

      setAddons: (addons) => set({ addons }),

      setVisits: (visits) => set({ visits }),
    }),
    {
      name: 'user-store',
      partialize: (state) => ({
        user: state.user,
        isAuthenticated: state.isAuthenticated,
        userId: state.userId,
      }),
    }
  )
);
