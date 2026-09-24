import { createOneKitClient, unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { DepartmentServiceClient } from '@voxire/voxie/hr/department/v1/client';
import { WorkScheduleServiceClient } from '@voxire/voxie/hr/work_schedule/v1/client';
import type {
  GetSetupStatusResponse,
  HRCompanyConfigPayload,
} from '@voxire/voxie/hr/setup/v1/types';
import { hrTransport } from '../../lib/clients';
import { useHRConfigStore } from '../../stores/hrConfigStore';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

type SetupStatus = Omit<GetSetupStatusResponse, 'config'> & {
  config: HRCompanyConfigPayload | null;
};

const workScheduleClient = createOneKitClient(WorkScheduleServiceClient, 'hr');
const departmentClient = createOneKitClient(DepartmentServiceClient, 'hr');

type SetupStepData = Record<string, unknown>;

function requiredString(data: SetupStepData, key: string): string {
  const value = data[key];
  if (typeof value !== 'string' || value.trim() === '') {
    throw new Error(`${key} is required`);
  }
  return value.trim();
}

function optionalString(data: SetupStepData, key: string): string | undefined {
  const value = data[key];
  return typeof value === 'string' && value.trim() ? value.trim() : undefined;
}

function requiredNumber(data: SetupStepData, key: string): number {
  const value = Number(data[key]);
  if (!Number.isFinite(value)) throw new Error(`${key} is required`);
  return value;
}

async function persistSetupStepData(stepKey: string, data?: SetupStepData): Promise<void> {
  if (!data) return;

  switch (stepKey) {
    case 'company_profile': {
      await unwrap(
        hrTransport.put('/hr/config', {
          companyName: requiredString(data, 'name'),
          defaultCurrency: requiredString(data, 'currency'),
          defaultTimezone: requiredString(data, 'timezone'),
        })
      );
      return;
    }
    case 'work_schedule': {
      const workingDays = data.workingDays;
      if (!Array.isArray(workingDays) || workingDays.length === 0) {
        throw new Error('workingDays is required');
      }
      const response = await workScheduleClient.createWorkSchedule({
        name: requiredString(data, 'name'),
        workingDays: workingDays.filter((day): day is string => typeof day === 'string'),
        startTime: requiredString(data, 'startTime'),
        endTime: requiredString(data, 'endTime'),
        breakMinutes: requiredNumber(data, 'breakMinutes'),
        contractedHoursPerWeek: requiredNumber(data, 'contractedHoursPerWeek'),
        dailyOtThresholdMinutes: requiredNumber(data, 'dailyOtThresholdMinutes'),
        weeklyOtThresholdMinutes: requiredNumber(data, 'weeklyOtThresholdMinutes'),
      });
      const scheduleId = response.schedule?.id;
      if (scheduleId) {
        await unwrap(hrTransport.put('/hr/config', { defaultWorkScheduleId: scheduleId }));
      }
      return;
    }
    case 'org_structure': {
      const departments = data.departments;
      if (!Array.isArray(departments)) throw new Error('departments is required');
      await Promise.all(
        departments.map((department) => {
          if (!department || typeof department !== 'object') {
            throw new Error('department is invalid');
          }
          return departmentClient.createDepartment({
            name: requiredString(department as SetupStepData, 'name'),
          });
        })
      );
      return;
    }
    case 'locations': {
      const locations = data.locations;
      if (!Array.isArray(locations)) throw new Error('locations is required');
      await Promise.all(
        locations.map((location, index) => {
          if (!location || typeof location !== 'object') {
            throw new Error('location is invalid');
          }
          const payload = location as SetupStepData;
          return unwrap(
            hrTransport.post('/hr/locations', {
              name: requiredString(payload, 'name'),
              address: requiredString(payload, 'address'),
              timezone: optionalString(payload, 'timezone'),
              isDefault: index === 0,
            })
          );
        })
      );
      return;
    }
    case 'payroll': {
      const period = requiredString(data, 'payrollPeriod');
      if (period !== 'weekly' && period !== 'biweekly' && period !== 'monthly') {
        throw new Error('payrollPeriod is invalid');
      }
      if (typeof data.enableOvertime !== 'boolean') {
        throw new Error('enableOvertime is required');
      }
      await unwrap(
        hrTransport.put('/hr/payroll/config', {
          periodType: period,
          payDayOfMonth: requiredNumber(data, 'payDayOfMonth'),
          overtimeEnabled: data.enableOvertime,
        })
      );
      return;
    }
    case 'notifications':
      // There is no HR notification-preference contract. The step is a
      // confirmation step now; never send a payload the server would discard.
      return;
    case 'roles_permissions':
    case 'employees':
      return;
    default:
      return;
  }
}

export const useGetSetupStatus = () => {
  return useQuery<SetupStatus>({
    queryKey: ['hr-setup-status'],
    queryFn: async (): Promise<SetupStatus> => {
      const [status, config] = await Promise.all([
        unwrap(hrTransport.get<Omit<GetSetupStatusResponse, 'config'>>('/hr/setup/status')),
        unwrap(hrTransport.get<{ config: HRCompanyConfigPayload | null }>('/hr/config')),
      ]);
      return { ...status, config: config.config };
    },
  });
};

export type CompleteStepInput = {
  stepKey: string;
  data?: Record<string, unknown>;
};

export const useCompleteSetupStep = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async ({ stepKey, data }: CompleteStepInput): Promise<void> => {
      await persistSetupStepData(stepKey, data);
      await unwrap(hrTransport.post(`/hr/setup/steps/${stepKey}/complete`, {}));
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['hr-setup-status'] });
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useGetHRConfig = () => {
  return useQuery<HRCompanyConfigPayload>({
    queryKey: ['hr-config'],
    queryFn: async (): Promise<HRCompanyConfigPayload> => {
      const res = await unwrap(hrTransport.get<{ config: HRCompanyConfigPayload }>('/hr/config'));
      return (res as { config: HRCompanyConfigPayload }).config;
    },
  });
};

export const useUpsertHRConfig = () => {
  const queryClient = useQueryClient();
  const setConfig = useHRConfigStore((s) => s.setConfig);
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: Partial<HRCompanyConfigPayload>): Promise<HRCompanyConfigPayload> => {
      const res = await unwrap(
        hrTransport.put<{ config: HRCompanyConfigPayload }>('/hr/config', data)
      );
      return (res as { config: HRCompanyConfigPayload }).config;
    },
    onSuccess: (config: HRCompanyConfigPayload) => {
      setConfig(config);
      queryClient.setQueryData(['hr-config'], config);
      queryClient.invalidateQueries({ queryKey: ['hr-config'] });
      toast.success(t('setup.configSaved'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};
