import { unwrap } from '@voxire/api';
import i18next from 'i18next';
import {
  useMutation,
  useQuery,
  useQueryClient,
  type UseMutationResult,
  type UseQueryResult,
} from '@tanstack/react-query';
import { toast } from 'sonner';
import { crmApi, toFiniteNumber } from '../../services/crmApi';
import { unwrapList } from '../../lib/unwrapList';
import type {
  PipelineColumn,
  Lead,
  LeadStage,
  Touchpoint,
  SalesContact,
  Contract,
  CreateLeadRequest,
  UpdateLeadRequest,
  CreateTouchpointRequest,
  CreateContactRequest,
  CreateContractRequest,
  SalesRep,
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
  SalesRep,
} from '../../services/crmApi';

// ── Pipeline ──────────────────────────────────────────────────────

export const usePipeline = (): UseQueryResult<PipelineColumn[]> =>
  useQuery<PipelineColumn[]>({
    queryKey: ['crm-sales-pipeline'],
    queryFn: async (): Promise<PipelineColumn[]> => {
      const res = await unwrap(crmApi.getSalesPipeline());
      // total_mrr arrives as a decimal string (money.Amount) — coerce or
      // stage sums concatenate strings.
      return unwrapList<PipelineColumn>(res, 'pipeline', 'columns').map((column) => {
        const raw = column as unknown as Record<string, unknown>;
        return {
          ...column,
          count: toFiniteNumber(raw.count ?? raw.cnt),
          totalMrr: toFiniteNumber(raw.totalMrr ?? raw.total_mrr),
        };
      });
    },
  });

// ── Leads ─────────────────────────────────────────────────────────

export const parseLead = (lead: Lead | (Record<string, unknown> & { id: unknown })): Lead => {
  const raw = lead as Record<string, unknown>;
  const createdAt = String(raw.createdAt ?? raw.created_at ?? '');
  const updatedAt = String(raw.updatedAt ?? raw.updated_at ?? createdAt);
  const companyName = String(raw.companyName ?? raw.company_name ?? raw.company ?? '');
  const company = companyName;
  const rawMrr = raw.mrrPotential ?? raw.mrr_potential ?? raw.dealValue ?? raw.deal_value;
  const mrrPotential = rawMrr === undefined || rawMrr === null ? undefined : toFiniteNumber(rawMrr);
  const dealValue = mrrPotential;
  const rawAssigned = raw.assignedTo ?? raw.assigned_to;
  const assignedTo =
    rawAssigned === null || rawAssigned === undefined ? null : toFiniteNumber(rawAssigned);

  return {
    ...(lead as Lead),
    id: toFiniteNumber(raw.id),
    company,
    companyName,
    stage: (raw.stage as LeadStage) ?? 'cold',
    industry: typeof raw.industry === 'string' ? raw.industry : '',
    size: typeof raw.size === 'string' ? raw.size : '',
    source: typeof raw.source === 'string' ? raw.source : '',
    notes: typeof raw.notes === 'string' ? raw.notes : '',
    assignedTo,
    mrrPotential,
    dealValue,
    createdAt,
    updatedAt,
    nextFollowUp:
      typeof (raw.nextFollowUp ?? raw.next_follow_up) === 'string' &&
      String(raw.nextFollowUp ?? raw.next_follow_up) !== ''
        ? String(raw.nextFollowUp ?? raw.next_follow_up)
        : undefined,
  };
};

/** Teammates a lead can be assigned to, for the rep picker and rep names. */
export const useSalesReps = (): UseQueryResult<SalesRep[]> =>
  useQuery<SalesRep[]>({
    queryKey: ['crm-sales-reps'],
    queryFn: async (): Promise<SalesRep[]> => {
      const res = await unwrap(crmApi.getSalesReps());
      return unwrapList<Record<string, unknown>>(res, 'reps').map((rep) => ({
        id: toFiniteNumber(rep.id),
        name: String(rep.name ?? ''),
      }));
    },
    staleTime: 5 * 60 * 1000,
  });

export const useLeads = (limit = 100, offset = 0): UseQueryResult<Lead[]> =>
  useQuery<Lead[]>({
    queryKey: ['crm-sales-leads', { limit, offset }],
    queryFn: async (): Promise<Lead[]> => {
      const res = await unwrap(crmApi.getLeads(limit, offset));
      return unwrapList<Lead>(res, 'leads').map(parseLead);
    },
  });

