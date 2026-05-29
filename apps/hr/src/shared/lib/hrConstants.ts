/**
 * Domain constants for the HR module.
 * Centralises magic numbers and label maps in one place.
 */

/** Standard number of working hours per week. */
export const STANDARD_WORK_HOURS_PER_WEEK = 40;

/** Maximum number of carry-over leave days allowed per year. */
export const MAX_LEAVE_CARRYOVER_DAYS = 5;

/** Default probation period in calendar days. */
export const DEFAULT_PROBATION_DAYS = 90;

/** Maximum PIN length for attendance clock-in. */
export const ATTENDANCE_PIN_LENGTH = 4;

/** Minimum wage per hour used for validation (USD). */
export const MINIMUM_HOURLY_WAGE_USD = 7.25;

/** Supported contract type values. */
export const CONTRACT_TYPES = ['full_time', 'part_time', 'freelancer', 'intern'] as const;

/** Supported employee status values. */
export const EMPLOYEE_STATUSES = ['active', 'probation', 'inactive', 'terminated'] as const;

/** Leave request status values. */
export const LEAVE_STATUSES = ['pending', 'approved', 'rejected', 'cancelled'] as const;

/** Timesheet status values. */
export const TIMESHEET_STATUSES = ['draft', 'submitted', 'approved', 'rejected'] as const;

/** Number of days before leave end to trigger a renewal reminder. */
export const LEAVE_RENEWAL_REMINDER_DAYS = 7;

/** Maximum number of consecutive overtime hours allowed per day. */
export const MAX_DAILY_OVERTIME_HOURS = 4;
