import { createOneKitClient } from '@voxire/api';
import { useMutation } from '@tanstack/react-query';
import { toast } from 'sonner';
import { ProductServiceClient } from '@voxire/voxie/erp/product/v1/client';
import type { Product } from '@voxire/voxie/erp/product/v1/types';

const productClient = createOneKitClient(ProductServiceClient, 'erp');

export function useBarcodeLookup() {
  return useMutation<Product | null, unknown, string>({
    mutationFn: async (barcode: string) => {
      const response = await productClient.getProductByBarcode({ code: barcode });
      return response.product ?? null;
    },
    onError: () => toast.error('Failed to look up barcode'),
  });
}