export const useLead = (id: number | undefined): UseQueryResult<Lead> =>
  useQuery<Lead>({
    queryKey: ['crm-sales-lead', id],
    queryFn: async (): Promise<Lead> => {
      const res = await unwrap(crmApi.getLead(id ?? 0));
      const envelope = (res ?? {}) as { lead?: Lead };
      return parseLead((envelope.lead ?? res) as Lead);
    },
    enabled: !!id,
  });

export const useCreateLead = (): UseMutationResult<Lead, unknown, CreateLeadRequest> => {
  const qc = useQueryClient();
  return useMutation<Lead, unknown, CreateLeadRequest>({
    mutationFn: async (data: CreateLeadRequest): Promise<Lead> => {
      const res = await unwrap(crmApi.createLead(data));
      const envelope = (res ?? {}) as { lead?: Lead };
      return parseLead((envelope.lead ?? res) as Lead);
    },
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-sales-leads'] });
      void qc.invalidateQueries({ queryKey: ['crm-sales-pipeline'] });
      toast.success(i18next.t('crm.sales.leadCreated'));
    },
    onError: (): void => {
      toast.error(i18next.t('crm.createLead.createFailed'));
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
      // The backend PUT replaces the record; merging over the cached lead
      // stops a stage-only update from wiping company/industry/etc. to "".
      let cached = qc.getQueryData<Lead>(['crm-sales-lead', payload.id]);
      if (!cached) {
        const leadQueries = qc.getQueriesData<Lead[]>({ queryKey: ['crm-sales-leads'] });
        for (const [, leads] of leadQueries) {
          const match = leads?.find((l) => l.id === payload.id);
          if (match) {
            cached = match;
            break;
          }
        }
      }
      const merged: UpdateLeadRequest = {
        company: payload.data.company ?? cached?.company,
        companyName: payload.data.companyName ?? cached?.companyName ?? cached?.company,
        industry: payload.data.industry ?? cached?.industry,
        size: payload.data.size ?? cached?.size,
        source: payload.data.source ?? cached?.source,
        stage: payload.data.stage ?? cached?.stage,
        mrrPotential: payload.data.mrrPotential ?? cached?.mrrPotential,
        // Explicit null unassigns; only an omitted rep keeps the cached one.
        assignedTo:
          payload.data.assignedTo !== undefined
            ? payload.data.assignedTo
            : (cached?.assignedTo ?? undefined),
        notes: payload.data.notes ?? cached?.notes,
      };
      const res = await unwrap(crmApi.updateLead(payload.id, merged));
      const envelope = (res ?? {}) as { lead?: Lead };
      return parseLead((envelope.lead ?? res) as Lead);
    },
    onSuccess: (_data: Lead, variables: UpdateLeadPayload): void => {
      void qc.invalidateQueries({ queryKey: ['crm-sales-leads'] });
      void qc.invalidateQueries({ queryKey: ['crm-sales-lead', variables.id] });
      void qc.invalidateQueries({ queryKey: ['crm-sales-pipeline'] });
      toast.success(i18next.t('crm.sales.leadUpdated'));
    },
    onError: (): void => {
      toast.error(i18next.t('crm.sales.leadUpdateFailed'));
    },
  });
};

export const useDeleteLead = (): UseMutationResult<unknown, unknown, number> => {
  const qc = useQueryClient();
  return useMutation<unknown, unknown, number>({
    mutationFn: (id: number): Promise<unknown> => unwrap(crmApi.deleteLead(id)),
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-sales-leads'] });
      void qc.invalidateQueries({ queryKey: ['crm-sales-pipeline'] });
      toast.success(i18next.t('crm.leads.deleted'));
    },
    onError: (): void => {
      toast.error(i18next.t('crm.leadDetail.deleteFailed'));
    },
  });
};

// ── Touchpoints ───────────────────────────────────────────────────

export const useLeadTouchpoints = (leadId: number | undefined): UseQueryResult<Touchpoint[]> =>
  useQuery<Touchpoint[]>({
    queryKey: ['crm-sales-touchpoints', leadId],
    queryFn: async (): Promise<Touchpoint[]> => {
      const res = await unwrap(crmApi.getLeadTouchpoints(leadId ?? 0));
      return unwrapList<Touchpoint>(res, 'touchpoints');
    },
    enabled: !!leadId,
  });

