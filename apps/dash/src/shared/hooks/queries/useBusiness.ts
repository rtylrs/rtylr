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
  const setWebsiteData = useConfigStore((s) => s.setWebsiteData);
  const setContactInfo = useConfigStore((s) => s.setContactInfo);

  return useMutation({
    mutationFn: async ({ data }: { data: Record<string, unknown> }) => {
      const payload: Record<string, unknown> = {};
      const has = (key: string) => Object.prototype.hasOwnProperty.call(data, key);
      const setIfPresent = (sourceKey: string, targetKey = sourceKey) => {
        if (has(sourceKey) && data[sourceKey] !== undefined) payload[targetKey] = data[sourceKey];
      };

      setIfPresent('storeName', 'siteName');
      if (!payload.siteName) setIfPresent('companyName', 'siteName');
      setIfPresent('contactEmail');
      setIfPresent('telephone', 'contactNumber');
      setIfPresent('aboutContent');
      setIfPresent('logoUrl');
      setIfPresent('instagram');
      setIfPresent('senderEmail');
      setIfPresent('mapsUrl');
      setIfPresent('website');
      setIfPresent('address');
      setIfPresent('city');
      setIfPresent('zipcode');
      setIfPresent('country');
      if (data.maintenanceMode !== undefined) payload.maintenanceMode = data.maintenanceMode;

      return unwrap<{ config?: Record<string, unknown> }>(safe.put('/config/update', payload));
    },
    onSuccess: (res) => {
      if (res.config) {
        const currentWebsiteData = useConfigStore.getState().websiteData ?? {};
        setWebsiteData({ ...currentWebsiteData, ...res.config });
        setContactInfo({
          contactEmail: (res.config.contactEmail as string | null | undefined) ?? null,
          contactNumber: (res.config.contactNumber as string | null | undefined) ?? null,
        });
      }
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
