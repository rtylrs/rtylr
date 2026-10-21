import {
  keepPreviousData,
  useInfiniteQuery,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query';
import { createOneKitClient } from '@voxire/api';
import {
  ActivityServiceClient,
  ContractServiceClient,
  DealServiceClient,
  LeadImportServiceClient,
  LeadServiceClient,
} from '@voxire/voxie/crm/sales/v1/client';
import type {
  ContractFields,
  CreateContractRequest,
  DealFields,
  ImportLeadRowsBody,
  LeadFields,
  LogActivityRequest,
  PipelineFields,
  SalesTaskFields,
  UpdateContractBody,
} from '@voxire/voxie/crm/sales/v1/types';
import i18next from 'i18next';
import { toast } from 'sonner';
import {
  type Billing,
  type DealSort,
  type DealStatus,
  type FollowUpWindow,
  type LeadSort,
  type LeadStatus,
  type Presence,
  type RepFilter,
  repFilterParam,
  todayISO,
  utcOffsetMinutes,
} from './model';

export const leadClient = createOneKitClient(LeadServiceClient, 'crm');
export const activityClient = createOneKitClient(ActivityServiceClient, 'crm');
export const dealClient = createOneKitClient(DealServiceClient, 'crm');
export const importClient = createOneKitClient(LeadImportServiceClient, 'crm');
export const contractClient = createOneKitClient(ContractServiceClient, 'crm');

const ROOT = 'crm-sales';
const PAGE = 50;

/** Finance prefills invoices from signed contracts and caches them here. */
const FINANCE_SIGNED_CONTRACTS = ['crm', 'sales-contracts', 'signed'];

// Known conflicts, by the server's message, so they read in the user's
// language. Anything else falls back to the caller's message.
const CONFLICTS: Record<string, string> = {
  'This lead is converted; its work continues on its deal': 'crmSales.errors.leadConverted',
  'This deal has a contract': 'crmSales.errors.dealHasContract',
  'This pipeline still has deals': 'crmSales.errors.pipelineInUse',
  'A removed stage still has deals': 'crmSales.errors.stageInUse',
  'This import is no longer running': 'crmSales.errors.importClosed',
  'Assignee is not a member of this business': 'crmSales.errors.badAssignee',
};

function serverMessage(error: unknown): string | undefined {
  const json = (error as { json?: unknown } | null)?.json;
  if (json && typeof json === 'object' && 'message' in json) {
    const message = (json as { message?: unknown }).message;
    if (typeof message === 'string') return message;
  }
  return undefined;
}

/** The message to show for a failed sales request. */
export function salesErrorMessage(error: unknown, fallbackKey: string): string {
  const message = serverMessage(error);
  const known = message ? CONFLICTS[message] : undefined;
  if (known) return i18next.t(known);
  const status = (error as { statusCode?: number } | null)?.statusCode;
  // Validation messages describe the input; the language may differ, but a
  // specific reason beats a generic one.
  if (status === 400 && message) return message;
  return i18next.t(fallbackKey);
}

function useSalesMutation<TInput, TResult>(
  run: (input: TInput) => Promise<TResult>,
  failureKey: string,
  options: { success?: string; alsoInvalidate?: unknown[][] } = {}
) {
  const qc = useQueryClient();
  return useMutation<TResult, unknown, TInput>({
    mutationFn: run,
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: [ROOT] });
      for (const key of options.alsoInvalidate ?? []) void qc.invalidateQueries({ queryKey: key });
      if (options.success) toast.success(i18next.t(options.success));
    },
    onError: (error) => {
      toast.error(salesErrorMessage(error, failureKey));
    },
  });
}

// ── Team ─────────────────────────────────────────────────────────────────

export interface SalesRep {
  id: string;
  name: string;
}

export function useReps() {
  return useQuery({
    queryKey: [ROOT, 'reps'],
    queryFn: async (): Promise<SalesRep[]> => {
      const res = await leadClient.listReps({});
      return (res.reps ?? []).map((rep) => ({
        id: rep.id ?? '',
        name: rep.name?.trim() || i18next.t('crmSales.team.unnamed', { id: rep.id }),
      }));
    },
    staleTime: 5 * 60_000,
  });
}

