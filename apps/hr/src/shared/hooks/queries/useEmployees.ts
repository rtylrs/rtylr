import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { EmployeeFormData } from '../../schemas/employeeSchema';
import { safe } from '../../lib/api';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

export interface EmployeeRecord {
  id: number;
  ownerId: number;
  userId: number;
  departmentId: number;
  locationId: number;
  workScheduleId: number;
  managerId: number;
  firstName: string;
  lastName: string;
  email: string;
  phone: string;
  employmentType: string;
  status: string;
  jobTitle: string;
  currency: string;
  accountCreationMode: string;
  salary: number;
  hourlyRate: number;
  overtimeRate: number;
  startDate: string;
  dateOfBirth?: string;
  probationEndDate?: string;
  endDate?: string;
  terminationDate?: string;
  address: string;
  emergencyContactName: string;
  emergencyContactPhone: string;
  notes: string;
  createdAt: string;
  updatedAt: string;
}

export interface EmployeeListResponse {
  employees: EmployeeRecord[];
  total: number;
  limit: number;
  offset: number;
}

export interface EmployeeMutationResponse {
  employee: EmployeeRecord;
}

export interface InviteEmployeePayload {
  firstName: string;
  lastName: string;
  email: string;
  phone?: string;
  employmentType?: string;
  departmentId?: number;
  locationId?: number;
}

export interface InviteEmployeeResult {
  token: string;
}

interface ListEmployeesParams {
  limit: number;
  offset: number;
  search?: string;
  departmentId?: number;
  locationId?: number;
  status?: string;
}

interface ListEmployeesOptions {
  enabled?: boolean;
}

export interface CreateEmployeePayload {
  userId: number;
  departmentId: number;
  locationId: number;
  workScheduleId?: number;
  managerId?: number;
  firstName: string;
  lastName: string;
  email: string;
  phone?: string;
  employmentType: string;
  status?: string;
  jobTitle: string;
  currency: string;
  accountCreationMode?: string;
  address?: string;
  salary?: number;
  hourlyRate?: number;
  overtimeRate?: number;
  startDate: string;
  dateOfBirth?: string;
  probationEndDate?: string;
  endDate?: string;
  terminationDate?: string;
}

export interface UpdateEmployeePayload extends Partial<CreateEmployeePayload> {
  id: number;
  status?: string;
  endDate?: string;
}

const asRecord = (value: unknown): Record<string, unknown> =>
  value && typeof value === 'object' ? (value as Record<string, unknown>) : {};

const valueOf = (row: Record<string, unknown>, snake: string, camel: string): unknown =>
  row[snake] ?? row[camel];

const toNumber = (value: unknown): number => {
  const parsed = Number(value ?? 0);
  return Number.isFinite(parsed) ? parsed : 0;
};

export const decodeEmployee = (value: unknown): EmployeeRecord => {
  const row = asRecord(value);
  return {
    id: toNumber(valueOf(row, 'id', 'id')),
    ownerId: toNumber(valueOf(row, 'owner_id', 'ownerId')),
    userId: toNumber(valueOf(row, 'user_id', 'userId')),
    departmentId: toNumber(valueOf(row, 'department_id', 'departmentId')),
    locationId: toNumber(valueOf(row, 'location_id', 'locationId')),
    workScheduleId: toNumber(valueOf(row, 'work_schedule_id', 'workScheduleId')),
    managerId: toNumber(valueOf(row, 'manager_id', 'managerId')),
    firstName: String(valueOf(row, 'first_name', 'firstName') ?? ''),
    lastName: String(valueOf(row, 'last_name', 'lastName') ?? ''),
    email: String(valueOf(row, 'email', 'email') ?? ''),
    phone: String(valueOf(row, 'phone', 'phone') ?? ''),
    employmentType: String(valueOf(row, 'employment_type', 'employmentType') ?? ''),
    status: String(valueOf(row, 'status', 'status') ?? ''),
    jobTitle: String(valueOf(row, 'job_title', 'jobTitle') ?? ''),
    currency: String(valueOf(row, 'currency', 'currency') ?? ''),
    accountCreationMode: String(valueOf(row, 'account_creation_mode', 'accountCreationMode') ?? ''),
    salary: toNumber(valueOf(row, 'salary', 'salary')),
    hourlyRate: toNumber(valueOf(row, 'hourly_rate', 'hourlyRate')),
    overtimeRate: toNumber(valueOf(row, 'overtime_rate', 'overtimeRate')),
    startDate: String(valueOf(row, 'start_date', 'startDate') ?? ''),
    dateOfBirth: String(valueOf(row, 'date_of_birth', 'dateOfBirth') ?? '') || undefined,
    probationEndDate:
      String(valueOf(row, 'probation_end_date', 'probationEndDate') ?? '') || undefined,
    endDate: String(valueOf(row, 'end_date', 'endDate') ?? '') || undefined,
    terminationDate:
      String(valueOf(row, 'termination_date', 'terminationDate') ?? '') || undefined,
    address: String(valueOf(row, 'address', 'address') ?? ''),
    emergencyContactName: String(
      valueOf(row, 'emergency_contact_name', 'emergencyContactName') ?? ''
    ),
    emergencyContactPhone: String(
      valueOf(row, 'emergency_contact_phone', 'emergencyContactPhone') ?? ''
    ),
    notes: String(valueOf(row, 'notes', 'notes') ?? ''),
    createdAt: String(valueOf(row, 'created_at', 'createdAt') ?? ''),
    updatedAt: String(valueOf(row, 'updated_at', 'updatedAt') ?? ''),
  };
};

