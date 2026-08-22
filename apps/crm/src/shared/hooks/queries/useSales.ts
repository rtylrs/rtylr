import { toOptionalFiniteNumber, unwrap } from '@voxire/api';
import {
  useMutation,
  useQuery,
  useQueryClient,
  type UseMutationResult,
  type UseQueryResult,
} from '@tanstack/react-query';
import { toast } from 'sonner';
import { safe } from '../../lib/api';
import { unwrapList } from '../../lib/unwrapList';
import {
  toContactMutationPayload,
  toLeadMutationPayload,
  toTouchpointMutationPayload,
} from '../../services/crmApi';
import type {
  PipelineColumn,
  Lead,
  Touchpoint,
  SalesContact,
  Contract,
  LeadStage,
  TouchpointType,
  TouchpointOutcome,
  CreateLeadRequest,
  UpdateLeadRequest,
  CreateTouchpointRequest,
  CreateContactRequest,
  UpdateContactRequest,
  CreateContractRequest,
  UpdateContractRequest,
  ContractPlan,
} from '../../services/crmApi';

// ── Re-exports ────────────────────────────────────────────────────

export type {
  PipelineColumn,
  Lead,
  Touchpoint,
  SalesContact,
  Contract,
  LeadStage,
  TouchpointType,
  TouchpointOutcome,
  ContractPlan,
} from '../../services/crmApi';

const CONTRACT_PLANS: readonly ContractPlan[] = ['free', 'pro', 'scale'];
const LEAD_STAGES: readonly LeadStage[] = [
  'cold',
  'contacted',
  'demo_booked',
  'demo_done',
  'proposal_sent',
  'contracted',
  'lost',
];
const TOUCHPOINT_TYPES: readonly TouchpointType[] = ['phone', 'email', 'meeting'];
const TOUCHPOINT_OUTCOMES: readonly TouchpointOutcome[] = [
  'positive',
  'neutral',
  'negative',
  'no_answer',
];

function asRecord(value: unknown): Record<string, unknown> {
  return value && typeof value === 'object' ? (value as Record<string, unknown>) : {};
}

function firstNonEmptyString(...values: unknown[]): string {
  for (const value of values) {
    const text = String(value ?? '').trim();
    if (text) return text;
  }
  return '';
}

function finiteNumber(value: unknown, fallback = 0): number {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : fallback;
}

function optionalNumber(value: unknown): number | undefined {
  if (value === null || value === undefined || value === '') return undefined;
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : undefined;
}

export function normalizeLead(value: unknown): Lead {
  const raw = asRecord(value);
  const company = firstNonEmptyString(raw.company, raw.companyName, raw.company_name);
  const companyName = firstNonEmptyString(raw.companyName, raw.company_name, raw.company);
  const stageValue = String(raw.stage ?? 'cold');
  const dealValue = optionalNumber(raw.dealValue ?? raw.deal_value);
  const mrrPotential = optionalNumber(raw.mrrPotential ?? raw.mrr_potential ?? dealValue);
  const assignedTo = raw.assignedTo ?? raw.assigned_to;

  return {
    id: finiteNumber(raw.id),
    company,
    companyName,
    industry: String(raw.industry ?? ''),
    size: String(raw.size ?? ''),
    source: String(raw.source ?? ''),
    stage: LEAD_STAGES.includes(stageValue as LeadStage) ? (stageValue as LeadStage) : 'cold',
    mrrPotential,
    dealValue,
    assignedTo: assignedTo == null ? null : optionalNumber(assignedTo),
    notes: String(raw.notes ?? ''),
    createdAt: String(raw.createdAt ?? raw.created_at ?? ''),
    updatedAt: String(raw.updatedAt ?? raw.updated_at ?? ''),
  };
}

