import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { erpApi, CreateSupplierInput, UpdateSupplierInput } from '../services/erpApi';

export function useSuppliers() {
  return useQuery({
    queryKey: ['suppliers'],
    queryFn: () => unwrap(erpApi.listSuppliers()),
  });
}

export function useCreateSupplier() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateSupplierInput) => unwrap(erpApi.createSupplier(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['suppliers'] }),
  });
}

export function useUpdateSupplier() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateSupplierInput) => unwrap(erpApi.updateSupplier(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['suppliers'] }),
  });
}
