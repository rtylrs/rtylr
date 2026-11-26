import React, { useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useParams, useNavigate, useRouter } from '@tanstack/react-router';
import { FullScreenLoader, HelmetContainer } from '@voxire/components';
import { FLAG_KEYS, useFeatureFlag } from '@voxire/config';
import {
  GridLayout,
  MenuCategoryData,
  BusinessInfo,
  type CustomerStampProgressData,
  type FulfillmentMode,
} from '@voxire/menu';
import { EmptyState } from '@voxire/ui';
import { SearchX, MapPinOff, UtensilsCrossed } from 'lucide-react';
import { usePublicBusinessBySlug, usePublicMenu } from '../../../shared/hooks/usePublicMenu';
import {
  capturePublicMenuGuestSafe,
  lookupPublicCustomerStampsSafe,
  submitPublicTableActionSafe,
  trackPublicMenuEventSafe,
  translationKey,
  usePublicBrandBundleSafe,
} from '../../../shared/hooks/usePublicBrandBundle';
import { useCartStore } from '../../../shared/stores/cartStore';
import { menuCheckoutPath, menuHomePath } from '../../../menuPaths';

/**
 * Mounted at `/$slug/menu/$location`. Resolves the business by slug, picks
 * the matching location, then reuses the existing POS-backed menu
 * fetch keyed by business_id. (For now items live per business, not
 * per location - when POS adds location-scoped menus this hook will
 * thread `location_id` through.)
 */

// Opt-in: the cart is off unless VITE_ONLINE_ORDERING_ENABLED is exactly 'true'.
const isOnlineOrderingEnabled = (): boolean =>
  import.meta.env.VITE_ONLINE_ORDERING_ENABLED === 'true';

