import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { erpApi } from '../../shared/services/erpApi';

export function useInventoryReport() {
  return useQuery({
    queryKey: ['inventory-report'],
    queryFn: () => unwrap(erpApi.getInventoryReport()),
  });
}

export function useLowStock() {
  return useQuery({
    queryKey: ['low-stock'],
    queryFn: () => unwrap(erpApi.getLowStock()),
  });
}

export function useAdjustInventory() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { productId: number; delta: number }) =>
      unwrap(erpApi.adjustInventory(data)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['inventory-report'] });
      qc.invalidateQueries({ queryKey: ['low-stock'] });
      qc.invalidateQueries({ queryKey: ['products'] });
    },
  });
}
