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

interface RawAddon {
  id?: string;
  name?: string;
  description?: string;
  icon?: string;
  price?: number | string;
  isFree?: boolean;
  is_free?: boolean;
  isEnabled?: boolean;
  is_enabled?: boolean;
}

function finiteNumber(value: unknown): number {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : 0;
}

export const ADDONS_QUERY_KEY = ['addons'] as const;

export const useAddons = () =>
  useQuery<AddonData[]>({
    queryKey: ADDONS_QUERY_KEY,
    queryFn: async () => {
      try {
        const res = await unwrap(safe.get<{ addons?: RawAddon[] }>('/api/hub/app/v1/addons'));
        const list = res.addons ?? [];
        return list.map((addon) => ({
          id: String(addon.id ?? ''),
          name: String(addon.name ?? ''),
          description: String(addon.description ?? ''),
          icon: String(addon.icon ?? ''),
          price: finiteNumber(addon.price),
          isFree: Boolean(addon.is_free ?? addon.isFree),
          isEnabled: Boolean(addon.is_enabled ?? addon.isEnabled),
        }));
      } catch {
        return [];
      }
    },
  });

export function isAddonEnabled(addons: AddonData[], addonId: string): boolean {
  return addons.some((addon) => addon.id === addonId && addon.isEnabled);
}
