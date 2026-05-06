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
      const items = (
        Array.isArray(productsData)
          ? productsData
          : ((productsData as Record<string, unknown>)?.items ?? [])
      ) as Record<string, unknown>[];
      if (items.length > 0) {
        await db.products.clear();
        await db.products.bulkPut(
          items.map((p) => ({
            id: p.id as number,
            name: (p.name as string) ?? '',
            sku: (p.sku as string) ?? '',
            barcode: (p.barcode as string) ?? undefined,
            categoryId: (p.categoryId ?? p.category_id ?? 0) as number,
            categoryName: (p.categoryName ?? p.category ?? '') as string,
            price: (p.price ?? p.salePrice ?? 0) as number,
            costPrice: (p.costPrice ?? p.cost_price ?? 0) as number,
            taxClass: (p.taxClass as string) ?? 'standard',
            imageUrl: (p.image ?? p.imageUrl ?? undefined) as string | undefined,
            available: p.available !== false && p.isActive !== false,
            syncedAt: Date.now(),
          }))
        );
      }

      // Sync modifiers
      const modsData = await unwrap(posService.getModifiers());
      const mods = (
        Array.isArray(modsData)
          ? modsData
          : ((modsData as Record<string, unknown>)?.modifiers ?? [])
      ) as Record<string, unknown>[];
      if (mods.length > 0) {
        await db.modifiers.clear();
        await db.modifiers.bulkPut(
          mods.map((m) => ({
            id: m.id as number,
            name: (m.name as string) ?? '',
            groupName: (m.groupName ?? m.group_name ?? '') as string,
            priceDelta: (m.priceDelta ?? m.price_delta ?? 0) as number,
            isRequired: (m.isRequired ?? m.is_required ?? false) as boolean,
            syncedAt: Date.now(),
          }))
        );
      }

      // Sync combos
      const combosData = await unwrap(posService.getCombos());
      const combos = (
        Array.isArray(combosData)
          ? combosData
          : ((combosData as Record<string, unknown>)?.combos ?? [])
      ) as Record<string, unknown>[];
      if (combos.length > 0) {
        await db.combos.clear();
        await db.combos.bulkPut(
          combos.map((c) => ({
            id: c.id as number,
            name: (c.name as string) ?? '',
            description: (c.description as string) ?? '',
            price: (c.price as number) ?? 0,
            items: (c.items ?? []) as { productId: number; name: string; quantity: number }[],
            isActive: c.isActive !== false,
            syncedAt: Date.now(),
          }))
        );
      }
    } catch (err) {}
  }, []);

  // Auto-sync on mount
  useEffect(() => {
    syncAll();
    // Re-sync every 5 minutes while online
    const interval = setInterval(
      () => {
        if (navigator.onLine) syncAll();
      },
      5 * 60 * 1000
    );
    return () => clearInterval(interval);
  }, [syncAll]);

  return { syncNow: syncAll };
}
