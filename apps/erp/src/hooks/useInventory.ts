import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { erpApi } from '../services/erpApi';

export function useInventoryReport() {
  return useQuery({
    queryKey: ['inventory-report'],
    queryFn: () => erpApi.getInventoryReport().then((r) => r.data || []),
  });
}

export function useLowStock() {
  return useQuery({
    queryKey: ['low-stock'],
    queryFn: () => erpApi.getLowStock().then((r) => r.data || []),
  });
}

export function useAdjustInventory() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: any) => erpApi.adjustInventory(data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['inventory-report'] });
      qc.invalidateQueries({ queryKey: ['low-stock'] });
      qc.invalidateQueries({ queryKey: ['products'] });
    },
  });
}
