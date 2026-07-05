import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { ListChecklistTemplatesResponse } from '@voxire/voxie/hr/onboarding/v1/list_checklist_templates';
import type {
  CreateChecklistTemplateRequest,
  CreateChecklistTemplateResponse,
} from '@voxire/voxie/hr/onboarding/v1/create_checklist_template';
import type { ListEmployeeChecklistsResponse } from '@voxire/voxie/hr/onboarding/v1/list_employee_checklists';
import type { CompleteChecklistItemResponse } from '@voxire/voxie/hr/onboarding/v1/complete_checklist_item';
import { safe } from '../../lib/api';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

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
        safe.get<ListChecklistTemplatesResponse>('/hr/onboarding/templates')
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
        safe.post<CreateChecklistTemplateResponse>('/hr/onboarding/templates', data)
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
        safe.get<ListEmployeeChecklistsResponse>('/hr/onboarding/checklists', {
          params: employeeId ? { employeeId } : undefined,
        })
      );
      return res as ListEmployeeChecklistsResponse;
    },
    enabled: !!employeeId,
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
        safe.post<CompleteChecklistItemResponse>(
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
