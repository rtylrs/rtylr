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
}

export const useSettings = create<ErpSettings>()(
  persist(
    (set) => ({
      currency: 'USD',
      taxRate: 8.25,
      lowStockThreshold: 10,
      defaultWarehouseId: 1,
      theme: 'dark',
      notifyOnLowStock: true,
      notifyOnExpiry: true,
      setCurrency: (currency) => set({ currency }),
      setTaxRate: (taxRate) => set({ taxRate }),
      setLowStockThreshold: (lowStockThreshold) => set({ lowStockThreshold }),
      setDefaultWarehouseId: (defaultWarehouseId) => set({ defaultWarehouseId }),
      setTheme: (theme) => set({ theme }),
      setNotifyOnLowStock: (notifyOnLowStock) => set({ notifyOnLowStock }),
      setNotifyOnExpiry: (notifyOnExpiry) => set({ notifyOnExpiry }),
    }),
    { name: 'rtylr-erp-settings' }
  )
);
