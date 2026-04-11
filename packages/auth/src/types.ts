export type UserRole = 'owner' | 'admin' | 'operator' | 'staff' | 'viewer';
export type AppId =
  | 'hub'
  | 'auth'
  | 'pos'
  | 'erp'
  | 'hr'
  | 'crm'
  | 'finance'
  | 'flow'
  | 'menu'
  | 'landing';

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
  visits: any[];
}

export interface UserActions {
  login: (user: User) => void;
  logout: () => void;
  setToken: (token: string) => void;
  setUser: (user: User) => void;
  setRole: (role: UserRole) => void;
  setUserId: (userId: string) => void;
  setAddons: (addons: Record<string, boolean>) => void;
  setVisits: (visits: any[]) => void;
}
