import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { EmployeePayload } from '@voxire/contracts/hr/employee/v1/types';
import type { ListEmployeesResponse } from '@voxire/contracts/hr/employee/v1/list_employees';
import type {
  CreateEmployeeRequest,
  CreateEmployeeResponse,
} from '@voxire/contracts/hr/employee/v1/create_employee';
import type {
  UpdateEmployeeRequest,
  UpdateEmployeeResponse,
} from '@voxire/contracts/hr/employee/v1/update_employee';
import type {
  InviteEmployeeRequest,
  InviteEmployeeResponse,
} from '@voxire/contracts/hr/employee/v1/invite_employee';
import { safe } from '../../lib/api';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

interface ListEmployeesParams {
  limit: number;
  offset: number;
  departmentId?: number;
  locationId?: number;
  status?: string;
}

export const useListEmployees = (params: ListEmployeesParams) => {
  return useQuery<ListEmployeesResponse>({
    queryKey: ['employees', 'filter', params],
    queryFn: async (): Promise<ListEmployeesResponse> => {
      const res = await unwrap(safe.get<ListEmployeesResponse>('/hr/employees', { params }));
      return res as ListEmployeesResponse;
    },
  });
};

export const useGetEmployee = (id: number | undefined) => {
  return useQuery<EmployeePayload>({
    queryKey: ['employee', id],
    queryFn: async (): Promise<EmployeePayload> => {
      const res = await unwrap(safe.get<{ employee: EmployeePayload }>(`/hr/employees/${id}`));
      return (res as { employee: EmployeePayload }).employee;
    },
    enabled: !!id,
  });
};

export const useCreateEmployee = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: CreateEmployeeRequest): Promise<CreateEmployeeResponse> => {
      const res = await unwrap(safe.post<CreateEmployeeResponse>('/hr/employees', data));
      return res as CreateEmployeeResponse;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['employees'] });
      toast.success(t('employees.created'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useUpdateEmployee = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: UpdateEmployeeRequest): Promise<UpdateEmployeeResponse> => {
      const res = await unwrap(safe.put<UpdateEmployeeResponse>(`/hr/employees/${data.id}`, data));
      return res as UpdateEmployeeResponse;
    },
    onSuccess: (_data: UpdateEmployeeResponse, variables: UpdateEmployeeRequest) => {
      queryClient.invalidateQueries({ queryKey: ['employees'] });
      queryClient.invalidateQueries({ queryKey: ['employee', variables.id] });
      toast.success(t('employees.updated'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useInviteEmployee = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: InviteEmployeeRequest): Promise<InviteEmployeeResponse> => {
      const res = await unwrap(safe.post<InviteEmployeeResponse>('/hr/employees/invite', data));
      return res as InviteEmployeeResponse;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['employees'] });
      toast.success(t('employees.inviteSent'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};