export function normalizeTouchpoint(value: unknown): Touchpoint {
  const raw = asRecord(value);
  const type = String(raw.type ?? 'phone');
  const outcome = String(raw.outcome ?? 'neutral');
  const contactId = raw.contactId ?? raw.contact_id;
  const nextFollowUp = raw.nextFollowUp ?? raw.next_follow_up;
  const repName = raw.repName ?? raw.rep_name ?? raw.createdBy ?? raw.created_by;
  return {
    id: finiteNumber(raw.id),
    leadId: finiteNumber(raw.leadId ?? raw.lead_id),
    type: TOUCHPOINT_TYPES.includes(type as TouchpointType) ? (type as TouchpointType) : 'phone',
    outcome: TOUCHPOINT_OUTCOMES.includes(outcome as TouchpointOutcome)
      ? (outcome as TouchpointOutcome)
      : 'neutral',
    notes: String(raw.notes ?? ''),
    repName: repName == null ? '' : String(repName),
    contactId: contactId == null ? null : (optionalNumber(contactId) ?? null),
    nextFollowUp: nextFollowUp == null ? null : String(nextFollowUp),
    createdAt: String(raw.createdAt ?? raw.created_at ?? ''),
  };
}

export function normalizeSalesContact(value: unknown): SalesContact {
  const raw = asRecord(value);
  return {
    id: finiteNumber(raw.id),
    leadId: finiteNumber(raw.leadId ?? raw.lead_id),
    name: String(raw.name ?? ''),
    role: String(raw.role ?? ''),
    phone: String(raw.phone ?? ''),
    email: String(raw.email ?? ''),
    createdAt: String(raw.createdAt ?? raw.created_at ?? ''),
  };
}

export function normalizeContract(value: unknown): Contract {
  const raw = (value ?? {}) as Record<string, unknown>;
  const rawPlan = String(raw.plan ?? 'pro');
  const plan: ContractPlan = CONTRACT_PLANS.includes(rawPlan as ContractPlan)
    ? (rawPlan as ContractPlan)
    : 'pro';
  const closedBy = raw.closedBy ?? raw.closed_by;

  return {
    id: finiteNumber(raw.id),
    leadId: finiteNumber(raw.leadId ?? raw.lead_id),
    company: firstNonEmptyString(raw.company, raw.companyName, raw.company_name),
    plan,
    mrr: finiteNumber(raw.mrr),
    seats: finiteNumber(raw.seats),
    signedAt: String(raw.signedAt ?? raw.signed_at ?? ''),
    signedBy: String(raw.signedBy ?? raw.signedByName ?? raw.signed_by_name ?? ''),
    closedBy: closedBy == null ? '' : String(closedBy),
    notes: String(raw.notes ?? ''),
    createdAt: String(raw.createdAt ?? raw.created_at ?? ''),
  };
}

export function mergeContactUpdate(
  current: SalesContact | undefined,
  data: UpdateContactRequest
): UpdateContactRequest {
  return {
    name: data.name ?? current?.name ?? '',
    role: data.role ?? current?.role ?? '',
    phone: data.phone ?? current?.phone ?? '',
    email: data.email ?? current?.email ?? '',
  };
}

export function mergeContractUpdate(
  current: Contract | undefined,
  data: UpdateContractRequest
): UpdateContractRequest {
  return {
    plan: data.plan ?? current?.plan,
    mrr: data.mrr ?? current?.mrr,
    seats: data.seats ?? current?.seats,
    signedAt: data.signedAt ?? (current?.signedAt || undefined),
    signedBy: data.signedBy ?? current?.signedBy,
    closedBy: data.closedBy ?? current?.closedBy,
    notes: data.notes ?? current?.notes,
  };
}

export function toContractMutationPayload(
  data: CreateContractRequest | UpdateContractRequest
): Record<string, unknown> {
  return {
    ...('leadId' in data ? { leadId: data.leadId } : {}),
    plan: data.plan,
    mrr: data.mrr,
    seats: data.seats,
    signedAt: data.signedAt?.trim() || undefined,
    signedByName: data.signedBy?.trim(),
    closedBy: toOptionalFiniteNumber(data.closedBy),
    notes: data.notes?.trim() ?? '',
  };
}

const SALES_PAGE_SIZE = 100;

export function normalizeSalesPageSize(value: number): number {
  const parsed = Number(value);
  return Number.isFinite(parsed)
    ? Math.max(1, Math.min(Math.floor(parsed), SALES_PAGE_SIZE))
    : SALES_PAGE_SIZE;
}

