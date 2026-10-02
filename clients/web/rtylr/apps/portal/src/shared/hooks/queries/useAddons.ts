import { useQuery } from '@tanstack/react-query';
import { createOneKitClient } from '@voxire/api';
import { AppServiceClient } from '@voxire/voxie/hub/app/v1/client';

export interface AddonData {
  id: string;
  name: string;
  description: string;
  icon: string;
  price: number;
  isFree: boolean;
  isEnabled: boolean;
}

const appClient = createOneKitClient(AppServiceClient, 'dash');

export const ADDONS_QUERY_KEY = ['addons'] as const;

export const useAddons = () =>
  useQuery<AddonData[]>({
    queryKey: ADDONS_QUERY_KEY,
    queryFn: async (): Promise<AddonData[]> => {
      const response = await appClient.listAddons({});
      return (response.addons ?? []).map((addon) => ({
        id: addon.id ?? '',
        name: addon.name ?? '',
        description: addon.description ?? '',
        icon: addon.icon ?? '',
        price: addon.price ?? 0,
        isFree: addon.isFree ?? false,
        isEnabled: addon.isEnabled ?? false,
      }));
    },
  });

export const enableAddon = (addonId: string) => appClient.enableAddon({ addonId });

export const disableAddon = (addonId: string) => appClient.disableAddon({ addonId });

export function isAddonEnabled(addons: AddonData[], addonId: string): boolean {
  return addons.some((addon) => addon.id === addonId && addon.isEnabled);
}
