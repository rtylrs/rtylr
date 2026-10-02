import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { EmployeePayload } from '@voxire/voxie/hr/employee/v1/types';
import type { ListEmployeesResponse } from '@voxire/voxie/hr/employee/v1/types';
import type { CreateEmployeeResponse } from '@voxire/voxie/hr/employee/v1/types';
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
    queryKey: ['hr', 'employees', 'filter', params],
    queryFn: async (): Promise<ListEmployeesResponse> => {
      const res = await unwrap(hrTransport.get<ListEmployeesResponse>('/hr/employees', { params }));
      return res as ListEmployeesResponse;
    },
  });
};

export const useGetEmployee = (id: number | undefined) => {
  return useQuery<EmployeePayload>({
    queryKey: ['hr', 'employee', id],
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
  const { t } = useTranslation('hrCommon');

  return useMutation({
    mutationFn: async (data: CreateEmployeePayload): Promise<CreateEmployeeResponse> => {
      const res = await unwrap(hrTransport.post<CreateEmployeeResponse>('/hr/employees', data));
      return res as CreateEmployeeResponse;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['hr', 'employees'] });
      toast.success(t('hrEmployees.created'));
    },
    onError: () => {
      toast.error(t('hrErrors.generic'));
    },
  });
};
