import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import api from '../../lib/api';
import { useUserStore } from '../../stores/userStore';
import { toast } from 'sonner';

export const useGetBusiness = (businessId: string | null) => {
  return useQuery({
    queryKey: ['business', businessId],
    queryFn: async () => {
      const response = await api.get(`/business/${businessId}`);
      return response.data;
    },
    enabled: !!businessId,
    onError: (error: any) => {
      toast.error('Failed to fetch business information');
      console.error('Error fetching business:', error);
    },
  });
};

export const useUpdateBusiness = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ businessId, data }: { businessId: string; data: any }) => {
      const response = await api.put(`/business/${businessId}`, data);
      return response.data;
    },
    onSuccess: (data, variables) => {
      queryClient.invalidateQueries({ queryKey: ['business', variables.businessId] });
      queryClient.invalidateQueries({ queryKey: ['user'] });
      toast.success('Business information updated successfully');
    },
    onError: (error: any) => {
      const message = error.response?.data?.message || 'Failed to update business information';
      toast.error(message);
      console.error('Error updating business:', error);
    },
  });
};

export const useUpdateTheme = () => {
  const queryClient = useQueryClient();
  const { user, setUser } = useUserStore();

  return useMutation({
    mutationFn: async ({ id, themeData }: { id: string; themeData: any }) => {
      const response = await api.put(`/business/${id}/theme`, themeData);
      return response.data;
    },
    onSuccess: (data, variables) => {
      if (user) {
        setUser({ ...user, theme: data.theme });
      }
      queryClient.invalidateQueries({ queryKey: ['business', variables.id] });
      queryClient.invalidateQueries({ queryKey: ['user'] });
      toast.success('Theme updated successfully');
    },
    onError: (error: any) => {
      const message = error.response?.data?.message || 'Failed to update theme';
      toast.error(message);
      console.error('Error updating theme:', error);
    },
  });
};

