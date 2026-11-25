import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import type {
  MenuCurrency,
  MenuPreset,
  MenuProfile,
  MenuStampCard,
  MenuTranslationSourceEntry,
} from '@voxire/voxie/hub/menu/v1/types';
import { menuClient, posLocationClient } from '../../lib/clients';

export interface MenuBranchOption {
  id: number;
  name: string;
  slug: string;
}

export const useMenuBranches = () =>
  useQuery<MenuBranchOption[]>({
    queryKey: ['menu-builder', 'branches'],
    staleTime: 60_000,
    queryFn: async () => {
      const res = await posLocationClient.listLocations({});
      return (res.locations ?? []).map((loc) => ({
        id: Number(loc.id ?? 0),
        name: loc.name ?? '',
        slug: (loc.name ?? '')
          .toLowerCase()
          .replace(/[^a-z0-9]+/g, '-')
          .replace(/^-|-$/g, ''),
      }));
    },
  });

export const useMenuProfile = (locationId: number) =>
  useQuery<MenuProfile | null>({
    queryKey: ['menu-builder', 'profile', locationId],
    queryFn: async () => {
      const res = await menuClient.getMenuProfile({
        ...(locationId > 0 ? { locationId: String(locationId) } : {}),
      });
      return res.profile ?? null;
    },
  });

export const useSaveMenuProfileDraft = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (params: {
      locationId: number;
      draftConfigJson: string;
      customCss?: string;
      whiteLabelHideBadge?: boolean;
      defaultLocale?: string;
      enabledLocales?: string[];
      defaultCurrency?: string;
    }) => {
      const res = await menuClient.saveMenuProfileDraft({
        ...(params.locationId > 0 ? { locationId: String(params.locationId) } : {}),
        draftConfigJson: params.draftConfigJson,
        customCss: params.customCss,
        whiteLabelHideBadge: params.whiteLabelHideBadge,
        defaultLocale: params.defaultLocale,
        enabledLocales: params.enabledLocales,
        defaultCurrency: params.defaultCurrency,
      });
      return res.profile;
    },
    onSuccess: (_data, vars) => {
      void queryClient.invalidateQueries({
        queryKey: ['menu-builder', 'profile', vars.locationId],
      });
      toast.success('Draft saved');
    },
    onError: (err: unknown) => {
      const msg = (err as { message?: string })?.message || 'Failed to save draft';
      toast.error(msg);
    },
  });
};

export const usePublishMenuProfile = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (locationId: number) => {
      const res = await menuClient.publishMenuProfile({
        ...(locationId > 0 ? { locationId: String(locationId) } : {}),
      });
      return res.profile;
    },
    onSuccess: (_data, locationId) => {
      void queryClient.invalidateQueries({ queryKey: ['menu-builder', 'profile', locationId] });
      toast.success('Branded menu published live');
    },
    onError: (err: unknown) => {
      const msg = (err as { message?: string })?.message || 'Failed to publish menu';
      toast.error(msg);
    },
  });
};

export const useCloneMenuProfile = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (params: { sourceLocationId: number; targetLocationId: number }) => {
      const res = await menuClient.cloneMenuProfile({
        sourceLocationId: String(params.sourceLocationId),
        targetLocationId: String(params.targetLocationId),
      });
      return res.profile;
    },
    onSuccess: (_data, vars) => {
      void queryClient.invalidateQueries({
        queryKey: ['menu-builder', 'profile', vars.targetLocationId],
      });
      void queryClient.invalidateQueries({
        queryKey: ['menu-builder', 'currencies', vars.targetLocationId],
      });
      toast.success('Copied branch menu configuration');
    },
    onError: (err: unknown) => {
      const msg = (err as { message?: string })?.message || 'Failed to copy branch menu';
      toast.error(msg);
    },
  });
};

export const useMenuPresets = () =>
  useQuery<MenuPreset[]>({
    queryKey: ['menu-builder', 'presets'],
    queryFn: async () => {
      const res = await menuClient.listMenuPresets({});
      return res.presets ?? [];
    },
  });

export const useCreateMenuPreset = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (params: { name: string; configJson: string }) => {
      const res = await menuClient.createMenuPreset(params);
      return res.preset;
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['menu-builder', 'presets'] });
    },
  });
};

