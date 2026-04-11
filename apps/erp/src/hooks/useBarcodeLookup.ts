import { unwrap } from '@voxire/api';
import { useMutation } from '@tanstack/react-query';
import { erpApi, Product } from '../services/erpApi';

export function useBarcodeLookup() {
  return useMutation<Product | null, unknown, string>({
    mutationFn: async (barcode: string) => {
      const res = await erpApi.lookupBarcode(barcode);
      return res.data ?? null;
    },
  });
}