// ── Leads ────────────────────────────────────────────────────────────────

/**
 * The leads list's filters. Sets are empty when unused, ranges are the
 * inputs' text, and reps may hold "0" for unassigned leads.
 */
export interface LeadListFilters {
  statuses: LeadStatus[];
  reps: string[];
  sources: string[];
  industries: string[];
  sizes: string[];
  countries: string[];
  cities: string[];
  billing: Billing[];
  importId?: string;
  followUp: FollowUpWindow | '';
  hasPhone: Presence;
  hasEmail: Presence;
  hasWebsite: Presence;
  hasLinkedin: Presence;
  hasCompany: Presence;
  hasValue: Presence;
  contacted: Presence;
  attemptsMin: string;
  attemptsMax: string;
  valueMin: string;
  valueMax: string;
  createdFrom: string;
  createdTo: string;
  activeFrom: string;
  activeTo: string;
  search: string;
  sort: LeadSort;
}

const wholeNumber = (value: string): number | undefined =>
  /^\d+$/.test(value.trim()) ? Number(value.trim()) : undefined;
const amount = (value: string): string | undefined =>
  /^\d+(\.\d{1,2})?$/.test(value.trim()) ? value.trim() : undefined;

/**
 * The filter's query parameters, shared by the list and its stats. Text sets
 * go as JSON arrays so values with commas survive; half-typed ranges are left
 * out until they parse.
 */
export function leadFilterParams(filters: LeadListFilters) {
  const set = (values: string[]) => (values.length > 0 ? JSON.stringify(values) : undefined);
  return {
    reps: filters.reps.join(',') || undefined,
    sources: set(filters.sources),
    industries: set(filters.industries),
    sizes: set(filters.sizes),
    countries: set(filters.countries),
    cities: set(filters.cities),
    billing: filters.billing.join(',') || undefined,
    importId: filters.importId || undefined,
    followUp: filters.followUp || undefined,
    hasPhone: filters.hasPhone || undefined,
    hasEmail: filters.hasEmail || undefined,
    hasWebsite: filters.hasWebsite || undefined,
    hasLinkedin: filters.hasLinkedin || undefined,
    hasCompany: filters.hasCompany || undefined,
    hasValue: filters.hasValue || undefined,
    contacted: filters.contacted || undefined,
    attemptsMin: wholeNumber(filters.attemptsMin),
    attemptsMax: wholeNumber(filters.attemptsMax),
    valueMin: amount(filters.valueMin),
    valueMax: amount(filters.valueMax),
    createdFrom: filters.createdFrom || undefined,
    createdTo: filters.createdTo || undefined,
    activeFrom: filters.activeFrom || undefined,
    activeTo: filters.activeTo || undefined,
    search: filters.search.trim() || undefined,
    today: todayISO(),
    utcOffsetMinutes: utcOffsetMinutes(),
  };
}

export function useLeadList(filters: LeadListFilters, pageSize = PAGE) {
  const params = leadFilterParams(filters);
  return useInfiniteQuery({
    queryKey: [ROOT, 'leads', params, filters.statuses, filters.sort, pageSize],
    initialPageParam: '',
    queryFn: ({ pageParam }) =>
      leadClient.listLeads({
        ...params,
        statuses: filters.statuses.join(',') || undefined,
        sort: filters.sort,
        cursor: pageParam || undefined,
        limit: pageSize,
      }),
    getNextPageParam: (last) => last.nextCursor || undefined,
    placeholderData: keepPreviousData,
  });
}

/** Counts per status and the header stats, for the same filters (statuses aside). */
export function useLeadStats(filters: LeadListFilters) {
  const params = leadFilterParams(filters);
  return useQuery({
    queryKey: [ROOT, 'lead-stats', params],
    queryFn: () => leadClient.getLeadStats(params),
    placeholderData: keepPreviousData,
  });
}

/** The distinct sources, industries, sizes and places leads have. */
export function useLeadFacets() {
  return useQuery({
    queryKey: [ROOT, 'lead-facets'],
    queryFn: () => leadClient.listLeadFacets({}),
    staleTime: 60_000,
  });
}

