import { z } from 'zod';

const finiteNumber = z.preprocess((value) => {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : 0;
}, z.number());

export const dashOrderItemSchema = z
  .object({
    id: z.union([z.string(), z.number()]).transform(String).optional(),
    name: z.string().optional(),
    title: z.string().optional(),
    quantity: finiteNumber.optional(),
    qty: finiteNumber.optional(),
    unitPrice: finiteNumber.optional(),
    unit_price: finiteNumber.optional(),
    price: finiteNumber.optional(),
  })
  .transform((raw) => ({
    id: raw.id ?? '',
    name: raw.name || raw.title || 'Product',
    quantity: raw.quantity ?? raw.qty ?? 1,
    price: raw.unitPrice ?? raw.unit_price ?? raw.price ?? 0,
  }));

export type HubOrderItem = z.infer<typeof dashOrderItemSchema>;

export const dashOrderSchema = z
  .object({
    id: z.union([z.string(), z.number()]).transform(String),
    ownerId: z.union([z.string(), z.number()]).transform(String).optional(),
    owner_id: z.union([z.string(), z.number()]).transform(String).optional(),
    customerEmail: z.string().optional(),
    customer_email: z.string().optional(),
    customerName: z.string().optional(),
    customer_name: z.string().optional(),
    customerPhone: z.string().optional(),
    customer_phone: z.string().optional(),
    status: z.string().catch('pending'),
    subtotal: finiteNumber,
    tax: finiteNumber,
    shippingCost: finiteNumber.optional(),
    shipping_cost: finiteNumber.optional(),
    total: finiteNumber,
    shippingAddress: z.string().optional(),
    shipping_address: z.string().optional(),
    notes: z.string().optional().default(''),
    items: z.array(dashOrderItemSchema).optional(),
    line_items: z.array(dashOrderItemSchema).optional(),
    createdAt: z.string().optional(),
    created_at: z.string().optional(),
    updatedAt: z.string().optional(),
    updated_at: z.string().optional(),
  })
  .transform((raw) => ({
    id: raw.id,
    ownerId: raw.ownerId ?? raw.owner_id,
    customerEmail: raw.customerEmail ?? raw.customer_email ?? '',
    customerName: raw.customerName ?? raw.customer_name ?? '',
    customerPhone: raw.customerPhone ?? raw.customer_phone ?? '',
    status: raw.status,
    subtotal: raw.subtotal,
    tax: raw.tax,
    shippingCost: raw.shippingCost ?? raw.shipping_cost ?? 0,
    total: raw.total,
    shippingAddress: raw.shippingAddress ?? raw.shipping_address ?? '',
    notes: raw.notes,
    items: raw.items ?? raw.line_items ?? [],
    createdAt: raw.createdAt ?? raw.created_at ?? '',
    updatedAt: raw.updatedAt ?? raw.updated_at ?? '',
  }));

export type HubOrder = z.infer<typeof dashOrderSchema>;
