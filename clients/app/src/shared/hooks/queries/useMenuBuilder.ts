import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { createOneKitClient } from '@voxire/api';
import i18next from 'i18next';
import { toast } from 'sonner';
import type {
  BranchMenuItemConfig,
  MenuCurrency,
  MenuPreset,
  MenuProfile,
  MenuReview,
  MenuStampCard,
  MenuTableAction,
  MenuTranslationSourceEntry,
} from '@voxire/voxie/hub/menu/v1/types';
import { MenuItemServiceClient } from '@voxire/voxie/pos/menu_item/v1/client';
import type { MenuItem as ApiMenuItem } from '@voxire/voxie/pos/menu_item/v1/types';
import { businessClient, menuClient, posLocationClient, tableClient } from '../../lib/clients';

const menuItemClient = createOneKitClient(MenuItemServiceClient, 'dash');

export interface MenuBranchOption {
  id: number;
  name: string;
  slug: string;
  address?: string;
}

export interface MenuTableOption {
  id: string;
  locationId: number;
  label: string;
}

/**
 * Resolves the business's public branches using the same HR location source
 * (`businessClient.lookupBusinessBySlug`) that `LocationMenu` renders, so
 * branch IDs and URL slugs always match across Studio, QR codes, and storefront.
 */
export const useMenuBranches = (businessSlug?: string) =>
  useQuery<MenuBranchOption[]>({
    queryKey: ['menu-builder', 'branches', businessSlug ?? ''],
    staleTime: 60_000,
    queryFn: async () => {
      const slug = (businessSlug ?? '').trim().toLowerCase();
      if (slug) {
        try {
          const res = await businessClient.lookupBusinessBySlug({ slug });
          const hrLocations = (res.locations ?? [])
            .map((loc) => ({
              id: Number(loc.id ?? 0),
              name: loc.name ?? '',
              slug: loc.slug ?? '',
              address: loc.address ?? '',
            }))
            .filter((loc) => loc.id > 0 && loc.slug);
          if (hrLocations.length > 0) return hrLocations;
        } catch {
          // Fall back to POS locations when public slug lookup is unavailable.
        }
      }
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

export const useMenuTables = () =>
  useQuery<MenuTableOption[]>({
    queryKey: ['menu-builder', 'tables'],
    staleTime: 60_000,
    queryFn: async () => {
      const res = await tableClient.listTables({});
      return (res.tables ?? []).map((tb) => ({
        id: String(tb.id ?? ''),
        locationId: Number(tb.locationId ?? 0),
        label: tb.label ?? `Table ${tb.id ?? ''}`,
      }));
    },
  });

export const useMenuCatalogItems = () =>
  useQuery<ApiMenuItem[]>({
    queryKey: ['menu-builder', 'catalog-items'],
    staleTime: 60_000,
    queryFn: async () => {
      const res = await menuItemClient.listMenuItems({});
      return res.items ?? [];
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
      defaultLocale?: string;
      enabledLocales?: string[];
      defaultCurrency?: string;
    }) => {
      const res = await menuClient.saveMenuProfileDraft({
        ...(params.locationId > 0 ? { locationId: String(params.locationId) } : {}),
        draftConfigJson: params.draftConfigJson,
        customCss: params.customCss,
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
      toast.success(i18next.t('design.toasts.draftSaved', 'Draft saved'));
    },
    onError: (err: unknown) => {
      const msg =
        (err as { message?: string })?.message ||
        i18next.t('design.toasts.draftSaveFailed', 'Failed to save draft');
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
      void queryClient.invalidateQueries({ queryKey: ['public-menu-brand-bundle'] });
      toast.success(i18next.t('design.toasts.published', 'Branded menu published live'));
    },
    onError: (err: unknown) => {
      const msg =
        (err as { message?: string })?.message ||
        i18next.t('design.toasts.publishFailed', 'Failed to publish menu');
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
      toast.success(i18next.t('design.toasts.cloned', 'Copied branch menu configuration'));
    },
    onError: (err: unknown) => {
      const msg =
        (err as { message?: string })?.message ||
        i18next.t('design.toasts.cloneFailed', 'Failed to copy branch menu');
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

export const useBranchMenuItems = (locationId: number) =>
  useQuery<BranchMenuItemConfig[]>({
    queryKey: ['menu-builder', 'branch-items', locationId],
    queryFn: async () => {
      const res = await menuClient.listBranchMenuItems({
        ...(locationId > 0 ? { locationId: String(locationId) } : {}),
      });
      return res.items ?? [];
    },
  });

export const useUpsertBranchMenuItem = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (params: {
      locationId: number;
      menuItemId: string;
      isVisible: boolean;
      hasPriceOverride?: boolean;
      priceOverride?: number;
      displayOrder?: number;
      featuredBadge?: string;
    }) => {
      const res = await menuClient.upsertBranchMenuItem({
        ...(params.locationId > 0 ? { locationId: String(params.locationId) } : {}),
        menuItemId: params.menuItemId,
        isVisible: params.isVisible,
        hasPriceOverride: params.hasPriceOverride ?? false,
        priceOverride: params.priceOverride ?? 0,
        displayOrder: params.displayOrder ?? 0,
        featuredBadge: params.featuredBadge ?? '',
      });
      return res.item;
    },
    onSuccess: (_data, vars) => {
      void queryClient.invalidateQueries({
        queryKey: ['menu-builder', 'branch-items', vars.locationId],
      });
      void queryClient.invalidateQueries({ queryKey: ['public-menu-brand-bundle'] });
      toast.success(i18next.t('design.toasts.branchItemSaved', 'Branch item updated'));
    },
    onError: (err: unknown) => {
      const msg =
        (err as { message?: string })?.message ||
        i18next.t('design.toasts.branchItemFailed', 'Failed to update branch item');
      toast.error(msg);
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
      toast.success(i18next.t('design.toasts.currencySaved', 'Currency saved'));
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
      toast.success(i18next.t('design.toasts.stampCardSaved', 'Stamp card saved'));
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
        i18next.t('design.toasts.translated', {
          count: res.translatedCount ?? 0,
          skipped: res.skippedManualCount ?? 0,
          defaultValue: 'Translated {{count}} entries ({{skipped}} manual kept)',
        })
      );
    },
    onError: (err: unknown) => {
      const msg =
        (err as { message?: string })?.message ||
        i18next.t('design.toasts.translationFailed', 'Translation failed');
      toast.error(msg);
    },
  });

export const useTableActions = (locationId: number, status?: string) =>
  useQuery<MenuTableAction[]>({
    queryKey: ['menu-builder', 'table-actions', locationId, status ?? ''],
    refetchInterval: 30_000,
    queryFn: async () => {
      const res = await menuClient.listTableActions({
        ...(locationId > 0 ? { locationId: String(locationId) } : {}),
        ...(status ? { status } : {}),
      });
      return res.actions ?? [];
    },
  });

export const useUpdateTableActionStatus = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (params: { id: string; status: 'open' | 'acknowledged' | 'resolved' }) => {
      await menuClient.updateTableActionStatus(params);
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['menu-builder', 'table-actions'] });
      toast.success(i18next.t('design.toasts.tableActionUpdated', 'Table request updated'));
    },
    onError: (err: unknown) => {
      const msg =
        (err as { message?: string })?.message ||
        i18next.t('design.toasts.tableActionFailed', 'Failed to update table request');
      toast.error(msg);
    },
  });
};