export function useLead(id: string | undefined) {
  return useQuery({
    queryKey: [ROOT, 'lead', id],
    queryFn: async () => (await leadClient.getLead({ id })).lead ?? {},
    enabled: Boolean(id),
  });
}

export function useCreateLead() {
  return useSalesMutation(
    async (lead: LeadFields) => (await leadClient.createLead({ lead })).lead ?? {},
    'crmSales.errors.saveLead',
    { success: 'crmSales.leads.created' }
  );
}

export interface LeadEdit {
  id: string;
  lead: LeadFields;
  status?: string;
  disqualifyReason?: string;
}

export function useUpdateLead() {
  return useSalesMutation(
    async ({ id, lead, status, disqualifyReason }: LeadEdit) =>
      (await leadClient.updateLead({ id, body: { lead, status, disqualifyReason } })).lead ?? {},
    'crmSales.errors.saveLead',
    { success: 'crmSales.leads.saved' }
  );
}

export function useDeleteLead() {
  return useSalesMutation(
    (id: string) => leadClient.deleteLead({ id }),
    'crmSales.errors.deleteLead',
    { success: 'crmSales.leads.deleted' }
  );
}

export interface BulkLeadChange {
  ids: string[];
  assignTo?: string;
  unassign?: boolean;
  status?: LeadStatus;
  disqualifyReason?: string;
}

export function useUpdateLeads() {
  return useSalesMutation(
    (change: BulkLeadChange) => leadClient.updateLeads(change),
    'crmSales.errors.bulk'
  );
}

export function useDeleteLeads() {
  return useSalesMutation(
    (ids: string[]) => leadClient.deleteLeads({ ids }),
    'crmSales.errors.bulk'
  );
}

export interface ConvertInput {
  id: string;
  createDeal: boolean;
  deal?: DealFields;
}

export function useConvertLead() {
  return useSalesMutation(
    ({ id, createDeal, deal }: ConvertInput) =>
      leadClient.convertLead({ id, body: { createDeal, deal } }),
    'crmSales.errors.convert',
    { success: 'crmSales.convert.done' }
  );
}

// ── Activities and tasks ─────────────────────────────────────────────────

export function useActivities(subject: { leadId?: string; dealId?: string }) {
  return useInfiniteQuery({
    queryKey: [ROOT, 'activities', subject],
    initialPageParam: '',
    queryFn: ({ pageParam }) =>
      subject.dealId
        ? activityClient.listDealActivities({
            dealId: subject.dealId,
            cursor: pageParam || undefined,
            limit: 30,
          })
        : activityClient.listLeadActivities({
            leadId: subject.leadId,
            cursor: pageParam || undefined,
            limit: 30,
          }),
    getNextPageParam: (last) => last.nextCursor || undefined,
    enabled: Boolean(subject.leadId || subject.dealId),
  });
}

export function useLogActivity() {
  return useSalesMutation(
    (activity: LogActivityRequest) => activityClient.logActivity(activity),
    'crmSales.errors.logActivity',
    { success: 'crmSales.activity.logged' }
  );
}

export function useDeleteActivity() {
  return useSalesMutation(
    (id: string) => activityClient.deleteActivity({ id }),
    'crmSales.errors.deleteActivity'
  );
}

export interface TaskListFilters {
  scope: 'open' | 'completed';
  rep: RepFilter;
  due: FollowUpWindow | '';
  leadId?: string;
  dealId?: string;
}

export function useTaskList(filters: TaskListFilters, pageSize = PAGE) {
  return useInfiniteQuery({
    queryKey: [ROOT, 'tasks', filters, pageSize],
    initialPageParam: '',
    queryFn: ({ pageParam }) =>
      activityClient.listTasks({
        scope: filters.scope,
        assignedTo: repFilterParam(filters.rep),
        due: filters.due || undefined,
        leadId: filters.leadId,
        dealId: filters.dealId,
        cursor: pageParam || undefined,
        limit: pageSize,
        today: todayISO(),
      }),
    getNextPageParam: (last) => last.nextCursor || undefined,
    placeholderData: keepPreviousData,
  });
}

export function useTaskCounts(rep: RepFilter) {
  return useQuery({
    queryKey: [ROOT, 'task-counts', rep],
    queryFn: () =>
      activityClient.getTaskCounts({ assignedTo: repFilterParam(rep), today: todayISO() }),
  });
}