function normalizeSalesOffset(value: number): number {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? Math.max(0, Math.floor(parsed)) : 0;
}

async function fetchLeadsPage(
  limit: number,
  offset: number
): Promise<{ leads: Lead[]; total?: number }> {
  const res = await unwrap(
    safe.get<unknown>('/crm/sales/leads', {
      params: {
        limit: normalizeSalesPageSize(limit),
        offset: normalizeSalesOffset(offset),
      },
    })
  );
  const leads = unwrapList<unknown>(res, 'leads').map(normalizeLead);
  const rawTotal = asRecord(res).total;
  const total = Number(rawTotal);
  return { leads, total: Number.isFinite(total) ? total : undefined };
}

// ── Pipeline ──────────────────────────────────────────────────────

export const usePipeline = (): UseQueryResult<PipelineColumn[]> =>
  useQuery<PipelineColumn[]>({
    queryKey: ['crm-sales-pipeline'],
    queryFn: async (): Promise<PipelineColumn[]> => {
      const res = await unwrap(safe.get<unknown>('/crm/sales/pipeline'));
      return unwrapList<Record<string, unknown>>(res, 'pipeline', 'columns').map((column) => ({
        stage: String(column.stage ?? 'cold') as PipelineColumn['stage'],
        leadCount: finiteNumber(column.leadCount ?? column.lead_count),
        totalMrr: finiteNumber(column.totalMrr ?? column.total_mrr),
      }));
    },
  });

// ── Leads ─────────────────────────────────────────────────────────

export const useLeads = (limit = 100, offset = 0): UseQueryResult<Lead[]> =>
  useQuery<Lead[]>({
    queryKey: ['crm-sales-leads', { limit, offset }],
    queryFn: async (): Promise<Lead[]> => (await fetchLeadsPage(limit, offset)).leads,
  });

/** Loads every lead page for the sales pipeline board. */
export const useListAllLeads = (): UseQueryResult<Lead[]> =>
  useQuery<Lead[]>({
    queryKey: ['crm-sales-leads-all'],
    queryFn: async (): Promise<Lead[]> => {
      const leads: Lead[] = [];
      let offset = 0;
      const maxPages = 100;

      for (let page = 0; page < maxPages; page += 1) {
        const result = await fetchLeadsPage(SALES_PAGE_SIZE, offset);
        leads.push(...result.leads);
        if (
          result.leads.length < SALES_PAGE_SIZE ||
          (result.total !== undefined && leads.length >= result.total)
        ) {
          break;
        }
        offset += SALES_PAGE_SIZE;
      }

      return leads;
    },
  });

export const useLead = (id: number | undefined): UseQueryResult<Lead> =>
  useQuery<Lead>({
    queryKey: ['crm-sales-lead', id],
    queryFn: async (): Promise<Lead> => {
      const res = await unwrap(safe.get<unknown>(`/crm/sales/leads/${id}`));
      const envelope = asRecord(res);
      return normalizeLead(envelope.lead ?? res);
    },
    enabled: id != null && Number.isInteger(id) && id > 0,
  });

export const useCreateLead = (): UseMutationResult<Lead, unknown, CreateLeadRequest> => {
  const qc = useQueryClient();
  return useMutation<Lead, unknown, CreateLeadRequest>({
    mutationFn: async (data: CreateLeadRequest): Promise<Lead> => {
      const res = await unwrap(safe.post<unknown>('/crm/sales/leads', toLeadMutationPayload(data)));
      const envelope = asRecord(res);
      return normalizeLead(envelope.lead ?? res);
    },
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-sales-leads'] });
      void qc.invalidateQueries({ queryKey: ['crm-sales-leads-all'] });
      void qc.invalidateQueries({ queryKey: ['crm-sales-pipeline'] });
      toast.success('Lead created');
    },
    onError: (): void => {
      toast.error('Failed to create lead');
    },
  });
};

interface UpdateLeadPayload {
  id: number;
  data: UpdateLeadRequest;
}

