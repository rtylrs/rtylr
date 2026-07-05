import { createOneKitClient } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import {
  WarehouseServiceClient,
  type CreateWarehouseRequest,
  type CreateWarehouseResponse,
  type DeleteWarehouseResponse,
  type UpdateWarehouseRequest,
  type UpdateWarehouseResponse,
  type Warehouse,
} from '@voxire/voxie/erp/warehouse/v1/service_client';

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
      const response = await warehouseClient.listWarehouses({});
      return response.warehouses;
    },
  });
}

export function useCreateWarehouse() {
  const qc = useQueryClient();
  return useMutation<CreateWarehouseResponse, Error, CreateWarehouseRequest>({
    mutationFn: (data) => warehouseClient.createWarehouse(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['warehouses'] }),
    onError: () => toast.error('Failed to create warehouse'),
  });
}

export function useUpdateWarehouse() {
  const qc = useQueryClient();
  return useMutation<UpdateWarehouseResponse, Error, UpdateWarehouseRequest>({
    mutationFn: (data) => warehouseClient.updateWarehouse(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['warehouses'] }),
    onError: () => toast.error('Failed to update warehouse'),
  });
}

export function useDeleteWarehouse() {
  const qc = useQueryClient();
  return useMutation<DeleteWarehouseResponse, Error, string>({
    mutationFn: (id) => warehouseClient.deleteWarehouse({ id }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['warehouses'] }),
    onError: () => toast.error('Failed to delete warehouse'),
  });
}