const LocationMenu: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation('locationmenu');
  const { slug: rawSlug, location: rawLoc } = useParams({ strict: false });
  const navigate = useNavigate();
  const router = useRouter();
  const slug = (rawSlug ?? '').toLowerCase();
  const locationSlug = (rawLoc ?? '').toLowerCase();

  const { data: biz, isLoading: bizLoading, isError: bizError } = usePublicBusinessBySlug(slug);
  const businessId = biz?.business.id ?? null;
  const location = biz?.locations.find((l) => l.slug === locationSlug);
  const {
    data: menu,
    isLoading: menuLoading,
    isError: menuError,
  } = usePublicMenu(businessId, location?.id ?? null);
  const {
    clearCart,
    setBusinessId,
    setLocationId,
    setTable,
    setAvailableFulfillmentModes,
    setCurrency,
    setCurrencyOption,
    addItem,
    updateQuantity,
    businessId: cartBusinessId,
  } = useCartStore();

  const platformOrderingEnabled = useFeatureFlag(FLAG_KEYS.APP_MENU_ORDERING);
  const platformCrmCaptureEnabled = useFeatureFlag(FLAG_KEYS.APP_MENU_CRM_CAPTURE);
  const platformTableActionsEnabled = useFeatureFlag(FLAG_KEYS.APP_MENU_TABLE_ACTIONS);
  const platformLoyaltyEnabled = useFeatureFlag(FLAG_KEYS.APP_MENU_LOYALTY);
  const platformBannersEnabled = useFeatureFlag(FLAG_KEYS.APP_MENU_BANNERS);

  const [activeLocale, setActiveLocale] = useState<string>('');
  const [activeCurrency, setActiveCurrency] = useState<string>('');
  const [customerStamps, setCustomerStamps] = useState<CustomerStampProgressData[]>([]);
  const brandBundle = usePublicBrandBundleSafe(
    businessId,
    location?.id ?? null,
    activeLocale || undefined
  );
  const activeCurrencyOption = useMemo(() => {
    const code = activeCurrency || brandBundle?.defaultCurrency;
    return (
      brandBundle?.currencies.find((c) => c.code === code) ??
      brandBundle?.currencies.find((c) => c.isDefault) ??
      brandBundle?.currencies[0]
    );
  }, [activeCurrency, brandBundle?.currencies, brandBundle?.defaultCurrency]);

  const tableContext = useMemo(() => {
    if (typeof window === 'undefined') return { tableId: undefined, tableLabel: undefined };
    const params = new URLSearchParams(window.location.search);
    const rawTable = params.get('table');
    const parsedTable = rawTable ? Number(rawTable) : NaN;
    const tableId = Number.isFinite(parsedTable) && parsedTable > 0 ? parsedTable : undefined;
    const tableLabel = params.get('tableLabel') || (tableId ? `Table ${tableId}` : undefined);
    return { tableId, tableLabel };
  }, []);

  const effectiveLocale = activeLocale || brandBundle?.defaultLocale || 'en';

  useEffect(() => {
    if (!businessId || !location?.id) return;
    trackPublicMenuEventSafe({
      businessId,
      locationId: location.id,
      tableId: tableContext.tableId,
      eventType: tableContext.tableId ? 'qr_scan' : 'menu_view',
      locale: effectiveLocale,
      currency: activeCurrency || brandBundle?.defaultCurrency || menu?.business.currency,
    });
  }, [businessId, location?.id, tableContext.tableId]);

  React.useEffect(() => {
    if (businessId) {
      // A persisted cart from another business must not be re-labelled: its
      // items and prices belong to that menu, while /checkout submits the
      // store's items under whatever businessId is set here.
      if (cartBusinessId !== 0 && cartBusinessId !== businessId) clearCart();
      setBusinessId(businessId);
    }
    if (location?.id) setLocationId(location.id);
    if (typeof setTable === 'function' && tableContext.tableId) {
      setTable(tableContext.tableId, tableContext.tableLabel);
    }
    if (
      typeof setAvailableFulfillmentModes === 'function' &&
      brandBundle?.config.features?.fulfillmentModes
    ) {
      const fm = brandBundle.config.features.fulfillmentModes;
      const enabled: FulfillmentMode[] = [];
      if (fm.dineIn) enabled.push('dine_in');
      if (fm.selfService) enabled.push('self_service');
      if (fm.pickup) enabled.push('pickup');
      if (fm.delivery) enabled.push('delivery');
      if (enabled.length > 0) setAvailableFulfillmentModes(enabled);
    }
    const resolvedCurrency =
      activeCurrency || brandBundle?.defaultCurrency || menu?.business.currency;
    if (resolvedCurrency) setCurrency(resolvedCurrency);
    if (typeof setCurrencyOption === 'function') setCurrencyOption(activeCurrencyOption);
  }, [
    businessId,
    cartBusinessId,
    location?.id,
    activeCurrency,
    activeCurrencyOption,
    brandBundle?.defaultCurrency,
    brandBundle?.config.features?.fulfillmentModes,
    menu?.business.currency,
    tableContext.tableId,
    tableContext.tableLabel,
    clearCart,
    setBusinessId,
    setLocationId,
    setTable,
    setAvailableFulfillmentModes,
    setCurrency,
    setCurrencyOption,
  ]);

  const categories = useMemo<MenuCategoryData[]>(() => {
    if (!menu) return [];
    const overrides = brandBundle?.branchItems;
    const tr = brandBundle?.translations;

    const visibleItems = menu.items
      .filter((i) => {
        const ov = overrides?.get(i.id);
        return ov ? ov.isVisible : true;
      })
      .sort((a, b) => {
        const orderA = overrides?.get(a.id)?.displayOrder ?? 0;
        const orderB = overrides?.get(b.id)?.displayOrder ?? 0;
        return orderA - orderB;
      });

    const grouped = new Map<string, MenuCategoryData['items']>();
    for (const i of visibleItems) {
      const ov = overrides?.get(i.id);
      const localizedName =
        tr?.get(translationKey('item', i.id, effectiveLocale, 'name')) || i.name;
      const localizedDesc =
        tr?.get(translationKey('item', i.id, effectiveLocale, 'description')) ||
        i.description ||
        undefined;
      const catName = i.category?.trim() || t('locationmenu.allItems');
      const list = grouped.get(catName) ?? [];
      list.push({
        id: String(i.id),
        name: localizedName,
        price: ov?.hasPriceOverride ? ov.priceOverride : i.salePrice,
        currency: activeCurrency || brandBundle?.defaultCurrency || menu.business.currency,
        description: localizedDesc,
        image: i.imageUrl,
        badges: ov?.featuredBadge ? [ov.featuredBadge] : undefined,
        dietaryTags: ov?.dietaryTags,
        allergens: ov?.allergens,
        calories: ov?.calories,
        available: i.isActive,
      });
      grouped.set(catName, list);
    }

    if (grouped.size <= 1) {
      const onlyItems = Array.from(grouped.values())[0] ?? [];
      return [{ id: 'all', name: t('locationmenu.allItems'), items: onlyItems }];
    }
    return Array.from(grouped.entries()).map(([catName, items], idx) => ({
      id: `cat-${idx}`,
      name: catName,
      items,
    }));
  }, [menu, brandBundle, effectiveLocale, activeCurrency, t]);

  if (bizLoading || menuLoading) return <FullScreenLoader />;
  if (bizError || !biz) {
    return (
      <div
        style={{
          minHeight: '100vh',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          padding: '0 20px',
        }}
      >
        <EmptyState
          icon={SearchX}
          title={t('locationmenu.businessNotFoundTitle')}
          description={t('locationmenu.businessNotFoundDescription')}
          action={{ label: t('locationmenu.goBack'), onClick: () => router.history.back() }}
        />
      </div>
    );
  }
  if (!location) {
    return (
      <div
        style={{
          minHeight: '100vh',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          padding: '0 20px',
        }}
      >
        <EmptyState
          icon={MapPinOff}
          title={t('locationmenu.locationNotFoundTitle')}
          description={t('locationmenu.locationNotFoundDescription')}
          action={{
            label: t('locationmenu.backToBusiness', { businessName: biz.business.name }),
            onClick: () => void navigate({ to: menuHomePath(slug) }),
          }}
        />
      </div>
    );
  }
  if (menuError) {
    return (
      <div
        style={{
          minHeight: '100vh',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          padding: '0 20px',
        }}
      >
        <EmptyState
          icon={UtensilsCrossed}
          title={t('locationmenu.loadFailedTitle')}
          description={t('locationmenu.loadFailedDescription')}
          action={{
            label: t('locationmenu.backToBusiness', { businessName: biz.business.name }),
            onClick: () => void navigate({ to: menuHomePath(slug) }),
          }}
        />
      </div>
    );
  }

  const businessInfo: BusinessInfo = {
    name: `${biz.business.name} - ${location.name}`,
    logo: biz.business.logoUrl || undefined,
  };

  const hasItems = categories.some((cat) => cat.items && cat.items.length > 0);
  if (!hasItems) {
    return (
      <div
        style={{
          minHeight: '100vh',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          padding: '0 20px',
        }}
      >
        <EmptyState
          icon={UtensilsCrossed}
          title={t('locationmenu.noItemsTitle')}
          description={t('locationmenu.noItemsDescription')}
          action={{
            label: t('locationmenu.backToBusiness', { businessName: biz.business.name }),
            onClick: () => void navigate({ to: menuHomePath(slug) }),
          }}
        />
      </div>
    );
  }

  const orderingEnabled =
    isOnlineOrderingEnabled() &&
    platformOrderingEnabled &&
    (brandBundle?.config.features?.enableOrdering ?? true);

  return (
    <>
      <HelmetContainer title={t('locationmenu.pageTitle')} />
      <GridLayout
        business={businessInfo}
        categories={categories}
        config={brandBundle?.config}
        banners={platformBannersEnabled ? brandBundle?.banners : []}
        currencies={brandBundle?.currencies}
        activeCurrencyCode={activeCurrency || brandBundle?.defaultCurrency}
        onCurrencyChange={(code) => setActiveCurrency(code)}
        availableLocales={brandBundle?.enabledLocales}
        activeLocale={effectiveLocale}
        onLocaleChange={(loc) => setActiveLocale(loc)}
        stampCards={platformLoyaltyEnabled ? brandBundle?.stampCards : []}
        customerStamps={customerStamps}
        onLookupStamps={
          businessId && platformLoyaltyEnabled
            ? async (phone) => {
                const progress = await lookupPublicCustomerStampsSafe(businessId, phone);
                setCustomerStamps(progress);
              }
            : undefined
        }
        onCaptureCrmGuest={
          businessId && platformCrmCaptureEnabled
            ? (guest) =>
                capturePublicMenuGuestSafe({
                  businessId,
                  locationId: location.id,
                  ...guest,
                })
            : undefined
        }
        tableLabel={tableContext.tableLabel}
        onTableAction={
          tableContext.tableId && businessId && platformTableActionsEnabled
            ? (actionType) =>
                submitPublicTableActionSafe({
                  businessId,
                  locationId: location.id,
                  tableId: tableContext.tableId,
                  tableLabel: tableContext.tableLabel,
                  actionType,
                })
            : undefined
        }
        onAddItem={(item) => {
          if (businessId) {
            trackPublicMenuEventSafe({
              businessId,
              locationId: location.id,
              tableId: tableContext.tableId,
              eventType: 'add_to_cart',
              itemId: Number(item.id),
              locale: effectiveLocale,
              currency: activeCurrency || brandBundle?.defaultCurrency,
            });
          }
        }}
        enableCart={orderingEnabled}
        onCheckout={
          orderingEnabled
            ? (cartItems) => {
                clearCart();
                setBusinessId(biz.business.id);
                if (location?.id) setLocationId(location.id);
                if (typeof setTable === 'function' && tableContext.tableId) {
                  setTable(tableContext.tableId, tableContext.tableLabel);
                }
                if (typeof setCurrencyOption === 'function' && activeCurrencyOption) {
                  setCurrencyOption(activeCurrencyOption);
                }
                cartItems.forEach((item) => {
                  const menuItemId = Number(item.id);
                  addItem({ menuItemId, name: item.name, unitPrice: item.price });
                  // addItem seeds quantity 1; carry the real line quantity over.
                  const qty = Number(item.quantity);
                  if (Number.isFinite(qty) && qty > 1) updateQuantity(menuItemId, qty);
                });
                // Carry the slug through checkout so the order-status page's
                // "order another" CTA returns to this business's menu instead
                // of the Portal root.
                void navigate({
                  to: `${menuCheckoutPath()}?slug=${encodeURIComponent(slug)}`,
                });
              }
            : undefined
        }
      />
    </>
  );
};

export default LocationMenu;
