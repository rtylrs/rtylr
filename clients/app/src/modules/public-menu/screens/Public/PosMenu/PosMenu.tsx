import React, { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useParams, useNavigate } from '@tanstack/react-router';
import { FullScreenLoader, HelmetContainer } from '@voxire/components';
import { FLAG_KEYS, useFeatureFlag } from '@voxire/config';
import {
  GridLayout,
  MenuCategoryData,
  BusinessInfo,
  CartItem,
  type CustomerStampProgressData,
} from '@voxire/menu';
import { usePublicMenu } from '../../../shared/hooks/usePublicMenu';
import {
  capturePublicMenuGuestSafe,
  lookupPublicCustomerStampsSafe,
  translationKey,
  usePublicBrandBundleSafe,
} from '../../../shared/hooks/usePublicBrandBundle';
import { useCartStore } from '../../../shared/stores/cartStore';
import { menuCheckoutPath } from '../../../menuPaths';

/**
 * PosMenu - public customer-facing menu for POS-backed businesses.
 *
 * Route: `/b/:businessId/menu` (numeric, e.g. `/b/42/menu`).
 *
 * Different from the legacy `/:id` flow (which routes to the old
 * mongo-backed menu). This one consumes the new `/pos/public/*`
 * anonymous endpoints shipped earlier this session. A single
 * business doesn't need categories today - the pos_menu table
 * stores flat items - so we synthesise one "All" category for the
 * GridLayout to render.
 *
 * Online ordering is gated by the VITE_ONLINE_ORDERING_ENABLED env
 * flag. When enabled the layout's built-in cart is wired to the
 * cartStore and navigates to /checkout on "Place Order".
 */

// Env kill-switch: online ordering stays enabled unless explicitly set to 'false'.
const isOnlineOrderingEnabled = (): boolean =>
  import.meta.env.VITE_ONLINE_ORDERING_ENABLED !== 'false';

const PosMenu: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation('posmenu');
  const { businessId: raw } = useParams({ strict: false });
  const navigate = useNavigate();
  const parsed = raw != null ? Number(raw) : NaN;
  const businessId = Number.isFinite(parsed) && parsed > 0 ? parsed : null;

  const platformOrderingEnabled = useFeatureFlag(FLAG_KEYS.APP_MENU_ORDERING);
  const platformCrmCaptureEnabled = useFeatureFlag(FLAG_KEYS.APP_MENU_CRM_CAPTURE);
  const platformLoyaltyEnabled = useFeatureFlag(FLAG_KEYS.APP_MENU_LOYALTY);
  const platformBannersEnabled = useFeatureFlag(FLAG_KEYS.APP_MENU_BANNERS);

  const { data, isLoading, isError } = usePublicMenu(businessId);
  const [activeLocale, setActiveLocale] = useState<string>('');
  const [activeCurrency, setActiveCurrency] = useState<string>('');
  const [customerStamps, setCustomerStamps] = useState<CustomerStampProgressData[]>([]);
  const brandBundle = usePublicBrandBundleSafe(businessId, null, activeLocale || undefined);
  const effectiveLocale = activeLocale || brandBundle?.defaultLocale || 'en';
  const activeCurrencyOption = useMemo(() => {
    const code = activeCurrency || brandBundle?.defaultCurrency;
    return (
      brandBundle?.currencies.find((c) => c.code === code) ??
      brandBundle?.currencies.find((c) => c.isDefault) ??
      brandBundle?.currencies[0]
    );
  }, [activeCurrency, brandBundle?.currencies, brandBundle?.defaultCurrency]);
  const { clearCart, setBusinessId, setCurrency, setCurrencyOption, addItem } = useCartStore();

  const categories = useMemo<MenuCategoryData[]>(() => {
    if (!data) return [];
    const overrides = brandBundle?.branchItems;
    const tr = brandBundle?.translations;
    const items = data.items
      .filter((i) => {
        const ov = overrides?.get(i.id);
        return ov ? ov.isVisible : true;
      })
      .map((i) => {
        const ov = overrides?.get(i.id);
        return {
          id: String(i.id),
          name: tr?.get(translationKey('item', i.id, effectiveLocale, 'name')) || i.name,
          price: ov?.hasPriceOverride ? ov.priceOverride : i.salePrice,
          currency: activeCurrency || brandBundle?.defaultCurrency || data.business.currency,
          description:
            tr?.get(translationKey('item', i.id, effectiveLocale, 'description')) ||
            i.description ||
            undefined,
          image: i.imageUrl,
          badges: ov?.featuredBadge ? [ov.featuredBadge] : [],
          dietaryTags: ov?.dietaryTags,
          allergens: ov?.allergens,
          calories: ov?.calories,
          available: i.isActive,
        };
      });
    if (items.length === 0) return [];
    return [
      {
        id: 'all',
        name: t('posmenu.allItems'),
        items,
      },
    ];
  }, [data, brandBundle, effectiveLocale, activeCurrency, t]);

  const business = useMemo<BusinessInfo | null>(() => {
    if (!data) return null;
    return {
      name: data.business.name,
      logo: data.business.logoUrl || undefined,
    };
  }, [data]);

  const handleCheckout = (cartItems: CartItem[]): void => {
    if (businessId == null) return;
    clearCart();
    setBusinessId(businessId);
    if (typeof setCurrencyOption === 'function' && activeCurrencyOption) {
      setCurrencyOption(activeCurrencyOption);
    } else if (data?.business.currency) {
      setCurrency(data.business.currency);
    }
    cartItems.forEach((item) => {
      addItem({
        menuItemId: Number(item.id),
        name: item.name,
        unitPrice: item.price,
        quantity: item.quantity,
      });
    });
    void navigate({ to: menuCheckoutPath() });
  };

  if (businessId == null) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center gap-4 bg-white">
        <h2 className="text-lg font-semibold text-gray-800">{t('posmenu.invalidUrlTitle')}</h2>
        <p className="text-sm text-gray-500">{t('posmenu.invalidUrlDescription')}</p>
      </div>
    );
  }

  if (isLoading) {
    return <FullScreenLoader />;
  }

  if (isError || !data || !business) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center gap-4 bg-white">
        <h2 className="text-lg font-semibold text-gray-800">{t('posmenu.loadFailedTitle')}</h2>
        <p className="text-sm text-gray-500">{t('posmenu.loadFailedDescription')}</p>
      </div>
    );
  }

  const hasItems = categories.some((cat) => cat.items && cat.items.length > 0);
  if (!hasItems) {
    return (
      <>
        <HelmetContainer title={t('posmenu.pageTitle')} />
        <div className="min-h-screen flex flex-col items-center justify-center gap-2 bg-white px-4 text-center">
          <h2 className="text-lg font-semibold text-gray-800">{t('posmenu.noItemsTitle')}</h2>
          <p className="text-sm text-gray-500">{t('posmenu.noItemsDescription')}</p>
        </div>
      </>
    );
  }

  const orderingEnabled =
    isOnlineOrderingEnabled() &&
    platformOrderingEnabled &&
    (brandBundle?.config.features?.enableOrdering ?? true);

  return (
    <>
      <HelmetContainer title={t('posmenu.pageTitle')} />
      <GridLayout
        categories={categories}
        business={business}
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
                  ...guest,
                })
            : undefined
        }
        enableCart={orderingEnabled}
        onCheckout={orderingEnabled ? handleCheckout : undefined}
      />
    </>
  );
};

export default PosMenu;
