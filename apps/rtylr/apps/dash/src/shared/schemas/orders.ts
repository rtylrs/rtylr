import { z } from 'zod';

const finiteNumber = z.preprocess((value) => {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : 0;
}, z.number());

const optionalFiniteNumber = z.preprocess((value) => {
  if (value === undefined || value === null || value === '') return undefined;
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : undefined;
}, z.number().optional());

export const dashOrderItemSchema = z
  .object({
    id: z.union([z.string(), z.number()]).transform(String).optional(),
    orderId: z.union([z.string(), z.number()]).transform(String).optional(),
    order_id: z.union([z.string(), z.number()]).transform(String).optional(),
    menuItemId: z.union([z.string(), z.number()]).transform(String).optional(),
    menu_item_id: z.union([z.string(), z.number()]).transform(String).optional(),
    name: z.string().optional(),
    title: z.string().optional(),
    sku: z.string().optional(),
    quantity: optionalFiniteNumber,
    qty: optionalFiniteNumber,
    unitPrice: optionalFiniteNumber,
    unit_price: optionalFiniteNumber,
    price: optionalFiniteNumber,
    total: optionalFiniteNumber,
    line_total: optionalFiniteNumber,
    discountPct: optionalFiniteNumber,
    discount_pct: optionalFiniteNumber,
    discountAmount: optionalFiniteNumber,
    discount_amount: optionalFiniteNumber,
    taxRate: optionalFiniteNumber,
    tax_rate: optionalFiniteNumber,
    notes: z.string().optional(),
    status: z.string().optional(),
    createdAt: z.string().optional(),
    created_at: z.string().optional(),
  })
  .transform((raw) => {
    const unitPrice = raw.unitPrice ?? raw.unit_price ?? raw.price ?? 0;
    return {
      id: raw.id ?? '',
      orderId: raw.orderId ?? raw.order_id ?? '',
      menuItemId: raw.menuItemId ?? raw.menu_item_id ?? '',
      name: raw.name || raw.title || 'Product',
      sku: raw.sku ?? '',
      quantity: raw.quantity ?? raw.qty ?? 1,
      unitPrice,
      price: unitPrice,
      total: raw.total ?? raw.line_total,
      discountPct: raw.discountPct ?? raw.discount_pct,
      discountAmount: raw.discountAmount ?? raw.discount_amount,
      taxRate: raw.taxRate ?? raw.tax_rate,
      notes: raw.notes ?? '',
      status: raw.status ?? '',
      createdAt: raw.createdAt ?? raw.created_at ?? '',
    };
  });

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
    taxAmount: optionalFiniteNumber,
    tax_amount: optionalFiniteNumber,
    discountAmount: optionalFiniteNumber,
    discount_amount: optionalFiniteNumber,
    shippingCost: optionalFiniteNumber,
    shipping_cost: optionalFiniteNumber,
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
    tax: raw.tax || raw.taxAmount || raw.tax_amount || 0,
    taxAmount: raw.taxAmount ?? raw.tax_amount ?? 0,
    discountAmount: raw.discountAmount ?? raw.discount_amount ?? 0,
    shippingCost: raw.shippingCost ?? raw.shipping_cost ?? 0,
    total: raw.total,
    shippingAddress: raw.shippingAddress ?? raw.shipping_address ?? '',
    notes: raw.notes,
    items: raw.items ?? raw.line_items ?? [],
    createdAt: raw.createdAt ?? raw.created_at ?? '',
    updatedAt: raw.updatedAt ?? raw.updated_at ?? '',
  }));

export type HubOrder = z.infer<typeof dashOrderSchema>;
