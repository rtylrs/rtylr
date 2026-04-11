import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export interface PrinterSettings {
  name: string;
  copies: number;
  cashDrawerKick: boolean;
}

interface SettingsState {
  taxRate: number; // 0.08 = 8%
  tipPresets: number[]; // [0.1, 0.15, 0.18, 0.2]
  currency: string;
  currencySymbol: string;
  receiptFooter: string;
  printer: PrinterSettings;
  autoLogoutMinutes: number;
  speedMode: boolean;
  barMode: boolean;
  quickAccessSidebar: boolean;
  kdsStation: string;
  language: string;
}

interface SettingsActions {
  setTaxRate: (rate: number) => void;
  setTipPresets: (presets: number[]) => void;
  setCurrency: (code: string, symbol: string) => void;
  setReceiptFooter: (footer: string) => void;
  setPrinter: (partial: Partial<PrinterSettings>) => void;
  setAutoLogoutMinutes: (min: number) => void;
  setSpeedMode: (on: boolean) => void;
  setBarMode: (on: boolean) => void;
  setQuickAccessSidebar: (on: boolean) => void;
  setKdsStation: (s: string) => void;
  reset: () => void;
}

const defaults: SettingsState = {
  taxRate: 0.08,
  tipPresets: [0.1, 0.15, 0.18, 0.2],
  currency: 'USD',
  currencySymbol: '$',
  receiptFooter: 'Thank you for your visit!',
  printer: { name: 'Default', copies: 1, cashDrawerKick: true },
  autoLogoutMinutes: 10,
  speedMode: false,
  barMode: false,
  quickAccessSidebar: true,
  kdsStation: 'all',
  language: 'en',
};

export const useSettingsStore = create<SettingsState & SettingsActions>()(
  persist(
    (set) => ({
      ...defaults,
      setTaxRate: (taxRate) => set({ taxRate }),
      setTipPresets: (tipPresets) => set({ tipPresets }),
      setCurrency: (currency, currencySymbol) => set({ currency, currencySymbol }),
      setReceiptFooter: (receiptFooter) => set({ receiptFooter }),
      setPrinter: (partial) => set((s) => ({ printer: { ...s.printer, ...partial } })),
      setAutoLogoutMinutes: (autoLogoutMinutes) => set({ autoLogoutMinutes }),
      setSpeedMode: (speedMode) => set({ speedMode }),
      setBarMode: (barMode) => set({ barMode }),
      setQuickAccessSidebar: (quickAccessSidebar) => set({ quickAccessSidebar }),
      setKdsStation: (kdsStation) => set({ kdsStation }),
      reset: () => set(defaults),
    }),
    { name: 'pos-settings' }
  )
);
