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

// /hr/employees returns hr.staff.v1.Staff (firstName/lastName/employmentType/startDate)
// while older callers typed it as EmployeePayload (jobTitle/contractType/hireDate).
export interface EmployeeRecord extends EmployeePayload {
  firstName?: string;
  lastName?: string;
  email?: string;
  employmentType?: string;
  startDate?: string;
  salary?: number;
  displayName?: string;
}

export interface NormalizedListEmployeesResponse extends Omit<ListEmployeesResponse, 'employees'> {
  employees: EmployeeRecord[];
}

export const KNOWN_CONTRACT_TYPES = new Set([
  'full_time',
  'part_time',
  'freelancer',
  'contractor',
  'intern',
]);

export const KNOWN_EMPLOYEE_STATUSES = new Set([
  'active',
  'probation',
  'inactive',
  'terminated',
  'invited',
]);

function cleanDate(value?: string): string {
  if (!value || value.startsWith('0001-01-01')) return '';
  return value;
}

export function normalizeEmployeeRecord(raw: EmployeeRecord): EmployeeRecord {
  const fullName = `${raw.firstName ?? ''} ${raw.lastName ?? ''}`.trim();
  const displayName = fullName || raw.jobTitle?.trim() || raw.email?.trim() || '';
  const contractType = raw.contractType || raw.employmentType || '';
  const hireDate = cleanDate(raw.hireDate || raw.startDate);
  const monthlySalary =
    raw.monthlySalary !== undefined
      ? raw.monthlySalary
      : raw.salary !== undefined && raw.salary > 0
        ? raw.salary
        : undefined;

  return {
    ...raw,
    displayName,
    jobTitle: raw.jobTitle || displayName,
    contractType,
    hireDate,
    monthlySalary,
  };
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
  return useQuery<NormalizedListEmployeesResponse>({
    queryKey: ['hr', 'employees', 'filter', params],
    queryFn: async (): Promise<NormalizedListEmployeesResponse> => {
      const res = await unwrap(
        hrTransport.get<NormalizedListEmployeesResponse>('/hr/employees', { params })
      );
      const body = res as NormalizedListEmployeesResponse;
      return {
        ...body,
        employees: (body.employees ?? []).map(normalizeEmployeeRecord),
      };
    },
  });
};

export const useGetEmployee = (id: number | undefined) => {
  return useQuery<EmployeeRecord>({
    queryKey: ['hr', 'employee', id],
    queryFn: async (): Promise<EmployeeRecord> => {
      const res = await unwrap(
        hrTransport.get<{ employee: EmployeeRecord }>(`/hr/employees/${id}`)
      );
      return normalizeEmployeeRecord((res as { employee: EmployeeRecord }).employee);
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
