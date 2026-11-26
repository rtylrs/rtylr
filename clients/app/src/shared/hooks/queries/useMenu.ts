import { createOneKitClient } from '@voxire/api';
import i18next from 'i18next';
import { useEffect } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { MenuItemServiceClient } from '@voxire/voxie/pos/menu_item/v1/client';
import type { MenuItem as ApiMenuItem } from '@voxire/voxie/pos/menu_item/v1/types';
import type { MenuBanner } from '@voxire/voxie/hub/menu/v1/types';
import { menuClient } from '../../lib/clients';
import { useMenuStore } from '../../stores/menuStore';

const menuItemClient = createOneKitClient(MenuItemServiceClient, 'dash');

type PosMenuItem = ApiMenuItem;

interface MenuListEntry {
  _id: string;
  name: string;
  items: StoreMenuItem[];
  itemCount: number;
  isActive: boolean;
  updatedAt?: string;
}

interface StoreMenuItem {
  _id: string;
  itemId: string;
  name: string;
  type: 'item';
  [key: string]: unknown;
  details: {
    name: string;
    description: string;
    price: number;
    sku: string;
    stock: number;
    image?: string;
    category?: string;
    isActive: boolean;
  };
}

function toStoreItems(items: PosMenuItem[]): StoreMenuItem[] {
  return items.map((item) => {
    const id = String(item.id);
    return {
      _id: id,
      itemId: id,
      name: item.name ?? '',
      type: 'item',
      details: {
        name: item.name ?? '',
        description: item.description || (item.sku ? `SKU ${item.sku}` : ''),
        price: item.salePrice ?? 0,
        sku: item.sku ?? '',
        stock: item.stock ?? 0,
        image: item.imageUrl || undefined,
        category: item.category || undefined,
        isActive: item.isActive ?? true,
      },
    };
  });
}

function latestTimestamp(items: PosMenuItem[]): string | undefined {
  let latest = '';
  for (const item of items) {
    const ts = item.updatedAt || item.createdAt || '';
    if (ts && ts > latest) latest = ts;
  }
  return latest || new Date().toISOString();
}

function toMenuList(items: PosMenuItem[]): MenuListEntry[] {
  const storeItems = toStoreItems(items);

  return [
    {
      _id: 'pos-menu',
      name: i18next.t('common.posMenuName'),
      items: storeItems,
      itemCount: items.length,
      isActive: true,
      updatedAt: latestTimestamp(items),
    },
  ];
}

export const useGetAllMenus = (restaurantId: string | null) => {
  return useQuery({
    queryKey: ['menus', restaurantId],
    queryFn: async () => {
      const res = await menuItemClient.listMenuItems({});
      return toMenuList(res.items ?? []);
    },
    enabled: !!restaurantId,
  });
};

export const useGetMenu = (userId: string | null, disabled?: boolean) => {
  const { setItems, setId } = useMenuStore();

  const query = useQuery({
    queryKey: ['menu', userId, disabled],
    queryFn: async () => {
      const res = await menuItemClient.listMenuItems({});
      return {
        _id: 'pos-menu',
        items: toStoreItems(res.items ?? []),
      };
    },
    enabled: !!userId && !disabled,
  });

  useEffect(() => {
    if (query.data) {
      setItems(query.data.items);
      setId(query.data._id);
    }
  }, [query.data, setItems, setId]);

  return query;
};

export const useCreateMenuItem = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (params: { name: string; sku: string; salePrice: number }) => {
      return menuItemClient.createMenuItem({
        name: params.name,
        sku: params.sku,
        salePrice: params.salePrice,
      });
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['menus'] });
      void queryClient.invalidateQueries({ queryKey: ['menu'] });
      void queryClient.invalidateQueries({ queryKey: ['menu-builder', 'catalog-items'] });
    },
  });
};

export const useUpdateMenuItem = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (params: {
      id: string;
      name: string;
      salePrice: number;
      isActive: boolean;
      stock: number;
    }) => {
      return menuItemClient.updateMenuItem({
        id: params.id,
        body: {
          name: params.name,
          salePrice: params.salePrice,
          isActive: params.isActive,
          stock: params.stock,
        },
      });
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['menus'] });
      void queryClient.invalidateQueries({ queryKey: ['menu'] });
      void queryClient.invalidateQueries({ queryKey: ['menu-builder', 'catalog-items'] });
    },
  });
};

export const useMenuBanners = () =>
  useQuery<MenuBanner[]>({
    queryKey: ['menu-builder', 'banners'],
    queryFn: async () => {
      const res = await menuClient.listMenuBanners({});
      return res.banners ?? [];
    },
  });

export const useCreateMenuBanner = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (params: {
      title: string;
      body?: string;
      imageUrl?: string;
      linkUrl?: string;
      position?: number;
      isActive?: boolean;
    }) => {
      const res = await menuClient.createMenuBanner({
        title: params.title,
        body: params.body ?? '',
        imageUrl: params.imageUrl ?? '',
        linkUrl: params.linkUrl ?? '',
        position: params.position ?? 0,
        isActive: params.isActive ?? true,
        startsAt: '',
        endsAt: '',
      });
      return res.banner;
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['menu-builder', 'banners'] });
      void queryClient.invalidateQueries({ queryKey: ['public-menu-brand-bundle'] });
    },
  });
};

export const useDeleteMenuBanner = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      await menuClient.deleteMenuBanner({ id });
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['menu-builder', 'banners'] });
      void queryClient.invalidateQueries({ queryKey: ['public-menu-brand-bundle'] });
    },
  });
};
