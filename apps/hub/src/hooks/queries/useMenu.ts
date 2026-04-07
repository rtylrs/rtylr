import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useEffect } from 'react';
import api from '../../lib/api';
import { useMenuStore } from '../../stores/menuStore';
import { toast } from 'sonner';

interface MenuItem {
  _id: string;
  name: string;
  [key: string]: any;
}

interface Category {
  _id: string;
  name: string;
  type: 'category';
  menuItems: MenuItem[];
  [key: string]: any;
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
      const response = await api.get(`/menu/menus/${restaurantId}`);
      if (response.data.message === 'No menus found') {
        return [];
      }
      return response.data;
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
      console.error('Error fetching menus:', query.error);
    }
  }, [query.error]);

  return query;
};

export const useGetMenu = (userId: string | null, disabled?: boolean) => {
  const { setItems, setId } = useMenuStore();

  const query = useQuery({
    queryKey: ['menu', userId, disabled],
    queryFn: async () => {
      const response = await api.get(`/menu/${userId}`, {
        params: { disabled },
      });
      return response.data;
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
      console.error('Error fetching menu:', query.error);
    }
  }, [query.error]);

  return query;
};

export const useGetMenuById = (menuId: string | null) => {
  const { setItems, setId } = useMenuStore();

  const query = useQuery({
    queryKey: ['menu', menuId],
    queryFn: async () => {
      const response = await api.get(`/menu/${menuId}`);
      return response.data;
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
      console.error('Error fetching menu:', query.error);
    }
  }, [query.error]);

  return query;
};

export const useCreateMenu = () => {
  const queryClient = useQueryClient();
  const { allMenus, setAllMenus } = useMenuStore();

  return useMutation({
    mutationFn: async ({ restaurantId, menuName }: { restaurantId: string; menuName: string }) => {
      const response = await api.post('/menu/createNewMenu', {
        restaurantId,
        menuName,
      });
      return response.data;
    },
    onSuccess: (data, variables) => {
      setAllMenus([...allMenus, data]);
      queryClient.invalidateQueries({ queryKey: ['menus', variables.restaurantId] });
      toast.success('Menu created successfully');
    },
    onError: (error: any) => {
      toast.error('Failed to create menu');
      console.error('Error creating menu:', error);
    },
  });
};

export const useSetActiveMenu = () => {
  const queryClient = useQueryClient();
  const { setActiveMenuId, setId, setItems } = useMenuStore();

  return useMutation({
    mutationFn: async ({ restaurantId, menuId }: { restaurantId: string; menuId: string }) => {
      await api.post('/menu/set-active-menu', { restaurantId, menuId });
      return { menuId };
    },
    onSuccess: (data) => {
      setActiveMenuId(data.menuId);
      setId(data.menuId);
      setItems([]);
      queryClient.invalidateQueries({ queryKey: ['menus'] });
      toast.success('Active menu updated');
    },
    onError: (error: any) => {
      toast.error('Failed to set active menu');
      console.error('Error setting active menu:', error);
    },
  });
};

export const useAddMenuItem = () => {
  const queryClient = useQueryClient();
  const { items, setItems } = useMenuStore();

  return useMutation({
    mutationFn: async ({ categoryId, item }: { categoryId: string; item: Record<string, any> }) => {
      const formData = new FormData();
      Object.keys(item).forEach((key) => {
        if (Array.isArray(item[key])) {
          item[key].forEach((subItem: any, index: number) => {
            formData.append(`${key}[${index}]`, subItem);
          });
        } else {
          formData.append(key, item[key]);
        }
      });

      const response = await api.post(`/menu/category/${categoryId}/item`, formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      });
      return { categoryId, item: response.data };
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
    onError: (error: any) => {
      toast.error('Failed to add item');
      console.error('Error adding menu item:', error);
    },
  });
};

export const useEditMenuItem = () => {
  const queryClient = useQueryClient();
  const { items, setItems } = useMenuStore();

  return useMutation({
    mutationFn: async ({ itemId, item }: { itemId: string; item: Partial<MenuItem> }) => {
      const response = await api.put(`/menu/item/${itemId}`, item);
      return { itemId, item: response.data };
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
    onError: (error: any) => {
      toast.error('Failed to update item');
      console.error('Error editing menu item:', error);
    },
  });
};

export const useDeleteMenuItem = () => {
  const queryClient = useQueryClient();
  const { items, setItems } = useMenuStore();

  return useMutation({
    mutationFn: async ({ categoryId, itemId }: { categoryId: string; itemId: string }) => {
      await api.delete(`/menu/category/${categoryId}/item/${itemId}`);
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
    onError: (error: any) => {
      toast.error('Failed to delete item');
      console.error('Error deleting menu item:', error);
    },
  });
};

export const useAddCategory = () => {
  const queryClient = useQueryClient();
  const { items, setItems } = useMenuStore();

  return useMutation({
    mutationFn: async (category: Partial<Category>) => {
      const response = await api.post('/menu/category', category);
      return response.data;
    },
    onSuccess: (data) => {
      setItems([...items, data]);
      queryClient.invalidateQueries({ queryKey: ['menu'] });
      toast.success('Category added successfully');
    },
    onError: (error: any) => {
      toast.error('Failed to add category');
      console.error('Error adding category:', error);
    },
  });
};

export const useEditCategory = () => {
  const queryClient = useQueryClient();
  const { items, setItems } = useMenuStore();

  return useMutation({
    mutationFn: async ({ categoryId, category }: { categoryId: string; category: Partial<Category> }) => {
      const response = await api.put(`/menu/category/${categoryId}`, category);
      return { categoryId, category: response.data };
    },
    onSuccess: (data) => {
      const updatedItems = items.map((item) =>
        'type' in item && item.type === 'category' && item._id === data.categoryId ? data.category : item
      );
      setItems(updatedItems);
      queryClient.invalidateQueries({ queryKey: ['menu'] });
      toast.success('Category updated successfully');
    },
    onError: (error: any) => {
      toast.error('Failed to update category');
      console.error('Error editing category:', error);
    },
  });
};

export const useDeleteCategory = () => {
  const queryClient = useQueryClient();
  const { items, setItems } = useMenuStore();

  return useMutation({
    mutationFn: async (categoryId: string) => {
      await api.delete(`/menu/category/${categoryId}`);
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
    onError: (error: any) => {
      toast.error('Failed to delete category');
      console.error('Error deleting category:', error);
      throw error;
    },
  });
};

export const useRenameCategory = () => {
  const queryClient = useQueryClient();
  const { items, setItems } = useMenuStore();

  return useMutation({
    mutationFn: async ({ categoryId, newName }: { categoryId: string; newName: string }) => {
      const response = await api.put(`/menu/category/${categoryId}`, { newName });
      return { categoryId, name: response.data.name };
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
    onError: (error: any) => {
      toast.error('Failed to rename category');
      console.error('Error renaming category:', error);
    },
  });
};

export const useAddBanner = () => {
  const queryClient = useQueryClient();
  const { items, setItems } = useMenuStore();

  return useMutation({
    mutationFn: async (formData: FormData) => {
      const response = await api.post('/menu/banner', formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      });
      return response.data;
    },
    onSuccess: (data) => {
      setItems([...items, data]);
      queryClient.invalidateQueries({ queryKey: ['menu'] });
      toast.success('Banner added successfully');
    },
    onError: (error: any) => {
      toast.error('Failed to add banner');
      console.error('Error adding banner:', error);
    },
  });
};

