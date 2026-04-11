import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import api from '../../lib/api';
import { useMenuStore, Menu, MenuItem, Category, Banner } from '../../stores/menuStore';

// Fetch all menus query
export const useAllMenus = (restaurantId: string) => {
  const { setAllMenus, setLoading, setError } = useMenuStore();

  return useQuery({
    queryKey: ['allMenus', restaurantId],
    queryFn: async () => {
      const response = await api.get(`/menu/menus/${restaurantId}`);
      if (response.data.message === 'No menus found') {
        return [];
      }
      return response.data;
    },
    enabled: !!restaurantId,
    onSuccess: (data: Menu[]) => {
      setAllMenus(data);
      setLoading(false);
      setError(null);
    },
    onError: (error: any) => {
      setAllMenus([]);
      setLoading(false);
      setError(error.message || 'Failed to fetch menus');
    },
  });
};

// Fetch menu by ID query
export const useMenuById = (menuId: string) => {
  const { setItems, setId, setLoading, setError } = useMenuStore();

  return useQuery({
    queryKey: ['menu', menuId],
    queryFn: async () => {
      const response = await api.get(`/menu/${menuId}`);
      return response.data;
    },
    enabled: !!menuId,
    onSuccess: (data: Menu) => {
      setItems(data.items);
      setId(data._id);
      setLoading(false);
      setError(null);
    },
    onError: (error: any) => {
      setItems([]);
      setId(null);
      setLoading(false);
      setError(error.message || 'Failed to fetch menu');
    },
  });
};

// Fetch menu query (with user ID)
export const useMenu = (userId: string, disabled: boolean = false) => {
  const { setItems, setId, setLoading, setError } = useMenuStore();

  return useQuery({
    queryKey: ['userMenu', userId, disabled],
    queryFn: async () => {
      const response = await api.get(`/menu/${userId}`, {
        params: { disabled },
      });
      return response.data;
    },
    enabled: !!userId,
    onSuccess: (data: Menu) => {
      setItems(data.items);
      setId(data._id);
      setLoading(false);
      setError(null);
    },
    onError: (error: any) => {
      setItems([]);
      setId(null);
      setLoading(false);
      setError(error.message || 'Failed to fetch menu');
    },
  });
};

// Set active menu mutation
export const useSetActiveMenu = () => {
  const { setActiveMenuId } = useMenuStore();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ restaurantId, menuId }: { restaurantId: string; menuId: string }) => {
      await api.post('/menu/set-active-menu', { restaurantId, menuId });
      return { restaurantId, menuId };
    },
    onSuccess: ({ menuId }) => {
      setActiveMenuId(menuId);
      // Invalidate related queries
      queryClient.invalidateQueries({ queryKey: ['allMenus'] });
    },
  });
};

// Create new menu mutation
export const useCreateMenu = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ restaurantId, menuName }: { restaurantId: string; menuName: string }) => {
      const response = await api.post('/menu/createNewMenu', { restaurantId, menuName });
      return response.data;
    },
    onSuccess: (newMenu: Menu) => {
      // Add to store
      const { allMenus, setAllMenus } = useMenuStore.getState();
      setAllMenus([...allMenus, newMenu]);

      // Invalidate queries
      queryClient.invalidateQueries({ queryKey: ['allMenus'] });
    },
  });
};

// Add menu item mutation
export const useAddMenuItem = () => {
  const { addMenuItem } = useMenuStore();
  const queryClient = useQueryClient();

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
        headers: { 'Content-Type': 'multipart/form-data' },
      });
      return { categoryId, item: response.data };
    },
    onSuccess: ({ categoryId, item }) => {
      addMenuItem(categoryId, item);
      queryClient.invalidateQueries({ queryKey: ['menu'] });
    },
  });
};

// Edit menu item mutation
export const useEditMenuItem = () => {
  const { editMenuItem } = useMenuStore();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ itemId, item }: { itemId: string; item: Partial<MenuItem> }) => {
      const response = await api.put(`/menu/item/${itemId}`, item);
      return { itemId, item: response.data };
    },
    onSuccess: ({ itemId, item }) => {
      editMenuItem(itemId, item as MenuItem);
      queryClient.invalidateQueries({ queryKey: ['menu'] });
    },
  });
};

// Delete menu item mutation
export const useDeleteMenuItem = () => {
  const { deleteMenuItem } = useMenuStore();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ categoryId, itemId }: { categoryId: string; itemId: string }) => {
      await api.delete(`/menu/category/${categoryId}/item/${itemId}`);
      return { categoryId, itemId };
    },
    onSuccess: ({ categoryId, itemId }) => {
      deleteMenuItem(categoryId, itemId);
      queryClient.invalidateQueries({ queryKey: ['menu'] });
    },
  });
};

// Add category mutation
export const useAddCategory = () => {
  const { addCategory } = useMenuStore();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (category: Partial<Category>) => {
      const response = await api.post('/menu/category', category);
      return response.data;
    },
    onSuccess: (newCategory: Category) => {
      addCategory(newCategory);
      queryClient.invalidateQueries({ queryKey: ['menu'] });
    },
  });
};

// Edit category mutation
export const useEditCategory = () => {
  const { editCategory } = useMenuStore();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      categoryId,
      category,
    }: {
      categoryId: string;
      category: Partial<Category>;
    }) => {
      const response = await api.put(`/menu/category/${categoryId}`, category);
      return { categoryId, category: response.data };
    },
    onSuccess: ({ categoryId, category }) => {
      editCategory(categoryId, category as Category);
      queryClient.invalidateQueries({ queryKey: ['menu'] });
    },
  });
};

// Delete category mutation
export const useDeleteCategory = () => {
  const { deleteCategory } = useMenuStore();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (categoryId: string) => {
      await api.delete(`/menu/category/${categoryId}`);
      return categoryId;
    },
    onSuccess: (categoryId) => {
      deleteCategory(categoryId);
      queryClient.invalidateQueries({ queryKey: ['menu'] });
    },
  });
};

// Rename category mutation
export const useRenameCategory = () => {
  const { renameCategory } = useMenuStore();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ categoryId, newName }: { categoryId: string; newName: string }) => {
      const response = await api.put(`/menu/category/${categoryId}`, { newName });
      return { categoryId, newName: response.data.name };
    },
    onSuccess: ({ categoryId, newName }) => {
      renameCategory(categoryId, newName);
      queryClient.invalidateQueries({ queryKey: ['menu'] });
    },
  });
};

// Add banner mutation
export const useAddBanner = () => {
  const { addBanner } = useMenuStore();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (formData: FormData) => {
      const response = await api.post('/menu/banner', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      });
      return response.data;
    },
    onSuccess: (newBanner: Banner) => {
      addBanner(newBanner);
      queryClient.invalidateQueries({ queryKey: ['menu'] });
    },
  });
};