export const useCreateTouchpoint = (): UseMutationResult<
  Touchpoint,
  unknown,
  CreateTouchpointRequest
> => {
  const qc = useQueryClient();
  return useMutation<Touchpoint, unknown, CreateTouchpointRequest>({
    mutationFn: async (data: CreateTouchpointRequest): Promise<Touchpoint> => {
      const res = await unwrap(crmApi.createTouchpoint(data));
      return res as Touchpoint;
    },
    onSuccess: (_data: Touchpoint, variables: CreateTouchpointRequest): void => {
      void qc.invalidateQueries({
        queryKey: ['crm-sales-touchpoints', variables.leadId],
      });
      toast.success(i18next.t('crm.sales.touchpointLogged'));
    },
    onError: (): void => {
      toast.error(i18next.t('crm.leadDetail.touchpointFailed'));
    },
  });
};

export const useDeleteTouchpoint = (): UseMutationResult<unknown, unknown, number> => {
  const qc = useQueryClient();
  return useMutation<unknown, unknown, number>({
    mutationFn: (id: number): Promise<unknown> => unwrap(crmApi.deleteTouchpoint(id)),
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-sales-touchpoints'] });
      toast.success(i18next.t('crm.sales.touchpointDeleted'));
    },
    onError: (): void => {
      toast.error(i18next.t('crm.sales.touchpointDeleteFailed'));
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
      const res = await unwrap(crmApi.getContacts(leadId, limit, offset));
      return unwrapList<SalesContact>(res, 'contacts');
    },
  });

export const useCreateContact = (): UseMutationResult<
  SalesContact,
  unknown,
  CreateContactRequest
> => {
  const qc = useQueryClient();
  return useMutation<SalesContact, unknown, CreateContactRequest>({
    mutationFn: async (data: CreateContactRequest): Promise<SalesContact> => {
      const res = await unwrap(crmApi.createContact(data));
      return res as SalesContact;
    },
    onSuccess: (_data: SalesContact, _variables: CreateContactRequest): void => {
      void qc.invalidateQueries({ queryKey: ['crm-sales-contacts'] });
      toast.success(i18next.t('crm.sales.contactAdded'));
    },
    onError: (): void => {
      toast.error(i18next.t('crm.leadDetail.addContactFailed'));
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
      unwrap(crmApi.deleteContact(payload.id)),
    onSuccess: (_data: unknown, _variables: DeleteContactPayload): void => {
      void qc.invalidateQueries({ queryKey: ['crm-sales-contacts'] });
      toast.success(i18next.t('crm.sales.contactRemoved'));
    },
    onError: (): void => {
      toast.error(i18next.t('crm.sales.contactRemoveFailed'));
    },
  });
};

// ── Contracts ─────────────────────────────────────────────────────

const parseContract = (contract: Contract): Contract => ({
  ...contract,
  id: toFiniteNumber(contract.id),
  leadId: toFiniteNumber(contract.leadId),
  mrr: toFiniteNumber(contract.mrr),
  seats: toFiniteNumber(contract.seats),
});

export const useContracts = (limit = 100, offset = 0): UseQueryResult<Contract[]> =>
  useQuery<Contract[]>({
    queryKey: ['crm-sales-contracts', { limit, offset }],
    queryFn: async (): Promise<Contract[]> => {
      const res = await unwrap(crmApi.getContracts(limit, offset));
      // mrr arrives as a decimal string (money.Amount) — coerce or totals
      // concatenate strings.
      return unwrapList<Contract>(res, 'contracts').map(parseContract);
    },
  });

export const useCreateContract = (): UseMutationResult<
  Contract,
  unknown,
  CreateContractRequest
> => {
  const qc = useQueryClient();
  return useMutation<Contract, unknown, CreateContractRequest>({
    mutationFn: async (data: CreateContractRequest): Promise<Contract> => {
      const res = await unwrap(crmApi.createContract(data));
      const envelope = (res ?? {}) as { contract?: Contract };
      return parseContract((envelope.contract ?? res) as Contract);
    },
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-sales-contracts'] });
      void qc.invalidateQueries({ queryKey: ['crm-sales-leads'] });
      toast.success(i18next.t('crm.sales.contractCreated'));
    },
    onError: (): void => {
      toast.error(i18next.t('crm.contracts.createFailed'));
    },
  });
};
