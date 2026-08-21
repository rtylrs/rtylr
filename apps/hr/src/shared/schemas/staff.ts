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

/**
 * The single-record projection. `getStaff` already returns phone, address,
 * start_date and both emergency_contact_* fields; `staffMemberSchema` is the
 * list shape and deliberately drops them. Before this existed, the detail screen
 * read those five values from `/hr/staff/detail`, which no onion module serves,
 * so they rendered empty while the data sat unused in the response the same
 * screen had already fetched. See #875.
 */
export const staffDetailSchema = staffMemberSchema.extend({
  photoUrl: z.string().optional(),
  bio: z.string().optional(),
  phone: z.string().optional(),
  address: z.string().optional(),
  startDate: z.string().optional(),
  emergencyContactName: z.string().optional(),
  emergencyContactPhone: z.string().optional(),
});

export const myEmployeeSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
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
export type StaffDetail = z.infer<typeof staffDetailSchema>;
export type Birthday = z.infer<typeof birthdaySchema>;
export type Anniversary = z.infer<typeof anniversarySchema>;
