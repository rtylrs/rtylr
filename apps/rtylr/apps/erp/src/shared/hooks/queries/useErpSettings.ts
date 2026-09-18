import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { UseMutationResult, UseQueryResult } from '@tanstack/react-query';
import { toast } from 'sonner';
import { SettingsServiceClient } from '@voxire/voxie/erp/settings/v1/client';
import type { ErpBusinessSettings as ApiErpBusinessSettings } from '@voxire/voxie/erp/settings/v1/types';
import { useSettings, type ErpBusinessSettings } from '../../lib/settingsStore';

const settingsClient = createOneKitClient(SettingsServiceClient, 'erp');

export const ERP_SETTINGS_QUERY_KEY = ['erp-settings'] as const;

const DEFAULT_SETTINGS: ErpBusinessSettings = {
  currency: 'USD',
  taxRate: 8.25,
  lowStockThreshold: 10,
  defaultWarehouseId: 1,
  notifyOnLowStock: true,
  notifyOnExpiry: true,
};

function fromApiSettings(settings?: ApiErpBusinessSettings): ErpBusinessSettings {
  return {
    currency: settings?.currency || DEFAULT_SETTINGS.currency,
    taxRate: settings?.taxRate ?? DEFAULT_SETTINGS.taxRate,
    lowStockThreshold: settings?.lowStockThreshold ?? DEFAULT_SETTINGS.lowStockThreshold,
    defaultWarehouseId: Number(settings?.defaultWarehouseId ?? DEFAULT_SETTINGS.defaultWarehouseId),
    notifyOnLowStock: settings?.notifyOnLowStock ?? DEFAULT_SETTINGS.notifyOnLowStock,
    notifyOnExpiry: settings?.notifyOnExpiry ?? DEFAULT_SETTINGS.notifyOnExpiry,
  };
}

export const useErpSettings = (): UseQueryResult<ErpBusinessSettings> =>
  useQuery<ErpBusinessSettings>({
    queryKey: ERP_SETTINGS_QUERY_KEY,
    queryFn: async (): Promise<ErpBusinessSettings> => {
      const response = await settingsClient.getErpSettings({});
      const settings = fromApiSettings(response.settings);
      useSettings.getState().setBusinessSettings(settings);
      return settings;
    },
    staleTime: 30_000,
  });

export const useUpdateErpSettings = (): UseMutationResult<
  ErpBusinessSettings,
  unknown,
  ErpBusinessSettings
> => {
  const queryClient = useQueryClient();

  return useMutation<ErpBusinessSettings, unknown, ErpBusinessSettings>({
    mutationFn: async (settings: ErpBusinessSettings): Promise<ErpBusinessSettings> => {
      const response = await settingsClient.updateErpSettings({
        currency: settings.currency,
        taxRate: settings.taxRate,
        lowStockThreshold: settings.lowStockThreshold,
        defaultWarehouseId: String(settings.defaultWarehouseId),
        notifyOnLowStock: settings.notifyOnLowStock,
        notifyOnExpiry: settings.notifyOnExpiry,
      });
      return fromApiSettings(response.settings);
    },
    onSuccess: (settings): void => {
      queryClient.setQueryData(ERP_SETTINGS_QUERY_KEY, settings);
      useSettings.getState().setBusinessSettings(settings);
      toast.success('ERP settings saved');
    },
    onError: (error): void => {
      toast.error(getOneKitErrorMessage(error, 'Failed to save ERP settings'));
    },
  });
};
