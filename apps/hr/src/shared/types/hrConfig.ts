/**
 * Configuration returned by the live `/hr/config` REST endpoint.
 *
 * The generated setup payload predates the core HR configuration route and
 * only describes the wizard's legacy capability flags. Keep the server-backed
 * fields here, while retaining optional legacy fields used by older widgets.
 */
export interface HRConfig {
  id?: string | number;
  ownerId?: string | number;
  companyName?: string;
  fiscalYearStartMonth?: number;
  defaultCurrency?: string;
  defaultTimezone?: string;
  defaultWorkScheduleId?: string | number;
  overtimeEnabled?: boolean;
  overtimeMultiplier?: number;
  emailNotificationsEnabled?: boolean;
  smsNotificationsEnabled?: boolean;
  timeTrackingModel?: string;
  timesheetPeriod?: string;
  payrollPeriod?: string;
  payrollGrossModel?: string;
  setupCompletedAt?: string | number;
  createdAt?: string | number;
  updatedAt?: string | number;
  clockInMethods?: string[];
  subscribedApps?: string[];
}
