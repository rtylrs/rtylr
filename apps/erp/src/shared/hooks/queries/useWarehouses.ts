import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { UseQueryResult, UseMutationResult } from '@tanstack/react-query';
import { safe } from '../../lib/api';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

// No proto type yet - define locally.
export interface Warehouse {
  id: number;
  name: string;
  address: string;
  manager: string;
  capacity: number;
  used: number;
  isDefault: boolean;
}

export interface CreateWarehouseInput {
  name: string;
  address: string;
  manager: string;
  capacity: number;
  isDefault?: boolean;
}

interface ListWarehousesResponse {
  warehouses: Warehouse[];
}

export const useListWarehouses = (): UseQueryResult<Warehouse[]> => {
  return useQuery<Warehouse[]>({
    queryKey: ['erp-warehouses'],
    queryFn: async (): Promise<Warehouse[]> => {
      const res = await unwrap(safe.get<ListWarehousesResponse>('/erp/warehouses'));
      return (res as ListWarehousesResponse).warehouses ?? [];
    },
  });
};

export const useCreateWarehouse = (): UseMutationResult<Warehouse, Error, CreateWarehouseInput> => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: CreateWarehouseInput): Promise<Warehouse> => {
      const res = await unwrap(
        safe.post<Warehouse>('/erp/warehouses', {
          name: data.name,
          address: data.address,
          managerName: data.manager,
          isDefault: data.isDefault,
        })
      );
      return res as Warehouse;
    },
    onSuccess: (): void => {
      queryClient.invalidateQueries({ queryKey: ['erp-warehouses'] });
      toast.success(t('erp.warehouseCreated'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};
