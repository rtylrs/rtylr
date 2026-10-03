import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { getOneKitErrorMessage } from '@voxire/api';
import { businessClient, configClient } from '../../lib/clients';
import { toast } from 'sonner';
import { normalizeConfigData, useConfigStore } from '../../stores/configStore';
import i18next from 'i18next';
import type { UpdateConfigRequest } from '@voxire/voxie/hub/config/v1/types';

export const useSetBusinessSlug = () => {
  const queryClient = useQueryClient();
  const setWebsiteData = useConfigStore((s) => s.setWebsiteData);
  return useMutation({
    mutationFn: async (slug: string) => {
      return businessClient.setSlug({ slug });
    },
    onSuccess: ({ slug }) => {
      const currentWebsiteData = useConfigStore.getState().websiteData ?? {};
      setWebsiteData({ ...currentWebsiteData, slug });
      queryClient.invalidateQueries({ queryKey: ['business'] });
      queryClient.invalidateQueries({ queryKey: ['user'] });
      toast.success(i18next.t('common.toasts.slugUpdated', 'Slug updated'));
    },
    onError: (error: unknown) => {
      const err = error as { message?: string };
      toast.error(
        err.message || i18next.t('common.toasts.failedUpdateSlug', 'Failed to update slug')
      );
    },
  });
};

export const useCustomDomain = () => {
  return useQuery({
    queryKey: ['custom-domain'],
    queryFn: async () => {
      const res = await configClient.getCustomDomain({});
      return res.domain?.domain ?? '';
    },
  });
};

export const useRegisterCustomDomain = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (domain: string) => configClient.registerCustomDomain({ domain }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['custom-domain'] });
      toast.success(
        i18next.t(
          'settings.domain.saved',
          'Custom domain saved. DNS propagation may take up to 24 hours.'
        )
      );
    },
    onError: (error: unknown) => {
      toast.error(
        getOneKitErrorMessage(
          error,
          i18next.t('settings.domain.saveFailed', 'Failed to save the custom domain')
        )
      );
    },
  });
};

export const useUpdateBusiness = () => {
  const queryClient = useQueryClient();
  const setWebsiteData = useConfigStore((s) => s.setWebsiteData);
  const setContactInfo = useConfigStore((s) => s.setContactInfo);

  return useMutation({
    mutationFn: async ({ data }: { data: Record<string, unknown> }) => {
      const payload: Partial<UpdateConfigRequest> = {};
      const has = (key: string) => Object.prototype.hasOwnProperty.call(data, key);
      const setIfPresent = (
        sourceKey: string,
        targetKey: keyof UpdateConfigRequest = sourceKey as keyof UpdateConfigRequest
      ) => {
        if (has(sourceKey) && data[sourceKey] !== undefined) {
          payload[targetKey] = data[sourceKey] as never;
        }
      };

      setIfPresent('storeName', 'siteName');
      if (!has('storeName') || data.storeName === undefined) {
        setIfPresent('companyName', 'siteName');
      }
      setIfPresent('contactEmail');
      setIfPresent('telephone', 'contactNumber');
      setIfPresent('aboutContent');
      setIfPresent('logoUrl');
      setIfPresent('instagram');
      setIfPresent('senderName');
      setIfPresent('senderEmail');
      setIfPresent('replyTo', 'replyToEmail');
      setIfPresent('smtpHost');
      setIfPresent('smtpPort');
      setIfPresent('smtpUser', 'smtpUsername');
      setIfPresent('smtpPassword');
      setIfPresent('smtpFrom', 'senderEmail');
      setIfPresent('mapsUrl');
      setIfPresent('website');
      setIfPresent('address');
      setIfPresent('city');
      setIfPresent('zipcode');
      setIfPresent('country');
      if (data.maintenanceMode !== undefined)
        payload.maintenanceMode = Boolean(data.maintenanceMode);

      return configClient.updateConfig(payload);
    },
    onSuccess: (res) => {
      if (res.config) {
        const config = normalizeConfigData(res.config as unknown as Record<string, unknown>);
        const currentWebsiteData = useConfigStore.getState().websiteData ?? {};
        setWebsiteData({ ...currentWebsiteData, ...config });
        setContactInfo({
          contactEmail: (config.contactEmail as string | null | undefined) ?? null,
          contactNumber: (config.contactNumber as string | null | undefined) ?? null,
        });
      }
      queryClient.invalidateQueries({ queryKey: ['business'] });
      queryClient.invalidateQueries({ queryKey: ['user'] });
      queryClient.invalidateQueries({ queryKey: ['hr'] });
      toast.success(
        i18next.t('common.toasts.businessInfoUpdated', 'Business information updated successfully')
      );
    },
    onError: (error: unknown) => {
      const err = error as { message?: string };
      toast.error(
        err.message ||
          i18next.t(
            'common.toasts.failedUpdateBusinessInfo',
            'Failed to update business information'
          )
      );
    },
  });
};

export const useUpdateBusinessConfig = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (payload: Record<string, unknown>) => {
      return configClient.updateConfig(payload as UpdateConfigRequest);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['business'] });
      queryClient.invalidateQueries({ queryKey: ['config'] });
      toast.success(i18next.t('common.toasts.designConfigSaved', 'Design configuration saved'));
    },
    onError: (error: unknown) => {
      const err = error as { message?: string };
      toast.error(
        err.message || i18next.t('common.toasts.failedSaveConfig', 'Failed to save configuration')
      );
    },
  });
};
