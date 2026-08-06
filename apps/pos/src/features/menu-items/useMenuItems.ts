import { createOneKitClient } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { MenuItemServiceClient } from '@voxire/voxie/pos/menu_item/v1/client';
import type { MenuItem } from '@voxire/voxie/pos/menu_item/v1/types';

const menuItemClient = createOneKitClient(MenuItemServiceClient, 'pos');

interface UseMenuItemsResult {
  items: MenuItem[];
  categories: string[];
  loading: boolean;
  error: string | null;
  reload: () => void;
}

export function useMenuItems(): UseMenuItemsResult {
  const query = useQuery({
    queryKey: ['pos-menu-items'],
    queryFn: async (): Promise<MenuItem[]> => {
      const res = await menuItemClient.listMenuItems({});
      return res.items ?? [];
    },
    staleTime: 2 * 60 * 1000,
  });

  const items = query.data ?? [];
  const categories = Array.from(new Set(items.map((i) => i.category ?? ''))).sort();

  return {
    items,
    categories,
    loading: query.isLoading,
    error: query.error
      ? query.error instanceof Error
        ? query.error.message
        : String(query.error)
      : null,
    reload: () => void query.refetch(),
  };
}
