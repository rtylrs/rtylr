export type UserRole = 'owner' | 'admin' | 'operator' | 'staff' | 'viewer';
export type AppId = 'hub' | 'auth' | 'pos' | 'erp' | 'hr' | 'crm' | 'finance' | 'flow' | 'menu' | 'landing' | 'store' | 'book' | 'support' | 'insights' | 'docs';

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
  apps?: AppId[];   // which apps this user has access to
  addons?: Record<string, boolean>;
  theme?: string;
  [key: string]: any;
}
export interface AuthResponse {
  token: string;
  userId: string;
  user: User;
  visits?: any[];
}
