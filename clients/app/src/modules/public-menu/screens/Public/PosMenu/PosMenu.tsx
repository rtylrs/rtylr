import React, { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { useParams, useNavigate } from '@tanstack/react-router';
import { FullScreenLoader, HelmetContainer } from '@voxire/components';
import { GridLayout, MenuCategoryData, BusinessInfo, CartItem } from '@voxire/menu';
import { usePublicMenu } from '../../../shared/hooks/usePublicMenu';
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

// Opt-in: the cart is off unless VITE_ONLINE_ORDERING_ENABLED is exactly 'true'.
const isOnlineOrderingEnabled = (): boolean =>
  import.meta.env.VITE_ONLINE_ORDERING_ENABLED === 'true';

const PosMenu: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation('posmenu');
  const { businessId: raw } = useParams({ strict: false });
  const navigate = useNavigate();
  const parsed = raw != null ? Number(raw) : NaN;
  const businessId = Number.isFinite(parsed) && parsed > 0 ? parsed : null;

  const { data, isLoading, isError } = usePublicMenu(businessId);
  const { clearCart, setBusinessId, setCurrency, addItem } = useCartStore();

  // GridLayout expects a category-grouped list. POS menu items are
  // ERP products today, so we bucket everything into one "Menu" group.
  const categories = useMemo<MenuCategoryData[]>(() => {
    if (!data) return [];
    const items = data.items.map((i) => ({
      id: String(i.id),
      name: i.name,
      price: i.salePrice,
      currency: data.business.currency,
      description: undefined,
      image: i.imageUrl,
      badges: [],
      available: i.isActive,
    }));
    if (items.length === 0) return [];
    return [
      {
        id: 'all',
        name: t('posmenu.allItems'),
        items,
      },
    ];
  }, [data, t]);

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
    if (data?.business.currency) {
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

  return (
    <>
      <HelmetContainer title={t('posmenu.pageTitle')} />
      <GridLayout
        categories={categories}
        business={business}
        enableCart={isOnlineOrderingEnabled()}
        onCheckout={isOnlineOrderingEnabled() ? handleCheckout : undefined}
      />
    </>
  );
};

export default PosMenu;
