import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import {
  erpApi,
  CreateSupplierInput,
  UpdateSupplierInput,
  Supplier,
} from '../../shared/services/erpApi';
import { supplierSchema } from '../../shared/schemas/suppliers';
import { unwrapList } from '../../shared/lib/unwrapList';

export interface SupplierListResult {
  suppliers: Supplier[];
  hasMore: boolean;
  page: number;
  pageSize: number;
}

export interface UseSupplierParams {
  page?: number;
  pageSize?: number;
}

export function useSuppliers(params: UseSupplierParams = {}) {
  const page = params.page ?? 1;
  const pageSize = params.pageSize ?? 50;
  return useQuery<SupplierListResult>({
    queryKey: ['suppliers', { page, pageSize }],
    queryFn: async () => {
      const res = await unwrap(erpApi.listSuppliers({ page, page_size: pageSize }));
      const suppliers = z.array(supplierSchema).parse(unwrapList(res, 'suppliers')) as Supplier[];
      const hasMore = (res as unknown as Record<string, unknown>)?.has_more === true;
      return { suppliers, hasMore, page, pageSize };
    },
  });
}

export function useCreateSupplier() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateSupplierInput) => unwrap(erpApi.createSupplier(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['suppliers'] }),
    onError: () => toast.error('Failed to create supplier'),
  });
}

export function useUpdateSupplier() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateSupplierInput) => unwrap(erpApi.updateSupplier(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['suppliers'] }),
    onError: () => toast.error('Failed to update supplier'),
  });
}
