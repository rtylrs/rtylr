import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export interface ErpSettings {
  currency: string;
  taxRate: number;
  lowStockThreshold: number;
  defaultWarehouseId: number;
  theme: 'dark' | 'light' | 'system';
  notifyOnLowStock: boolean;
  notifyOnExpiry: boolean;
  setCurrency: (currency: string) => void;
  setTaxRate: (rate: number) => void;
  setLowStockThreshold: (n: number) => void;
  setDefaultWarehouseId: (id: number) => void;
  setTheme: (theme: 'dark' | 'light' | 'system') => void;
  setNotifyOnLowStock: (v: boolean) => void;
  setNotifyOnExpiry: (v: boolean) => void;
  setBusinessSettings: (settings: ErpBusinessSettings) => void;
}

export type ErpBusinessSettings = Omit<
  ErpSettings,
  | 'theme'
  | 'setCurrency'
  | 'setTaxRate'
  | 'setLowStockThreshold'
  | 'setDefaultWarehouseId'
  | 'setTheme'
  | 'setNotifyOnLowStock'
  | 'setNotifyOnExpiry'
  | 'setBusinessSettings'
>;

const DEFAULT_BUSINESS_SETTINGS: ErpBusinessSettings = {
  currency: 'USD',
  taxRate: 8.25,
  lowStockThreshold: 10,
  defaultWarehouseId: 1,
  notifyOnLowStock: true,
  notifyOnExpiry: true,
};

export const useSettings = create<ErpSettings>()(
  persist(
    (set) => ({
      ...DEFAULT_BUSINESS_SETTINGS,
      theme: 'dark',
      setCurrency: (currency) => set({ currency }),
      setTaxRate: (taxRate) => set({ taxRate }),
      setLowStockThreshold: (lowStockThreshold) => set({ lowStockThreshold }),
      setDefaultWarehouseId: (defaultWarehouseId) => set({ defaultWarehouseId }),
      setTheme: (theme) => set({ theme }),
      setNotifyOnLowStock: (notifyOnLowStock) => set({ notifyOnLowStock }),
      setNotifyOnExpiry: (notifyOnExpiry) => set({ notifyOnExpiry }),
      setBusinessSettings: (settings) => set(settings),
    }),
    {
      name: 'rtylr-erp-settings',
      partialize: (state) => ({ theme: state.theme }),
      merge: (persisted, current) => {
        const theme = (persisted as { theme?: ErpSettings['theme'] } | undefined)?.theme;
        return { ...current, ...(theme ? { theme } : {}) };
      },
    }
  )
);
