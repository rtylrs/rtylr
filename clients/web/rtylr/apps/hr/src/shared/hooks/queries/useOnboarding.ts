import { createOneKitClient } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { OnboardingServiceClient } from '@voxire/voxie/hr/onboarding/v1/client';
import { StaffServiceClient } from '@voxire/voxie/hr/staff/v1/client';
import type {
  ListChecklistTemplatesResponse,
  CreateChecklistTemplateRequest,
  CreateChecklistTemplateResponse,
  ListEmployeeChecklistsResponse,
  EmployeeChecklistPayload,
} from '@voxire/voxie/hr/onboarding/v1/types';
import { todayIso } from '../../lib/dateUtils';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import {
  employeeChecklistSummaryListSchema,
  type EmployeeChecklistSummary,
  type EmployeeChecklistSummaryList,
} from '../../schemas/onboarding';

const onboardingClient = createOneKitClient(OnboardingServiceClient, 'hr');
const staffClient = createOneKitClient(StaffServiceClient, 'hr');

export interface ListEmployeeChecklistSummariesParams {
  limit: number;
  offset: number;
}

function today(): string {
  return todayIso();
}

function toSummary(
  employeeId: number,
  checklist: EmployeeChecklistPayload
): EmployeeChecklistSummary {
  const items = checklist.items ?? [];
  const completedItems = items.filter((item) => item.status === 'complete').length;
  const overdueItemsCount = items.filter((item) => {
    if (item.status === 'complete' || !item.dueDate) return false;
    return item.dueDate.slice(0, 10) < today();
  }).length;
  const totalItems = items.length;
  const completionPercentage =
    totalItems === 0 ? 0 : Math.round((completedItems / totalItems) * 100);

  return {
    employeeId,
    checklistId: Number(checklist.id ?? 0),
    templateId: Number(checklist.templateId ?? 0),
    status:
      completionPercentage === 100 ? 'complete' : completedItems > 0 ? 'in_progress' : 'incomplete',
    totalItems,
    completedItems,
    completionPercentage,
    overdueItemsCount,
    triggeredAt: Number(checklist.triggeredAt ?? checklist.createdAt ?? 0),
    updatedAt: Number(checklist.updatedAt ?? checklist.createdAt ?? 0),
  };
}

export const useListChecklistTemplates = () => {
  return useQuery<ListChecklistTemplatesResponse>({
    queryKey: ['checklist-templates'],
    queryFn: () => onboardingClient.listChecklistTemplates({}),
  });
};

export const useCreateChecklistTemplate = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: (data: CreateChecklistTemplateRequest): Promise<CreateChecklistTemplateResponse> =>
      onboardingClient.createChecklistTemplate(data),
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
    queryFn: () =>
      onboardingClient.listEmployeeChecklists({
        employeeId: String(employeeId),
        limit: 100,
        offset: 0,
      }),
    enabled: employeeId !== undefined && employeeId > 0,
  });
};

export const useListEmployeeChecklistSummaries = (params: ListEmployeeChecklistSummariesParams) => {
  return useQuery<EmployeeChecklistSummaryList>({
    queryKey: ['employee-checklist-summaries', 'filter', params],
    queryFn: async (): Promise<EmployeeChecklistSummaryList> => {
      const staffResponse = await staffClient.listStaff({
        search: '',
        departmentId: '',
        status: '',
      });
      const staff = staffResponse.staff ?? [];
      const page = staff.slice(params.offset, params.offset + params.limit);
      const summaries = (
        await Promise.all(
          page.map(async (member) => {
            const employeeId = Number(member.id);
            if (!Number.isSafeInteger(employeeId) || employeeId <= 0) return [];

            const response = await onboardingClient.listEmployeeChecklists({
              employeeId: String(employeeId),
              limit: 100,
              offset: 0,
            });
            const checklist = response.checklists?.[0];
            return checklist ? [toSummary(employeeId, checklist)] : [];
          })
        )
      ).flat();

      return employeeChecklistSummaryListSchema.parse({
        summaries,
        total: staff.length,
        limit: params.limit,
        offset: params.offset,
      });
    },
  });
};
