/**
 * ERP settings validation and transformation helpers.
 * All functions are pure with no side-effects.
 */

export type CurrencyCode = 'USD' | 'EUR' | 'GBP' | 'AED' | 'SAR' | 'CAD' | 'AUD';

export interface ErpSettings {
  currency: CurrencyCode;
  taxRate: number;
  lowStockThreshold: number;
  defaultWarehouseId?: number;
  fiscalYearStart: number;
  autoReorder: boolean;
  reorderLeadDays: number;
  valuationMethod: 'FIFO' | 'LIFO' | 'WAC';
}

export interface SettingsValidationError {
  field: string;
  message: string;
}

/** Validate ERP settings and return a list of validation errors. */
export function validateErpSettings(settings: Partial<ErpSettings>): SettingsValidationError[] {
  const errors: SettingsValidationError[] = [];

  if (settings.taxRate !== undefined && (settings.taxRate < 0 || settings.taxRate > 100)) {
    errors.push({ field: 'taxRate', message: 'Tax rate must be between 0 and 100.' });
  }

  if (settings.lowStockThreshold !== undefined && settings.lowStockThreshold < 0) {
    errors.push({ field: 'lowStockThreshold', message: 'Low stock threshold cannot be negative.' });
  }

  if (
    settings.fiscalYearStart !== undefined &&
    (settings.fiscalYearStart < 1 || settings.fiscalYearStart > 12)
  ) {
    errors.push({
      field: 'fiscalYearStart',
      message: 'Fiscal year start must be a month between 1 and 12.',
    });
  }

  if (settings.reorderLeadDays !== undefined && settings.reorderLeadDays < 0) {
    errors.push({ field: 'reorderLeadDays', message: 'Reorder lead days cannot be negative.' });
  }

  return errors;
}

/** Return whether settings are valid (no errors). */
export function isSettingsValid(settings: Partial<ErpSettings>): boolean {
  return validateErpSettings(settings).length === 0;
}

/** Merge partial settings into existing settings (non-destructive). */
export function mergeSettings(current: ErpSettings, updates: Partial<ErpSettings>): ErpSettings {
  return { ...current, ...updates };
}

/** Return the currency symbol for a currency code. */
export function currencySymbol(code: CurrencyCode): string {
  const symbols: Record<CurrencyCode, string> = {
    USD: '$',
    EUR: '€',
    GBP: '£',
    AED: 'AED',
    SAR: 'SAR',
    CAD: 'CA$',
    AUD: 'A$',
  };
  return symbols[code] ?? code;
}

/** Return the fiscal year end month given the start month (1-based). */
export function fiscalYearEndMonth(startMonth: number): number {
  return ((startMonth - 2 + 12) % 12) + 1;
}

/** Return default ERP settings. */
export function defaultErpSettings(): ErpSettings {
  return {
    currency: 'USD',
    taxRate: 0,
    lowStockThreshold: 10,
    fiscalYearStart: 1,
    autoReorder: false,
    reorderLeadDays: 7,
    valuationMethod: 'WAC',
  };
}
