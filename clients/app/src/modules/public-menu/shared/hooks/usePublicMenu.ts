import { useQuery } from '@tanstack/react-query';
import { cdnImageUrl } from '@voxire/config';
import {
  DEFAULT_LAYOUT_CONFIG,
  type MenuBannerData,
  type MenuCurrencyOption,
  type MenuLayoutConfig,
  type MenuStampCardSummary,
  type TableActionType,
} from '@voxire/menu';
import { businessClient, menuBuilderClient, menuItemClient } from '../lib/clients';

export interface PublicBusiness {
  id: number;
  name: string;
  logoUrl: string;
  currency: string;
  timezone: string;
}

export interface PublicMenuItem {
  id: number;
  ownerId: number;
  categoryId: number;
  category?: string;
  name: string;
  description?: string;
  sku: string;
  salePrice: number;
  imageUrl?: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface BranchItemOverride {
  menuItemId: number;
  isVisible: boolean;
  hasPriceOverride: boolean;
  priceOverride: number;
  displayOrder: number;
  featuredBadge: string;
  dietaryTags: string[];
  allergens: string[];
  calories: number;
  mediaUrls: string[];
  upsellItemIds: number[];
}

export interface PublicBrandBundle {
  config: MenuLayoutConfig;
  defaultLocale: string;
  enabledLocales: string[];
  defaultCurrency: string;
  banners: MenuBannerData[];
  branchItems: Map<number, BranchItemOverride>;
  currencies: MenuCurrencyOption[];
  translations: Map<string, string>;
  stampCards: MenuStampCardSummary[];
}

const toNumber = (value: string | undefined, fallback = 0): number => {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : fallback;
};

async function fetchPublicBusiness(businessId: number): Promise<PublicBusiness> {
  const response = await businessClient.getPublicBusiness({ businessId: String(businessId) });
  const business = response.business;
  if (!business) throw new Error('Public business was not returned');

  return {
    id: toNumber(business.id, businessId),
    name: business.name ?? '',
    logoUrl: business.logoUrl ? (cdnImageUrl(business.logoUrl) ?? '') : '',
    currency: business.currency ?? 'USD',
    timezone: business.timezone ?? 'UTC',
  };
}

async function fetchPublicMenu(
  businessId: number,
  locationId?: number | null
): Promise<PublicMenuItem[]> {
  const response = await menuItemClient.listPublicMenuItems({
    businessId: String(businessId),
    ...(locationId && locationId > 0 ? { locationId: String(locationId) } : {}),
  });

  return (response.items ?? []).map((item) => ({
    id: toNumber(item.id),
    ownerId: toNumber(item.ownerId),
    categoryId: toNumber(item.categoryId),
    ...(item.category ? { category: item.category } : {}),
    name: item.name ?? '',
    ...(item.description ? { description: item.description } : {}),
    sku: item.sku ?? '',
    salePrice: item.salePrice ?? 0,
    imageUrl: item.imageUrl ? (cdnImageUrl(item.imageUrl) ?? undefined) : undefined,
    isActive: item.isActive ?? true,
    createdAt: item.createdAt ?? '',
    updatedAt: item.updatedAt ?? '',
  }));
}

export interface PublicMenuBundle {
  business: PublicBusiness;
  items: PublicMenuItem[];
}

export interface PublicLocationSummary {
  id: number;
  name: string;
  slug: string;
  address: string;
  timezone: string;
  isActive: boolean;
}

export interface PublicBusinessBySlugBundle {
  business: PublicBusiness & { slug: string };
  locations: PublicLocationSummary[];
}

export function usePublicBusinessBySlug(slug: string | null): {
  data: PublicBusinessBySlugBundle | undefined;
  isLoading: boolean;
  isError: boolean;
} {
  const query = useQuery({
    queryKey: ['public-business-by-slug', slug],
    queryFn: async (): Promise<PublicBusinessBySlugBundle> => {
      if (!slug) throw new Error('slug required');

      const response = await businessClient.lookupBusinessBySlug({ slug });
      const business = response.business;
      if (!business) throw new Error('Public business was not returned');

      return {
        business: {
          id: toNumber(business.id),
          name: business.name ?? '',
          slug: business.slug ?? slug,
          logoUrl: business.logoUrl ? (cdnImageUrl(business.logoUrl) ?? '') : '',
          currency: business.currency ?? 'USD',
          timezone: business.timezone ?? 'UTC',
        },
        locations: (response.locations ?? []).map((location) => ({
          id: toNumber(location.id),
          name: location.name ?? '',
          slug: location.slug ?? '',
          address: location.address ?? '',
          timezone: location.timezone ?? 'UTC',
          isActive: location.isActive ?? true,
        })),
      };
    },
    enabled: !!slug,
    staleTime: 60_000,
  });

  return {
    data: query.data,
    isLoading: query.isLoading,
    isError: query.isError,
  };
}

export function usePublicMenu(
  businessId: number | null,
  locationId?: number | null
): {
  data: PublicMenuBundle | undefined;
  isLoading: boolean;
  isError: boolean;
} {
  const query = useQuery({
    queryKey: ['public-menu', businessId, locationId ?? null],
    queryFn: async (): Promise<PublicMenuBundle> => {
      if (businessId == null) {
        throw new Error('business_id required');
      }

      // Two-trip parallel. If either fails the whole bundle fails - a menu
      // without a business header renders poorly and vice versa.
      const [business, items] = await Promise.all([
        fetchPublicBusiness(businessId),
        fetchPublicMenu(businessId, locationId),
      ]);
      return { business, items };
    },
    enabled: businessId != null && Number.isFinite(businessId),
    staleTime: 60_000,
  });

  return {
    data: query.data,
    isLoading: query.isLoading,
    isError: query.isError,
  };
}

function parseProfileConfig(raw?: string): Partial<MenuLayoutConfig> {
  if (!raw || !raw.trim()) return {};
  try {
    return JSON.parse(raw) as Partial<MenuLayoutConfig>;
  } catch {
    return {};
  }
}

export function translationLookupKey(
  entityType: string,
  entityId: string | number,
  locale: string,
  field: string
): string {
  return `${entityType.toLowerCase()}|${String(entityId)}|${locale.toLowerCase()}|${field.toLowerCase()}`;
}

export function usePublicBrandBundle(
  businessId: number | null,
  locationId?: number | null,
  locale?: string
) {
  return useQuery<PublicBrandBundle>({
    queryKey: ['public-menu-brand-bundle', businessId, locationId ?? 0, locale ?? ''],
    enabled: businessId != null && Number.isFinite(businessId) && businessId > 0,
    staleTime: 60_000,
    retry: false,
    queryFn: async () => {
      const res = await menuBuilderClient.getPublicMenuBundle({
        businessId: String(businessId),
        ...(locationId && locationId > 0 ? { locationId: String(locationId) } : {}),
        ...(locale ? { locale } : {}),
      });
      const profile = res.profile;
      const rawCfg = profile?.isPublished
        ? profile.publishedConfigJson || profile.draftConfigJson
        : profile?.draftConfigJson;
      const parsed = parseProfileConfig(rawCfg);
      const customCss = profile?.isPublished
        ? profile.publishedCustomCss || profile.customCss || ''
        : profile?.customCss || '';

      const config: MenuLayoutConfig = {
        ...DEFAULT_LAYOUT_CONFIG,
        ...parsed,
        header: { ...DEFAULT_LAYOUT_CONFIG.header, ...parsed.header },
        nav: { ...DEFAULT_LAYOUT_CONFIG.nav, ...parsed.nav },
        items: { ...DEFAULT_LAYOUT_CONFIG.items, ...parsed.items },
        pages: { menu: true, ...parsed.pages },
        typography: { ...DEFAULT_LAYOUT_CONFIG.typography!, ...parsed.typography },
        customCss: customCss || parsed.customCss,
        hidePoweredBy: profile?.whiteLabelHideBadge ?? parsed.hidePoweredBy ?? false,
      };

      const branchItems = new Map<number, BranchItemOverride>();
      for (const bi of res.branchItems ?? []) {
        const menuItemId = toNumber(bi.menuItemId);
        if (menuItemId > 0) {
          branchItems.set(menuItemId, {
            menuItemId,
            isVisible: bi.isVisible !== false,
            hasPriceOverride: Boolean(bi.hasPriceOverride),
            priceOverride: bi.priceOverride ?? 0,
            displayOrder: bi.displayOrder ?? 0,
            featuredBadge: bi.featuredBadge ?? '',
            dietaryTags: bi.dietaryTags ?? [],
            allergens: bi.allergens ?? [],
            calories: bi.calories ?? 0,
            mediaUrls: bi.mediaUrls ?? [],
            upsellItemIds: (bi.upsellItemIds ?? [])
              .map((id) => toNumber(id))
              .filter((id) => id > 0),
          });
        }
      }

      const translations = new Map<string, string>();
      for (const tr of res.translations ?? []) {
        translations.set(
          translationLookupKey(
            tr.entityType ?? '',
            tr.entityId ?? '',
            tr.locale ?? '',
            tr.field ?? ''
          ),
          tr.value ?? ''
        );
      }

      return {
        config,
        defaultLocale: profile?.defaultLocale || 'en',
        enabledLocales:
          profile?.enabledLocales && profile.enabledLocales.length > 0
            ? profile.enabledLocales
            : ['en'],
        defaultCurrency: profile?.defaultCurrency || 'USD',
        banners: (res.banners ?? []).map((b) => ({
          id: String(b.id ?? ''),
          title: b.title ?? '',
          body: b.body ?? '',
          imageUrl: b.imageUrl ?? undefined,
          linkUrl: b.linkUrl ?? undefined,
        })),
        branchItems,
        currencies: (res.currencies ?? [])
          .filter((c) => c.isActive !== false)
          .map((c) => ({
            code: c.code ?? 'USD',
            symbol: c.symbol ?? '$',
            exchangeRate: c.exchangeRate ?? 1,
            decimalPlaces: c.decimalPlaces ?? 2,
            isDefault: Boolean(c.isDefault),
          })),
        translations,
        stampCards: (res.stampCards ?? []).map((sc) => ({
          id: String(sc.id ?? ''),
          title: sc.title ?? '',
          description: sc.description ?? '',
          stampsRequired: sc.stampsRequired ?? 10,
          rewardLabel: sc.rewardLabel ?? '',
        })),
      };
    },
  });
}

export async function submitPublicTableAction(params: {
  businessId: number;
  locationId?: number;
  tableId?: number;
  tableLabel?: string;
  actionType: TableActionType;
  note?: string;
}): Promise<void> {
  await menuBuilderClient.createPublicTableAction({
    businessId: String(params.businessId),
    ...(params.locationId ? { locationId: String(params.locationId) } : {}),
    ...(params.tableId ? { tableId: String(params.tableId) } : {}),
    ...(params.tableLabel ? { tableLabel: params.tableLabel } : {}),
    actionType: params.actionType,
    ...(params.note ? { note: params.note } : {}),
  });
}

export function trackPublicMenuEvent(params: {
  businessId: number;
  locationId?: number;
  tableId?: number;
  eventType:
    'qr_scan' | 'menu_view' | 'item_view' | 'add_to_cart' | 'checkout_start' | 'order_placed';
  itemId?: number;
  locale?: string;
  currency?: string;
}): void {
  const sessionId =
    typeof sessionStorage !== 'undefined'
      ? (sessionStorage.getItem('rtylr-menu-sid') ??
        (() => {
          const sid = `m_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`;
          sessionStorage.setItem('rtylr-menu-sid', sid);
          return sid;
        })())
      : '';
  void menuBuilderClient
    .trackPublicMenuEvents({
      businessId: String(params.businessId),
      events: [
        {
          ...(params.locationId ? { locationId: String(params.locationId) } : {}),
          ...(params.tableId ? { tableId: String(params.tableId) } : {}),
          sessionId,
          eventType: params.eventType,
          ...(params.itemId ? { itemId: String(params.itemId) } : {}),
          ...(params.locale ? { locale: params.locale } : {}),
          ...(params.currency ? { currency: params.currency } : {}),
        },
      ],
    })
    .catch(() => undefined);
}
