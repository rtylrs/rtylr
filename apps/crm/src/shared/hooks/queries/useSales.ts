import { unwrap } from '@voxire/api';
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
  toContractUpdatePayload,
  toCreateContractPayload,
  toLeadMutationPayload,
} from '../../services/crmApi';
import type {
  PipelineColumn,
  Lead,
  Touchpoint,
  SalesContact,
  Contract,
  CreateLeadRequest,
  UpdateLeadRequest,
  CreateTouchpointRequest,
  CreateContactRequest,
  UpdateContactRequest,
  CreateContractRequest,
  UpdateContractRequest,
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

// ── Pipeline ──────────────────────────────────────────────────────

export const usePipeline = (): UseQueryResult<PipelineColumn[]> =>
  useQuery<PipelineColumn[]>({
    queryKey: ['crm-sales-pipeline'],
    queryFn: async (): Promise<PipelineColumn[]> => {
      const res = await unwrap(safe.get<unknown>('/crm/sales/pipeline'));
      return unwrapList<PipelineColumn>(res, 'pipeline', 'columns');
    },
  });

// ── Leads ─────────────────────────────────────────────────────────

export const useLeads = (limit = 100, offset = 0): UseQueryResult<Lead[]> =>
  useQuery<Lead[]>({
    queryKey: ['crm-sales-leads', { limit, offset }],
    queryFn: async (): Promise<Lead[]> => {
      const res = await unwrap(
        safe.get<unknown>('/crm/sales/leads', { params: { limit, offset } })
      );
      return unwrapList<Lead>(res, 'leads');
    },
  });

export const useLead = (id: number | undefined): UseQueryResult<Lead> =>
  useQuery<Lead>({
    queryKey: ['crm-sales-lead', id],
    queryFn: async (): Promise<Lead> => {
      const res = await unwrap(safe.get<unknown>(`/crm/sales/leads/${id}`));
      const envelope = (res ?? {}) as { lead?: Lead };
      return (envelope.lead ?? res) as Lead;
    },
    enabled: !!id,
  });

export const useCreateLead = (): UseMutationResult<Lead, unknown, CreateLeadRequest> => {
  const qc = useQueryClient();
  return useMutation<Lead, unknown, CreateLeadRequest>({
    mutationFn: async (data: CreateLeadRequest): Promise<Lead> => {
      const res = await unwrap(safe.post<Lead>('/crm/sales/leads', toLeadMutationPayload(data)));
      return res as Lead;
    },
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-sales-leads'] });
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
      const res = await unwrap(
        safe.put<Lead>(`/crm/sales/leads/${payload.id}`, toLeadMutationPayload(payload.data))
      );
      return res as Lead;
    },
    onSuccess: (_data: Lead, variables: UpdateLeadPayload): void => {
      void qc.invalidateQueries({ queryKey: ['crm-sales-leads'] });
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
      const res = await unwrap(safe.post<Touchpoint>('/crm/sales/touchpoints', data));
      return res as Touchpoint;
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
          params: { ...(leadId ? { lead_id: leadId } : {}), limit, offset },
        })
      );
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
      const res = await unwrap(safe.post<SalesContact>('/crm/sales/contacts', data));
      return res as SalesContact;
    },
    onSuccess: (_data: SalesContact, variables: CreateContactRequest): void => {
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
      const res = await unwrap(
        safe.put<SalesContact>(`/crm/sales/contacts/${payload.id}`, payload.data)
      );
      return res as SalesContact;
    },
    onSuccess: (_data: SalesContact, variables: UpdateContactPayload): void => {
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
    onSuccess: (_data: unknown, variables: DeleteContactPayload): void => {
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
    queryFn: async (): Promise<Contract[]> => {
      const res = await unwrap(
        safe.get<unknown>('/crm/sales/contracts', { params: { limit, offset } })
      );
      return unwrapList<Contract>(res, 'contracts');
    },
  });

export const useContract = (id: number | undefined): UseQueryResult<Contract> =>
  useQuery<Contract>({
    queryKey: ['crm-sales-contract', id],
    queryFn: async (): Promise<Contract> => {
      const res = await unwrap(safe.get<Contract>(`/crm/sales/contracts/${id}`));
      const envelope = (res ?? {}) as { contract?: Contract };
      return (envelope.contract ?? res) as Contract;
    },
    enabled: !!id,
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
        safe.post<Contract>('/crm/sales/contracts', toCreateContractPayload(data))
      );
      return res as Contract;
    },
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-sales-contracts'] });
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
      const res = await unwrap(
        safe.put<Contract>(
          `/crm/sales/contracts/${payload.id}`,
          toContractUpdatePayload(payload.data)
        )
      );
      return res as Contract;
    },
    onSuccess: (_data: Contract, variables: UpdateContractPayload): void => {
      void qc.invalidateQueries({ queryKey: ['crm-sales-contracts'] });
      void qc.invalidateQueries({ queryKey: ['crm-sales-contract', variables.id] });
      toast.success('Contract updated');
    },
    onError: (): void => {
      toast.error('Failed to update contract');
    },
  });
};
