import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { WorkDay, WorkDayLine, WorkType } from '@voxire/voxie/flow/time/v2/types';
import { toast } from 'sonner';
import { timeV2 } from '../../shared/lib/clients';
import i18next from '../../shared/services/i18next';

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
    queryFn: () => timeV2.getWorkDay({ date }),
    enabled: !!date,
  });
}

const workTypesKey = ['work-types'];

export function useWorkTypes(includeArchived = false) {
  return useQuery({
    queryKey: [...workTypesKey, includeArchived],
    queryFn: () => timeV2.listWorkTypes({ includeArchived }),
  });
}

interface WorkTypeInput {
  name: string;
  description: string;
}

export function useCreateWorkType() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: WorkTypeInput) =>
      timeV2.createWorkType({ name: input.name, description: input.description }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: workTypesKey });
    },
    onError: () => toast.error(t('flow.settings.workTypes.saveFailed')),
  });
}

export function useUpdateWorkType() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: WorkTypeInput & { id: string; archived: boolean }) =>
      timeV2.updateWorkType({
        id: input.id,
        name: input.name,
        description: input.description,
        archived: input.archived,
      }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: workTypesKey });
    },
    onError: () => toast.error(t('flow.settings.workTypes.saveFailed')),
  });
}

export function useDeleteWorkType() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => timeV2.deleteWorkType({ id }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: workTypesKey });
    },
    onError: () => toast.error(t('flow.settings.workTypes.deleteFailed')),
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
    queryFn: () => timeV2.listWorkDays({ fromDate: from, toDate: to, limit: 7 }),
    enabled: !!from && !!to,
  });
}

/**
 * Every member's work days across a range, for the team view and reports.
 * Requires at least admin on the workspace - enforced server-side, not just
 * hidden in the UI: a caller without it gets a real 403 from this same call.
 */
// Business-scoped, not workspace-scoped. v1's listWorkDaysForWorkspace required a
// workspace id, and activeWorkspaceId is '' outside the /w/ routes - so on
// /time/team and /time/reports the query was permanently disabled and every row
// read as not-started. v2's listTeamWorkDays takes no workspace at all.
export function useTeamWorkDays(from: string, to: string) {
  return useQuery({
    queryKey: ['work-days-team', from, to],
    queryFn: () => timeV2.listTeamWorkDays({ fromDate: from, toDate: to, limit: 500 }),
    enabled: !!from && !!to,
  });
}

interface UpsertDayInput {
  date: string;
  workMode: string;
  startTime?: string;
  endTime?: string;
  breakMinutes: number;
}

interface MutationOptions {
  /**
   * Suppresses the default toast. Add Hours shows its own inline failure
   * banner per its design (never a toast on top of a persistent banner), so
   * it opts into silence and handles the error itself; every other caller
   * keeps today's toast behaviour unchanged.
   */
  silent?: boolean;
}

export function useUpsertWorkDay(options?: MutationOptions) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: UpsertDayInput) =>
      timeV2.upsertWorkDay({
        date: input.date,
        timezone: localTimezone(),
        workMode: input.workMode,
        startTime: input.startTime,
        endTime: input.endTime,
        breakMinutes: input.breakMinutes,
      }),
    onSuccess: (_data, input) => {
      void queryClient.invalidateQueries({ queryKey: workDayKey(input.date) });
      void queryClient.invalidateQueries({ queryKey: ['work-days-range'] });
    },
    onError: () => {
      if (!options?.silent) toast.error(t('flow.time.daySaveFailed'));
    },
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

export function useCreateWorkDayLine(options?: MutationOptions) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateLineInput) =>
      timeV2.createWorkDayLine({
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
      void queryClient.invalidateQueries({ queryKey: ['work-days-range'] });
    },
    onError: () => {
      if (!options?.silent) toast.error(t('flow.time.lineSaveFailed'));
    },
  });
}

export function useDeleteWorkDayLine() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id }: { id: string; date: string }) => timeV2.deleteWorkDayLine({ id }),
    onSuccess: (_data, input) => {
      void queryClient.invalidateQueries({ queryKey: workDayKey(input.date) });
      void queryClient.invalidateQueries({ queryKey: ['work-days-range'] });
    },
    onError: () => toast.error(t('flow.time.lineDeleteFailed')),
  });
}
