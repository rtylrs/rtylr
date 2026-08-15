import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { createOneKitClient } from '@voxire/api';
import { TimeServiceClient } from '@voxire/voxie/flow/time/v1/client';
import type { WorkDay, WorkDayLine, WorkType } from '@voxire/voxie/flow/time/v1/types';
import { toast } from 'sonner';
import i18next from '../../shared/services/i18next';

const timeClient = createOneKitClient(TimeServiceClient, 'flow');

const t = (key: string): string => i18next.t(key) as string;

export type { WorkDay, WorkDayLine, WorkType };

/** The day the browser is in, as YYYY-MM-DD, in the person's own timezone. */
export function todayLocalDate(): string {
  const now = new Date();
  const month = String(now.getMonth() + 1).padStart(2, '0');
  const day = String(now.getDate()).padStart(2, '0');
  return `${now.getFullYear()}-${month}-${day}`;
}

/**
 * The IANA zone this browser is in. Sent with every write so the day is stored
 * against the zone it was recorded in: a travelling worker's day boundary is
 * their own, and the zone cannot be recovered after the fact.
 */
export function localTimezone(): string {
  return Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC';
}

/**
 * A key that identifies this line attempt, so a retry cannot create a second
 * entry. Generated before the request leaves, which is the whole point: a key
 * the server assigns cannot make the client's retry idempotent.
 */
export function newClientKey(): string {
  if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) {
    return crypto.randomUUID();
  }
  return `${Date.now()}-${Math.random().toString(36).slice(2)}`;
}

const workDayKey = (date: string): [string, string] => ['work-day', date];

export function useWorkDay(date: string) {
  return useQuery({
    queryKey: workDayKey(date),
    queryFn: () => timeClient.getWorkDay({ date }),
    enabled: !!date,
  });
}

export function useWorkTypes() {
  return useQuery({
    queryKey: ['work-types'],
    queryFn: () => timeClient.listWorkTypes({}),
  });
}

/**
 * The caller's own days across a range, for the week strip. No new endpoint:
 * ListWorkDays already takes from/to, so this is the same call the single-day
 * view could have made, just with a wider range.
 */
export function useWorkDaysRange(from: string, to: string) {
  return useQuery({
    queryKey: ['work-days-range', from, to],
    queryFn: () => timeClient.listWorkDays({ from, to, limit: 7 }),
    enabled: !!from && !!to,
  });
}

/**
 * Every member's work day for one date, for the team view. Requires at least
 * admin on the workspace - enforced server-side, not just hidden in the UI:
 * a caller without it gets a real 403 from this same call.
 */
export function useWorkDaysForWorkspace(workspaceId: string | undefined, date: string) {
  return useQuery({
    queryKey: ['work-days-team', workspaceId, date],
    queryFn: () =>
      timeClient.listWorkDaysForWorkspace({
        workspaceId: workspaceId!,
        from: date,
        to: date,
        limit: 200,
      }),
    enabled: !!workspaceId && !!date,
  });
}

interface UpsertDayInput {
  date: string;
  workMode: string;
  startTime?: string;
  endTime?: string;
  breakMinutes: number;
}

export function useUpsertWorkDay() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: UpsertDayInput) =>
      timeClient.upsertWorkDay({
        date: input.date,
        timezone: localTimezone(),
        workMode: input.workMode,
        startTime: input.startTime,
        endTime: input.endTime,
        breakMinutes: input.breakMinutes,
      }),
    onSuccess: (_data, input) => {
      void queryClient.invalidateQueries({ queryKey: workDayKey(input.date) });
    },
    onError: () => toast.error(t('flow.time.daySaveFailed')),
  });
}

interface CreateLineInput {
  date: string;
  workDayId: string;
  clientKey: string;
  durationMinutes: number;
  description: string;
  projectId?: string;
  taskId?: string;
  workTypeId?: string;
  workMode?: string;
  isBillable: boolean;
}

export function useCreateWorkDayLine() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateLineInput) =>
      timeClient.createWorkDayLine({
        workDayId: input.workDayId,
        clientKey: input.clientKey,
        durationMinutes: input.durationMinutes,
        description: input.description,
        projectId: input.projectId,
        taskId: input.taskId,
        workTypeId: input.workTypeId,
        workMode: input.workMode,
        isBillable: input.isBillable,
      }),
    onSuccess: (_data, input) => {
      void queryClient.invalidateQueries({ queryKey: workDayKey(input.date) });
    },
    onError: () => toast.error(t('flow.time.lineSaveFailed')),
  });
}

export function useDeleteWorkDayLine() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id }: { id: string; date: string }) => timeClient.deleteWorkDayLine({ id }),
    onSuccess: (_data, input) => {
      void queryClient.invalidateQueries({ queryKey: workDayKey(input.date) });
    },
    onError: () => toast.error(t('flow.time.lineDeleteFailed')),
  });
}
