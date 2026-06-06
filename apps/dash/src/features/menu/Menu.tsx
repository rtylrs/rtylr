import { Spinner } from '@voxire/components';
import { useEffect, useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import CreateCategoryModal from './Components/CreateCategoryModal';
import { useMenuStore } from '../../shared/stores/menuStore';
import { useUserStore } from '../../shared/stores/userStore';
import { useGetMenu } from '../../shared/hooks/queries/useMenu';
import CreateParentCategoryModal from './Components/CreateParentCategoryModal';

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

const Menu: React.FC<{}> = () => {
  const navigate = useNavigate();
  const { userId } = useUserStore();
  const { items, status, id: menuId, deleteCategory } = useMenuStore();
  useGetMenu(userId as string);

  const [isCreateCategoryModalOpen, setCreateCategoryModalOpen] = useState(false);
  const [isCreateParentCategoryModalOpen, setCreateParentCategoryModalOpen] = useState(false);

  useEffect(() => {
    const url = new URL(window.location.href);
    if (!url.searchParams.has('menuId')) {
      url.searchParams.append('menuId', menuId as string);
      window.history.replaceState({}, '', url.toString());
    }
  }, [menuId]);

  const handleDeleteCategory = (categoryId: string) => {
    deleteCategory(categoryId);
  };

  const renderItems = (items: StoreItem[]) => {
    return items.map((item) => (
      <div
        key={item._id}
        className="flex items-center justify-between p-3 border border-black/10 rounded-lg bg-white"
      >
        <div className="flex items-center gap-3">
          {item?.details?.image && (
            <img
              src={item.details.image}
              alt={item.details.name}
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
        {item?.details?.price !== undefined && (
          <span className="text-sm font-semibold text-primary">
            ${(item.details.price / 100).toFixed(2)}
          </span>
        )}
      </div>
    ));
  };

  const renderCategory = (category: StoreItem) => {
    return (
      <div key={category._id} className="border border-black/10 rounded-xl p-4 bg-white mb-3">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center gap-2">
            <span className="font-semibold text-primary">{category?.details?.name}</span>
            {category?.details?.icon && (
              <img
                src={category?.details?.icon}
                alt={category?.details?.name}
                className="w-5 h-5 object-cover"
              />
            )}
          </div>
          <div className="flex gap-2">
            <button
              className="px-3 py-1.5 text-sm bg-accent text-white rounded-lg hover:opacity-90 transition-opacity"
              onClick={() => {
                navigate({ to: '/create-item', search: { categoryId: category.itemId } });
              }}
            >
              Add a product
            </button>
            <button
              className="px-3 py-1.5 text-sm text-red-500 border border-red-200 rounded-lg hover:bg-red-50 transition-colors"
              onClick={() => handleDeleteCategory(category.itemId ?? '')}
            >
              Delete
            </button>
          </div>
        </div>
        <div className="space-y-2">{renderItems(category.details?.menuItems ?? [])}</div>
      </div>
    );
  };

  const renderParentCategory = (parentCategory: StoreItem, childCategories: StoreItem[]) => {
    return (
      <div key={parentCategory._id} className="border border-black/10 rounded-xl p-4 bg-white mb-4">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center gap-2">
            <span className="font-bold text-primary">{parentCategory?.details?.name}</span>
            {parentCategory?.details?.icon && (
              <img
                src={parentCategory?.details?.icon}
                alt={parentCategory?.details?.name}
                className="w-5 h-5 object-cover"
              />
            )}
          </div>
          <div className="flex gap-2">
            <button
              className="px-3 py-1.5 text-sm text-red-500 border border-red-200 rounded-lg hover:bg-red-50 transition-colors"
              onClick={() => handleDeleteCategory(parentCategory.itemId ?? '')}
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
          <div key={item._id} className="border border-black/10 rounded-xl p-4 bg-white mb-4">
            <div className="flex items-center justify-between mb-2">
              <p className="font-semibold text-primary">{item?.details?.banner?.title}</p>
              <div>
                <button
                  className="px-3 py-1.5 text-sm text-red-500 border border-red-200 rounded-lg hover:bg-red-50 transition-colors"
                  onClick={() => setCreateCategoryModalOpen(true)}
                >
                  Delete
                </button>
              </div>
            </div>
            <img
              style={{ width: '50%', height: '50%', objectFit: 'cover' }}
              src={item?.details?.banner?.bannerImage}
              alt={item?.details?.banner?.title}
            />
          </div>
        );
      } else if (item.type === 'category') {
        if (item.details?.isParentCategory) {
          const childCategories = (groupedCategories[item.itemId ?? ''] as GroupedCategory)
            .children;
          return renderParentCategory(item, childCategories);
        } else if (item.details?.parentCategory) {
          return null;
        } else {
          return renderCategory(item);
        }
      }
      return null;
    });
  };

  if (status === 'loading') {
    return (
      <div className="flex justify-center items-center h-[80vh]">
        <Spinner />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-primary">Menu</h1>
        <div className="flex gap-2">
          <button
            onClick={() => setCreateParentCategoryModalOpen(true)}
            className="px-4 py-2 text-sm font-medium bg-white border border-black/10 rounded-lg hover:bg-black/5 transition-colors"
          >
            Create Group
          </button>
          <button
            onClick={() => setCreateCategoryModalOpen(true)}
            className="px-4 py-2 text-sm font-medium bg-white border border-black/10 rounded-lg hover:bg-black/5 transition-colors"
          >
            Create Category
          </button>
          <button
            onClick={() => navigate({ to: '/create-banner' })}
            className="px-4 py-2 text-sm font-medium bg-accent text-white rounded-lg hover:opacity-90 transition-opacity"
          >
            Create Banner
          </button>
        </div>
      </div>
      <div>{renderSections(items as StoreItem[])}</div>
      <CreateCategoryModal
        isOpen={isCreateCategoryModalOpen}
        onClose={() => setCreateCategoryModalOpen(false)}
        addOrEdit="add"
      />
      <CreateParentCategoryModal
        isOpen={isCreateParentCategoryModalOpen}
        onClose={() => setCreateParentCategoryModalOpen(false)}
        addOrEdit="add"
      />
    </div>
  );
};

export default Menu;
