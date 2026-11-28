import { Spinner, HelmetContainer } from '@voxire/components';
import { cdnImageUrl } from '@voxire/config';
import { useEffect } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { toast } from 'sonner';
import { useMenuStore } from '../../shared/stores/menuStore';
import { useUserStore } from '../../shared/stores/userStore';
import { useGetMenu, useUpdateMenuItem } from '../../shared/hooks/queries/useMenu';
import { Button } from '../../shared/components/ui';
import { toFiniteNumber } from '@voxire/api';
import { formatMoney } from '../../shared/lib/money';

interface StoreItemDetails {
  name?: string;
  description?: string;
  image?: string;
  price?: number;
  isParentCategory?: boolean;
  parentCategory?: string;
  icon?: string;
  menuItems?: StoreItem[];
  banner?: { title?: string; bannerImage?: string };
  sku?: string;
  isActive?: boolean;
}

interface StoreItem {
  _id?: string;
  itemId?: string;
  type?: string;
  details?: StoreItemDetails;
}

interface GroupedCategory {
  parent: StoreItem | null;
  children: StoreItem[];
}

const Menu: React.FC = () => {
  const navigate = useNavigate();
  const { user, userId } = useUserStore();
  const {
    items: storeItems,
    status,
    id: menuId,
    deleteCategory,
    deleteBanner,
    deleteMenuItem,
  } = useMenuStore();
  const menuQueryId =
    user?.businessId != null
      ? String(user.businessId)
      : user?.id != null
        ? String(user.id)
        : userId;
  const { data: fetchedMenu, isLoading: isMenuLoading } = useGetMenu(menuQueryId);
  const updateMenuItemMutation = useUpdateMenuItem();
  const items = fetchedMenu?.items?.length ? fetchedMenu.items : storeItems;

  const handleToggleItemActive = (item: StoreItem) => {
    const id = String(item.itemId || item._id || '');
    if (!id) return;
    const nextActive = item.details?.isActive === false;
    updateMenuItemMutation.mutate(
      {
        id,
        name: item.details?.name ?? '',
        salePrice: toFiniteNumber(item.details?.price),
        isActive: nextActive,
        stock: toFiniteNumber((item.details as { stock?: number } | undefined)?.stock),
      },
      {
        onSuccess: () => {
          toast.success(nextActive ? 'Item activated on menu' : 'Item hidden from menu');
        },
        onError: () => {
          toast.error('Failed to update menu item');
        },
      }
    );
  };

  useEffect(() => {
    if (menuId) {
      const url = new URL(window.location.href);
      if (!url.searchParams.has('menuId')) {
        url.searchParams.append('menuId', menuId as string);
        window.history.replaceState({}, '', url.toString());
      }
    }
  }, [menuId]);

  const renderItems = (items: StoreItem[]) => {
    return items.map((item) => {
      const imageUrl = cdnImageUrl(item?.details?.image);
      return (
        <div
          key={item._id}
          className="flex items-center justify-between p-3 border rounded-lg"
          style={{ background: 'var(--dash-card)', borderColor: 'var(--dash-border)' }}
        >
          <div className="flex items-center gap-3">
            {imageUrl && (
              <img
                src={imageUrl}
                alt={item?.details?.name ?? ''}
                className="w-12 h-12 rounded-lg object-cover"
              />
            )}
            <div>
              <p className="text-sm font-medium text-primary">{item?.details?.name}</p>
              {item?.details?.description && (
                <p className="text-xs text-muted line-clamp-1">{item.details.description}</p>
              )}
            </div>
          </div>
          <div className="flex items-center gap-2">
            {item?.details?.price !== undefined && (
              <span className="text-sm font-semibold text-primary">
                {formatMoney(toFiniteNumber(item.details.price))}
              </span>
            )}
            <button
              type="button"
              className="text-xs text-blue-500 hover:underline px-2 py-1"
              onClick={() => {
                navigate({
                  to: '/create-item',
                  search: { itemId: (item.itemId || item._id) as string },
                });
              }}
            >
              Edit
            </button>
            <button
              type="button"
              className="text-xs text-red-500 hover:underline px-2 py-1"
              onClick={() => deleteMenuItem(item._id as string)}
            >
              Delete
            </button>
          </div>
        </div>
      );
    });
  };

  const renderFlatItem = (item: StoreItem) => (
    <div
      key={item._id}
      className="flex items-center justify-between py-2.5 px-3 rounded-lg hover:bg-[var(--dash-hover)] transition-colors"
      style={{ borderBottom: '1px solid var(--dash-border)' }}
    >
      <div className="min-w-0">
        <p className="text-sm font-semibold text-primary truncate">{item.details?.name}</p>
        {item.details?.sku && <p className="text-xs text-muted mt-1">SKU {item.details.sku}</p>}
      </div>
      <div className="flex items-center gap-3">
        <span className="text-sm font-semibold text-primary">
          {formatMoney(toFiniteNumber(item.details?.price))}
        </span>
        <span
          className="px-2 py-1 rounded-full text-[11px] font-semibold"
          style={{
            color: item.details?.isActive === false ? '#8a6a00' : '#047857',
            background: item.details?.isActive === false ? '#fef3c7' : '#d1fae5',
          }}
        >
          {item.details?.isActive === false ? 'Inactive' : 'Active'}
        </span>
        <button
          type="button"
          className="text-xs text-blue-500 hover:underline px-2 py-1"
          onClick={() => {
            navigate({
              to: '/create-item',
              search: { itemId: (item.itemId || item._id) as string },
            });
          }}
        >
          Edit
        </button>
        <button
          type="button"
          className="text-xs hover:underline px-2 py-1"
          style={{ color: item.details?.isActive === false ? '#047857' : '#b45309' }}
          disabled={updateMenuItemMutation.isPending}
          onClick={() => handleToggleItemActive(item)}
        >
          {item.details?.isActive === false ? 'Activate' : 'Hide'}
        </button>
      </div>
    </div>
  );

  const renderCategory = (category: StoreItem) => {
    return (
      <div
        key={category._id}
        className="ps-4 py-2 my-2 border-s-2"
        style={{ borderColor: 'var(--dash-border)' }}
      >
        <div className="flex items-center justify-between mb-2">
          <div className="flex items-center gap-2">
            <span className="font-semibold text-primary">{category?.details?.name}</span>
            {cdnImageUrl(category?.details?.icon) && (
              <img
                src={cdnImageUrl(category?.details?.icon)}
                alt={category?.details?.name}
                className="w-5 h-5 object-cover"
              />
            )}
          </div>
          <div className="flex gap-2">
            <button
              type="button"
              className="px-3 py-1 text-xs bg-[var(--dash-active)] text-[var(--dash-text)] rounded-lg hover:opacity-90 transition-opacity"
              onClick={() => {
                navigate({ to: '/create-item', search: { categoryId: category.itemId } });
              }}
            >
              Add a product
            </button>
            <button
              type="button"
              className="px-3 py-1 text-xs text-red-500 border border-red-200 rounded-lg hover:bg-red-50 transition-colors"
              onClick={() => deleteCategory((category.itemId || category._id) as string)}
            >
              Delete
            </button>
          </div>
        </div>
        <div className="space-y-1">{renderItems(category.details?.menuItems ?? [])}</div>
      </div>
    );
  };

  const renderParentCategory = (parentCategory: StoreItem, childCategories: StoreItem[]) => {
    return (
      <div
        key={parentCategory._id}
        className="border rounded-xl p-4 mb-4"
        style={{ background: 'var(--dash-card)', borderColor: 'var(--dash-border)' }}
      >
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center gap-2">
            <span className="font-bold text-primary">{parentCategory?.details?.name}</span>
            {cdnImageUrl(parentCategory?.details?.icon) && (
              <img
                src={cdnImageUrl(parentCategory?.details?.icon)}
                alt={parentCategory?.details?.name}
                className="w-5 h-5 object-cover"
              />
            )}
          </div>
          <div className="flex gap-2">
            <button
              type="button"
              className="px-3 py-1.5 text-sm text-red-500 border border-red-200 rounded-lg hover:bg-red-50 transition-colors"
              onClick={() =>
                deleteCategory((parentCategory.itemId || parentCategory._id) as string)
              }
            >
              Delete
            </button>
          </div>
        </div>
        <div className="pl-4 space-y-2">{childCategories.map(renderCategory)}</div>
      </div>
    );
  };

  const groupedCategories = (items as StoreItem[]).reduce(
    (acc: Record<string, GroupedCategory | StoreItem>, item) => {
      if (item.type === 'category') {
        if (item.details?.isParentCategory) {
          acc[item.itemId ?? ''] = { parent: item, children: [] };
        } else if (item.details?.parentCategory) {
          if (!acc[item.details.parentCategory]) {
            acc[item.details.parentCategory] = { parent: null, children: [] };
          }
          (acc[item.details.parentCategory] as GroupedCategory).children.push(item);
        } else {
          acc[item.itemId ?? ''] = { parent: item, children: [] };
        }
      } else if (item.type === 'banner') {
        acc[item.itemId ?? ''] = item;
      }
      return acc;
    },
    {}
  );

  const renderSections = (items: StoreItem[]) => {
    return items.map((item) => {
      if (item.type === 'banner') {
        return (
          <div
            key={item._id}
            className="border rounded-xl p-4 mb-4"
            style={{ background: 'var(--dash-card)', borderColor: 'var(--dash-border)' }}
          >
            <div className="flex items-center justify-between mb-2">
              <p className="font-semibold text-primary">{item?.details?.banner?.title}</p>
              <div>
                <button
                  type="button"
                  className="px-3 py-1.5 text-sm text-red-500 border border-red-200 rounded-lg hover:bg-red-50 transition-colors"
                  onClick={() => deleteBanner(item._id as string)}
                >
                  Delete
                </button>
              </div>
            </div>
            {cdnImageUrl(item?.details?.banner?.bannerImage) && (
              <img
                style={{ width: '50%', height: '50%', objectFit: 'cover' }}
                src={cdnImageUrl(item?.details?.banner?.bannerImage)}
                alt={item?.details?.banner?.title}
              />
            )}
          </div>
        );
      } else if (item.type === 'category') {
        if (item.details?.isParentCategory) {
          const group = groupedCategories[item.itemId ?? ''] as GroupedCategory | undefined;
          if (!group) return null;
          const childCategories = group.children;
          return renderParentCategory(item, childCategories);
        } else if (item.details?.parentCategory) {
          return null;
        } else {
          return renderCategory(item);
        }
      }
      if (item.type === 'item') {
        return renderFlatItem(item);
      }
      return null;
    });
  };

  if (status === 'loading' || isMenuLoading) {
    return (
      <div className="flex justify-center items-center h-[80vh]">
        <Spinner />
      </div>
    );
  }

  return (
    <>
      <HelmetContainer title="Menu · rtylr" />
      <div className="space-y-6">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h1 className="text-2xl font-bold text-primary">Menu</h1>
          <div className="flex flex-wrap items-center gap-2">
            <Button size="sm" variant="outline" onClick={() => navigate({ to: '/menus/studio' })}>
              Menu Studio
            </Button>
            <Button size="sm" variant="outline" onClick={() => navigate({ to: '/menus/qr-codes' })}>
              QR &amp; Links
            </Button>
            <Button size="sm" onClick={() => navigate({ to: '/create-item' })}>
              Add Item
            </Button>
          </div>
        </div>
        {items.length > 0 ? (
          <div className="space-y-2">{renderSections(items as StoreItem[])}</div>
        ) : (
          <div
            className="rounded-xl border p-8 text-center text-sm space-y-3"
            style={{
              background: 'var(--dash-card)',
              borderColor: 'var(--dash-border)',
              color: 'var(--dash-text-muted)',
            }}
          >
            <p>No POS menu items yet.</p>
            <Button size="sm" onClick={() => navigate({ to: '/create-item' })}>
              Add your first item
            </Button>
          </div>
        )}
      </div>
    </>
  );
};

export default Menu;