const encodeEmployeePayload = (
  data: Partial<CreateEmployeePayload>
): Record<string, unknown> => {
  const row: Record<string, unknown> = {};
  const set = (key: string, value: unknown): void => {
    if (value !== undefined) row[key] = value;
  };
  set('user_id', data.userId);
  set('department_id', data.departmentId);
  set('location_id', data.locationId);
  set('work_schedule_id', data.workScheduleId);
  set('manager_id', data.managerId);
  set('first_name', data.firstName);
  set('last_name', data.lastName);
  set('email', data.email);
  set('phone', data.phone);
  set('employment_type', data.employmentType);
  set('status', data.status);
  set('job_title', data.jobTitle);
  set('currency', data.currency);
  set('account_creation_mode', data.accountCreationMode);
  set('salary', data.salary);
  set('hourly_rate', data.hourlyRate);
  set('overtime_rate', data.overtimeRate);
  set('start_date', data.startDate);
  set('date_of_birth', data.dateOfBirth);
  set('probation_end_date', data.probationEndDate);
  set('end_date', data.endDate);
  set('termination_date', data.terminationDate);
  set('address', data.address);
  return row;
};

const decodeEmployeeResponse = (value: unknown): EmployeeMutationResponse => {
  const row = asRecord(value);
  return { employee: decodeEmployee(row.employee) };
};

const decodeEmployeeList = (value: unknown): EmployeeListResponse => {
  const row = asRecord(value);
  return {
    employees: Array.isArray(row.employees) ? row.employees.map(decodeEmployee) : [],
    total: toNumber(row.total),
    limit: toNumber(row.limit),
    offset: toNumber(row.offset),
  };
};

export const toCreateEmployeePayload = (data: EmployeeFormData): CreateEmployeePayload => ({
  userId: data.userId,
  departmentId: data.departmentId,
  locationId: data.locationId,
  workScheduleId: data.workScheduleId,
  managerId: data.managerId,
  firstName: data.firstName,
  lastName: data.lastName,
  email: data.email,
  employmentType: data.contractType,
  status: data.status,
  jobTitle: data.jobTitle,
  currency: data.currency,
  accountCreationMode: data.accountCreationMode,
  salary: data.monthlySalary,
  hourlyRate: data.hourlyRate,
  startDate: data.hireDate,
  dateOfBirth: data.dateOfBirth,
  probationEndDate: data.probationEndDate,
  terminationDate: data.terminationDate,
});

export const useListEmployees = (params: ListEmployeesParams, options?: ListEmployeesOptions) => {
  return useQuery<EmployeeListResponse>({
    queryKey: ['employees', 'filter', params],
    queryFn: async (): Promise<EmployeeListResponse> => {
      const res = await unwrap(safe.get<unknown>('/hr/employees', { params }));
      return decodeEmployeeList(res);
    },
    enabled: options?.enabled ?? true,
  });
};

export const useGetEmployee = (id: number | undefined) => {
  return useQuery<EmployeeRecord>({
    queryKey: ['employee', id],
    queryFn: async (): Promise<EmployeeRecord> => {
      const res = await unwrap(safe.get<unknown>(`/hr/employees/${id}`));
      return decodeEmployeeResponse(res).employee;
    },
    enabled: !!id,
  });
};

export const useCreateEmployee = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: CreateEmployeePayload): Promise<EmployeeMutationResponse> => {
      const res = await unwrap(safe.post<unknown>('/hr/employees', encodeEmployeePayload(data)));
      return decodeEmployeeResponse(res);
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
    mutationFn: async (data: UpdateEmployeePayload): Promise<EmployeeMutationResponse> => {
      const { id, ...body } = data;
      const res = await unwrap(
        safe.put<unknown>(`/hr/employees/${id}`, encodeEmployeePayload(body))
      );
      return decodeEmployeeResponse(res);
    },
    onSuccess: (_data: EmployeeMutationResponse, variables: UpdateEmployeePayload) => {
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
    mutationFn: async (data: InviteEmployeePayload): Promise<InviteEmployeeResult> => {
      const res = await unwrap(safe.post<InviteEmployeeResult>('/hr/employees/invite', data));
      return res as InviteEmployeeResult;
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
