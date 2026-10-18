import { useSettings } from '../lib/settingsStore';

/**
 * The business's ERP currency. ErpSettingsLoader fills the settings store when
 * the ERP layout mounts, so screens read it without a query of their own.
 */
export function useErpCurrency(): string {
  return useSettings((state) => state.currency || 'USD');
}
