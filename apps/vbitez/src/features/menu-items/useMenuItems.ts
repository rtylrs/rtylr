import { useEffect, useState } from 'react';
import { posService } from '../../shared/services/posService';
import type { MenuItem } from '../../shared/types/pos';

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
      const res = await posService.getMenuItems();
      const raw = res?.data?.items ?? res?.data ?? [];
      const mapped: MenuItem[] = (Array.isArray(raw) ? raw : []).map((it: any) => ({
        id: String(it.id ?? it.sku ?? Math.random().toString(36).slice(2)),
        name: it.name ?? 'Item',
        price: Number(it.price ?? 0),
        category: it.category ?? 'Other',
        image: it.image ?? it.imageUrl,
        available: it.available !== false,
        sku: it.sku,
        stock: it.stock,
        description: it.description,
        taxRate: it.taxRate,
        barcode: it.barcode,
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
