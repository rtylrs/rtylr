import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useEffect, useState } from 'react';
import { posService } from '../../shared/services/posService';
import type { MenuItem } from '../../shared/types/pos';
import { menuItemSchema } from '../../shared/schemas/menuItems';

interface UseMenuItemsResult {
  items: MenuItem[];
  categories: string[];
  loading: boolean;
  error: string | null;
  reload: () => Promise<void>;
}

export function useMenuItems(): UseMenuItemsResult {
  const [items, setItems] = useState<MenuItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await unwrap(posService.getMenuItems());
      const raw = res?.items ?? res?.data ?? [];
      const parsed = z.array(menuItemSchema).parse(Array.isArray(raw) ? raw : []);
      const mapped: MenuItem[] = parsed.map((it) => ({
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
      setItems(mapped);
    } catch (e: any) {
      setError(e?.response?.data?.message ?? e?.message ?? 'Failed to load menu');
      setItems([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const categories = Array.from(new Set(items.map((i) => i.category))).sort();
  return { items, categories, loading, error, reload: load };
}
