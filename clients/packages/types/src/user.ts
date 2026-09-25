/**
 * Every role the platform can store, matching the database exactly:
 * shared/user_apps.sql declares
 *   role ENUM('owner','admin','operator','billing','menu_manager','store_manager','staff','viewer')
 *
 * UserRole listed five of the eight until #795. The three it omitted are real
 * and storable, so a user whose row said 'billing' arrived in the frontend as a
 * value the type claimed could not exist, and every exhaustive switch over
 * UserRole was quietly incomplete.
 *
 * The type is derived from this array rather than declared alongside it, so the
 * runtime list and the compile-time union cannot drift. Adding a role to the
 * ENUM means adding it here, once.
 *
 * Keep in step with PlatformRole in @voxire/permissions, which is where a
 * role's meaning lives. This is only the set of names.
 */
export const USER_ROLES = [
  'owner',
  'admin',
  'operator',
  'billing',
  'menu_manager',
  'store_manager',
  'staff',
  'viewer',
] as const;

export type UserRole = (typeof USER_ROLES)[number];

/**
 * Narrows an untrusted value to a UserRole, or null if it is not one.
 *
 * The wire role used to reach the app through an unchecked cast, so a backend
 * role the frontend had never heard of became a valid UserRole at the type level
 * with nothing complaining at runtime (#795). Callers are expected to fail
 * closed on null rather than substitute a default: a role we cannot interpret is
 * not the same as the least-privileged role we can.
 *
 * Returns null for undefined and for the empty string as well, so "no role" and
 * "a role I do not recognise" are distinguishable only by the caller checking
 * the input, which AppGuard does.
 */
export function parseUserRole(value: unknown): UserRole | null {
  return typeof value === 'string' && (USER_ROLES as readonly string[]).includes(value)
    ? (value as UserRole)
    : null;
}
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
  role?: UserRole;
  businessName?: string;
  storeName?: string;
  planType?: string;
  apps?: AppId[]; // which apps this user has access to
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
