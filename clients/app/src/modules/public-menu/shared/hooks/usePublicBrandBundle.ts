import { useContext, useEffect, useState } from 'react';
import { QueryClientContext } from '@tanstack/react-query';
import {
  DEFAULT_LAYOUT_CONFIG,
  type CustomerStampProgressData,
  type MenuBannerData,
  type MenuCurrencyOption,
  type MenuLayoutConfig,
  type MenuStampCardSummary,
  type TableActionType,
} from '@voxire/menu';
import { menuBuilderClient } from '../lib/clients';
import type { BranchItemOverride, PublicBrandBundle } from './usePublicMenu';

const toNumber = (value: string | undefined, fallback = 0): number => {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : fallback;
};

function parseProfileConfig(raw?: string): Partial<MenuLayoutConfig> {
  if (!raw || !raw.trim()) return {};
  try {
    return JSON.parse(raw) as Partial<MenuLayoutConfig>;
  } catch {
    return {};
  }
}

export function translationKey(
  entityType: string,
  entityId: string | number,
  locale: string,
  field: string
): string {
  return `${entityType.toLowerCase()}|${String(entityId)}|${locale.toLowerCase()}|${field.toLowerCase()}`;
}

export async function fetchPublicBrandBundle(
  businessId: number,
  locationId?: number | null,
  locale?: string
): Promise<PublicBrandBundle> {
  const res = await menuBuilderClient.getPublicMenuBundle({
    businessId: String(businessId),
    ...(locationId && locationId > 0 ? { locationId: String(locationId) } : {}),
    ...(locale ? { locale } : {}),
  });
  const profile = res.profile;
  const rawCfg = profile?.isPublished ? profile.publishedConfigJson : undefined;
  const parsed = parseProfileConfig(rawCfg);
  const customCss = profile?.isPublished ? profile.publishedCustomCss || '' : '';

  const config: MenuLayoutConfig = {
    ...DEFAULT_LAYOUT_CONFIG,
    ...parsed,
    header: { ...DEFAULT_LAYOUT_CONFIG.header, ...parsed.header },
    nav: { ...DEFAULT_LAYOUT_CONFIG.nav, ...parsed.nav },
    items: { ...DEFAULT_LAYOUT_CONFIG.items, ...parsed.items },
    pages: { menu: true, ...parsed.pages },
    typography: { ...DEFAULT_LAYOUT_CONFIG.typography!, ...parsed.typography },
    customCss: customCss || undefined,
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
        upsellItemIds: (bi.upsellItemIds ?? []).map((id) => toNumber(id)).filter((id) => id > 0),
      });
    }
  }

  const translations = new Map<string, string>();
  for (const tr of res.translations ?? []) {
    translations.set(
      translationKey(tr.entityType ?? '', tr.entityId ?? '', tr.locale ?? '', tr.field ?? ''),
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
    banners: (res.banners ?? []).map<MenuBannerData>((b) => ({
      id: String(b.id ?? ''),
      title: b.title ?? '',
      body: b.body ?? '',
      imageUrl: b.imageUrl ?? undefined,
      linkUrl: b.linkUrl ?? undefined,
    })),
    branchItems,
    currencies: (res.currencies ?? [])
      .filter((c) => c.isActive !== false)
      .map<MenuCurrencyOption>((c) => ({
        code: c.code ?? 'USD',
        symbol: c.symbol ?? '$',
        exchangeRate: c.exchangeRate ?? 1,
        decimalPlaces: c.decimalPlaces ?? 2,
        isDefault: Boolean(c.isDefault),
      })),
    translations,
    stampCards: (res.stampCards ?? []).map<MenuStampCardSummary>((sc) => ({
      id: String(sc.id ?? ''),
      title: sc.title ?? '',
      description: sc.description ?? '',
      stampsRequired: sc.stampsRequired ?? 10,
      rewardLabel: sc.rewardLabel ?? '',
    })),
  };
}

/**
 * Loads the published branded menu bundle when a QueryClient is mounted.
 * Gracefully degrades to undefined in isolated unit tests that render without
 * a QueryClientProvider or backend transport.
 */
export function usePublicBrandBundleSafe(
  businessId: number | null,
  locationId?: number | null,
  locale?: string
): PublicBrandBundle | undefined {
  const queryClient = useContext(QueryClientContext);
  const [bundle, setBundle] = useState<PublicBrandBundle | undefined>(undefined);

  useEffect(() => {
    if (!queryClient || businessId == null || !Number.isFinite(businessId) || businessId <= 0) {
      return;
    }
    let cancelled = false;
    void queryClient
      .fetchQuery({
        queryKey: ['public-menu-brand-bundle', businessId, locationId ?? 0, locale ?? ''],
        queryFn: () => fetchPublicBrandBundle(businessId, locationId, locale),
        staleTime: 60_000,
      })
      .then((data) => {
        if (!cancelled) setBundle(data);
      })
      .catch(() => undefined);
    return () => {
      cancelled = true;
    };
  }, [queryClient, businessId, locationId, locale]);

  return bundle;
}

export async function submitPublicTableActionSafe(params: {
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

export function trackPublicMenuEventSafe(params: {
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

export async function lookupPublicCustomerStampsSafe(
  businessId: number,
  phone: string
): Promise<CustomerStampProgressData[]> {
  const res = await menuBuilderClient.getPublicCustomerStamps({
    businessId: String(businessId),
    phone: phone.trim(),
  });
  return (res.progress ?? []).map((p) => ({
    stampCardId: String(p.stampCardId ?? ''),
    stampsCount: p.stampsCount ?? 0,
    stampsRequired: p.stampsRequired ?? 10,
    rewardsUnlocked: p.rewardsUnlocked ?? 0,
    rewardLabel: p.rewardLabel ?? '',
  }));
}

export async function resolvePublicCustomDomain(host: string): Promise<{
  found: boolean;
  businessId: number;
  businessSlug: string;
}> {
  const res = await menuBuilderClient.resolvePublicMenuHost({ host });
  return {
    found: Boolean(res.found),
    businessId: toNumber(res.businessId),
    businessSlug: res.businessSlug ?? '',
  };
}

export async function capturePublicMenuGuestSafe(params: {
  businessId: number;
  locationId?: number;
  name?: string;
  phone?: string;
  email?: string;
  notes?: string;
  consent: boolean;
}): Promise<void> {
  await menuBuilderClient.capturePublicMenuGuest({
    businessId: String(params.businessId),
    ...(params.locationId ? { locationId: String(params.locationId) } : {}),
    ...(params.name ? { name: params.name } : {}),
    ...(params.phone ? { phone: params.phone } : {}),
    ...(params.email ? { email: params.email } : {}),
    ...(params.notes ? { notes: params.notes } : {}),
    consent: params.consent,
  });
}
