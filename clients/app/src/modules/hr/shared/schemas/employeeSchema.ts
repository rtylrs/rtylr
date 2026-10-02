import { z } from 'zod';

export const employeeSchema = z.object({
  userId: z.number().int().min(0),
  departmentId: z.number().int().positive(),
  locationId: z.number().int().positive(),
  firstName: z.string().trim().min(1),
  lastName: z.string().trim().min(1),
  email: z.string().trim().email(),
  managerId: z.number().int().positive().optional(),
  contractType: z.enum(['full_time', 'part_time', 'freelancer']),
  status: z.enum(['active', 'probation', 'inactive', 'terminated']),
  hireDate: z.string().min(1),
  probationEndDate: z.string().optional(),
  terminationDate: z.string().optional(),
  hourlyRate: z.number().positive().optional(),
  monthlySalary: z.number().positive().optional(),
  currency: z.string().min(3).max(3),
  workScheduleId: z.number().int().positive().optional(),
  accountCreationMode: z.enum(['hr_created', 'invite_link']),
  jobTitle: z.string().min(1),
});

export type EmployeeFormData = z.infer<typeof employeeSchema>;