export const useDeleteMenuPreset = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      await menuClient.deleteMenuPreset({ id });
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['menu-builder', 'presets'] });
    },
  });
};

export const useMenuCurrencies = (locationId: number) =>
  useQuery<MenuCurrency[]>({
    queryKey: ['menu-builder', 'currencies', locationId],
    queryFn: async () => {
      const res = await menuClient.listMenuCurrencies({
        ...(locationId > 0 ? { locationId: String(locationId) } : {}),
      });
      return res.currencies ?? [];
    },
  });

export const useUpsertMenuCurrency = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (params: {
      locationId: number;
      code: string;
      symbol: string;
      exchangeRate: number;
      decimalPlaces?: number;
      isDefault?: boolean;
      isActive?: boolean;
    }) => {
      const res = await menuClient.upsertMenuCurrency({
        ...(params.locationId > 0 ? { locationId: String(params.locationId) } : {}),
        code: params.code,
        symbol: params.symbol,
        exchangeRate: params.exchangeRate,
        decimalPlaces: params.decimalPlaces ?? 2,
        isDefault: params.isDefault ?? false,
        isActive: params.isActive ?? true,
      });
      return res.currencies ?? [];
    },
    onSuccess: (_data, vars) => {
      void queryClient.invalidateQueries({
        queryKey: ['menu-builder', 'currencies', vars.locationId],
      });
      toast.success('Currency saved');
    },
  });
};

export const useDeleteMenuCurrency = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (params: { locationId: number; code: string }) => {
      await menuClient.deleteMenuCurrency({
        ...(params.locationId > 0 ? { locationId: String(params.locationId) } : {}),
        code: params.code,
      });
    },
    onSuccess: (_data, vars) => {
      void queryClient.invalidateQueries({
        queryKey: ['menu-builder', 'currencies', vars.locationId],
      });
    },
  });
};

export const useStampCards = (locationId: number) =>
  useQuery<MenuStampCard[]>({
    queryKey: ['menu-builder', 'stamp-cards', locationId],
    queryFn: async () => {
      const res = await menuClient.listStampCards({
        ...(locationId > 0 ? { locationId: String(locationId) } : {}),
      });
      return res.cards ?? [];
    },
  });

export const useUpsertStampCard = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (params: {
      id?: string;
      locationId: number;
      title: string;
      description?: string;
      stampsRequired: number;
      rewardLabel: string;
      qualifyingMinAmount?: number;
      isActive?: boolean;
    }) => {
      const res = await menuClient.upsertStampCard({
        ...(params.id ? { id: params.id } : {}),
        ...(params.locationId > 0 ? { locationId: String(params.locationId) } : {}),
        title: params.title,
        description: params.description ?? '',
        stampsRequired: params.stampsRequired,
        rewardType: 'free_item',
        rewardLabel: params.rewardLabel,
        qualifyingMinAmount: params.qualifyingMinAmount ?? 0,
        isActive: params.isActive ?? true,
      });
      return res.card;
    },
    onSuccess: (_data, vars) => {
      void queryClient.invalidateQueries({
        queryKey: ['menu-builder', 'stamp-cards', vars.locationId],
      });
      toast.success('Stamp card saved');
    },
  });
};

export const useTranslateMenu = () =>
  useMutation({
    mutationFn: async (params: {
      locationId: number;
      targetLocales: string[];
      overwriteManual?: boolean;
      entries: MenuTranslationSourceEntry[];
    }) => {
      return menuClient.translateMenu({
        ...(params.locationId > 0 ? { locationId: String(params.locationId) } : {}),
        targetLocales: params.targetLocales,
        overwriteManual: params.overwriteManual ?? false,
        entries: params.entries,
      });
    },
    onSuccess: (res) => {
      toast.success(
        `Translated ${res.translatedCount ?? 0} entries (${res.skippedManualCount ?? 0} manual kept)`
      );
    },
    onError: (err: unknown) => {
      const msg = (err as { message?: string })?.message || 'Translation failed';
      toast.error(msg);
    },
  });

export const useMenuAnalyticsSummary = (locationId: number, days = 30) =>
  useQuery({
    queryKey: ['menu-builder', 'analytics', locationId, days],
    queryFn: () =>
      menuClient.getMenuAnalyticsSummary({
        ...(locationId > 0 ? { locationId: String(locationId) } : {}),
        days,
      }),
  });
