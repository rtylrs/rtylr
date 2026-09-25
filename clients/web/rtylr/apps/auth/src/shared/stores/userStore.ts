import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { removeCookie, getCookie, getDefaultCookieDomain } from '@voxire/auth';

export interface UserState {
  token: string | null;
  isAuthenticated: boolean;
  loading: boolean;
  error: string | null;
  role: string | null;
}

interface UserActions {
  login: (role: string) => void;
  logout: () => void;
  clearError: () => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  getToken: () => string | null;
}

type UserStore = UserState & UserActions;

export const useUserStore = create<UserStore>()(
  persist(
    (set, _get) => ({
      // State
      token: null,
      isAuthenticated: false,
      loading: false,
      error: null,
      role: null,

      // Actions
      login: (role: string) =>
        set({
          token: null, // Token is stored in cookie only
          role,
          isAuthenticated: true,
          loading: false,
          error: null,
        }),

      logout: () => {
        removeCookie('token', { domain: getDefaultCookieDomain() });
        set({
          isAuthenticated: false,
          loading: false,
          error: null,
          role: null,
          token: null,
        });
      },

      clearError: () => set({ error: null }),

      setLoading: (loading: boolean) => set({ loading }),

      setError: (error: string | null) => set({ error }),

      getToken: () => getCookie('token'),
    }),
    {
      name: 'rtylr:auth:app:user',
      partialize: (state) => ({
        isAuthenticated: state.isAuthenticated,
        role: state.role,
        // Don't persist token, loading, or error
      }),
    }
  )
);
