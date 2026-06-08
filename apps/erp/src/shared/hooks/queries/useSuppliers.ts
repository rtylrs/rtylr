import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { Supplier } from '@voxire/contracts/erp/supplier/v1/types';
import type { ListSuppliersResponse } from '@voxire/contracts/erp/supplier/v1/list_suppliers';
import type {
  CreateSupplierRequest,
  CreateSupplierResponse,
} from '@voxire/contracts/erp/supplier/v1/create_supplier';
import type {
  UpdateSupplierRequest,
  UpdateSupplierResponse,
} from '@voxire/contracts/erp/supplier/v1/update_supplier';
import type { UseQueryResult, UseMutationResult } from '@tanstack/react-query';
import { safe } from '../../lib/api';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

export const useListSuppliers = (): UseQueryResult<Supplier[]> => {
  return useQuery<Supplier[]>({
    queryKey: ['erp-suppliers'],
    queryFn: async (): Promise<Supplier[]> => {
      const res = await unwrap(safe.get<ListSuppliersResponse>('/erp/suppliers'));
      return (res as ListSuppliersResponse).suppliers ?? [];
    },
  });
};

export const useCreateSupplier = (): UseMutationResult<
  CreateSupplierResponse,
  Error,
  CreateSupplierRequest
> => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: CreateSupplierRequest): Promise<CreateSupplierResponse> => {
      const res = await unwrap(safe.post<CreateSupplierResponse>('/erp/suppliers', data));
      return res as CreateSupplierResponse;
    },
    onSuccess: (): void => {
      queryClient.invalidateQueries({ queryKey: ['erp-suppliers'] });
      toast.success(t('erp.supplierCreated'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useUpdateSupplier = (): UseMutationResult<
  UpdateSupplierResponse,
  Error,
  UpdateSupplierRequest
> => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: UpdateSupplierRequest): Promise<UpdateSupplierResponse> => {
      const { id, ...body } = data;
      const res = await unwrap(safe.put<UpdateSupplierResponse>(`/erp/suppliers/${id}`, body));
      return res as UpdateSupplierResponse;
    },
    onSuccess: (_res: UpdateSupplierResponse, variables: UpdateSupplierRequest): void => {
      queryClient.invalidateQueries({ queryKey: ['erp-suppliers'] });
      queryClient.invalidateQueries({ queryKey: ['erp-supplier', variables.id] });
      toast.success(t('erp.supplierUpdated'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useDeleteSupplier = (): UseMutationResult<{ message: string }, Error, number> => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (id: number): Promise<{ message: string }> => {
      const res = await unwrap(safe.delete<{ message: string }>(`/erp/suppliers/${id}`));
      return res as { message: string };
    },
    onSuccess: (): void => {
      queryClient.invalidateQueries({ queryKey: ['erp-suppliers'] });
      toast.success(t('erp.supplierDeleted'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};
