import { createOneKitClient } from '@voxire/api';
import { useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { MenuItemServiceClient } from '@voxire/voxie/pos/menu_item/v1/service_client';
import type { MenuItem as ApiMenuItem } from '@voxire/voxie/pos/menu_item/v1/service_client';
import { useMenuStore } from '../../stores/menuStore';

const menuItemClient = createOneKitClient(MenuItemServiceClient, 'dash');

type PosMenuItem = ApiMenuItem;

interface MenuListEntry {
  _id: string;
  name: string;
  items: StoreMenuItem[];
  itemCount: number;
  isActive: boolean;
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
    isActive: boolean;
  };
}

function toStoreItems(items: PosMenuItem[]): StoreMenuItem[] {
  return items.map((item) => {
    const id = String(item.id);
    return {
      _id: id,
      itemId: id,
      name: item.name,
      type: 'item',
      details: {
        name: item.name,
        description: item.sku ? `SKU ${item.sku}` : '',
        price: item.salePrice,
        sku: item.sku,
        isActive: item.isActive,
      },
    };
  });
}

function toMenuList(items: PosMenuItem[]): MenuListEntry[] {
  const storeItems = toStoreItems(items);

  return [
    {
      _id: 'pos-menu',
      name: 'POS Menu',
      items: storeItems,
      itemCount: items.length,
      isActive: true,
    },
  ];
}

export const useGetAllMenus = (restaurantId: string | null) => {
  const { setAllMenus } = useMenuStore();

  const query = useQuery({
    queryKey: ['menus', restaurantId],
    queryFn: async () => {
      const res = await menuItemClient.listMenuItems({});
      return toMenuList(res.items);
    },
    enabled: !!restaurantId,
  });

  useEffect(() => {
    if (query.data) {
      setAllMenus(query.data);
    }
  }, [query.data, setAllMenus]);

  return query;
};

export const useGetMenu = (userId: string | null, disabled?: boolean) => {
  const { setItems, setId } = useMenuStore();

  const query = useQuery({
    queryKey: ['menu', userId, disabled],
    queryFn: async () => {
      const res = await menuItemClient.listMenuItems({});
      return {
        _id: 'pos-menu',
        items: toStoreItems(res.items),
      };
    },
    enabled: !!userId,
  });

  useEffect(() => {
    if (query.data) {
      setItems(query.data.items);
      setId(query.data._id);
    }
  }, [query.data, setItems, setId]);

  return query;
};

export const useGetMenuById = (menuId: string | null) => {
  return useGetMenu(menuId);
};
