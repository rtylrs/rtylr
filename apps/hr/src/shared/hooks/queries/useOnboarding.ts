import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { ListChecklistTemplatesResponse } from '@voxire/voxie/hr/onboarding/v1/types';
import type {
  CreateChecklistTemplateRequest,
  CreateChecklistTemplateResponse,
} from '@voxire/voxie/hr/onboarding/v1/types';
import type { ListEmployeeChecklistsResponse } from '@voxire/voxie/hr/onboarding/v1/types';
import type { CompleteChecklistItemResponse } from '@voxire/voxie/hr/onboarding/v1/types';
import { hrTransport } from '../../lib/clients';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import {
  employeeChecklistSummaryListSchema,
  type EmployeeChecklistSummaryList,
} from '../../schemas/onboarding';

// EmployeeChecklistSummaryPayload has no generated TS type yet - the backend
// route (GET /hr/onboarding/checklists/summary) isn't on a bumped voxie
// pointer at the time this hook was written. The response is validated with
// employeeChecklistSummaryListSchema instead of a generated contract type.
interface ListEmployeeChecklistSummariesParams {
  limit: number;
  offset: number;
}

// The generated CompleteChecklistItemRequest body is intentionally empty —
// the item id travels in the URL path, not the body. attachmentUrl is an
// optional body field. Model the call inputs locally.
interface CompleteChecklistItemInput {
  itemId: number;
  attachmentUrl?: string;
}

export const useListChecklistTemplates = () => {
  return useQuery<ListChecklistTemplatesResponse>({
    queryKey: ['checklist-templates'],
    queryFn: async (): Promise<ListChecklistTemplatesResponse> => {
      const res = await unwrap(
        hrTransport.get<ListChecklistTemplatesResponse>('/hr/onboarding/templates')
      );
      return res as ListChecklistTemplatesResponse;
    },
  });
};

export const useCreateChecklistTemplate = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (
      data: CreateChecklistTemplateRequest
    ): Promise<CreateChecklistTemplateResponse> => {
      const res = await unwrap(
        hrTransport.post<CreateChecklistTemplateResponse>('/hr/onboarding/templates', data)
      );
      return res as CreateChecklistTemplateResponse;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['checklist-templates'] });
      toast.success(t('onboardingMgmt.templateCreated'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useGetEmployeeChecklist = (employeeId?: number) => {
  return useQuery<ListEmployeeChecklistsResponse>({
    queryKey: ['employee-checklists', employeeId],
    queryFn: async (): Promise<ListEmployeeChecklistsResponse> => {
      const res = await unwrap(
        hrTransport.get<ListEmployeeChecklistsResponse>('/hr/onboarding/checklists', {
          params: employeeId ? { employeeId } : undefined,
        })
      );
      return res as ListEmployeeChecklistsResponse;
    },
    enabled: !!employeeId,
  });
};

export const useListEmployeeChecklistSummaries = (params: ListEmployeeChecklistSummariesParams) => {
  return useQuery<EmployeeChecklistSummaryList>({
    queryKey: ['employee-checklist-summaries', 'filter', params],
    queryFn: async (): Promise<EmployeeChecklistSummaryList> => {
      const res = await unwrap(
        hrTransport.get<unknown>('/hr/onboarding/checklists/summary', { params })
      );
      return employeeChecklistSummaryListSchema.parse(res);
    },
  });
};

export const useCompleteChecklistItem = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (
      data: CompleteChecklistItemInput
    ): Promise<CompleteChecklistItemResponse> => {
      const res = await unwrap(
        hrTransport.post<CompleteChecklistItemResponse>(
          `/hr/onboarding/checklists/items/${data.itemId}/complete`,
          data
        )
      );
      return res as CompleteChecklistItemResponse;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['employee-checklists'] });
      toast.success(t('onboardingMgmt.itemCompleted'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};
