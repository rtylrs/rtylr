import { z } from 'zod';

export const appointmentSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  serviceId: z.union([z.string(), z.number()]).transform(String).optional(),
  resourceId: z.union([z.string(), z.number()]).transform(String).optional().nullable(),
  customerId: z.union([z.string(), z.number()]).transform(String).optional().nullable(),
  staffId: z.union([z.string(), z.number()]).transform(String).optional().nullable(),
  customerName: z.string().optional(),
  customerEmail: z.string().optional().nullable(),
  customerPhone: z.string().optional().nullable(),
  serviceName: z.string().optional(),
  staffName: z.string().optional().nullable(),
  status: z.enum(['pending', 'confirmed', 'completed', 'cancelled', 'no_show']).catch('pending'),
  startTime: z.string(),
  endTime: z.string(),
  date: z.string().optional(),
  duration: z.number().optional(),
  notes: z.string().optional().nullable(),
  price: z.number().optional().nullable(),
  createdAt: z.string().optional(),
  updatedAt: z.string().optional(),
});

export type Appointment = z.infer<typeof appointmentSchema>;
