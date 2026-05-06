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

function clampNumber(value: number, min: number, max: number): number {
  if (!Number.isFinite(value)) return min;
  return Math.min(max, Math.max(min, value));
}

function normalizePrinter(partial: Partial<PrinterSettings>): Partial<PrinterSettings> {
  if (partial.copies === undefined) return partial;
  return { ...partial, copies: Math.round(clampNumber(partial.copies, 1, 5)) };
}

function normalizeTipPresets(presets: number[]): number[] {
  return presets.map((preset) => clampNumber(preset, 0, 1));
}

function normalizeCurrency(currency: string, currencySymbol: string) {
  const normalizedCurrency = currency.trim().toUpperCase();
  const normalizedSymbol = currencySymbol.trim();
  if (!/^[A-Z]{3}$/.test(normalizedCurrency) || !normalizedSymbol) {
    return { currency: defaults.currency, currencySymbol: defaults.currencySymbol };
  }
  return { currency: normalizedCurrency, currencySymbol: normalizedSymbol };
}

export const useSettingsStore = create<SettingsState & SettingsActions>()(
  persist(
    (set) => ({
      ...defaults,
      setTaxRate: (taxRate) => set({ taxRate: clampNumber(taxRate, 0, 1) }),
      setTipPresets: (tipPresets) => set({ tipPresets: normalizeTipPresets(tipPresets) }),
      setCurrency: (currency, currencySymbol) => set(normalizeCurrency(currency, currencySymbol)),
      setReceiptFooter: (receiptFooter) => set({ receiptFooter }),
      setPrinter: (partial) =>
        set((s) => ({ printer: { ...s.printer, ...normalizePrinter(partial) } })),
      setAutoLogoutMinutes: (autoLogoutMinutes) =>
        set({ autoLogoutMinutes: Math.round(clampNumber(autoLogoutMinutes, 0, 24 * 60)) }),
      setSpeedMode: (speedMode) => set({ speedMode }),
      setBarMode: (barMode) => set({ barMode }),
      setQuickAccessSidebar: (quickAccessSidebar) => set({ quickAccessSidebar }),
      setKdsStation: (kdsStation) => set({ kdsStation }),
      reset: () => set(defaults),
    }),
    { name: 'pos-settings' }
  )
);
