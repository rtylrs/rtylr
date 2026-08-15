import { unwrap } from '@voxire/api';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { safe } from '../../lib/api';
import { toast } from 'sonner';
import { normalizeConfigData, useConfigStore } from '../../stores/configStore';

export const useSetBusinessSlug = () => {
  const queryClient = useQueryClient();
  const setWebsiteData = useConfigStore((s) => s.setWebsiteData);
  return useMutation({
    mutationFn: async (slug: string) => {
      return unwrap<{ slug: string }>(safe.patch('/hub/business/v1/businesses/slug', { slug }));
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

      setIfPresent('storeName', 'site_name');
      if (!payload.site_name) setIfPresent('companyName', 'site_name');
      setIfPresent('contactEmail', 'contact_email');
      setIfPresent('telephone', 'contact_number');
      setIfPresent('aboutContent', 'about_content');
      setIfPresent('logoUrl', 'logo_url');
      setIfPresent('instagram');
      setIfPresent('senderEmail', 'sender_email');
      setIfPresent('mapsUrl', 'maps_url');
      setIfPresent('website');
      setIfPresent('address');
      setIfPresent('city');
      setIfPresent('zipcode');
      setIfPresent('country');
      setIfPresent('businessType', 'business_type');
      setIfPresent('business_type');
      if (data.maintenanceMode !== undefined) payload.maintenance_mode = data.maintenanceMode;

      return unwrap<{ config?: Record<string, unknown> }>(
        safe.put('/hub/config/v1/config', payload)
      );
    },
    onSuccess: (res) => {
      if (res.config) {
        const config = normalizeConfigData(res.config);
        const currentWebsiteData = useConfigStore.getState().websiteData ?? {};
        setWebsiteData({ ...currentWebsiteData, ...config });
        setContactInfo({
          contactEmail: (config.contactEmail as string | null | undefined) ?? null,
          contactNumber: (config.contactNumber as string | null | undefined) ?? null,
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

export const useUpdateBusinessConfig = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (payload: Record<string, unknown>) => {
      return unwrap<{ config?: Record<string, unknown> }>(
        safe.put('/hub/config/v1/config', payload)
      );
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['business'] });
      queryClient.invalidateQueries({ queryKey: ['config'] });
      toast.success('Design configuration saved');
    },
    onError: (error: unknown) => {
      const err = error as { message?: string };
      toast.error(err.message || 'Failed to save configuration');
    },
  });
};
