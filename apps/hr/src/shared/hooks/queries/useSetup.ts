import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { GetSetupStatusResponse } from '@voxire/voxie/hr/setup/v1/types';
import { safe } from '../../lib/api';
import { useHRConfigStore } from '../../stores/hrConfigStore';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import type { HRConfig } from '../../types/hrConfig';
import type {
  LocationsFormData,
  NotificationsFormData,
  OrgStructureFormData,
  RolesPermissionsFormData,
  WorkScheduleFormData,
} from '../../schemas/wizardSchemas';

export const useGetSetupStatus = () => {
  return useQuery<GetSetupStatusResponse>({
    queryKey: ['hr-setup-status'],
    queryFn: async (): Promise<GetSetupStatusResponse> => {
      const res = await unwrap(safe.get<GetSetupStatusResponse>('/hr/setup/status'));
      return res as GetSetupStatusResponse;
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
      await unwrap(safe.post(`/hr/setup/steps/${stepKey}/complete`, data ?? {}));
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
  return useQuery<HRConfig>({
    queryKey: ['hr-config'],
    queryFn: async (): Promise<HRConfig> => {
      const res = await unwrap(safe.get<{ config?: HRConfig }>('/hr/config'));
      const config = (res as { config?: HRConfig }).config ?? (res as unknown as HRConfig);
      return {
        ...config,
        timeTrackingModel: config.timeTrackingModel ?? 'both',
        clockInMethods: config.clockInMethods ?? ['app'],
        subscribedApps: config.subscribedApps ?? [],
      };
    },
  });
};

export const useUpsertHRConfig = () => {
  const queryClient = useQueryClient();
  const setConfig = useHRConfigStore((s) => s.setConfig);
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: HRConfigUpdate): Promise<HRConfig> => {
      const res = await unwrap(
        safe.put<{ config: HRConfig }>('/hr/config', {
          companyName: data.companyName,
          fiscalYearStartMonth: data.fiscalYearStartMonth ?? 1,
          defaultCurrency: data.defaultCurrency ?? 'USD',
          defaultTimezone: data.defaultTimezone ?? 'UTC',
          defaultWorkScheduleId: data.defaultWorkScheduleId ?? 0,
          overtimeEnabled: data.overtimeEnabled ?? false,
          overtimeMultiplier: data.overtimeMultiplier ?? 1.5,
          emailNotificationsEnabled: data.emailNotificationsEnabled ?? true,
          smsNotificationsEnabled: data.smsNotificationsEnabled ?? false,
        })
      );
      return (res as { config: HRConfig }).config;
    },
    onSuccess: (config: HRConfig) => {
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

export interface HRLocation {
  id: number;
  name: string;
  isDefault: boolean;
}

export const useListHRLocations = () => {
  return useQuery<HRLocation[]>({
    queryKey: ['hr-locations'],
    queryFn: async (): Promise<HRLocation[]> => {
      const response = await unwrap(safe.get<unknown>('/hr/locations'));
      const records = Array.isArray(response)
        ? response
        : ((response as { locations?: unknown[] } | null)?.locations ?? []);
      return records.flatMap((record) => {
        if (record === null || typeof record !== 'object') return [];
        const value = record as { id?: number | string; name?: string; is_default?: boolean };
        const id = Number(value.id);
        if (!Number.isInteger(id) || id <= 0) return [];
        return [{ id, name: value.name ?? `#${id}`, isDefault: value.is_default ?? false }];
      });
    },
  });
};

export interface HRConfigUpdate {
  companyName: string;
  fiscalYearStartMonth?: number;
  defaultCurrency?: string;
  defaultTimezone?: string;
  defaultWorkScheduleId?: number;
  overtimeEnabled?: boolean;
  overtimeMultiplier?: number;
  emailNotificationsEnabled?: boolean;
  smsNotificationsEnabled?: boolean;
}

type SetupRecord = {
  id?: number;
  name?: string;
  description?: string;
  permissions?: string;
  address?: string;
  city?: string;
  country?: string;
  timezone?: string;
  isDefault?: boolean;
  scheduleData?: string;
  locationId?: number;
  parentDepartmentId?: number;
  defaultWorkScheduleId?: number;
};

const recordsFrom = (value: unknown, key: string): SetupRecord[] => {
  if (Array.isArray(value)) return value as SetupRecord[];
  if (value !== null && typeof value === 'object') {
    const records = (value as Record<string, unknown>)[key];
    if (Array.isArray(records)) return records as SetupRecord[];
  }
  return [];
};

const recordFrom = (value: unknown, key: string): SetupRecord | undefined => {
  if (value !== null && typeof value === 'object') {
    const record = (value as Record<string, unknown>)[key];
    if (record !== null && typeof record === 'object') return record as SetupRecord;
  }
  return undefined;
};

const sameName = (left: string | undefined, right: string): boolean =>
  left?.trim().toLocaleLowerCase() === right.trim().toLocaleLowerCase();

const requireRecordId = (record: SetupRecord | undefined, label: string): number => {
  if (!record?.id || record.id <= 0) throw new Error(`${label} response did not include an id`);
  return record.id;
};

const readHRConfig = async (): Promise<HRConfig | undefined> => {
  const cached = useHRConfigStore.getState().config;
  if (cached?.companyName) return cached;

  const response = await unwrap(safe.get<{ config?: HRConfig }>('/hr/config'));
  const config = response.config ?? (response as unknown as HRConfig);
  return config.companyName ? config : undefined;
};

const persistWorkSchedule = async (data: WorkScheduleFormData): Promise<HRConfig | void> => {
  const listed = await unwrap(safe.get<unknown>('/hr/work-schedules'));
  const existing = recordsFrom(listed, 'workSchedules').find((item) =>
    sameName(item.name, data.name)
  );
  const payload = {
    name: data.name.trim(),
    scheduleData: JSON.stringify({
      workingDays: data.workingDays,
      startTime: data.startTime,
      endTime: data.endTime,
      breakMinutes: data.breakMinutes,
      contractedHoursPerWeek: data.contractedHoursPerWeek,
      dailyOtThresholdMinutes: data.dailyOtThresholdMinutes,
      weeklyOtThresholdMinutes: data.weeklyOtThresholdMinutes,
    }),
    isDefault: true,
  };
  const response = existing?.id
    ? await unwrap(safe.put<unknown>(`/hr/work-schedules/${existing.id}`, payload))
    : await unwrap(safe.post<unknown>('/hr/work-schedules', payload));
  const schedule = recordFrom(response, 'workSchedule');
  const scheduleId = requireRecordId(schedule ?? existing, 'work schedule');

  const config = await readHRConfig();
  if (!config) return;

  const saved = await unwrap(
    safe.put<{ config: HRConfig }>('/hr/config', {
      companyName: config.companyName,
      fiscalYearStartMonth: config.fiscalYearStartMonth ?? 1,
      defaultCurrency: config.defaultCurrency ?? 'USD',
      defaultTimezone: config.defaultTimezone ?? 'UTC',
      defaultWorkScheduleId: scheduleId,
      overtimeEnabled: config.overtimeEnabled ?? false,
      overtimeMultiplier: config.overtimeMultiplier ?? 1.5,
      emailNotificationsEnabled: config.emailNotificationsEnabled ?? true,
      smsNotificationsEnabled: config.smsNotificationsEnabled ?? false,
    })
  );
  return saved.config ?? { ...config, defaultWorkScheduleId: scheduleId };
};

const persistDepartments = async (
  departments: OrgStructureFormData['departments']
): Promise<void> => {
  const [listed, schedules] = await Promise.all([
    unwrap(safe.get<unknown>('/hr/core/departments')),
    unwrap(safe.get<unknown>('/hr/work-schedules')),
  ]);
  const existingDepartments = recordsFrom(listed, 'departments');
  const defaultSchedule = recordsFrom(schedules, 'workSchedules').find((item) => item.isDefault);

  for (const department of departments) {
    const existing = existingDepartments.find((item) => sameName(item.name, department.name));
    const payload = {
      name: department.name.trim(),
      locationId: department.locationId ?? existing?.locationId ?? 0,
      parentDepartmentId: department.parentDepartmentId ?? existing?.parentDepartmentId ?? 0,
      defaultWorkScheduleId: existing?.defaultWorkScheduleId ?? defaultSchedule?.id ?? 0,
    };
    if (existing?.id) {
      await unwrap(safe.put(`/hr/core/departments/${existing.id}`, payload));
    } else {
      await unwrap(safe.post('/hr/core/departments', payload));
    }
  }
};

const persistLocations = async (data: LocationsFormData): Promise<void> => {
  const listed = await unwrap(safe.get<unknown>('/hr/locations'));
  const existingLocations = recordsFrom(listed, 'locations');
  const timezone = useHRConfigStore.getState().config?.defaultTimezone ?? 'UTC';

  for (const [index, location] of data.locations.entries()) {
    const existing = existingLocations.find((item) => sameName(item.name, location.name));
    const payload = {
      name: location.name.trim(),
      address: location.address.trim(),
      city: existing?.city ?? '',
      country: existing?.country ?? '',
      timezone: location.timezone ?? existing?.timezone ?? timezone,
      isDefault: index === 0,
    };
    if (existing?.id) {
      await unwrap(safe.put(`/hr/locations/${existing.id}`, payload));
    } else {
      await unwrap(safe.post('/hr/locations', payload));
    }
  }
};

const persistRoles = async (roles: RolesPermissionsFormData['roles']): Promise<void> => {
  const listed = await unwrap(safe.get<unknown>('/hr/roles'));
  const existingRoles = recordsFrom(listed, 'roles');

  for (const role of roles) {
    const existing = existingRoles.find((item) => sameName(item.name, role.name));
    const payload = {
      name: role.name.trim(),
      description: existing?.description ?? '',
      permissions: JSON.stringify(
        role.permissions
          .filter((permission) => permission.isGranted)
          .map((permission) => permission.key)
      ),
    };
    if (existing?.id) {
      await unwrap(safe.put(`/hr/roles/${existing.id}`, payload));
    } else {
      await unwrap(safe.post('/hr/roles', payload));
    }
  }
};

const persistNotifications = async (data: NotificationsFormData): Promise<HRConfig> => {
  const response = await readHRConfig();
  if (!response?.companyName) throw new Error('HR configuration is required before notifications');

  const saved = await unwrap(
    safe.put<{ config: HRConfig }>('/hr/config', {
      companyName: response.companyName,
      fiscalYearStartMonth: response.fiscalYearStartMonth ?? 1,
      defaultCurrency: response.defaultCurrency ?? 'USD',
      defaultTimezone: response.defaultTimezone ?? 'UTC',
      defaultWorkScheduleId: Number(response.defaultWorkScheduleId ?? 0),
      overtimeEnabled: response.overtimeEnabled ?? false,
      overtimeMultiplier: response.overtimeMultiplier ?? 1.5,
      emailNotificationsEnabled: data.emailEnabled,
      smsNotificationsEnabled: data.smsEnabled,
    })
  );
  return saved.config;
};

export type PersistSetupStepInput =
  | { stepKey: 'work_schedule'; data: WorkScheduleFormData }
  | { stepKey: 'org_structure'; data: OrgStructureFormData['departments'] }
  | { stepKey: 'locations'; data: LocationsFormData }
  | { stepKey: 'roles_permissions'; data: RolesPermissionsFormData['roles'] }
  | { stepKey: 'notifications'; data: NotificationsFormData };

export const usePersistSetupStep = () => {
  const queryClient = useQueryClient();
  const setConfig = useHRConfigStore((state) => state.setConfig);
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (input: PersistSetupStepInput): Promise<HRConfig | void> => {
      switch (input.stepKey) {
        case 'work_schedule':
          return persistWorkSchedule(input.data);
        case 'org_structure':
          await persistDepartments(input.data);
          return;
        case 'locations':
          await persistLocations(input.data);
          return;
        case 'roles_permissions':
          await persistRoles(input.data);
          return;
        case 'notifications':
          return persistNotifications(input.data);
      }
    },
    onSuccess: (config) => {
      if (config) {
        setConfig(config);
        queryClient.setQueryData(['hr-config'], config);
      }
      queryClient.invalidateQueries({ queryKey: ['hr-config'] });
      queryClient.invalidateQueries({ queryKey: ['hr-work-schedules'] });
      queryClient.invalidateQueries({ queryKey: ['hr-locations'] });
      queryClient.invalidateQueries({ queryKey: ['hr-departments'] });
      queryClient.invalidateQueries({ queryKey: ['hr-roles'] });
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};
