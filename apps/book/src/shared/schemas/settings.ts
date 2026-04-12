import { z } from 'zod';

export const bookSettingsSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String).optional(),
  businessName: z.string().optional(),
  timezone: z.string().optional(),
  bookingUrl: z.string().optional().nullable(),
  onlineBookingEnabled: z.boolean().optional(),
  bookingLeadTime: z.number().optional(),
  bookingWindow: z.number().optional(),
  slotDuration: z.number().optional(),
  bufferTime: z.number().optional().nullable(),
  businessHoursStart: z.string().optional().nullable(),
  businessHoursEnd: z.string().optional().nullable(),
  allowCancellations: z.boolean().optional(),
  cancellationWindow: z.number().optional(),
  emailConfirmations: z.boolean().optional(),
  smsReminders: z.boolean().optional(),
  reminderAdvance: z.number().optional(),
  reminderEnabled: z.boolean().optional(),
  reminderHours: z.number().optional(),
  cancellationPolicy: z.string().optional().nullable(),
  maxBookingsPerDay: z.number().optional().nullable(),
  workingHours: z.record(z.string(), z.unknown()).optional().nullable(),
});

export type BookSettings = z.infer<typeof bookSettingsSchema>;
