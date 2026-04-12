import { z } from 'zod';

export const storeOrderSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  orderNumber: z.string(),
  customerName: z.string(),
  customerEmail: z.string(),
  status: z
    .enum(['pending', 'confirmed', 'shipped', 'delivered', 'cancelled', 'refunded'])
    .catch('pending'),
  total: z.number(),
  subtotal: z.number(),
  shippingCost: z.number(),
  tax: z.number(),
  itemCount: z.number(),
  paymentMethod: z.string(),
  shippingMethod: z.string().optional(),
  trackingNumber: z.string().optional(),
  createdAt: z.string(),
  updatedAt: z.string(),
});

export type StoreOrderParsed = z.infer<typeof storeOrderSchema>;