export function useCreateTask() {
  return useSalesMutation(
    (task: SalesTaskFields) => activityClient.createTask({ task }),
    'crmSales.errors.saveTask',
    { success: 'crmSales.tasks.created' }
  );
}

export function useCompleteTask() {
  return useSalesMutation(
    ({ id, done }: { id: string; done: boolean }) =>
      activityClient.completeTask({ id, body: { done } }),
    'crmSales.errors.saveTask'
  );
}

export function useDeleteTask() {
  return useSalesMutation(
    (id: string) => activityClient.deleteTask({ id }),
    'crmSales.errors.saveTask'
  );
}

// ── Pipelines and deals ──────────────────────────────────────────────────

export function usePipelines() {
  return useQuery({
    queryKey: [ROOT, 'pipelines'],
    queryFn: async () => (await dealClient.listPipelines({})).pipelines ?? [],
    staleTime: 60_000,
  });
}

export function useSavePipeline() {
  return useSalesMutation(
    ({ id, pipeline }: { id?: string; pipeline: PipelineFields }) =>
      id
        ? dealClient.updatePipeline({ id, body: pipeline })
        : dealClient.createPipeline({ pipeline }),
    'crmSales.errors.savePipeline',
    { success: 'crmSales.pipelines.saved' }
  );
}

export function useDeletePipeline() {
  return useSalesMutation(
    (id: string) => dealClient.deletePipeline({ id }),
    'crmSales.errors.deletePipeline',
    { success: 'crmSales.pipelines.deleted' }
  );
}

export interface BoardFilters {
  pipelineId?: string;
  rep: RepFilter;
  search: string;
}

export function useDealBoard(filters: BoardFilters) {
  return useQuery({
    queryKey: [ROOT, 'board', filters],
    queryFn: () =>
      dealClient.getDealBoard({
        pipelineId: filters.pipelineId,
        assignedTo: repFilterParam(filters.rep),
        search: filters.search.trim() || undefined,
      }),
    placeholderData: keepPreviousData,
  });
}

export function useStageDeals(stageId: string | undefined, filters: BoardFilters, pageSize = 20) {
  return useInfiniteQuery({
    queryKey: [ROOT, 'stage-deals', stageId, filters, pageSize],
    initialPageParam: '',
    queryFn: ({ pageParam }) =>
      dealClient.listStageDeals({
        stageId,
        assignedTo: repFilterParam(filters.rep),
        search: filters.search.trim() || undefined,
        cursor: pageParam || undefined,
        limit: pageSize,
      }),
    getNextPageParam: (last) => last.nextCursor || undefined,
    enabled: Boolean(stageId),
  });
}

export function useDealStats(pipelineId: string | undefined, rep: RepFilter) {
  return useQuery({
    queryKey: [ROOT, 'deal-stats', pipelineId, rep],
    queryFn: () =>
      dealClient.getDealStats({
        pipelineId,
        assignedTo: repFilterParam(rep),
        today: todayISO(),
        utcOffsetMinutes: utcOffsetMinutes(),
      }),
  });
}

/** The dashboard's windows, in days. */
export const ACTIVITY_WINDOWS = [7, 30, 90] as const;
export type ActivityWindow = (typeof ACTIVITY_WINDOWS)[number];

/** Calls, new leads and won deals per day over the window, with each rep's share. */
export function useSalesActivity(days: ActivityWindow, rep: RepFilter) {
  return useQuery({
    queryKey: [ROOT, 'activity', days, rep],
    queryFn: () =>
      dealClient.getSalesActivity({
        days,
        assignedTo: repFilterParam(rep),
        today: todayISO(),
        utcOffsetMinutes: utcOffsetMinutes(),
      }),
    placeholderData: keepPreviousData,
  });
}

export interface DealListFilters {
  pipelineId?: string;
  statuses: DealStatus[];
  rep: RepFilter;
  search: string;
  sort: DealSort;
}