export const useUpdateLead = (): UseMutationResult<Lead, unknown, UpdateLeadPayload> => {
  const qc = useQueryClient();
  return useMutation<Lead, unknown, UpdateLeadPayload>({
    mutationFn: async (payload: UpdateLeadPayload): Promise<Lead> => {
      const cached =
        qc.getQueryData<Lead>(['crm-sales-lead', payload.id]) ??
        qc
          .getQueriesData<Lead[]>({ queryKey: ['crm-sales-leads'] })
          .flatMap(([, rows]) => rows ?? [])
          .find((lead) => lead.id === payload.id);
      const companyName =
        payload.data.companyName ??
        payload.data.company ??
        cached?.companyName ??
        cached?.company ??
        '';
      const merged: UpdateLeadRequest = {
        company: companyName,
        companyName,
        industry: payload.data.industry ?? cached?.industry,
        size: payload.data.size ?? cached?.size,
        source: payload.data.source ?? cached?.source,
        stage: payload.data.stage ?? cached?.stage,
        mrrPotential: payload.data.mrrPotential ?? cached?.mrrPotential ?? cached?.dealValue,
        assignedTo:
          payload.data.assignedTo ??
          (typeof cached?.assignedTo === 'number' ? cached.assignedTo : undefined),
        notes: payload.data.notes ?? cached?.notes,
      };
      const res = await unwrap(
        safe.put<unknown>(`/crm/sales/leads/${payload.id}`, toLeadMutationPayload(merged))
      );
      const envelope = asRecord(res);
      return normalizeLead(envelope.lead ?? res);
    },
    onSuccess: (_data: Lead, variables: UpdateLeadPayload): void => {
      void qc.invalidateQueries({ queryKey: ['crm-sales-leads'] });
      void qc.invalidateQueries({ queryKey: ['crm-sales-leads-all'] });
      void qc.invalidateQueries({ queryKey: ['crm-sales-lead', variables.id] });
      void qc.invalidateQueries({ queryKey: ['crm-sales-pipeline'] });
      toast.success('Lead updated');
    },
    onError: (): void => {
      toast.error('Failed to update lead');
    },
  });
};

export const useDeleteLead = (): UseMutationResult<unknown, unknown, number> => {
  const qc = useQueryClient();
  return useMutation<unknown, unknown, number>({
    mutationFn: (id: number): Promise<unknown> => unwrap(safe.delete(`/crm/sales/leads/${id}`)),
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-sales-leads'] });
      void qc.invalidateQueries({ queryKey: ['crm-sales-leads-all'] });
      void qc.invalidateQueries({ queryKey: ['crm-sales-pipeline'] });
      toast.success('Lead deleted');
    },
    onError: (): void => {
      toast.error('Failed to delete lead');
    },
  });
};

// ── Touchpoints ───────────────────────────────────────────────────

export const useLeadTouchpoints = (leadId: number | undefined): UseQueryResult<Touchpoint[]> =>
  useQuery<Touchpoint[]>({
    queryKey: ['crm-sales-touchpoints', leadId],
    queryFn: async (): Promise<Touchpoint[]> => {
      const res = await unwrap(safe.get<unknown>(`/crm/sales/leads/${leadId}/touchpoints`));
      return unwrapList<unknown>(res, 'touchpoints').map(normalizeTouchpoint);
    },
    enabled: leadId != null && Number.isInteger(leadId) && leadId > 0,
  });

export const useCreateTouchpoint = (): UseMutationResult<
  Touchpoint,
  unknown,
  CreateTouchpointRequest
> => {
  const qc = useQueryClient();
  return useMutation<Touchpoint, unknown, CreateTouchpointRequest>({
    mutationFn: async (data: CreateTouchpointRequest): Promise<Touchpoint> => {
      const res = await unwrap(
        safe.post<unknown>('/crm/sales/touchpoints', toTouchpointMutationPayload(data))
      );
      const envelope = asRecord(res);
      return normalizeTouchpoint(envelope.touchpoint ?? res);
    },
    onSuccess: (_data: Touchpoint, variables: CreateTouchpointRequest): void => {
      void qc.invalidateQueries({
        queryKey: ['crm-sales-touchpoints', variables.leadId],
      });
      toast.success('Touchpoint logged');
    },
    onError: (): void => {
      toast.error('Failed to log touchpoint');
    },
  });
};

