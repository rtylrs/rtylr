import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface Addons {
  feedback: boolean;
  about: boolean;
  contact: boolean;
  landing: boolean;
}

export interface StoredUser {
  // Auth fields
  email?: string;
  emailAddress?: string;
  role?: string;
  fullName?: string;
  name?: string;
  firstName?: string;
  lastName?: string;
  // Business fields
  businessName?: string;
  storeName?: string;
  businessDescription?: string;
  coverImage?: string;
  logo?: string;
  address?: string;
  phone?: string;
  hours?: string;
  averageRating?: number;
  reviewCount?: number;
  // Permissions
  isOwner?: boolean;
  apps?: string[];
  // Onboarding
  onboarded?: boolean;
  // Server can send additional fields
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