export function useDealList(filters: DealListFilters, pageSize = PAGE) {
  return useInfiniteQuery({
    queryKey: [ROOT, 'deals', filters, pageSize],
    initialPageParam: '',
    queryFn: ({ pageParam }) =>
      dealClient.listDeals({
        pipelineId: filters.pipelineId,
        statuses: filters.statuses.join(',') || undefined,
        assignedTo: repFilterParam(filters.rep),
        search: filters.search.trim() || undefined,
        sort: filters.sort,
        cursor: pageParam || undefined,
        limit: pageSize,
      }),
    getNextPageParam: (last) => last.nextCursor || undefined,
    placeholderData: keepPreviousData,
  });
}

export function useDeal(id: string | undefined) {
  return useQuery({
    queryKey: [ROOT, 'deal', id],
    queryFn: () => dealClient.getDeal({ id }),
    enabled: Boolean(id),
  });
}

export function useCreateDeal() {
  return useSalesMutation(
    async (deal: DealFields) => (await dealClient.createDeal({ deal })).deal ?? {},
    'crmSales.errors.saveDeal',
    { success: 'crmSales.deals.created' }
  );
}

export function useUpdateDeal() {
  return useSalesMutation(
    ({ id, deal }: { id: string; deal: DealFields }) => dealClient.updateDeal({ id, body: deal }),
    'crmSales.errors.saveDeal',
    { success: 'crmSales.deals.saved' }
  );
}

export function useMoveDeal() {
  return useSalesMutation(
    ({ id, stageId }: { id: string; stageId: string }) =>
      dealClient.moveDeal({ id, body: { stageId } }),
    'crmSales.errors.moveDeal'
  );
}

export function useWinDeal() {
  return useSalesMutation(
    ({ id, contract }: { id: string; contract?: ContractFields }) =>
      dealClient.winDeal({ id, body: { contract } }),
    'crmSales.errors.winDeal',
    { success: 'crmSales.deals.won', alsoInvalidate: [FINANCE_SIGNED_CONTRACTS] }
  );
}

export function useLoseDeal() {
  return useSalesMutation(
    ({ id, reason }: { id: string; reason: string }) =>
      dealClient.loseDeal({ id, body: { reason } }),
    'crmSales.errors.loseDeal',
    { success: 'crmSales.deals.lost' }
  );
}

export function useReopenDeal() {
  return useSalesMutation(
    ({ id, stageId }: { id: string; stageId?: string }) =>
      dealClient.reopenDeal({ id, body: { stageId } }),
    'crmSales.errors.moveDeal',
    { success: 'crmSales.deals.reopened' }
  );
}

export function useDeleteDeal() {
  return useSalesMutation(
    (id: string) => dealClient.deleteDeal({ id }),
    'crmSales.errors.deleteDeal',
    { success: 'crmSales.deals.deleted' }
  );
}

// ── Imports ──────────────────────────────────────────────────────────────

export function useLeadImports() {
  return useQuery({
    queryKey: [ROOT, 'imports'],
    queryFn: async () => (await importClient.listLeadImports({})).imports ?? [],
  });
}

export function useUndoImport() {
  return useSalesMutation(
    (id: string) => importClient.undoLeadImport({ id }),
    'crmSales.errors.undoImport'
  );
}

export async function sendImportChunk(id: string, body: ImportLeadRowsBody) {
  return importClient.importLeadRows({ id, body });
}

// ── Contracts ────────────────────────────────────────────────────────────

export function useContracts(signed: boolean | undefined, offset: number, limit = 25) {
  return useQuery({
    queryKey: [ROOT, 'contracts', signed, offset, limit],
    queryFn: () => contractClient.listContracts({ signed, limit, offset }),
    placeholderData: keepPreviousData,
  });
}

export function useCreateContract() {
  return useSalesMutation(
    (contract: CreateContractRequest) => contractClient.createContract(contract),
    'crmSales.errors.saveContract',
    { success: 'crmSales.contracts.created', alsoInvalidate: [FINANCE_SIGNED_CONTRACTS] }
  );
}

export function useUpdateContract() {
  return useSalesMutation(
    ({ id, body }: { id: string; body: UpdateContractBody }) =>
      contractClient.updateContract({ id, body }),
    'crmSales.errors.saveContract',
    { success: 'crmSales.contracts.saved', alsoInvalidate: [FINANCE_SIGNED_CONTRACTS] }
  );
}
