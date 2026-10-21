import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { useUserStore } from '../../../../shared/stores/userStore';
import type { Currency } from './money';

export const DEFAULT_CATEGORY_COLOR = '#14b8a6';

export interface FinanceSettings {
  currency: Currency;
  defaultTaxRate: number;
  invoicePrefix: string;
  fiscalYearStartMonth: number; // 1 - 12
  lateFeePercent: number;
  lateFeeGraceDays: number;
  companyName: string;
  companyEmail: string;
  brandColor: string;
  logoUrl: string;
  invoiceNotes: string;
  setSetting: <K extends keyof FinanceSettings>(key: K, value: FinanceSettings[K]) => void;
  reset: () => void;
}

const DEFAULTS = {
  currency: 'USD' as Currency,
  defaultTaxRate: 8.5,
  invoicePrefix: 'INV',
  fiscalYearStartMonth: 1,
  lateFeePercent: 1.5,
  lateFeeGraceDays: 7,
  companyName: '',
  companyEmail: '',
  brandColor: '#10b981',
  logoUrl: '',
  invoiceNotes: 'Thank you for your business!',
};

export const useSettings = create<FinanceSettings>()(
  persist(
    (set) => ({
      ...DEFAULTS,
      setSetting: (key, value) => set({ [key]: value } as Partial<FinanceSettings>),
      reset: () => set({ ...DEFAULTS }),
    }),
    { name: `rtylr-finance-settings:${useUserStore.getState().user?.businessId ?? 'none'}` }
  )
);

useUserStore.subscribe((state, previous) => {
  const businessId = state.user?.businessId;
  if (businessId === previous.user?.businessId) return;
  const name = `rtylr-finance-settings:${businessId ?? 'none'}`;
  const saved = localStorage.getItem(name);
  useSettings.persist.setOptions({ name });
  useSettings.setState(useSettings.getInitialState(), true);
  if (saved) localStorage.setItem(name, saved);
  void useSettings.persist.rehydrate();
});
