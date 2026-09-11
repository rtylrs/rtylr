import { createOneKitClient } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { WarehouseServiceClient } from '@voxire/voxie/erp/warehouse/v1/client';
import type {
  CreateWarehouseRequest,
  CreateWarehouseResponse,
  DeleteWarehouseResponse,
  UpdateWarehouseRequest,
  UpdateWarehouseResponse,
  Warehouse,
} from '@voxire/voxie/erp/warehouse/v1/types';
import { withErpListBounds } from '../../shared/lib/erpListBounds';

const warehouseClient = createOneKitClient(WarehouseServiceClient, 'erp');

export type {
  CreateWarehouseRequest as CreateWarehouseInput,
  UpdateWarehouseRequest as UpdateWarehouseInput,
  Warehouse,
};

export function useWarehouses() {
  return useQuery<Warehouse[]>({
    queryKey: ['warehouses'],
    queryFn: async () => {
      const response = await warehouseClient.listWarehouses(withErpListBounds({}));
      return response.warehouses ?? [];
    },
  });
}

export function useCreateWarehouse() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation<CreateWarehouseResponse, Error, CreateWarehouseRequest>({
    mutationFn: (data) => warehouseClient.createWarehouse(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['warehouses'] }),
    onError: () => toast.error(t('erp.warehouses.createFailed')),
  });
}

export function useUpdateWarehouse() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation<UpdateWarehouseResponse, Error, UpdateWarehouseRequest>({
    mutationFn: (data) => warehouseClient.updateWarehouse(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['warehouses'] }),
    onError: () => toast.error(t('erp.warehouses.updateFailed')),
  });
}

export function useDeleteWarehouse() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation<DeleteWarehouseResponse, Error, string>({
    mutationFn: (id) => warehouseClient.deleteWarehouse({ id }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['warehouses'] }),
    onError: () => toast.error(t('erp.warehouses.deleteFailed')),
  });
}
