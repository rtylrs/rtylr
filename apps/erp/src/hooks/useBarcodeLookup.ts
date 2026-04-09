import { useMutation } from '@tanstack/react-query';
import { erpApi, Product } from '../services/erpApi';
import { mockProducts } from '../lib/mockData';

export function useBarcodeLookup() {
  return useMutation<Product | null, unknown, string>({
    mutationFn: async (barcode: string) => {
      try {
        const res = await erpApi.lookupBarcode(barcode);
        if (res.data) return res.data;
      } catch {
        // fall through to mock
      }
      const match = mockProducts.find((p) => p.barcode === barcode || p.sku === barcode);
      return match || null;
    },
  });
}
