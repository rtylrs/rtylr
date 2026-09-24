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

export async function fetchAllSuppliers(pageSize: number): Promise<ListSuppliersResponse> {
  const suppliers: Supplier[] = [];
  const seenPages = new Set<string>();
  let page = 1;
  let firstResponse: ListSuppliersResponse | undefined;

  while (true) {
    const response = await supplierClient.listSuppliers({ page, pageSize });
    firstResponse ??= response;
    const pageSuppliers = response.suppliers ?? [];
    const pageSignature = pageSuppliers
      .map((supplier) => supplier.id ?? `${supplier.name ?? ''}:${supplier.email ?? ''}`)
      .join('|');
    if (pageSuppliers.length > 0 && seenPages.has(pageSignature)) {
      throw new Error('Supplier pagination did not advance');
    }
    if (pageSuppliers.length > 0) seenPages.add(pageSignature);
    suppliers.push(...pageSuppliers);

    if (!response.hasMore || pageSuppliers.length === 0) {
      return {
        ...(firstResponse ?? {}),
        suppliers,
        hasMore: false,
        page: 1,
        pageSize: suppliers.length || pageSize,
      };
    }
    page += 1;
  }
}

export function useSuppliers(params: UseSupplierParams = {}) {
  const hasExplicitPaging = params.page !== undefined || params.pageSize !== undefined;
  const page = params.page ?? 1;
  const pageSize = params.pageSize ?? 50;
  return useQuery<ListSuppliersResponse>({
    queryKey: hasExplicitPaging ? ['suppliers', { page, pageSize }] : ['suppliers', 'all'],
    queryFn: () =>
      hasExplicitPaging
        ? supplierClient.listSuppliers({ page, pageSize })
        : fetchAllSuppliers(pageSize),
  });
}

export function useCreateSupplier() {
  const qc = useQueryClient();
  return useMutation<CreateSupplierResponse, Error, CreateSupplierRequest>({
    mutationFn: (data) => supplierClient.createSupplier(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['suppliers'] }),
    onError: () => {
      void qc.invalidateQueries({ queryKey: ['suppliers'] });
      toast.error('Failed to create supplier');
    },
  });
}

export function useUpdateSupplier() {
  const qc = useQueryClient();
  return useMutation<UpdateSupplierResponse, Error, UpdateSupplierRequest>({
    mutationFn: (data) => supplierClient.updateSupplier(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['suppliers'] }),
    onError: () => {
      void qc.invalidateQueries({ queryKey: ['suppliers'] });
      toast.error('Failed to update supplier');
    },
  });
}