export const useDeleteTouchpoint = (): UseMutationResult<unknown, unknown, number> => {
  const qc = useQueryClient();
  return useMutation<unknown, unknown, number>({
    mutationFn: (id: number): Promise<unknown> =>
      unwrap(safe.delete(`/crm/sales/touchpoints/${id}`)),
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-sales-touchpoints'] });
      toast.success('Touchpoint deleted');
    },
    onError: (): void => {
      toast.error('Failed to delete touchpoint');
    },
  });
};

// ── Contacts ──────────────────────────────────────────────────────

export const useContacts = (
  leadId?: number,
  limit = 100,
  offset = 0
): UseQueryResult<SalesContact[]> =>
  useQuery<SalesContact[]>({
    queryKey: ['crm-sales-contacts', { leadId: leadId ?? 'all', limit, offset }],
    queryFn: async (): Promise<SalesContact[]> => {
      const res = await unwrap(
        safe.get<unknown>('/crm/sales/contacts', {
          params: {
            ...(leadId != null && leadId > 0 ? { lead_id: leadId } : {}),
            limit,
            offset,
          },
        })
      );
      return unwrapList<unknown>(res, 'contacts').map(normalizeSalesContact);
    },
    enabled: leadId === undefined || (Number.isInteger(leadId) && leadId > 0),
  });

export const useCreateContact = (): UseMutationResult<
  SalesContact,
  unknown,
  CreateContactRequest
> => {
  const qc = useQueryClient();
  return useMutation<SalesContact, unknown, CreateContactRequest>({
    mutationFn: async (data: CreateContactRequest): Promise<SalesContact> => {
      const res = await unwrap(
        safe.post<unknown>('/crm/sales/contacts', toContactMutationPayload(data))
      );
      const envelope = asRecord(res);
      return normalizeSalesContact(envelope.contact ?? res);
    },
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-sales-contacts'] });
      toast.success('Contact added');
    },
    onError: (): void => {
      toast.error('Failed to add contact');
    },
  });
};

interface UpdateContactPayload {
  id: number;
  leadId: number;
  data: UpdateContactRequest;
}

export const useUpdateContact = (): UseMutationResult<
  SalesContact,
  unknown,
  UpdateContactPayload
> => {
  const qc = useQueryClient();
  return useMutation<SalesContact, unknown, UpdateContactPayload>({
    mutationFn: async (payload: UpdateContactPayload): Promise<SalesContact> => {
      const cached = qc
        .getQueriesData<SalesContact[]>({ queryKey: ['crm-sales-contacts'] })
        .flatMap(([, rows]) => rows ?? [])
        .find((contact) => contact.id === payload.id);
      const merged = mergeContactUpdate(cached, payload.data);
      const res = await unwrap(
        safe.put<unknown>(`/crm/sales/contacts/${payload.id}`, toContactMutationPayload(merged))
      );
      const envelope = asRecord(res);
      return normalizeSalesContact(envelope.contact ?? res);
    },
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-sales-contacts'] });
      toast.success('Contact updated');
    },
    onError: (): void => {
      toast.error('Failed to update contact');
    },
  });
};

interface DeleteContactPayload {
  id: number;
  leadId: number;
}

export const useDeleteContact = (): UseMutationResult<unknown, unknown, DeleteContactPayload> => {
  const qc = useQueryClient();
  return useMutation<unknown, unknown, DeleteContactPayload>({
    mutationFn: (payload: DeleteContactPayload): Promise<unknown> =>
      unwrap(safe.delete(`/crm/sales/contacts/${payload.id}`)),
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-sales-contacts'] });
      toast.success('Contact removed');
    },
    onError: (): void => {
      toast.error('Failed to remove contact');
    },
  });
};

// ── Contracts ─────────────────────────────────────────────────────

export const useContracts = (limit = 100, offset = 0): UseQueryResult<Contract[]> =>
  useQuery<Contract[]>({
    queryKey: ['crm-sales-contracts', { limit, offset }],
    queryFn: async (): Promise<Contract[]> => (await fetchContractsPage(limit, offset)).contracts,
  });

