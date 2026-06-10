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
  isOwner?: boolean;
  isConsoleAdmin?: boolean;
  is_console_admin?: boolean;
  apps?: string[];
  onboarded?: boolean;
  planSlug?: string;
  subscriptionStatus?: string;
  effectiveResources?: Record<string, string>;
  [key: string]: unknown;
}

interface UserState {
  user: StoredUser | null;
  token: string | null;
  isAuthenticated: boolean;
  role: string | null;
  userId: string | null;
  addons: Addons;
  visits: unknown[];
}

interface UserActions {
  login: (user: StoredUser) => void;
  logout: () => void;
  setToken: (token: string) => void;
  setUser: (user: StoredUser) => void;
  setRole: (role: string) => void;
  setUserId: (userId: string) => void;
  setAddons: (addons: Addons) => void;
  setVisits: (visits: unknown[]) => void;
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
        // Hand-rolled `document.cookie = ...` without a Domain attr
        // only clears the cookie for the current host (dash.rtylr.com).
        // The shared `.rtylr.com` cookie survives, so auth.rtylr.com
        // still sees the token on the next navigation and bounces
        // you right back in. Use the shared helper which matches the
        // exact domain set at login.
        removeCookie('token', { domain: getDefaultCookieDomain() });
        // Belt + suspenders: also clear host-only variant in case an
        // older session persisted one with no Domain.
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
