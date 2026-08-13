import { createOneKitClient } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { InventoryServiceClient } from '@voxire/voxie/erp/inventory/v1/client';
import type { GetValuationResponse, ValuationItem } from '@voxire/voxie/erp/inventory/v1/types';
import { withErpListBounds } from '../../shared/lib/erpListBounds';

const inventoryClient = createOneKitClient(InventoryServiceClient, 'erp');

export type { ValuationItem };

export function useValuation() {
  return useQuery<GetValuationResponse>({
    queryKey: ['inventory-valuation'],
    queryFn: () => inventoryClient.getValuation(withErpListBounds({})),
  });
}
