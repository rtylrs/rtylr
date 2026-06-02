import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { erpApi } from '../../shared/services/erpApi';
import { unwrapList } from '../../shared/lib/unwrapList';

/**
 * Backend contract: GET /erp/inventory/valuation returns
 * `{ items: ValuationItem[], totalValue }`. After @voxire/api converts
 * snake_case -> camelCase the row shape matches erp/v1 ValuationItem.
 */
export const valuationItemSchema = z.object({
  categoryName: z.string().default(''),
  warehouseName: z.string().default(''),
  totalItems: z.number().default(0),
  totalValue: z.number().default(0),
  totalCost: z.number().default(0),
});

export type ValuationItem = z.infer<typeof valuationItemSchema>;

export interface ValuationResult {
  items: ValuationItem[];
  totalValue: number;
  totalCost: number;
  totalItems: number;
}

export function useValuation() {
  return useQuery({
    queryKey: ['inventory-valuation'],
    queryFn: async (): Promise<ValuationResult> => {
      const res = await unwrap(erpApi.getValuation());
      const items = z.array(valuationItemSchema).parse(unwrapList(res, 'items', 'valuation'));
      const totalValue = items.reduce((sum, i) => sum + i.totalValue, 0);
      const totalCost = items.reduce((sum, i) => sum + i.totalCost, 0);
      const totalItems = items.reduce((sum, i) => sum + i.totalItems, 0);
      return { items, totalValue, totalCost, totalItems };
    },
  });
}
