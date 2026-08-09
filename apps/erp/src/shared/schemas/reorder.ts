import { z } from 'zod';

export const reorderSuggestionSchema = z.object({
  productId: z.number(),
  name: z.string(),
  sku: z.string(),
  currentStock: z.number().default(0),
  avgDailySales: z.number().default(0),
  suggestedQuantity: z.number().default(0),
  leadTimeDays: z.number().default(0),
  confidence: z.enum(['high', 'medium', 'low']).catch('low'),
});

export type ReorderSuggestionRaw = z.infer<typeof reorderSuggestionSchema>;
