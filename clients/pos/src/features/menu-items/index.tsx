import { createOneKitClient } from '@voxire/api';
import { MenuServiceClient } from '@voxire/voxie/hub/menu/v1/client';
import type { BranchMenuItemConfig } from '@voxire/voxie/hub/menu/v1/types';
import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { AlertTriangle, Eye, EyeOff, Search, SlidersHorizontal } from 'lucide-react';
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
import { useTerminalContextStore } from '../../shared/stores/terminalContextStore';
import type { MenuItem } from '@voxire/voxie/pos/menu_item/v1/types';
import { buildMenuItemCartProduct } from '../../shared/utils/cart/cartProduct';

const menuClient = createOneKitClient(MenuServiceClient, 'pos');

const MenuItemsScreen: React.FC = () => {
  const { items, categories, loading, error, reload } = useMenuItems();
  const [search, setSearch] = useState('');
  const [category, setCategory] = useState('All');
  const [manageMode, setManageMode] = useState(false);
  const [branchOverrides, setBranchOverrides] = useState<Map<string, BranchMenuItemConfig>>(
    () => new Map()
  );
  const [priceDrafts, setPriceDrafts] = useState<Record<string, string>>({});
  const [savingItemId, setSavingItemId] = useState<string | null>(null);

  const addItem = useCartStore((s) => s.addItem);
  const businessProfile = useSettingsStore((s) => s.businessProfile);
  const locationId = useTerminalContextStore((s) => s.locationId) ?? 0;
  const catalogLabel = businessProfile === 'retail' ? 'Products' : 'Menu Items';
  const catalogSearchLabel = businessProfile === 'retail' ? 'products' : 'menu';

  const loadOverrides = useCallback(async () => {
    try {
      const res = await menuClient.listBranchMenuItems({
        ...(locationId > 0 ? { locationId: String(locationId) } : {}),
      });
      const next = new Map<string, BranchMenuItemConfig>();
      for (const bi of res.items ?? []) {
        if (bi.menuItemId) next.set(String(bi.menuItemId), bi);
      }
      setBranchOverrides(next);
    } catch {
      // Non-fatal: keep base catalog usable if branch overrides fail to load.
    }
  }, [locationId]);

  useEffect(() => {
    void loadOverrides();
  }, [loadOverrides]);

  const saveBranchOverride = useCallback(
    async (
      itemId: string,
      nextVisible: boolean,
      hasPriceOverride: boolean,
      priceOverride: number
    ) => {
      setSavingItemId(itemId);
      try {
        const res = await menuClient.upsertBranchMenuItem({
          ...(locationId > 0 ? { locationId: String(locationId) } : {}),
          menuItemId: itemId,
          isVisible: nextVisible,
          hasPriceOverride,
          priceOverride: hasPriceOverride ? priceOverride : 0,
          upsellItemIds: [],
        });
        if (res.item) {
          setBranchOverrides((prev) => {
            const copy = new Map(prev);
            copy.set(itemId, res.item!);
            return copy;
          });
        }
        toast.success(nextVisible ? 'Branch menu updated' : 'Item marked 86’d (hidden on menu)');
      } catch {
        toast.error('Failed to update branch menu override');
      } finally {
        setSavingItemId(null);
      }
    },
    [locationId]
  );

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
      if (
        item.trackInventory === true &&
        item.allowBackorder !== true &&
        typeof item.stock === 'number' &&
        Number.isFinite(item.stock) &&
        item.stock <= 0
      ) {
        toast.error(`${item.name ?? 'Item'} is out of stock`);
        return;
      }
      const ov = branchOverrides.get(String(item.id ?? ''));
      const effectiveItem =
        ov?.hasPriceOverride && ov.priceOverride != null
          ? { ...item, salePrice: ov.priceOverride }
          : item;
      try {
        addItem(buildMenuItemCartProduct(effectiveItem));
      } catch (err) {
        toast.error(err instanceof Error ? err.message : `Failed to add ${catalogSearchLabel}`);
        return;
      }
      toast.success(`Added ${item.name}`);
    },
    [addItem, branchOverrides, catalogSearchLabel]
  );

  return (
    <ScreenLayout
      title={catalogLabel}
      subtitle={`${items.length} items`}
      actions={
        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={() => setManageMode((m) => !m)}
            className={`h-10 px-3.5 rounded-lg text-sm font-semibold inline-flex items-center gap-1.5 transition-colors ${
              manageMode
                ? 'bg-amber-500 text-black hover:bg-amber-400'
                : 'bg-white/10 text-pos-text hover:bg-white/20'
            }`}
          >
            <SlidersHorizontal size={14} />
            {manageMode ? 'Done (86 & Prices)' : 'Branch 86 & Prices'}
          </button>
          <button
            type="button"
            onClick={() => {
              reload();
              void loadOverrides();
            }}
            className="h-10 px-4 rounded-lg bg-white/10 text-pos-text text-sm font-semibold hover:bg-white/20"
          >
            Refresh
          </button>
        </div>
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
        ) : manageMode ? (
          <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-3">
            {filtered.map((it) => {
              const itemId = String(it.id ?? '');
              const ov = branchOverrides.get(itemId);
              const visible = ov ? ov.isVisible !== false : true;
              const draft =
                priceDrafts[itemId] ?? (ov?.hasPriceOverride ? String(ov.priceOverride ?? '') : '');
              const busy = savingItemId === itemId;
              return (
                <div
                  key={itemId}
                  className={`rounded-xl border p-3.5 space-y-3 ${
                    visible
                      ? 'border-white/10 bg-white/5'
                      : 'border-red-500/30 bg-red-500/10 opacity-85'
                  }`}
                >
                  <div className="flex items-start justify-between gap-2">
                    <div className="min-w-0">
                      <p className="text-sm font-semibold text-pos-text truncate">{it.name}</p>
                      <p className="text-xs text-white/60">
                        Base price: ${(it.salePrice ?? 0).toFixed(2)}
                        {ov?.hasPriceOverride ? ` · Branch: $${(ov.priceOverride ?? 0).toFixed(2)}` : ''}
                      </p>
                    </div>
                    <button
                      type="button"
                      disabled={busy}
                      onClick={() =>
                        void saveBranchOverride(
                          itemId,
                          !visible,
                          Boolean(ov?.hasPriceOverride),
                          ov?.priceOverride ?? 0
                        )
                      }
                      className={`h-8 px-2.5 rounded-lg text-xs font-semibold inline-flex items-center gap-1 shrink-0 disabled:opacity-50 ${
                        visible
                          ? 'bg-emerald-600/80 hover:bg-emerald-600 text-white'
                          : 'bg-red-600 hover:bg-red-500 text-white'
                      }`}
                    >
                      {visible ? <Eye size={13} /> : <EyeOff size={13} />}
                      {visible ? 'Available' : '86’d'}
                    </button>
                  </div>
                  <div className="flex items-center gap-2">
                    <input
                      type="number"
                      step="0.01"
                      min="0"
                      value={draft}
                      onChange={(e) =>
                        setPriceDrafts((prev) => ({ ...prev, [itemId]: e.target.value }))
                      }
                      placeholder={`Branch price (default ${(it.salePrice ?? 0).toFixed(2)})`}
                      className="flex-1 h-8 rounded-lg border border-white/15 bg-black/30 px-2.5 text-xs text-pos-text focus:outline-none focus:border-white/40"
                    />
                    <button
                      type="button"
                      disabled={busy}
                      onClick={() => {
                        const raw = (priceDrafts[itemId] ?? '').trim();
                        const parsed = Number(raw);
                        const hasOverride = raw !== '' && Number.isFinite(parsed) && parsed >= 0;
                        void saveBranchOverride(itemId, visible, hasOverride, hasOverride ? parsed : 0);
                      }}
                      className="h-8 px-3 rounded-lg bg-white/15 hover:bg-white/25 text-xs font-semibold text-pos-text disabled:opacity-50"
                    >
                      Save
                    </button>
                  </div>
                </div>
              );
            })}
          </div>
        ) : (
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-3">
            {filtered.map((it) => {
              const ov = branchOverrides.get(String(it.id ?? ''));
              const isHidden = ov?.isVisible === false;
              const effectiveItem =
                ov?.hasPriceOverride && ov.priceOverride != null
                  ? { ...it, salePrice: ov.priceOverride }
                  : it;
              return (
                <div key={it.id} className="relative">
                  <ItemTile item={effectiveItem} onClick={quickAdd} />
                  {isHidden && (
                    <span className="pointer-events-none absolute top-2 right-2 rounded-full bg-red-600 px-2 py-0.5 text-[10px] font-bold text-white">
                      86’d
                    </span>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </div>
    </ScreenLayout>
  );
};

export default MenuItemsScreen;
