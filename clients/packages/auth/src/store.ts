import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { removeCookie, getDefaultCookieDomain } from './cookies';
import type { User, UserState, UserActions } from './types';

type UserStore = UserState & UserActions;

export const useUserStore = create<UserStore>()(
  persist(
    (set) => ({
      // State
      user: null,
      token: null,
      isAuthenticated: false,
      userId: null,
      addons: {},
      effectiveAccess: null,
      visits: [],

      // Actions
      login: (user: User) =>
        set({
          user,
          isAuthenticated: true,
          userId: user.id ?? null,
          addons: user.addons ?? {},
          effectiveAccess: user.effectiveAccess ?? null,
        }),

      logout: () => {
        removeCookie('token', { domain: getDefaultCookieDomain() });
        removeCookie('token');
        set({
          user: null,
          token: null,
          isAuthenticated: false,
          userId: null,
          addons: {},
          effectiveAccess: null,
          visits: [],
        });
      },

      setToken: (token: string) => set({ token }),

      setUser: (user: User) =>
        set({
          user,
          userId: user.id ?? null,
          addons: user.addons ?? {},
          effectiveAccess: user.effectiveAccess ?? null,
        }),

      setUserId: (userId: string) => set({ userId }),

      setAddons: (addons: Record<string, boolean>) => set({ addons }),

      setEffectiveAccess: (effectiveAccess) => set({ effectiveAccess }),

      setVisits: (visits: unknown[]) => set({ visits }),
    }),
    {
      name: 'rtylr:voxire:user',
      partialize: (state) => ({
        user: state.user,
        isAuthenticated: state.isAuthenticated,
        userId: state.userId,
        addons: state.addons,
        effectiveAccess: state.effectiveAccess,
        visits: state.visits,
        // token is intentionally excluded - stored in cookie only
      }),
    }
  )
);
