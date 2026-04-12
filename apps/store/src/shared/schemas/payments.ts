import { z } from 'zod';

export const paymentMethodSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  name: z.string(),
  provider: z.string(),
  type: z
    .enum(['credit_card', 'paypal', 'apple_pay', 'google_pay', 'cod', 'bank_transfer'])
    .catch('credit_card'),
  enabled: z.boolean(),
  testMode: z.boolean(),
  configuredAt: z.string().optional(),
});

export type PaymentMethodParsed = z.infer<typeof paymentMethodSchema>;
