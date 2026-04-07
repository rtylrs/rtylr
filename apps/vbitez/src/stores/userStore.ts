import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import api from '../lib/api';

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
  loading: boolean;
  error: string | null;
  role: string | null;
  userId: string | null;
  addons: Addons;
  visits: any[];
}

interface UserActions {
  login: (user: any) => void;
  logout: () => void;
  clearError: () => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  setToken: (token: string) => void;
  setUser: (user: any) => void;
  setRole: (role: string) => void;
  setUserId: (userId: string) => void;
  setAddons: (addons: Addons) => void;
  setVisits: (visits: any[]) => void;
  loginAsync: (payload: { email: string; password: string }) => Promise<void>;
  registerRestaurantAsync: (payload: { email: string; password: string; name: string; referred: string }) => Promise<void>;
  restaurantWithGoogleAsync: (payload: { idToken: string }) => Promise<void>;
  getUserInfo: () => Promise<void>;
  setThemeAsync: (payload: { id: string; themeData: any }) => Promise<void>;
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

      login: (user) => set({ user, isAuthenticated: true }),

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

      setLoading: (loading) => set({ loading }),

      setError: (error) => set({ error }),

      setToken: (token) => set({ token }),

      setUser: (user) => set({ user }),

      setRole: (role) => set({ role }),

      setUserId: (userId) => set({ userId }),

      setAddons: (addons) => set({ addons }),

      setVisits: (visits) => set({ visits }),

      loginAsync: async (payload) => {
        set({ loading: true, error: null });
        try {
          const response = await api.post('/auth/login', payload);
          const data = response.data;
          localStorage.setItem('token', data.token);
          set({
            token: data.token,
            isAuthenticated: true,
            loading: false,
            error: null,
          });
        } catch (error: any) {
          set({
            loading: false,
            error: error.response?.data?.message || error.response?.data?.error || 'Login failed',
          });
          throw error;
        }
      },

      registerRestaurantAsync: async (payload) => {
        set({ loading: true, error: null });
        try {
          const response = await api.post('/auth/create/business', payload);
          const data = response.data;
          localStorage.setItem('token', data.token);
          set({
            token: data.token,
            isAuthenticated: true,
            loading: false,
            error: null,
          });
        } catch (error: any) {
          set({
            loading: false,
            error: error.response?.data?.message || error.response?.data?.error || 'Registration failed',
          });
          throw error;
        }
      },

      restaurantWithGoogleAsync: async (payload) => {
        set({ loading: true, error: null });
        try {
          const response = await api.post('/auth/create/business/google', payload);
          const data = response.data;
          localStorage.setItem('token', data.token);
          set({
            token: data.token,
            isAuthenticated: true,
            loading: false,
            error: null,
          });
        } catch (error: any) {
          set({
            loading: false,
            error: error.response?.data?.message || error.response?.data?.error || 'Google login failed',
          });
          throw error;
        }
      },

      getUserInfo: async () => {
        set({ loading: true, error: null });
        try {
          const response = await api.get('/business/me');
          const { userId, token, user, visits } = response.data;
          set({
            role: user.role,
            userId: userId,
            token: token,
            user: user,
            isAuthenticated: true,
            loading: false,
            addons: user.addons,
            visits: visits,
          });
        } catch (error: any) {
          set({
            loading: false,
            error: error.response?.data?.message || error.response?.data?.error || 'Failed to get user info',
          });
          throw error;
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
        } catch (error: any) {
          set({
            loading: false,
            error: error.response?.data?.message || error.response?.data?.error || 'Failed to update theme',
          });
          throw error;
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