async function fetchContractsPage(
  limit: number,
  offset: number
): Promise<{ contracts: Contract[]; total?: number }> {
  const pageLimit = normalizeSalesPageSize(limit);
  const pageOffset = normalizeSalesOffset(offset);
  const res = await unwrap(
    safe.get<unknown>('/crm/sales/contracts', {
      params: { limit: pageLimit, offset: pageOffset },
    })
  );
  const contracts = unwrapList<Record<string, unknown>>(res, 'contracts').map(normalizeContract);
  const total = Number(asRecord(res).total);
  return { contracts, total: Number.isFinite(total) ? total : undefined };
}

/** Loads all contracts for detail lookups that must not stop at the first page. */
export const useListAllContracts = (): UseQueryResult<Contract[]> =>
  useQuery<Contract[]>({
    queryKey: ['crm-sales-contracts-all'],
    queryFn: async (): Promise<Contract[]> => {
      const contracts: Contract[] = [];
      let offset = 0;
      const maxPages = 100;

      for (let page = 0; page < maxPages; page += 1) {
        const result = await fetchContractsPage(SALES_PAGE_SIZE, offset);
        contracts.push(...result.contracts);
        if (
          result.contracts.length < SALES_PAGE_SIZE ||
          (result.total !== undefined && contracts.length >= result.total)
        ) {
          break;
        }
        offset += SALES_PAGE_SIZE;
      }

      return contracts;
    },
  });

export const useContract = (id: number | undefined): UseQueryResult<Contract> =>
  useQuery<Contract>({
    queryKey: ['crm-sales-contract', id],
    queryFn: async (): Promise<Contract> => {
      const res = await unwrap(safe.get<unknown>(`/crm/sales/contracts/${id}`));
      const envelope = (res ?? {}) as { contract?: unknown };
      return normalizeContract(envelope.contract ?? res);
    },
    enabled: id != null && Number.isInteger(id) && id > 0,
  });

export const useCreateContract = (): UseMutationResult<
  Contract,
  unknown,
  CreateContractRequest
> => {
  const qc = useQueryClient();
  return useMutation<Contract, unknown, CreateContractRequest>({
    mutationFn: async (data: CreateContractRequest): Promise<Contract> => {
      const res = await unwrap(
        safe.post<unknown>('/crm/sales/contracts', toContractMutationPayload(data))
      );
      const envelope = (res ?? {}) as { contract?: unknown };
      return normalizeContract(envelope.contract ?? res);
    },
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-sales-contracts'] });
      void qc.invalidateQueries({ queryKey: ['crm-sales-contracts-all'] });
      void qc.invalidateQueries({ queryKey: ['crm-sales-leads'] });
      toast.success('Contract created');
    },
    onError: (): void => {
      toast.error('Failed to create contract');
    },
  });
};

interface UpdateContractPayload {
  id: number;
  data: UpdateContractRequest;
}

export const useUpdateContract = (): UseMutationResult<
  Contract,
  unknown,
  UpdateContractPayload
> => {
  const qc = useQueryClient();
  return useMutation<Contract, unknown, UpdateContractPayload>({
    mutationFn: async (payload: UpdateContractPayload): Promise<Contract> => {
      const cached =
        qc.getQueryData<Contract>(['crm-sales-contract', payload.id]) ??
        qc
          .getQueriesData<Contract[]>({ queryKey: ['crm-sales-contracts'] })
          .flatMap(([, rows]) => rows ?? [])
          .find((contract) => contract.id === payload.id);
      const merged = mergeContractUpdate(cached, payload.data);
      const res = await unwrap(
        safe.put<unknown>(`/crm/sales/contracts/${payload.id}`, toContractMutationPayload(merged))
      );
      const envelope = (res ?? {}) as { contract?: unknown };
      return normalizeContract(envelope.contract ?? res);
    },
    onSuccess: (_data: Contract, variables: UpdateContractPayload): void => {
      void qc.invalidateQueries({ queryKey: ['crm-sales-contracts'] });
      void qc.invalidateQueries({ queryKey: ['crm-sales-contracts-all'] });
      void qc.invalidateQueries({ queryKey: ['crm-sales-contract', variables.id] });
      toast.success('Contract updated');
    },
    onError: (): void => {
      toast.error('Failed to update contract');
    },
  });
};
