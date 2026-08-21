import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { removeCookie, getDefaultCookieDomain } from './cookies';
import type { User, UserState, UserActions, UserRole } from './types';

type UserStore = UserState & UserActions;

export const useUserStore = create<UserStore>()(
  persist(
    (set) => ({
      // State
      user: null,
      token: null,
      isAuthenticated: false,
      role: null,
      userId: null,
      addons: {},
      visits: [],

      // Actions
      login: (user: User) =>
        set({
          user,
          isAuthenticated: true,
          role: user.role ?? null,
          userId: user.id ?? null,
          addons: user.addons ?? {},
        }),

      logout: () => {
        removeCookie('token', { domain: getDefaultCookieDomain() });
        removeCookie('token');
        set({
          user: null,
          token: null,
          isAuthenticated: false,
          role: null,
          userId: null,
          addons: {},
          visits: [],
        });
      },

      setToken: (token: string) => set({ token }),

      setUser: (user: User) =>
        set({
          user,
          role: user.role ?? null,
          userId: user.id ?? null,
          addons: user.addons ?? {},
        }),

      setRole: (role: UserRole) => set({ role }),

      setUserId: (userId: string) => set({ userId }),

      setAddons: (addons: Record<string, boolean>) => set({ addons }),

      setVisits: (visits: unknown[]) => set({ visits }),
    }),
    {
      name: 'rtylr:voxire:user',
      partialize: (state) => ({
        user: state.user,
        isAuthenticated: state.isAuthenticated,
        role: state.role,
        userId: state.userId,
        addons: state.addons,
        visits: state.visits,
        // token is intentionally excluded - stored in cookie only
      }),
    }
  )
);
