import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { erpApi, CreateSupplierInput, UpdateSupplierInput } from '../services/erpApi';

export function useSuppliers() {
  return useQuery({
    queryKey: ['suppliers'],
    queryFn: () => erpApi.listSuppliers().then((r) => r.data || []),
  });
}

export function useCreateSupplier() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateSupplierInput) => erpApi.createSupplier(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['suppliers'] }),
  });
}

export function useUpdateSupplier() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateSupplierInput) => erpApi.updateSupplier(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['suppliers'] }),
  });
}
