import { create } from 'zustand';

export interface ErpSettings {
  currency: string;
  taxRate: number;
  lowStockThreshold: number;
  defaultWarehouseId: number;
  notifyOnLowStock: boolean;
  notifyOnExpiry: boolean;
  setCurrency: (currency: string) => void;
  setTaxRate: (rate: number) => void;
  setLowStockThreshold: (n: number) => void;
  setDefaultWarehouseId: (id: number) => void;
  setNotifyOnLowStock: (v: boolean) => void;
  setNotifyOnExpiry: (v: boolean) => void;
  setBusinessSettings: (settings: ErpBusinessSettings) => void;
}

export type ErpBusinessSettings = Omit<
  ErpSettings,
  | 'setCurrency'
  | 'setTaxRate'
  | 'setLowStockThreshold'
  | 'setDefaultWarehouseId'
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

export const useSettings = create<ErpSettings>((set) => ({
  ...DEFAULT_BUSINESS_SETTINGS,
  setCurrency: (currency) => set({ currency }),
  setTaxRate: (taxRate) => set({ taxRate }),
  setLowStockThreshold: (lowStockThreshold) => set({ lowStockThreshold }),
  setDefaultWarehouseId: (defaultWarehouseId) => set({ defaultWarehouseId }),
  setNotifyOnLowStock: (notifyOnLowStock) => set({ notifyOnLowStock }),
  setNotifyOnExpiry: (notifyOnExpiry) => set({ notifyOnExpiry }),
  setBusinessSettings: (settings) => set(settings),
}));
