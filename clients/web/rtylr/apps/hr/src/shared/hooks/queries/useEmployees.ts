import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { EmployeePayload } from '@voxire/voxie/hr/employee/v1/types';
import type { ListEmployeesResponse } from '@voxire/voxie/hr/employee/v1/types';
import type { CreateEmployeeResponse } from '@voxire/voxie/hr/employee/v1/types';
import type { UpdateEmployeeResponse } from '@voxire/voxie/hr/employee/v1/types';
import type {
  InviteEmployeeRequest,
  InviteEmployeeResponse,
} from '@voxire/voxie/hr/employee/v1/types';
import type { EmployeeFormData } from '../../schemas/employeeSchema';
import { hrTransport } from '../../lib/clients';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

interface ListEmployeesParams {
  limit: number;
  offset: number;
  departmentId?: number;
  locationId?: number;
  status?: string;
}

export interface CreateEmployeePayload {
  userId?: number;
  departmentId: number;
  locationId: number;
  workScheduleId?: number;
  firstName: string;
  lastName: string;
  email: string;
  employmentType: string;
  salary?: number;
  hourlyRate?: number;
  startDate: string;
}

export interface UpdateEmployeePayload extends Partial<CreateEmployeePayload> {
  id: number;
  status?: string;
  endDate?: string;
}

export const toCreateEmployeePayload = (data: EmployeeFormData): CreateEmployeePayload => ({
  userId: data.userId || undefined,
  departmentId: data.departmentId,
  locationId: data.locationId,
  workScheduleId: data.workScheduleId,
  firstName: data.firstName,
  lastName: data.lastName,
  email: data.email,
  employmentType: data.contractType,
  salary: data.monthlySalary,
  hourlyRate: data.hourlyRate,
  startDate: data.hireDate,
});

export const useListEmployees = (params: ListEmployeesParams) => {
  return useQuery<ListEmployeesResponse>({
    queryKey: ['employees', 'filter', params],
    queryFn: async (): Promise<ListEmployeesResponse> => {
      const res = await unwrap(hrTransport.get<ListEmployeesResponse>('/hr/employees', { params }));
      return res as ListEmployeesResponse;
    },
  });
};

export const useGetEmployee = (id: number | undefined) => {
  return useQuery<EmployeePayload>({
    queryKey: ['employee', id],
    queryFn: async (): Promise<EmployeePayload> => {
      const res = await unwrap(
        hrTransport.get<{ employee: EmployeePayload }>(`/hr/employees/${id}`)
      );
      return (res as { employee: EmployeePayload }).employee;
    },
    enabled: !!id,
  });
};

export const useCreateEmployee = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: CreateEmployeePayload): Promise<CreateEmployeeResponse> => {
      const res = await unwrap(hrTransport.post<CreateEmployeeResponse>('/hr/employees', data));
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
    mutationFn: async (data: UpdateEmployeePayload): Promise<UpdateEmployeeResponse> => {
      const { id, ...body } = data;
      const res = await unwrap(
        hrTransport.put<UpdateEmployeeResponse>(`/hr/employees/${id}`, body)
      );
      return res as UpdateEmployeeResponse;
    },
    onSuccess: (_data: UpdateEmployeeResponse, variables: UpdateEmployeePayload) => {
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
      const source = data as InviteEmployeeRequest & {
        firstName?: string;
        lastName?: string;
        email?: string;
        phone?: string;
        employmentType?: string;
        departmentId?: number;
        locationId?: number;
      };
      const res = await unwrap(
        hrTransport.post<InviteEmployeeResponse>('/hr/employees/invite', {
          firstName: source.firstName,
          lastName: source.lastName,
          email: source.email,
          phone: source.phone,
          employmentType: source.employmentType,
          departmentId: source.departmentId,
          locationId: source.locationId,
        })
      );
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
