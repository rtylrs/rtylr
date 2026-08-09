import { z } from 'zod';

export const REPORT_TYPES = [
  'attendance_summary',
  'leave_balance',
  'payroll_summary',
  'overtime_report',
  'headcount',
  'turnover',
  'timesheet_summary',
  'shift_coverage',
  'department_cost',
  'employee_directory',
] as const;

export type ReportType = (typeof REPORT_TYPES)[number];

export const reportRequestSchema = z.object({
  reportType: z.enum(REPORT_TYPES),
  dateFrom: z.string(),
  dateTo: z.string(),
  locationId: z.number().optional(),
  departmentId: z.number().optional(),
  employeeId: z.number().optional(),
  format: z.enum(['json', 'csv', 'pdf']).default('json'),
});

export type ReportRequest = z.infer<typeof reportRequestSchema>;

export const reportRowSchema = z.record(z.string(), z.unknown());

export const reportPayloadSchema = z.object({
  reportType: z.string(),
  generatedAt: z.number(),
  columns: z.array(z.string()).default([]),
  rows: z.array(reportRowSchema).default([]),
  summary: z.record(z.string(), z.unknown()).default({}),
});

export type ReportPayload = z.infer<typeof reportPayloadSchema>;
