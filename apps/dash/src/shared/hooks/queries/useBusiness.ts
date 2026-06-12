import { unwrap } from '@voxire/api';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { safe } from '../../lib/api';
import { toast } from 'sonner';
import { useConfigStore } from '../../stores/configStore';

export const useSetBusinessSlug = () => {
  const queryClient = useQueryClient();
  const setWebsiteData = useConfigStore((s) => s.setWebsiteData);
  return useMutation({
    mutationFn: async (slug: string) => {
      return unwrap<{ slug: string }>(safe.put('/business/me/slug', { slug }));
    },
    onSuccess: ({ slug }) => {
      const currentWebsiteData = useConfigStore.getState().websiteData ?? {};
      setWebsiteData({ ...currentWebsiteData, slug });
      queryClient.invalidateQueries({ queryKey: ['business'] });
      queryClient.invalidateQueries({ queryKey: ['user'] });
      toast.success('Slug updated');
    },
    onError: (error: unknown) => {
      const err = error as { message?: string };
      toast.error(err.message || 'Failed to update slug');
    },
  });
};

export const useUpdateBusiness = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ data }: { data: Record<string, unknown> }) => {
      const payload: Record<string, unknown> = {};
      if (data.storeName) payload.siteName = data.storeName;
      if (data.companyName && !payload.siteName) payload.siteName = data.companyName;
      if (data.contactEmail) payload.contactEmail = data.contactEmail;
      if (data.telephone) payload.contactNumber = data.telephone;
      if (data.aboutContent) payload.aboutContent = data.aboutContent;
      if (data.logoUrl) payload.logoUrl = data.logoUrl;
      if (data.instagram) payload.instagram = data.instagram;
      if (data.senderEmail) payload.senderEmail = data.senderEmail;
      if (data.mapsUrl) payload.mapsUrl = data.mapsUrl;
      if (data.website) payload.website = data.website;
      if (data.address) payload.address = data.address;
      if (data.city) payload.city = data.city;
      if (data.zipcode) payload.zipcode = data.zipcode;
      if (data.country) payload.country = data.country;
      if (data.maintenanceMode !== undefined) payload.maintenanceMode = data.maintenanceMode;

      return unwrap(safe.put('/config/update', payload));
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['business'] });
      queryClient.invalidateQueries({ queryKey: ['user'] });
      toast.success('Business information updated successfully');
    },
    onError: (error: unknown) => {
      const err = error as { message?: string };
      toast.error(err.message || 'Failed to update business information');
    },
  });
};
