import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type ErpTheme = 'dark' | 'light' | 'system';
export const ERP_THEME_STORAGE_KEY = 'rtylr-erp-theme';

export interface ErpSettings {
  currency: string;
  taxRate: number;
  lowStockThreshold: number;
  defaultWarehouseId: number;
  theme: ErpTheme;
  notifyOnLowStock: boolean;
  notifyOnExpiry: boolean;
  setCurrency: (currency: string) => void;
  setTaxRate: (rate: number) => void;
  setLowStockThreshold: (n: number) => void;
  setDefaultWarehouseId: (id: number) => void;
  setTheme: (theme: ErpTheme) => void;
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

function isErpTheme(value: unknown): value is ErpTheme {
  return value === 'dark' || value === 'light' || value === 'system';
}

function initialTheme(): ErpTheme {
  if (typeof window === 'undefined') return 'dark';
  const stored = window.localStorage.getItem(ERP_THEME_STORAGE_KEY);
  return isErpTheme(stored) ? stored : 'dark';
}

export function applyErpTheme(theme: ErpTheme): void {
  if (typeof document === 'undefined') return;
  const resolved =
    theme === 'system'
      ? window.matchMedia?.('(prefers-color-scheme: dark)').matches
        ? 'dark'
        : 'light'
      : theme;
  document.documentElement.setAttribute('data-theme', resolved);
}

export const useSettings = create<ErpSettings>()(
  persist(
    (set) => ({
      ...DEFAULT_BUSINESS_SETTINGS,
      theme: initialTheme(),
      setCurrency: (currency) => set({ currency }),
      setTaxRate: (taxRate) => set({ taxRate }),
      setLowStockThreshold: (lowStockThreshold) => set({ lowStockThreshold }),
      setDefaultWarehouseId: (defaultWarehouseId) => set({ defaultWarehouseId }),
      setTheme: (theme) => {
        set({ theme });
        if (typeof window !== 'undefined') {
          window.localStorage.setItem(ERP_THEME_STORAGE_KEY, theme);
          applyErpTheme(theme);
        }
      },
      setNotifyOnLowStock: (notifyOnLowStock) => set({ notifyOnLowStock }),
      setNotifyOnExpiry: (notifyOnExpiry) => set({ notifyOnExpiry }),
      setBusinessSettings: (settings) => set(settings),
    }),
    {
      name: 'rtylr-erp-settings',
      partialize: (state) => ({ theme: state.theme }),
      merge: (persisted, current) => {
        const theme = (persisted as { theme?: ErpTheme } | undefined)?.theme;
        return { ...current, ...(isErpTheme(theme) ? { theme } : {}) };
      },
    }
  )
);
