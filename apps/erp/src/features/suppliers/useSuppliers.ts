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

export function useSuppliers() {
  return useQuery({
    queryKey: ['suppliers'],
    queryFn: async () => {
      const res = await unwrap(erpApi.listSuppliers());
      return z.array(supplierSchema).parse(unwrapList(res, 'suppliers')) as Supplier[];
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
