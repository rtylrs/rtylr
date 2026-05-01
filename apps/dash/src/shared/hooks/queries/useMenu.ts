import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useEffect } from 'react';
import { z } from 'zod';
import { safe } from '../../lib/api';
import { menuSchema } from '../../schemas/menu';
import { useMenuStore } from '../../stores/menuStore';
import { toast } from 'sonner';

interface MenuItem {
  _id: string;
  name: string;
  [key: string]: unknown;
}

interface Category {
  _id: string;
  name: string;
  type: 'category';
  menuItems: MenuItem[];
  [key: string]: unknown;
}

interface Banner {
  _id: string;
  bannerImage: string;
  menuId: string;
}

export const useGetAllMenus = (restaurantId: string | null) => {
  const { setAllMenus } = useMenuStore();

  const query = useQuery({
    queryKey: ['menus', restaurantId],
    queryFn: async () => {
      const data = await unwrap(safe.get(`/menu/menus/${restaurantId}`));
      if (data?.message === 'No menus found') {
        return [];
      }
      return z.array(menuSchema).parse(Array.isArray(data) ? data : []);
    },
    enabled: !!restaurantId,
  });

  useEffect(() => {
    if (query.data) {
      setAllMenus(query.data);
    }
  }, [query.data, setAllMenus]);

  useEffect(() => {
    if (query.error) {
      toast.error('Failed to fetch menus');
    }
  }, [query.error]);

  return query;
};

export const useGetMenu = (userId: string | null, disabled?: boolean) => {
  const { setItems, setId } = useMenuStore();

  const query = useQuery({
    queryKey: ['menu', userId, disabled],
    queryFn: async () => {
      const response = await unwrap(
        safe.get(`/menu/${userId}`, {
          params: { disabled },
        })
      );
      return menuSchema.parse(response);
    },
    enabled: !!userId,
  });

  useEffect(() => {
    if (query.data) {
      setItems(query.data.items);
      setId(query.data._id);
    }
  }, [query.data, setItems, setId]);

  useEffect(() => {
    if (query.error) {
      toast.error('Failed to fetch menu');
    }
  }, [query.error]);

  return query;
};

export const useGetMenuById = (menuId: string | null) => {
  const { setItems, setId } = useMenuStore();

  const query = useQuery({
    queryKey: ['menu', menuId],
    queryFn: async () => {
      const response = await unwrap(safe.get(`/menu/${menuId}`));
      return menuSchema.parse(response);
    },
    enabled: !!menuId,
  });

  useEffect(() => {
    if (query.data) {
      setItems(query.data.items);
      setId(query.data._id);
    }
  }, [query.data, setItems, setId]);

  useEffect(() => {
    if (query.error) {
      toast.error('Failed to fetch menu');
    }
  }, [query.error]);

  return query;
};

export const useCreateMenu = () => {
  const queryClient = useQueryClient();
  const { allMenus, setAllMenus } = useMenuStore();

  return useMutation({
    mutationFn: async ({ restaurantId, menuName }: { restaurantId: string; menuName: string }) => {
      return unwrap(
        safe.post('/menu/createNewMenu', {
          restaurantId,
          menuName,
        })
      );
    },
    onSuccess: (data, variables) => {
      setAllMenus([...allMenus, data]);
      queryClient.invalidateQueries({ queryKey: ['menus', variables.restaurantId] });
      toast.success('Menu created successfully');
    },
    onError: (error: Error) => {
      toast.error('Failed to create menu');
    },
  });
};

export const useSetActiveMenu = () => {
  const queryClient = useQueryClient();
  const { setActiveMenuId, setId, setItems } = useMenuStore();

  return useMutation({
    mutationFn: async ({ restaurantId, menuId }: { restaurantId: string; menuId: string }) => {
      await unwrap(safe.post('/menu/set-active-menu', { restaurantId, menuId }));
      return { menuId };
    },
    onSuccess: (data) => {
      setActiveMenuId(data.menuId);
      setId(data.menuId);
      setItems([]);
      queryClient.invalidateQueries({ queryKey: ['menus'] });
      toast.success('Active menu updated');
    },
    onError: (error: Error) => {
      toast.error('Failed to set active menu');
    },
  });
};

export const useAddMenuItem = () => {
  const queryClient = useQueryClient();
  const { items, setItems } = useMenuStore();

  return useMutation({
    mutationFn: async ({
      categoryId,
      item,
    }: {
      categoryId: string;
      item: Record<string, unknown>;
    }) => {
      const formData = new FormData();
      Object.keys(item).forEach((key) => {
        if (Array.isArray(item[key])) {
          (item[key] as unknown[]).forEach((subItem: unknown, index: number) => {
            formData.append(`${key}[${index}]`, subItem as string | Blob);
          });
        } else {
          formData.append(key, item[key] as string | Blob);
        }
      });

      const created = await unwrap(
        safe.post(`/menu/category/${categoryId}/item`, formData, {
          headers: {
            'Content-Type': 'multipart/form-data',
          },
        })
      );
      return { categoryId, item: created };
    },
    onSuccess: (data) => {
      const updatedItems = items.map((cat) => {
        if ('type' in cat && cat.type === 'category' && cat._id === data.categoryId) {
          return {
            ...cat,
            menuItems: [...cat.menuItems, data.item],
          };
        }
        return cat;
      });
      setItems(updatedItems);
      queryClient.invalidateQueries({ queryKey: ['menu'] });
      toast.success('Item added successfully');
    },
    onError: (error: Error) => {
      toast.error('Failed to add item');
    },
  });
};

