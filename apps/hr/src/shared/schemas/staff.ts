import { z } from 'zod';

export const staffMemberSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  name: z.string(),
  role: z.string(),
  department: z.string(),
  email: z.string(),
  employmentType: z.enum(['Full-time', 'Part-time', 'Contractor']).catch('Full-time'),
  status: z.enum(['Active', 'Inactive']).catch('Active'),
});

export const employeeProfileSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  photoUrl: z.string().optional(),
  phone: z.string().optional(),
  address: z.string().optional(),
  birthday: z.string().optional(),
  hiredAt: z.string().optional(),
  emergency: z
    .object({
      name: z.string(),
      relation: z.string(),
      phone: z.string(),
    })
    .optional(),
  bio: z.string().optional(),
  managerId: z.number().optional(),
  skills: z.array(z.string()).optional(),
});

export const birthdaySchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  staffId: z.number(),
  name: z.string(),
  date: z.string(),
  department: z.string().optional(),
});

export const anniversarySchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  staffId: z.number(),
  name: z.string(),
  date: z.string(),
  years: z.number(),
});

export type StaffMember = z.infer<typeof staffMemberSchema>;
export type EmployeeProfile = z.infer<typeof employeeProfileSchema>;
export type Birthday = z.infer<typeof birthdaySchema>;
export type Anniversary = z.infer<typeof anniversarySchema>;
