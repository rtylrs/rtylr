import { unwrap } from '@voxire/api';
import React, { useEffect, useState } from 'react';
import { AlertTriangle, Package } from 'lucide-react';
import { toast } from 'sonner';
import ScreenLayout from '../../layouts/ScreenLayout';
import SearchBar from '../../shared/components/SearchBar';
import Skeleton from '../../shared/components/Skeleton';
import EmptyState from '../../shared/components/EmptyState';
import { useMenuItems } from '../menu-items/useMenuItems';
import { posService } from '../../shared/services/api/posService';
import type { MenuItem } from '../../shared/types/pos';
import { formatCurrency } from '../../shared/utils/format';
import { buildMenuItemUpdate } from './productQuickEditWorkflow';

const ProductQuickEditScreen: React.FC = () => {
  const { items, loading, error, reload } = useMenuItems();
  const [search, setSearch] = useState('');
  const [drafts, setDrafts] = useState<
    Record<string, { price: number; stock?: number; available: boolean }>
  >({});

  useEffect(() => {
    setDrafts((prev) => {
      const next = { ...prev };
      items.forEach((it) => {
        if (!next[it.id]) {
          next[it.id] = {
            price: it.price,
            stock: it.stock,
            available: it.available !== false,
          };
        }
      });
      return next;
    });
  }, [items]);

  const filtered = items.filter(
    (it) => !search || it.name.toLowerCase().includes(search.toLowerCase())
  );

  const update = (id: string, field: keyof (typeof drafts)[string], value: unknown) =>
    setDrafts((prev) => ({ ...prev, [id]: { ...prev[id], [field]: value } }));

  const save = async (item: MenuItem) => {
    const draft = drafts[item.id];
    try {
      const update = buildMenuItemUpdate(item.id, draft);
      await unwrap(posService.updateMenuItem(update.itemId, update.payload));
      toast.success(`Updated ${item.name}`);
      reload();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Update failed');
    }
  };

  return (
    <ScreenLayout title="Product Quick Edit">
      <div className="p-5 space-y-4 max-w-5xl">
        <SearchBar value={search} onChange={setSearch} placeholder="Search products..." />
        {loading ? (
          <Skeleton rows={8} />
        ) : error ? (
          <EmptyState icon={AlertTriangle} title="Failed to load" message={error} />
        ) : filtered.length === 0 ? (
          <EmptyState icon={Package} title="No products" />
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="text-left text-xs text-pos-muted uppercase border-b border-pos-border">
                <th className="py-2">Item</th>
                <th className="py-2">Current</th>
                <th className="py-2">New price</th>
                <th className="py-2">Stock</th>
                <th className="py-2">Available</th>
                <th className="py-2"></th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((it) => {
                const draft = drafts[it.id] ?? { price: it.price, available: true };
                return (
                  <tr key={it.id} className="border-b border-pos-border">
                    <td className="py-2 text-pos-text">{it.name}</td>
                    <td className="py-2 text-pos-muted">{formatCurrency(it.price)}</td>
                    <td className="py-2">
                      <input
                        type="number"
                        value={draft.price}
                        onChange={(e) => update(it.id, 'price', Number(e.target.value))}
                        className="h-10 w-24 rounded-lg bg-white/8 border border-pos-border px-2 text-pos-text"
                      />
                    </td>
                    <td className="py-2">
                      <input
                        type="number"
                        value={draft.stock ?? ''}
                        onChange={(e) =>
                          update(
                            it.id,
                            'stock',
                            e.target.value ? Number(e.target.value) : undefined
                          )
                        }
                        className="h-10 w-20 rounded-lg bg-white/8 border border-pos-border px-2 text-pos-text"
                      />
                    </td>
                    <td className="py-2">
                      <input
                        type="checkbox"
                        checked={draft.available}
                        onChange={(e) => update(it.id, 'available', e.target.checked)}
                        className="w-5 h-5"
                      />
                    </td>
                    <td className="py-2">
                      <button
                        onClick={() => save(it)}
                        className="h-10 px-4 rounded-lg bg-pos-text text-pos-bg font-semibold text-xs"
                      >
                        Save
                      </button>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
      </div>
    </ScreenLayout>
  );
};

export default ProductQuickEditScreen;