export const useMenuAnalyticsSummary = (locationId: number, days = 30) =>
  useQuery({
    queryKey: ['menu-builder', 'analytics', locationId, days],
    queryFn: () =>
      menuClient.getMenuAnalyticsSummary({
        ...(locationId > 0 ? { locationId: String(locationId) } : {}),
        days,
      }),
  });

export const useMenuReviews = (locationId: number, status?: string) =>
  useQuery<MenuReview[]>({
    queryKey: ['menu-builder', 'reviews', locationId, status ?? ''],
    queryFn: async () => {
      const res = await menuClient.listMenuReviews({
        ...(locationId > 0 ? { locationId: String(locationId) } : {}),
        ...(status ? { status } : {}),
      });
      return res.reviews ?? [];
    },
  });

export const useUpdateMenuReviewStatus = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (params: { id: string; status: 'pending' | 'approved' | 'hidden' }) => {
      await menuClient.updateMenuReviewStatus(params);
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['menu-builder', 'reviews'] });
      void queryClient.invalidateQueries({ queryKey: ['public-menu-brand-bundle'] });
      toast.success(i18next.t('design.toasts.reviewUpdated', 'Review updated'));
    },
    onError: (err: unknown) => {
      const msg =
        (err as { message?: string })?.message ||
        i18next.t('design.toasts.reviewUpdateFailed', 'Failed to update review');
      toast.error(msg);
    },
  });
};

export const useDeleteMenuReview = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      await menuClient.deleteMenuReview({ id });
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['menu-builder', 'reviews'] });
      void queryClient.invalidateQueries({ queryKey: ['public-menu-brand-bundle'] });
      toast.success(i18next.t('design.toasts.reviewDeleted', 'Review deleted'));
    },
    onError: (err: unknown) => {
      const msg =
        (err as { message?: string })?.message ||
        i18next.t('design.toasts.reviewDeleteFailed', 'Failed to delete review');
      toast.error(msg);
    },
  });
};
