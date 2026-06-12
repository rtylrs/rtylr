import { useQuery } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { safe } from '../../lib/api';

export interface AddonData {
  id: string;
  name: string;
  description: string;
  icon: string;
  price: number;
  isFree: boolean;
  isEnabled: boolean;
}

export const ADDONS_QUERY_KEY = ['addons'] as const;

export const useAddons = () =>
  useQuery<AddonData[]>({
    queryKey: ADDONS_QUERY_KEY,
    queryFn: async () => {
      try {
        const res = await unwrap(safe.get<{ addons?: AddonData[] }>('/addons'));
        return res.addons ?? [];
      } catch {
        return [];
      }
    },
  });

export function isAddonEnabled(addons: AddonData[], addonId: string): boolean {
  return addons.some((addon) => addon.id === addonId && addon.isEnabled);
}
