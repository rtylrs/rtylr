import { describe, it, expect, beforeEach } from 'vitest';
import { useSettingsStore } from './settingsStore';

const { getState, setState } = useSettingsStore;
const act = (fn: () => void) => {
  fn();
};

const defaults = {
  taxRate: 0.08,
  tipPresets: [0.1, 0.15, 0.18, 0.2],
  currency: 'USD',
  currencySymbol: '$',
  receiptFooter: 'Thank you for your purchase!',
  businessProfile: 'retail',
  features: {
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
  autoLogoutMinutes: 10,
  speedMode: false,
  barMode: false,
  quickAccessSidebar: true,
  kdsStation: 'all',
  language: 'en',
};

beforeEach(() => {
  act(() => getState().reset());
});

describe('settingsStore', () => {
  // ── Initial state ──────────────────────────────────────────────

  it('has correct default values', () => {
    expect(getState().taxRate).toBe(0.08);
    expect(getState().tipPresets).toEqual([0.1, 0.15, 0.18, 0.2]);
    expect(getState().currency).toBe('USD');
    expect(getState().currencySymbol).toBe('$');
    expect(getState().receiptFooter).toBe('Thank you for your purchase!');
    expect(getState().businessProfile).toBe('retail');
    expect(getState().features).toEqual({
      tables: false,
      kitchenDisplay: false,
      modifiers: false,
      combos: false,
      tips: false,
      giftCards: true,
      customers: true,
      inventory: true,
      cashDrawer: true,
    });
    expect(getState().printer).toEqual({
      name: 'Default',
      copies: 1,
      cashDrawerKick: true,
      drawerPin: 2,
      transport: 'windows',
      serialPort: '',
      serialBaudRate: 9600,
      networkHost: '',
      networkPort: 9100,
    });
    expect(getState().autoLogoutMinutes).toBe(10);
    expect(getState().speedMode).toBe(false);
    expect(getState().barMode).toBe(false);
    expect(getState().quickAccessSidebar).toBe(true);
    expect(getState().kdsStation).toBe('all');
    expect(getState().language).toBe('en');
  });

  // ── setTaxRate ─────────────────────────────────────────────────

  it('sets tax rate', () => {
    act(() => getState().setTaxRate(0.1));
    expect(getState().taxRate).toBe(0.1);
  });

  it('sets tax rate to zero', () => {
    act(() => getState().setTaxRate(0));
    expect(getState().taxRate).toBe(0);
  });

  it('clamps invalid tax rates to the production range', () => {
    act(() => getState().setTaxRate(-0.2));
    expect(getState().taxRate).toBe(0);

    act(() => getState().setTaxRate(2));
    expect(getState().taxRate).toBe(1);
  });

  // ── setTipPresets ──────────────────────────────────────────────

  it('sets tip presets', () => {
    act(() => getState().setTipPresets([0.15, 0.2, 0.25]));
    expect(getState().tipPresets).toEqual([0.15, 0.2, 0.25]);
  });

  it('sets empty tip presets', () => {
    act(() => getState().setTipPresets([]));
    expect(getState().tipPresets).toEqual([]);
  });

  it('clamps tip presets to the production range', () => {
    act(() => getState().setTipPresets([-0.1, 0.18, 2]));
    expect(getState().tipPresets).toEqual([0, 0.18, 1]);
  });

  it('normalizes non-finite tip presets to zero', () => {
    act(() => getState().setTipPresets([Number.NaN, Number.POSITIVE_INFINITY]));
    expect(getState().tipPresets).toEqual([0, 0]);
  });

  // ── setCurrency ────────────────────────────────────────────────

  it('sets currency code and symbol', () => {
    act(() => getState().setCurrency('EUR', '€'));
    expect(getState().currency).toBe('EUR');
    expect(getState().currencySymbol).toBe('€');
  });

  it('normalizes currency code before persisting', () => {
    act(() => getState().setCurrency(' eur ', '€'));
    expect(getState().currency).toBe('EUR');
  });

  it('falls back to default currency when code or symbol is blank', () => {
    act(() => getState().setCurrency('', ''));
    expect(getState().currency).toBe('USD');
    expect(getState().currencySymbol).toBe('$');
  });

  // ── setReceiptFooter ──────────────────────────────────────────

  it('sets receipt footer', () => {
    act(() => getState().setReceiptFooter('Come again!'));
    expect(getState().receiptFooter).toBe('Come again!');
  });

  it('sets empty receipt footer', () => {
    act(() => getState().setReceiptFooter(''));
    expect(getState().receiptFooter).toBe('');
  });

  // ── POS Experience ────────────────────────────────────────────

  it('applies restaurant profile feature defaults', () => {
    act(() => getState().setBusinessProfile('restaurant'));
    expect(getState().businessProfile).toBe('restaurant');
    expect(getState().features.tables).toBe(true);
    expect(getState().features.kitchenDisplay).toBe(true);
    expect(getState().features.tips).toBe(true);
  });

  it('updates individual POS features', () => {
    act(() => getState().setFeature('tables', true));
    expect(getState().features.tables).toBe(true);
    expect(getState().features.kitchenDisplay).toBe(false);
  });

  // ── setPrinter ─────────────────────────────────────────────────

  it('updates printer partially', () => {
    act(() => getState().setPrinter({ copies: 3 }));
    expect(getState().printer.copies).toBe(3);
    expect(getState().printer.name).toBe('Default');
    expect(getState().printer.cashDrawerKick).toBe(true);
  });

  it('updates printer name only', () => {
    act(() => getState().setPrinter({ name: 'Kitchen Printer' }));
    expect(getState().printer.name).toBe('Kitchen Printer');
    expect(getState().printer.copies).toBe(1);
  });

  it('updates cashDrawerKick', () => {
    act(() => getState().setPrinter({ cashDrawerKick: false }));
    expect(getState().printer.cashDrawerKick).toBe(false);
  });

  it('clamps printer copies between 1 and 5', () => {
    act(() => getState().setPrinter({ copies: 0 }));
    expect(getState().printer.copies).toBe(1);

    act(() => getState().setPrinter({ copies: 99 }));
    expect(getState().printer.copies).toBe(5);
  });

  // ── setAutoLogoutMinutes ──────────────────────────────────────

  it('sets auto logout minutes', () => {
    act(() => getState().setAutoLogoutMinutes(30));
    expect(getState().autoLogoutMinutes).toBe(30);
  });

  it('clamps negative auto logout minutes to zero', () => {
    act(() => getState().setAutoLogoutMinutes(-5));
    expect(getState().autoLogoutMinutes).toBe(0);
  });

  // ── Boolean toggles ──────────────────────────────────────────

  it('sets speed mode', () => {
    act(() => getState().setSpeedMode(true));
    expect(getState().speedMode).toBe(true);
  });

  it('sets bar mode', () => {
    act(() => getState().setBarMode(true));
    expect(getState().barMode).toBe(true);
  });

  it('sets quick access sidebar', () => {
    act(() => getState().setQuickAccessSidebar(false));
    expect(getState().quickAccessSidebar).toBe(false);
  });

  // ── setKdsStation ─────────────────────────────────────────────

  it('sets KDS station', () => {
    act(() => getState().setKdsStation('bar'));
    expect(getState().kdsStation).toBe('bar');
  });

  // ── reset ─────────────────────────────────────────────────────

  it('resets all settings to defaults', () => {
    act(() => {
      getState().setTaxRate(0.15);
      getState().setSpeedMode(true);
      getState().setBarMode(true);
      getState().setCurrency('EUR', '€');
      getState().reset();
    });

    expect(getState().taxRate).toBe(0.08);
    expect(getState().speedMode).toBe(false);
    expect(getState().barMode).toBe(false);
    expect(getState().currency).toBe('USD');
    expect(getState().currencySymbol).toBe('$');
  });
});
