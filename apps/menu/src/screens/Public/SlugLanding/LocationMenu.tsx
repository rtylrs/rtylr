import React, { useMemo } from 'react';
import { useParams, useNavigate } from '@tanstack/react-router';
import { FullScreenLoader, HelmetContainer } from '@voxire/components';
import { GridLayout, MenuCategoryData, BusinessInfo } from '@voxire/menu';
import { usePublicBusinessBySlug, usePublicMenu } from '../../../shared/hooks/usePublicMenu';
import { useCartStore } from '../../../shared/stores/cartStore';

/**
 * Mounted at `/:slug/:location`. Resolves the business by slug, picks
 * the matching location, then reuses the existing POS-backed menu
 * fetch keyed by business_id. (For now items live per business, not
 * per location - when POS adds location-scoped menus this hook will
 * thread `location_id` through.)
 */

const onlineOrderingEnabled = import.meta.env.VITE_ONLINE_ORDERING_ENABLED === 'true';

const LocationMenu: React.FC = (): React.JSX.Element => {
  const { slug: rawSlug, location: rawLoc } = useParams({ strict: false });
  const navigate = useNavigate();
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
  const { clearCart, setBusinessId, addItem, updateQuantity } = useCartStore();

  const categories = useMemo<MenuCategoryData[]>(() => {
    if (!menu) return [];
    const items = menu.items.map((i) => ({
      id: String(i.id),
      name: i.name,
      price: i.salePrice,
      currency: menu.business.currency,
      description: undefined,
      image: i.imageUrl,
      available: i.isActive,
    }));
    return [{ id: 'all', name: 'Menu', items }];
  }, [menu]);

  if (bizLoading || menuLoading) return <FullScreenLoader />;
  if (bizError || !biz) {
    return (
      <div style={{ padding: 32, textAlign: 'center' }}>
        <h1 style={{ fontSize: 18, fontWeight: 600 }}>Business not found</h1>
      </div>
    );
  }
  if (!location) {
    return (
      <div style={{ padding: 32, textAlign: 'center' }}>
        <h1 style={{ fontSize: 18, fontWeight: 600 }}>Location not found</h1>
        <p style={{ fontSize: 13, opacity: 0.7 }}>
          <a href={`/${slug}`}>Back to {biz.business.name}</a>
        </p>
      </div>
    );
  }
  if (menuError) {
    return (
      <div style={{ padding: 32, textAlign: 'center' }}>
        <h1 style={{ fontSize: 18, fontWeight: 600 }}>Couldn't load the menu</h1>
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
      <div style={{ padding: 48, textAlign: 'center' }}>
        <h1 style={{ fontSize: 18, fontWeight: 600, marginBottom: 8 }}>No menu items available</h1>
        <p style={{ fontSize: 14, opacity: 0.7 }}>
          This menu is currently empty. Please check back later.
        </p>
      </div>
    );
  }

  return (
    <>
      <HelmetContainer title="Menu · rtylr" />
      <GridLayout
        business={businessInfo}
        categories={categories}
        enableCart={onlineOrderingEnabled}
        onCheckout={
          onlineOrderingEnabled
            ? (cartItems) => {
                clearCart();
                setBusinessId(biz.business.id);
                cartItems.forEach((item) => {
                  const menuItemId = Number(item.id);
                  addItem({ menuItemId, name: item.name, unitPrice: item.price });
                  // addItem seeds quantity 1; carry the real line quantity over.
                  const qty = Number(item.quantity);
                  if (Number.isFinite(qty) && qty > 1) updateQuantity(menuItemId, qty);
                });
                void navigate({ to: '/checkout' });
              }
            : undefined
        }
      />
    </>
  );
};

export default LocationMenu;
