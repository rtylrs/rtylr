import { z } from 'zod';

export const storeSettingsSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  storeName: z.string(),
  currency: z.string(),
  taxRate: z.number(),
  taxInclusive: z.boolean(),
  orderPrefix: z.string(),
  lowStockThreshold: z.number(),
  emailNotifications: z.boolean(),
  orderConfirmationEmail: z.boolean(),
  shippingNotificationEmail: z.boolean(),
  timezone: z.string(),
});

export type StoreSettingsParsed = z.infer<typeof storeSettingsSchema>;
