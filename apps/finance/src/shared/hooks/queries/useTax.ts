import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import type { ListTaxConfigResponse } from '@voxire/voxie/finance/report/v1/list_tax_config';
import type {
  CreateTaxConfigRequest,
  CreateTaxConfigResponse,
} from '@voxire/voxie/finance/report/v1/create_tax_config';
import type { GetTaxReportResponse } from '@voxire/voxie/finance/report/v1/get_tax_report';
import { safe } from '../../lib/api';

interface TaxReportRange {
  from: string;
  to: string;
}

export const useTaxConfig = () => {
  return useQuery<ListTaxConfigResponse>({
    queryKey: ['tax-config'],
    queryFn: async (): Promise<ListTaxConfigResponse> => {
      const res = await unwrap(safe.get<ListTaxConfigResponse>('/finance/tax'));
      return res as ListTaxConfigResponse;
    },
  });
};

export const useUpdateTaxConfig = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: CreateTaxConfigRequest): Promise<CreateTaxConfigResponse> => {
      const source = data as CreateTaxConfigRequest & {
        taxName?: string;
        taxRate?: number;
        isInclusive?: boolean;
        isActive?: boolean;
      };
      const res = await unwrap(
        safe.put<CreateTaxConfigResponse>('/finance/tax', {
          taxName: source.taxName ?? data.name,
          taxRate: source.taxRate ?? data.rate,
          isInclusive: source.isInclusive ?? false,
          isActive: source.isActive ?? true,
        })
      );
      return res as CreateTaxConfigResponse;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tax-config'] });
      queryClient.invalidateQueries({ queryKey: ['tax-report'] });
      toast.success(t('tax.configUpdated'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useTaxReport = (range: TaxReportRange) => {
  return useQuery<GetTaxReportResponse>({
    queryKey: ['tax-report', range],
    queryFn: async (): Promise<GetTaxReportResponse> => {
      const res = await unwrap(
        safe.get<GetTaxReportResponse>('/finance/tax/report', { params: range })
      );
      return res as GetTaxReportResponse;
    },
    enabled: !!range.from && !!range.to,
  });
};
