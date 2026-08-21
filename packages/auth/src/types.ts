// Re-exported, not redeclared. This file carried its own five-value copy of
// UserRole while @voxire/types carried another and the database stored eight,
// and every app imports the role type from here rather than from @voxire/types
// (see the *Permissions.ts bridges in erp, crm, hr, finance and insights). So
// the copy that mattered to consumers was the one that was wrong (#795).
import type { UserRole } from '@voxire/types';

export type { UserRole };
export { USER_ROLES, parseUserRole } from '@voxire/types';
export type AppId =
  | 'dash'
  | 'auth'
  | 'pos'
  | 'erp'
  | 'hr'
  | 'crm'
  | 'finance'
  | 'flow'
  | 'insights'
  | 'menu'
  | 'landing'
  | 'recruit';

export interface User {
  id?: string;
  name?: string;
  firstName?: string;
  lastName?: string;
  email?: string;
  emailAddress?: string;
  role?: UserRole;
  businessName?: string;
  storeName?: string;
  planType?: string;
  apps?: AppId[];
  addons?: Record<string, boolean>;
  theme?: string;
  [key: string]: any;
}

export interface UserState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  role: UserRole | null;
  userId: string | null;
  addons: Record<string, boolean>;
  visits: unknown[];
}

export interface UserActions {
  login: (user: User) => void;
  logout: () => void;
  setToken: (token: string) => void;
  setUser: (user: User) => void;
  setRole: (role: UserRole) => void;
  setUserId: (userId: string) => void;
  setAddons: (addons: Record<string, boolean>) => void;
  setVisits: (visits: unknown[]) => void;
}
