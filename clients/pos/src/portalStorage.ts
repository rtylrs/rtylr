import type { StoreApi } from 'zustand';
import type { PersistOptions } from 'zustand/middleware';
import { useReceiptsStore } from './shared/stores/receiptsStore';
import { useCartStore } from './shared/stores/cartStore';
import { useSettingsStore } from './shared/stores/settingsStore';
import { useSessionCloseQueueStore } from './shared/stores/sessionCloseQueueStore';
import { useHeldOrdersStore } from './shared/stores/heldOrdersStore';
import { useCashMovementQueueStore } from './shared/stores/cashMovementQueueStore';
import { useCashDrawerStore } from './shared/stores/cashDrawerStore';
import { useUserStore } from './shared/stores/userStore';
import { useGiftCardStore } from './shared/stores/giftCardStore';
import { useThemeStore } from './shared/stores/themeStore';
import { useOfflineQueueStore } from './shared/stores/offlineQueueStore';
import { useTerminalContextStore } from './shared/stores/terminalContextStore';
import { useEmployeeSessionStore } from './shared/stores/employeeSessionStore';

interface PersistedStore<T, P> extends StoreApi<T> {
 persist: {
 getOptions: () => Partial<PersistOptions<T, P>>;
 setOptions: (options: Partial<PersistOptions<T, P>>) => void;
 rehydrate: () => Promise<void> | void;
 };
}
async function partition<T, P>(store: PersistedStore<T, P>, businessId: number) {
 const options = store.persist.getOptions();
 if (!options.name) throw new Error('Terminal storage has no name');
 const name = `portal-pos:${businessId}:${options.name.replace(/^portal-pos:\d+:/, '')}`;
 const saved = await options.storage?.getItem(name);
 store.persist.setOptions({ name });
 store.setState(store.getInitialState(), true);
 if (saved) await options.storage?.setItem(name, saved);
 await store.persist.rehydrate();
}
async function restoreTerminal(businessId: number): Promise<void> {
 await partition(useReceiptsStore, businessId);
 await partition(useCartStore, businessId);
 await partition(useSettingsStore, businessId);
 await partition(useSessionCloseQueueStore, businessId);
 await partition(useHeldOrdersStore, businessId);
 await partition(useCashMovementQueueStore, businessId);
 await partition(useCashDrawerStore, businessId);
 await partition(useUserStore, businessId);
 await partition(useGiftCardStore, businessId);
 await partition(useThemeStore, businessId);
 await partition(useOfflineQueueStore, businessId);
 await partition(useTerminalContextStore, businessId);
 await partition(useEmployeeSessionStore, businessId);
}

let activeBusinessId: number | undefined;
let preparation: Promise<void> = Promise.resolve();

export function preparePortalTerminal(businessId: number): Promise<void> {
  const next = preparation.catch(() => undefined).then(async () => {
    if (activeBusinessId === businessId) return;
    await restoreTerminal(businessId);
    activeBusinessId = businessId;
  });
  preparation = next;
  return next;
}
