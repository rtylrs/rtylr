import { createOneKitClient } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import {
  InventoryServiceClient,
  type GetValuationResponse,
  type ValuationItem,
} from '@voxire/voxie/erp/inventory/v1/service_client';

const inventoryClient = createOneKitClient(InventoryServiceClient, 'erp');

export type { ValuationItem };

export function useValuation() {
  return useQuery<GetValuationResponse>({
    queryKey: ['inventory-valuation'],
    queryFn: () => inventoryClient.getValuation({}),
  });
}
