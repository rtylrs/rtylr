import { useEffect, useCallback } from 'react';
import { unwrap } from '@voxire/api';
import { db } from '../db';
import { posService } from '../services/posService';

/**
 * Syncs products, modifiers, combos, and customers from the API
 * into the local Dexie (IndexedDB) cache for offline-first access.
 *
 * Call this once after the POS employee logs in.
 */
export function useSyncCache() {
  const syncAll = useCallback(async () => {
    try {
      // Sync products
      const productsData = await unwrap(posService.getMenuItems());
      const items = Array.isArray(productsData) ? productsData : (productsData as any)?.items ?? [];
      if (items.length > 0) {
        await db.products.clear();
        await db.products.bulkPut(
          items.map((p: any) => ({
            id: p.id,
            name: p.name,
            sku: p.sku ?? '',
            barcode: p.barcode ?? undefined,
            categoryId: p.categoryId ?? p.category_id ?? 0,
            categoryName: p.categoryName ?? p.category ?? '',
            price: p.price ?? p.salePrice ?? 0,
            costPrice: p.costPrice ?? p.cost_price ?? 0,
            taxClass: p.taxClass ?? 'standard',
            imageUrl: p.image ?? p.imageUrl ?? undefined,
            available: p.available !== false && p.isActive !== false,
            syncedAt: Date.now(),
          }))
        );
      }

      // Sync modifiers
      const modsData = await unwrap(posService.getModifiers());
      const mods = Array.isArray(modsData) ? modsData : (modsData as any)?.modifiers ?? [];
      if (mods.length > 0) {
        await db.modifiers.clear();
        await db.modifiers.bulkPut(
          mods.map((m: any) => ({
            id: m.id,
            name: m.name,
            groupName: m.groupName ?? m.group_name ?? '',
            priceDelta: m.priceDelta ?? m.price_delta ?? 0,
            isRequired: m.isRequired ?? m.is_required ?? false,
            syncedAt: Date.now(),
          }))
        );
      }

      // Sync combos
      const combosData = await unwrap(posService.getCombos());
      const combos = Array.isArray(combosData) ? combosData : (combosData as any)?.combos ?? [];
      if (combos.length > 0) {
        await db.combos.clear();
        await db.combos.bulkPut(
          combos.map((c: any) => ({
            id: c.id,
            name: c.name,
            description: c.description ?? '',
            price: c.price ?? 0,
            items: c.items ?? [],
            isActive: c.isActive !== false,
            syncedAt: Date.now(),
          }))
        );
      }

      console.log('[POS] Cache synced: products, modifiers, combos');
    } catch (err) {
      console.warn('[POS] Cache sync failed (offline mode):', err);
    }
  }, []);

  // Auto-sync on mount
  useEffect(() => {
    syncAll();
    // Re-sync every 5 minutes while online
    const interval = setInterval(() => {
      if (navigator.onLine) syncAll();
    }, 5 * 60 * 1000);
    return () => clearInterval(interval);
  }, [syncAll]);

  return { syncNow: syncAll };
}