export const useEditMenuItem = () => {
  const queryClient = useQueryClient();
  const { items, setItems } = useMenuStore();

  return useMutation({
    mutationFn: async ({ itemId, item }: { itemId: string; item: Partial<MenuItem> }) => {
      const response = await unwrap(safe.put(`/menu/item/${itemId}`, item));
      return { itemId, item: response };
    },
    onSuccess: (data) => {
      const updatedItems = items.map((cat) => {
        if ('type' in cat && cat.type === 'category') {
          const updatedMenuItems = cat.menuItems.map((menuItem) =>
            menuItem._id === data.itemId ? data.item : menuItem
          );
          return { ...cat, menuItems: updatedMenuItems };
        }
        return cat;
      });
      setItems(updatedItems);
      queryClient.invalidateQueries({ queryKey: ['menu'] });
      toast.success('Item updated successfully');
    },
    onError: (error: Error) => {
      toast.error('Failed to update item');
    },
  });
};

export const useDeleteMenuItem = () => {
  const queryClient = useQueryClient();
  const { items, setItems } = useMenuStore();

  return useMutation({
    mutationFn: async ({ categoryId, itemId }: { categoryId: string; itemId: string }) => {
      await unwrap(safe.delete(`/menu/category/${categoryId}/item/${itemId}`));
      return { categoryId, itemId };
    },
    onSuccess: (data) => {
      const updatedItems = items.map((cat) => {
        if ('type' in cat && cat.type === 'category' && cat._id === data.categoryId) {
          return {
            ...cat,
            menuItems: cat.menuItems.filter((item) => item._id !== data.itemId),
          };
        }
        return cat;
      });
      setItems(updatedItems);
      queryClient.invalidateQueries({ queryKey: ['menu'] });
      toast.success('Item deleted successfully');
    },
    onError: (error: Error) => {
      toast.error('Failed to delete item');
    },
  });
};

export const useAddCategory = () => {
  const queryClient = useQueryClient();
  const { items, setItems } = useMenuStore();

  return useMutation({
    mutationFn: async (category: Partial<Category>) => {
      const response = await unwrap(safe.post('/menu/category', category));
      return response;
    },
    onSuccess: (data) => {
      setItems([...items, data]);
      queryClient.invalidateQueries({ queryKey: ['menu'] });
      toast.success('Category added successfully');
    },
    onError: (error: Error) => {
      toast.error('Failed to add category');
    },
  });
};

export const useEditCategory = () => {
  const queryClient = useQueryClient();
  const { items, setItems } = useMenuStore();

  return useMutation({
    mutationFn: async ({
      categoryId,
      category,
    }: {
      categoryId: string;
      category: Partial<Category>;
    }) => {
      const response = await unwrap(safe.put(`/menu/category/${categoryId}`, category));
      return { categoryId, category: response };
    },
    onSuccess: (data) => {
      const updatedItems = items.map((item) =>
        'type' in item && item.type === 'category' && item._id === data.categoryId
          ? data.category
          : item
      );
      setItems(updatedItems);
      queryClient.invalidateQueries({ queryKey: ['menu'] });
      toast.success('Category updated successfully');
    },
    onError: (error: Error) => {
      toast.error('Failed to update category');
    },
  });
};

export const useDeleteCategory = () => {
  const queryClient = useQueryClient();
  const { items, setItems } = useMenuStore();

  return useMutation({
    mutationFn: async (categoryId: string) => {
      await unwrap(safe.delete(`/menu/category/${categoryId}`));
      return categoryId;
    },
    onSuccess: (data) => {
      const updatedItems = items.filter(
        (item) => !('type' in item && item.type === 'category' && item._id === data)
      );
      setItems(updatedItems);
      queryClient.invalidateQueries({ queryKey: ['menu'] });
      toast.success('Category deleted successfully');
    },
    onError: (error: Error) => {
      toast.error('Failed to delete category');
      throw error;
    },
  });
};

export const useRenameCategory = () => {
  const queryClient = useQueryClient();
  const { items, setItems } = useMenuStore();

  return useMutation({
    mutationFn: async ({ categoryId, newName }: { categoryId: string; newName: string }) => {
      const response = await unwrap(safe.put(`/menu/category/${categoryId}`, { newName }));
      return { categoryId, name: response.name };
    },
    onSuccess: (data) => {
      const updatedItems = items.map((cat) => {
        if ('type' in cat && cat.type === 'category' && cat._id === data.categoryId) {
          return { ...cat, name: data.name };
        }
        return cat;
      });
      setItems(updatedItems);
      queryClient.invalidateQueries({ queryKey: ['menu'] });
      toast.success('Category renamed successfully');
    },
    onError: (error: Error) => {
      toast.error('Failed to rename category');
    },
  });
};

export const useAddBanner = () => {
  const queryClient = useQueryClient();
  const { items, setItems } = useMenuStore();

  return useMutation({
    mutationFn: async (formData: FormData) => {
      return unwrap(
        safe.post('/menu/banner', formData, {
          headers: {
            'Content-Type': 'multipart/form-data',
          },
        })
      );
    },
    onSuccess: (data) => {
      setItems([...items, data]);
      queryClient.invalidateQueries({ queryKey: ['menu'] });
      toast.success('Banner added successfully');
    },
    onError: (error: Error) => {
      toast.error('Failed to add banner');
    },
  });
};
