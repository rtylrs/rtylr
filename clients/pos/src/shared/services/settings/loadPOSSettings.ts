import { createOneKitClient } from '@voxire/api';
import { SettingsServiceClient } from '@voxire/voxie/pos/settings/v1/client';
import { useSettingsStore, type POSBusinessProfile } from '../../stores/settingsStore';
import { useThemeStore } from '../../stores/themeStore';

const client = createOneKitClient(SettingsServiceClient, 'pos');
function isProfile(value: string): value is POSBusinessProfile {
  return value === 'retail' || value === 'restaurant' || value === 'quick_service';
}
export async function loadPOSSettings(isCurrent: () => boolean = () => true): Promise<void> {
  const { settings } = await client.getPOSSettings({});
  if (!isCurrent()) return;
  if (!settings) throw new Error('Could not load terminal configuration');
  const store = useSettingsStore.getState();
  const profile = settings.businessProfile ?? '';
  if (isProfile(profile)) store.setBusinessProfile(profile);
  if (settings.features) store.setFeatures(settings.features);
  store.setReceiptFooter(settings.receiptFooter ?? '');
  useThemeStore.getState().setTheme(settings.theme ?? '');
}
