import React, { useCallback, useMemo, useState } from 'react';
import { AlertTriangle, Search } from 'lucide-react';
import { toast } from 'sonner';
import ScreenLayout from '../../layouts/ScreenLayout';
import SearchBar from '../../shared/components/SearchBar';
import CategoryTabs from '../../shared/components/CategoryTabs';
import ItemTile from '../../shared/components/ItemTile';
import EmptyState from '../../shared/components/EmptyState';
import { SkeletonGrid } from '../../shared/components/Skeleton';
import { useMenuItems } from './useMenuItems';
import { useCartStore } from '../../shared/stores/cartStore';
import { useSettingsStore } from '../../shared/stores/settingsStore';
import type { MenuItem } from '@voxire/voxie/pos/menu_item/v1/types';
import { buildMenuItemCartProduct } from '../../shared/utils/cart/cartProduct';

const MenuItemsScreen: React.FC = () => {
  const { items, categories, loading, error, reload } = useMenuItems();
  const [search, setSearch] = useState('');
  const [category, setCategory] = useState('All');
  const addItem = useCartStore((s) => s.addItem);
  const businessProfile = useSettingsStore((s) => s.businessProfile);
  const catalogLabel = businessProfile === 'retail' ? 'Products' : 'Menu Items';
  const catalogSearchLabel = businessProfile === 'retail' ? 'products' : 'menu';

  const searchTerm = search.trim().toLowerCase();
  const filtered = useMemo(
    () =>
      items.filter((it) => {
        const matchesCat = category === 'All' || it.category === category;
        const matchesSearch = !searchTerm || (it.name ?? '').toLowerCase().includes(searchTerm);
        return matchesCat && matchesSearch;
      }),
    [items, category, searchTerm]
  );

  const quickAdd = useCallback(
    (item: MenuItem) => {
      try {
        addItem(buildMenuItemCartProduct(item));
      } catch (error) {
        toast.error(error instanceof Error ? error.message : `Failed to add ${catalogSearchLabel}`);
        return;
      }
      toast.success(`Added ${item.name}`);
    },
    [addItem, catalogSearchLabel]
  );

  return (
    <ScreenLayout
      title={catalogLabel}
      subtitle={`${items.length} items`}
      actions={
        <button
          onClick={reload}
          className="h-10 px-4 rounded-lg bg-white/10 text-pos-text text-sm font-semibold hover:bg-white/20"
        >
          Refresh
        </button>
      }
    >
      <div className="p-5 space-y-4">
        <SearchBar
          value={search}
          onChange={setSearch}
          placeholder={`Search ${catalogSearchLabel}...`}
        />
        <CategoryTabs categories={categories} active={category} onChange={setCategory} />
        {loading ? (
          <SkeletonGrid count={12} />
        ) : error ? (
          <EmptyState
            icon={AlertTriangle}
            title={`Failed to load ${catalogSearchLabel}`}
            message={error}
          />
        ) : filtered.length === 0 ? (
          <EmptyState
            icon={Search}
            title="No items match"
            message="Try a different search or category."
          />
        ) : (
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-3">
            {filtered.map((it) => (
              <ItemTile key={it.id} item={it} onClick={quickAdd} />
            ))}
          </div>
        )}
      </div>
    </ScreenLayout>
  );
};

export default MenuItemsScreen;
