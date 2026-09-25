import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type PrinterTransportType = 'windows' | 'serial' | 'network';

export interface PrinterSettings {
  /** Display name (used as Windows spooler printer name when transport='windows') */
  name: string;
  copies: number;
  cashDrawerKick: boolean;
  drawerPin: 2 | 5;
  /** How to reach the printer */
  transport: PrinterTransportType;
  /** Serial: e.g. /dev/tty.usbserial-1420 or COM3 */
  serialPort: string;
  /** Serial baud rate (default 9600) */
  serialBaudRate: number;
  /** Network: hostname or IP address */
  networkHost: string;
  /** Network: TCP port (default 9100 for most thermal printers) */
  networkPort: number;
}

export interface BusinessInfo {
  name: string;
  address: string;
  phone: string;
}

export type POSBusinessProfile = 'retail' | 'restaurant' | 'quick_service';

export interface POSFeatureSettings {
  tables: boolean;
  kitchenDisplay: boolean;
  modifiers: boolean;
  combos: boolean;
  tips: boolean;
  giftCards: boolean;
  customers: boolean;
  inventory: boolean;
  cashDrawer: boolean;
}

export type POSFeatureKey = keyof POSFeatureSettings;

export type KdsHardwareTransport = 'none' | 'serial' | 'network';

export interface KdsHardwareSettings {
  /** 'none' = software-only KDS (browser), 'serial'/'network' = hardware KDS display */
  transport: KdsHardwareTransport;
  serialPort: string;
  serialBaudRate: number;
  networkHost: string;
  networkPort: number;
}

interface SettingsState {
  taxRate: number; // 0.08 = 8%
  tipPresets: number[]; // [0.1, 0.15, 0.18, 0.2]
  currency: string;
  currencySymbol: string;
  receiptFooter: string;
  businessProfile: POSBusinessProfile;
  features: POSFeatureSettings;
  printer: PrinterSettings;
  business: BusinessInfo;
  kdsHardware: KdsHardwareSettings;
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
  setBusinessProfile: (profile: POSBusinessProfile) => void;
  setFeature: (feature: POSFeatureKey, enabled: boolean) => void;
  setFeatures: (partial: Partial<POSFeatureSettings>) => void;
  setPrinter: (partial: Partial<PrinterSettings>) => void;
  setBusiness: (partial: Partial<BusinessInfo>) => void;
  setKdsHardware: (partial: Partial<KdsHardwareSettings>) => void;
  setAutoLogoutMinutes: (min: number) => void;
  setSpeedMode: (on: boolean) => void;
  setBarMode: (on: boolean) => void;
  setQuickAccessSidebar: (on: boolean) => void;
  setKdsStation: (s: string) => void;
  reset: () => void;
}

const profileFeatureDefaults: Record<POSBusinessProfile, POSFeatureSettings> = {
  retail: {
    tables: false,
    kitchenDisplay: false,
    modifiers: false,
    combos: false,
    tips: false,
    giftCards: true,
    customers: true,
    inventory: true,
    cashDrawer: true,
  },
  restaurant: {
    tables: true,
    kitchenDisplay: true,
    modifiers: true,
    combos: true,
    tips: true,
    giftCards: true,
    customers: true,
    inventory: true,
    cashDrawer: true,
  },
  quick_service: {
    tables: false,
    kitchenDisplay: true,
    modifiers: true,
    combos: true,
    tips: true,
    giftCards: true,
    customers: true,
    inventory: true,
    cashDrawer: true,
  },
};

const defaults: SettingsState = {
  taxRate: 0.08,
  tipPresets: [0.1, 0.15, 0.18, 0.2],
  currency: 'USD',
  currencySymbol: '$',
  receiptFooter: 'Thank you for your purchase!',
  businessProfile: 'retail',
  features: { ...profileFeatureDefaults.retail },
  printer: {
    name: 'Default',
    copies: 1,
    cashDrawerKick: true,
    drawerPin: 2,
    transport: 'windows',
    serialPort: '',
    serialBaudRate: 9600,
    networkHost: '',
    networkPort: 9100,
  },
  business: {
    name: 'My Business',
    address: '',
    phone: '',
  },
  kdsHardware: {
    transport: 'none',
    serialPort: '',
    serialBaudRate: 9600,
    networkHost: '',
    networkPort: 9000,
  },
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

function isPOSBusinessProfile(value: unknown): value is POSBusinessProfile {
  return value === 'retail' || value === 'restaurant' || value === 'quick_service';
}

function mergeSettings(
  persistedState: unknown,
  currentState: SettingsState & SettingsActions
): SettingsState & SettingsActions {
  const persisted = (persistedState ?? {}) as Partial<SettingsState>;
  const businessProfile = isPOSBusinessProfile(persisted.businessProfile)
    ? persisted.businessProfile
    : currentState.businessProfile;
  return {
    ...currentState,
    ...persisted,
    businessProfile,
    printer: { ...currentState.printer, ...(persisted.printer ?? {}) },
    business: { ...currentState.business, ...(persisted.business ?? {}) },
    kdsHardware: { ...currentState.kdsHardware, ...(persisted.kdsHardware ?? {}) },
    features: { ...profileFeatureDefaults[businessProfile], ...(persisted.features ?? {}) },
  };
}

function freshDefaults(): SettingsState {
  return {
    ...defaults,
    printer: { ...defaults.printer },
    business: { ...defaults.business },
    kdsHardware: { ...defaults.kdsHardware },
    features: { ...defaults.features },
  };
}

export const useSettingsStore = create<SettingsState & SettingsActions>()(
  persist(
    (set) => ({
      ...defaults,
      setTaxRate: (taxRate) => set({ taxRate: clampNumber(taxRate, 0, 1) }),
      setTipPresets: (tipPresets) => set({ tipPresets: normalizeTipPresets(tipPresets) }),
      setCurrency: (currency, currencySymbol) => set(normalizeCurrency(currency, currencySymbol)),
      setReceiptFooter: (receiptFooter) => set({ receiptFooter }),
      setBusinessProfile: (businessProfile) =>
        set({ businessProfile, features: { ...profileFeatureDefaults[businessProfile] } }),
      setFeature: (feature, enabled) =>
        set((s) => ({ features: { ...s.features, [feature]: enabled } })),
      setFeatures: (partial) =>
        set((s) => ({ features: { ...s.features, ...partial } })),
      setPrinter: (partial) =>
        set((s) => ({ printer: { ...s.printer, ...normalizePrinter(partial) } })),
      setBusiness: (partial) =>
        set((s) => ({ business: { ...s.business, ...partial } })),
      setKdsHardware: (partial) =>
        set((s) => ({ kdsHardware: { ...s.kdsHardware, ...partial } })),
      setAutoLogoutMinutes: (autoLogoutMinutes) =>
        set({ autoLogoutMinutes: Math.round(clampNumber(autoLogoutMinutes, 0, 24 * 60)) }),
      setSpeedMode: (speedMode) => set({ speedMode }),
      setBarMode: (barMode) => set({ barMode }),
      setQuickAccessSidebar: (quickAccessSidebar) => set({ quickAccessSidebar }),
      setKdsStation: (kdsStation) => set({ kdsStation }),
      reset: () => set(freshDefaults()),
    }),
    { name: 'pos-settings', merge: mergeSettings }
  )
);
