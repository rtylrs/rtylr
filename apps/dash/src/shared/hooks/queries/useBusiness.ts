import { unwrap } from '@voxire/api';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { safe } from '../../lib/api';
import { toast } from 'sonner';

export const useUpdateBusiness = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ data }: { businessId: string; data: Record<string, unknown> }) => {
      const payload: Record<string, unknown> = {};
      if (data.storeName) payload.siteName = data.storeName;
      if (data.companyName && !payload.siteName) payload.siteName = data.companyName;
      if (data.contactEmail) payload.contactEmail = data.contactEmail;
      if (data.telephone) payload.contactNumber = data.telephone;
      if (data.aboutContent) payload.aboutContent = data.aboutContent;
      if (data.logoUrl) payload.logoUrl = data.logoUrl;
      if (data.instagram) payload.instagram = data.instagram;
      if (data.maintenanceMode !== undefined) payload.maintenanceMode = data.maintenanceMode;

      return unwrap(safe.put('/config/update', payload));
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: ['business', variables.businessId] });
      queryClient.invalidateQueries({ queryKey: ['user'] });
      toast.success('Business information updated successfully');
    },
    onError: (error: unknown) => {
      const err = error as { message?: string };
      toast.error(err.message || 'Failed to update business information');
    },
  });
};
