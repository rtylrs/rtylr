import { createOneKitClient } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { SupplierServiceClient } from '@voxire/voxie/erp/supplier/v1/client';
import type {
  CreateSupplierRequest,
  CreateSupplierResponse,
  ListSuppliersResponse,
  Supplier,
  UpdateSupplierRequest,
  UpdateSupplierResponse,
} from '@voxire/voxie/erp/supplier/v1/types';

const supplierClient = createOneKitClient(SupplierServiceClient, 'erp');

export type { CreateSupplierRequest, Supplier, UpdateSupplierRequest };

export interface UseSupplierParams {
  page?: number;
  pageSize?: number;
}

export function useSuppliers(params: UseSupplierParams = {}) {
  const page = params.page ?? 1;
  const pageSize = params.pageSize ?? 50;
  return useQuery<ListSuppliersResponse>({
    queryKey: ['suppliers', { page, pageSize }],
    queryFn: () => supplierClient.listSuppliers({ page, pageSize }),
  });
}

export function useCreateSupplier() {
  const qc = useQueryClient();
  return useMutation<CreateSupplierResponse, Error, CreateSupplierRequest>({
    mutationFn: (data) => supplierClient.createSupplier(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['suppliers'] }),
    onError: () => toast.error('Failed to create supplier'),
  });
}

export function useUpdateSupplier() {
  const qc = useQueryClient();
  return useMutation<UpdateSupplierResponse, Error, UpdateSupplierRequest>({
    mutationFn: (data) => supplierClient.updateSupplier(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['suppliers'] }),
    onError: () => toast.error('Failed to update supplier'),
  });
}
