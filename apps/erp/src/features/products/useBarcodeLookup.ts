import { unwrap } from '@voxire/api';
import { useMutation } from '@tanstack/react-query';
import { toast } from 'sonner';
import { erpApi, Product } from '../../shared/services/erpApi';

export function useBarcodeLookup() {
  return useMutation<Product | null, unknown, string>({
    mutationFn: (barcode: string) => unwrap(erpApi.lookupBarcode(barcode)),
    onError: () => toast.error('Failed to look up barcode'),
  });
}
