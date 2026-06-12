import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { posService } from '../../shared/services/api/posService';
import type { MenuItem } from '../../shared/types/pos';
import { menuItemSchema } from '../../shared/utils/schemas/menuItems';

interface UseMenuItemsResult {
  items: MenuItem[];
  categories: string[];
  loading: boolean;
  error: string | null;
  reload: () => void;
}

export function useMenuItems(): UseMenuItemsResult {
  const query = useQuery({
    queryKey: ['pos-menu-items'],
    queryFn: async (): Promise<MenuItem[]> => {
      const res = await unwrap(posService.getMenuItems());
      const raw = res?.items ?? res?.data ?? [];
      const parsed = z.array(menuItemSchema).parse(Array.isArray(raw) ? raw : []);
      return parsed.map((it) => ({
        id: it.id,
        name: it.name,
        price: it.price,
        category: it.category,
        image: it.image ?? undefined,
        available: it.available,
        sku: it.sku ?? undefined,
        stock: it.stock ?? undefined,
        description: it.description ?? undefined,
        taxRate: it.taxRate ?? undefined,
        barcode: it.barcode ?? undefined,
      }));
    },
    staleTime: 2 * 60 * 1000,
  });

  const items = query.data ?? [];
  const categories = Array.from(new Set(items.map((i) => i.category))).sort();

  return {
    items,
    categories,
    loading: query.isLoading,
    error: query.error
      ? query.error instanceof Error
        ? query.error.message
        : String(query.error)
      : null,
    reload: () => void query.refetch(),
  };
}
