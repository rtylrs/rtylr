import { createOneKitClient } from '@voxire/api';
import React, { useEffect, useMemo, useState } from 'react';
import { AlertTriangle, Package } from 'lucide-react';
import { toast } from 'sonner';
import { MenuItemServiceClient } from '@voxire/voxie/pos/menu_item/v1/client';
import type { MenuItem } from '@voxire/voxie/pos/menu_item/v1/types';
import ScreenLayout from '../../layouts/ScreenLayout';
import SearchBar from '../../shared/components/SearchBar';
import Skeleton from '../../shared/components/Skeleton';
import EmptyState from '../../shared/components/EmptyState';
import { useMenuItems } from '../menu-items/useMenuItems';
import { formatCurrency } from '../../shared/utils/format';

const menuItemClient = createOneKitClient(MenuItemServiceClient, 'pos');

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
        const id = it.id ?? '';
        if (!next[id]) {
          next[id] = {
            price: it.salePrice ?? 0,
            stock: it.stock,
            available: it.isActive !== false,
          };
        }
      });
      return next;
    });
  }, [items]);

  const searchTerm = search.trim().toLowerCase();
  const filtered = useMemo(
    () => items.filter((it) => !searchTerm || (it.name ?? '').toLowerCase().includes(searchTerm)),
    [items, searchTerm]
  );

  const update = (id: string, field: keyof (typeof drafts)[string], value: unknown) =>
    setDrafts((prev) => ({ ...prev, [id]: { ...prev[id], [field]: value } }));

  const save = async (item: MenuItem) => {
    const draft = drafts[item.id ?? ''];
    try {
      if (!draft) {
        throw new Error('Select a valid product before saving.');
      }
      // Reject invalid input instead of silently publishing a free/untracked
      // product (Number('') === 0).
      if (!Number.isFinite(draft.price) || draft.price < 0) {
        throw new Error('Enter a valid price before saving.');
      }
      if (!Number.isFinite(draft.stock) || draft.stock == null || draft.stock < 0) {
        throw new Error('Enter a valid stock count before saving.');
      }
      await menuItemClient.updateMenuItem({
        id: item.id ?? '',
        body: {
          name: item.name,
          salePrice: draft.price,
          stock: draft.stock,
          isActive: draft.available,
        },
      });
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
                const itemId = it.id ?? '';
                const draft = drafts[itemId] ?? { price: it.salePrice ?? 0, available: true };
                return (
                  <tr key={itemId} className="border-b border-pos-border">
                    <td className="py-2 text-pos-text">{it.name}</td>
                    <td className="py-2 text-pos-muted">{formatCurrency(it.salePrice ?? 0)}</td>
                    <td className="py-2">
                      <input
                        type="number"
                        value={draft.price}
                        onChange={(e) => update(itemId, 'price', Number(e.target.value))}
                        className="h-10 w-24 rounded-lg bg-white/8 border border-pos-border px-2 text-pos-text"
                      />
                    </td>
                    <td className="py-2">
                      <input
                        type="number"
                        value={draft.stock ?? ''}
                        onChange={(e) =>
                          update(
                            itemId,
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
                        onChange={(e) => update(itemId, 'available', e.target.checked)}
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
