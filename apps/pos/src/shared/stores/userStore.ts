import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import api from '../services/api/api';

interface Addons {
  feedback: boolean;
  about: boolean;
  contact: boolean;
  landing: boolean;
}

interface ApiError extends Error {
  response?: { data?: { message?: string } };
}

export interface StoredUser {
  role?: string;
  addons?: Addons;
  name?: string;
  email?: string;
  theme?: Record<string, unknown>;
  [key: string]: unknown;
}

interface UserState {
  user: StoredUser | null;
  token: string | null;
  isAuthenticated: boolean;
  loading: boolean;
  error: string | null;
  role: string | null;
  userId: string | null;
  addons: Addons;
  visits: unknown[];
}

interface UserActions {
  logout: () => void;
  clearError: () => void;
  setToken: (token: string) => void;
  setUser: (user: StoredUser) => void;
  getUserInfo: () => Promise<void>;
  setThemeAsync: (payload: { id: string; themeData: Record<string, unknown> }) => Promise<void>;
}

const initialState: UserState = {
  user: null,
  token: null,
  isAuthenticated: false,
  loading: false,
  error: null,
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

      logout: () => {
        localStorage.removeItem('token');
        set({
          user: null,
          isAuthenticated: false,
          loading: false,
          error: null,
          role: null,
          userId: null,
          token: null,
        });
      },

      clearError: () => set({ error: null }),

      setToken: (token) => set({ token }),

      setUser: (user) => set({ user }),

      getUserInfo: async () => {
        set({ loading: true, error: null });
        try {
          const response = await api.get('/business/me');
          const { userId, token, user, visits } = response.data;
          set({
            role: user.role,
            userId,
            token,
            user,
            isAuthenticated: true,
            loading: false,
            addons: user.addons,
            visits,
          });
        } catch (err) {
          const error = err as ApiError;
          set({
            loading: false,
            error: error.response?.data?.message || 'Failed to restore session',
          });
          throw err;
        }
      },

      setThemeAsync: async (payload) => {
        set({ loading: true, error: null });
        try {
          const response = await api.put(`/business/${payload.id}/theme`, payload.themeData);
          const state = get();
          if (state.user) {
            set({
              user: { ...state.user, theme: response.data.theme },
              loading: false,
              error: null,
            });
          }
        } catch (err) {
          const error = err as ApiError;
          set({
            loading: false,
            error: error.response?.data?.message || 'Failed to update theme',
          });
          throw err;
        }
      },
    }),
    {
      name: 'user-store',
      partialize: (state) => ({
        user: state.user,
        token: state.token,
        isAuthenticated: state.isAuthenticated,
        role: state.role,
        userId: state.userId,
        addons: state.addons,
        visits: state.visits,
      }),
    }
  )
);
