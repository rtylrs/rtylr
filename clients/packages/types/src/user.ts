export type AppId =
  | 'dash'
  | 'auth'
  | 'pos'
  | 'erp'
  | 'hr'
  | 'crm'
  | 'finance'
  | 'flow'
  | 'menu'
  | 'landing'
  | 'insights'
  | 'recruit';

export interface User {
  id?: string;
  userId?: string;
  name?: string;
  firstName?: string;
  lastName?: string;
  email?: string;
  emailAddress?: string;
  businessId?: string;
  businessName?: string;
  storeName?: string;
  planType?: string;
  effectiveAccess?: import('./role-access').EffectiveAccess;
  provisioningStatus?: import('./role-access').ProvisioningStatus;
  isConsoleAdmin?: boolean;
  addons?: Record<string, boolean>;
  theme?: string;
  [key: string]: unknown;
}
export interface AuthResponse {
  token: string;
  userId: string;
  user: User;
  visits?: unknown[];
}
