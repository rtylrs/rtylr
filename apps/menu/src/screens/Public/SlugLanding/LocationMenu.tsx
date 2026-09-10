import React, { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { useParams, useNavigate, useRouter } from '@tanstack/react-router';
import { FullScreenLoader, HelmetContainer } from '@voxire/components';
import { GridLayout, MenuCategoryData, BusinessInfo } from '@voxire/menu';
import { EmptyState } from '@voxire/ui';
import { SearchX, MapPinOff, UtensilsCrossed } from 'lucide-react';
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
  const { clearCart, setBusinessId, setLocationId, addItem, updateQuantity } = useCartStore();

  React.useEffect(() => {
    if (businessId) setBusinessId(businessId);
    if (location?.id) setLocationId(location.id);
  }, [businessId, location?.id, setBusinessId, setLocationId]);

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
            onClick: () => void navigate({ to: '/$slug', params: { slug } }),
          }}
        />
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
            onClick: () => void navigate({ to: '/$slug', params: { slug } }),
          }}
        />
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
