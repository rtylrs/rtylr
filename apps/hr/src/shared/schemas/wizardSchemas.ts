import { z } from 'zod';

export const companyProfileSchema = z.object({
  name: z.string().min(1),
  industry: z.string().min(1),
  country: z.string().min(1),
  timezone: z.string().min(1),
  currency: z.string().min(3).max(3),
});

export const workScheduleSchema = z.object({
  name: z.string().min(1),
  workingDays: z.array(z.string()).min(1),
  startTime: z.string().regex(/^\d{2}:\d{2}$/),
  endTime: z.string().regex(/^\d{2}:\d{2}$/),
  breakMinutes: z.number().int().min(0),
  contractedHoursPerWeek: z.number().positive(),
  dailyOtThresholdMinutes: z.number().int().min(0),
  weeklyOtThresholdMinutes: z.number().int().min(0),
});

export const orgStructureSchema = z.object({
  departments: z.array(
    z.object({
      name: z.string().min(1),
      locationId: z.number().int().positive().optional(),
      parentDepartmentId: z.number().int().positive().optional(),
    })
  ),
});

export const locationsSchema = z.object({
  singleLocation: z.boolean(),
  locations: z.array(
    z.object({
      name: z.string().min(1),
      address: z.string().min(1),
      timezone: z.string().optional(),
    })
  ),
});

export const payrollConfigSchema = z.object({
  payrollGrossModel: z.enum(['hourly', 'salary', 'both']),
  payrollPeriod: z.enum(['weekly', 'biweekly', 'monthly']),
  enableOvertime: z.boolean(),
  enableDeductions: z.boolean(),
  enableBonuses: z.boolean(),
});

export const notificationsSchema = z.object({
  emailEnabled: z.boolean(),
  smsEnabled: z.boolean(),
});

export type CompanyProfileFormData = z.infer<typeof companyProfileSchema>;
export type WorkScheduleFormData = z.infer<typeof workScheduleSchema>;
export type OrgStructureFormData = z.infer<typeof orgStructureSchema>;
export type LocationsFormData = z.infer<typeof locationsSchema>;
export type PayrollConfigFormData = z.infer<typeof payrollConfigSchema>;
export type NotificationsFormData = z.infer<typeof notificationsSchema>;
