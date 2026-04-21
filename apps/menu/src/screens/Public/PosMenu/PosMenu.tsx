import React, { useMemo } from 'react';
import { useParams } from 'react-router-dom';
import { FullScreenLoader } from '@voxire/components';
import { GridLayout, MenuCategoryData, BusinessInfo } from '@voxire/menu';
import { usePublicMenu } from '../../../shared/hooks/usePublicMenu';

/**
 * PosMenu — public customer-facing menu for POS-backed businesses.
 *
 * Route: `/b/:businessId` (numeric, e.g. `/b/42`).
 *
 * Different from the legacy `/:id` flow (which routes to the old
 * mongo-backed menu). This one consumes the new `/pos/public/*`
 * anonymous endpoints shipped earlier this session. A single
 * business doesn't need categories today — the pos_menu table
 * stores flat items — so we synthesise one "All" category for the
 * GridLayout to render.
 */

const PosMenu: React.FC = (): React.JSX.Element => {
  const { businessId: raw } = useParams<{ businessId: string }>();
  const parsed = raw != null ? Number(raw) : NaN;
  const businessId = Number.isFinite(parsed) && parsed > 0 ? parsed : null;

  const { data, isLoading, isError } = usePublicMenu(businessId);

  // GridLayout expects a category-grouped list. pos_menu_items is
  // flat today, so we bucket everything into one "Menu" group and
  // let the layout render it. When POS adds a category column we
  // can fan this out.
  const categories = useMemo<MenuCategoryData[]>(() => {
    if (!data) return [];
    const items = data.items
      .filter((i) => i.isActive)
      .map((i) => ({
        id: String(i.id),
        name: i.name,
        price: i.salePrice,
        currency: data.business.currency,
        description: undefined,
        image: undefined,
        badges: [],
        available: i.isActive,
      }));
    if (items.length === 0) return [];
    return [
      {
        id: 'all',
        name: 'Menu',
        items,
      },
    ];
  }, [data]);

  const business = useMemo<BusinessInfo | null>(() => {
    if (!data) return null;
    return {
      name: data.business.name,
      logo: data.business.logoUrl || undefined,
    };
  }, [data]);

  if (businessId == null) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center gap-4 bg-white">
        <h2 className="text-lg font-semibold text-gray-800">Invalid menu URL</h2>
        <p className="text-sm text-gray-500">This link is missing a business ID.</p>
      </div>
    );
  }

  if (isLoading) {
    return <FullScreenLoader />;
  }

  if (isError || !data || !business) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center gap-4 bg-white">
        <h2 className="text-lg font-semibold text-gray-800">Failed to load menu</h2>
        <p className="text-sm text-gray-500">Please try refreshing the page.</p>
      </div>
    );
  }

  return <GridLayout categories={categories} business={business} />;
};

export default PosMenu;
