import { unwrap } from '@voxire/api';
import { useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { safe } from '../../lib/api';
import { useMenuStore } from '../../stores/menuStore';

interface PosMenuItem {
  id: number | string;
  name: string;
  sku?: string;
  salePrice?: number;
  sale_price?: number;
  isActive?: boolean;
  is_active?: boolean;
}

interface PosMenuItemsResponse {
  items?: PosMenuItem[];
}

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

function itemPrice(item: PosMenuItem): number {
  return item.salePrice ?? item.sale_price ?? 0;
}

function itemActive(item: PosMenuItem): boolean {
  return item.isActive ?? item.is_active ?? true;
}

function normalizeItems(response: PosMenuItemsResponse | PosMenuItem[]): PosMenuItem[] {
  if (Array.isArray(response)) return response;
  return response.items ?? [];
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
        price: itemPrice(item),
        sku: item.sku ?? '',
        isActive: itemActive(item),
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
      const response = await unwrap(safe.get<PosMenuItemsResponse>('/pos/menu-items'));
      return toMenuList(normalizeItems(response));
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
      const response = await unwrap(safe.get<PosMenuItemsResponse>('/pos/menu-items'));
      const items = normalizeItems(response);
      return {
        _id: 'pos-menu',
        items: toStoreItems(items),
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
