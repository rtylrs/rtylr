import React, { useState } from 'react';
import { toast } from 'sonner';
import ScreenLayout from '../../layouts/ScreenLayout';
import SearchBar from '../../shared/components/SearchBar';
import CategoryTabs from '../../shared/components/CategoryTabs';
import ItemTile from '../../shared/components/ItemTile';
import EmptyState from '../../shared/components/EmptyState';
import { SkeletonGrid } from '../../shared/components/Skeleton';
import { useMenuItems } from './useMenuItems';
import { useCartStore } from '../../shared/stores/cartStore';
import type { MenuItem } from '../../shared/types/pos';
import { buildMenuItemCartProduct } from '../../shared/utils/cart/cartProduct';

const MenuItemsScreen: React.FC = () => {
  const { items, categories, loading, error, reload } = useMenuItems();
  const [search, setSearch] = useState('');
  const [category, setCategory] = useState('All');
  const addItem = useCartStore((s) => s.addItem);

  const filtered = items.filter((it) => {
    const matchesCat = category === 'All' || it.category === category;
    const matchesSearch = !search || it.name.toLowerCase().includes(search.toLowerCase());
    return matchesCat && matchesSearch;
  });

  const quickAdd = (item: MenuItem) => {
    try {
      addItem(buildMenuItemCartProduct(item));
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to add menu item');
      return;
    }
    toast.success(`Added ${item.name}`);
  };

  return (
    <ScreenLayout
      title="Menu Items"
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
        <SearchBar value={search} onChange={setSearch} placeholder="Search menu..." />
        <CategoryTabs categories={categories} active={category} onChange={setCategory} />
        {loading ? (
          <SkeletonGrid count={12} />
        ) : error ? (
          <EmptyState icon="⚠" title="Failed to load menu" message={error} />
        ) : filtered.length === 0 ? (
          <EmptyState
            icon="🔍"
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
