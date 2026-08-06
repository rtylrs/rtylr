import { useEffect, useCallback } from 'react';
import { createOneKitClient } from '@voxire/api';
import { ComboServiceClient } from '@voxire/voxie/pos/combo/v1/client';
import { MenuItemServiceClient } from '@voxire/voxie/pos/menu_item/v1/client';
import { ModifierServiceClient } from '@voxire/voxie/pos/modifier/v1/client';
import { db } from '../services/db';
import {
  normalizeCachedCombos,
  normalizeCachedModifiers,
  normalizeCachedProducts,
  responseArray,
} from './syncCacheWorkflow';

const menuItemClient = createOneKitClient(MenuItemServiceClient, 'pos');
const modifierClient = createOneKitClient(ModifierServiceClient, 'pos');
const comboClient = createOneKitClient(ComboServiceClient, 'pos');

/**
 * Syncs products, modifiers, combos, and customers from the API
 * into the local Dexie (IndexedDB) cache for offline-first access.
 *
 * Call this once after the POS employee logs in.
 */
export function useSyncCache() {
  const syncAll = useCallback(async () => {
    try {
      const syncedAt = Date.now();

      // Sync products
      const productsData = await menuItemClient.listMenuItems({});
      const items = normalizeCachedProducts(
        responseArray(productsData, ['items', 'products', 'data']),
        syncedAt,
      );
      if (items.length > 0) {
        await db.products.clear();
        await db.products.bulkPut(items);
      }

      // Sync modifiers
      const modsData = await modifierClient.listModifiers({});
      const mods = normalizeCachedModifiers(
        responseArray(modsData, ['modifiers', 'groups', 'data']),
        syncedAt,
      );
      if (mods.length > 0) {
        await db.modifiers.clear();
        await db.modifiers.bulkPut(mods);
      }

      // Sync combos
      const combosData = await comboClient.listCombos({});
      const combos = normalizeCachedCombos(responseArray(combosData, ['combos', 'data']), syncedAt);
      if (combos.length > 0) {
        await db.combos.clear();
        await db.combos.bulkPut(combos);
      }
    } catch (err) {
      console.error('Failed to sync offline POS cache', err);
    }
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
