import { useErpSettings } from '../hooks/queries/useErpSettings';

/** Loads the business's ERP settings (currency, tax rate) into the settings store. */
const ErpSettingsLoader = (): null => {
  useErpSettings();
  return null;
};

export default ErpSettingsLoader;
