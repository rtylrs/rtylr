import { z } from 'zod';

export const bookCustomerSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  name: z.string(),
  email: z.string().optional().nullable(),
  phone: z.string().optional().nullable(),
  notes: z.string().optional().nullable(),
  totalBookings: z.number().optional(),
  totalAppointments: z.number().optional(),
  preferences: z.string().optional().nullable(),
  lastAppointmentAt: z.string().optional().nullable(),
  lastVisit: z.string().optional().nullable(),
  createdAt: z.string().optional(),
  updatedAt: z.string().optional(),
});

export type BookCustomer = z.infer<typeof bookCustomerSchema>;
